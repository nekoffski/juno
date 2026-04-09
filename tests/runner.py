import sys
import json
import os
import signal
import subprocess
import threading


REPO_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))


def _load_env_file(path):
    """Parse KEY=VALUE lines from an env file, skipping comments and blanks."""
    env = {}
    with open(path) as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#") or "=" not in line:
                continue
            key, _, value = line.partition("=")
            env[key.strip()] = value.strip()
    return env


class Runner(object):
    def __init__(self, config):
        self.cfg = config
        self.use_postgres = config.get("postgres-enabled", False)
        self.conductor_config = os.path.join(
            REPO_ROOT, config["conductor-config"])
        self.conductor_cmd = os.path.join(REPO_ROOT, config["conductor-cmd"])
        env_file = config.get(
            "env-file", os.path.join(REPO_ROOT, "conf", ".env.example"))
        self.env_file = os.path.abspath(env_file)
        self.env = _load_env_file(self.env_file)
        self.pytest_args = config.get("pytest-args", [])
        self.pytest_path = os.path.join(
            REPO_ROOT, config.get("pytest-path", "tests"))
        self.log_dir = os.path.join(REPO_ROOT, "logs")
        gocoverdir = config.get("gocoverdir", "")
        self.gocoverdir = os.path.abspath(gocoverdir) if gocoverdir else ""
        lan = config.get("lan-agent-cmd", "")
        self.lan_agent_cmd = os.path.join(REPO_ROOT, lan) if lan else ""
        self._lan_agent_proc = None
        self._lan_agent_log = None
        self._conductor_proc = None
        self._conductor_log = None
        self._postgres_log = None
        self._postgres_logs_proc = None

        assert os.path.isfile(
            self.env_file), f"Env file {self.env_file} does not exist"

        for key, value in self.env.items():
            print(f"  {key}={value}")

    def start(self):
        try:
            self._init()
            self._run()
        finally:
            self._cleanup()

    def _init(self):
        os.makedirs(self.log_dir, exist_ok=True)
        if self.use_postgres:
            self._start_postgres()
        if self.lan_agent_cmd:
            self._start_lan_agent()
        self._start_conductor()

    def _start_postgres(self):
        print("$ Starting postgres via docker compose...")
        env = {**os.environ, "ENV_FILE": self.env_file}
        log_path = os.path.join(self.log_dir, "postgres.log")
        self._postgres_log = open(log_path, "w")

        subprocess.run(
            ["docker", "compose", "--env-file", self.env_file,
                "up", "-d", "--wait", "postgres"],
            cwd=REPO_ROOT,
            env=env,
            stdout=self._postgres_log,
            stderr=self._postgres_log,
            check=True,
        )

        # Stream live container logs into the same file
        self._postgres_logs_proc = subprocess.Popen(
            ["docker", "compose", "--env-file", self.env_file,
                "logs", "-f", "--no-color", "postgres"],
            cwd=REPO_ROOT,
            env=env,
            stdout=self._postgres_log,
            stderr=self._postgres_log,
        )
        print(f"$ Postgres is ready, logs: {log_path}.")

    def _start_lan_agent(self):
        print(f"$ Starting juno-lan-agent ({self.lan_agent_cmd})...")
        env = {**os.environ, **self.env}
        if self.gocoverdir:
            env["GOCOVERDIR"] = self.gocoverdir
        log_path = os.path.join(self.log_dir, "lan-agent.log")
        self._lan_agent_log = open(log_path, "w")
        self._lan_agent_proc = subprocess.Popen(
            [self.lan_agent_cmd],
            cwd=REPO_ROOT,
            env=env,
            stdout=self._lan_agent_log,
            stderr=self._lan_agent_log,
        )
        print(
            f"$ juno-lan-agent started (pid={self._lan_agent_proc.pid}), logs: {log_path}.")

    def _start_conductor(self):
        print(f"$ Starting juno-conductor ({self.conductor_cmd})...")
        env = {**os.environ, **self.env}
        if self.gocoverdir:
            os.makedirs(self.gocoverdir, exist_ok=True)
            env["GOCOVERDIR"] = self.gocoverdir
        log_path = os.path.join(self.log_dir, "conductor.log")
        self._conductor_log = open(log_path, "w")
        self._conductor_proc = subprocess.Popen(
            [self.conductor_cmd, "-config", self.conductor_config],
            cwd=REPO_ROOT,
            env=env,
            stdout=self._conductor_log,
            stderr=self._conductor_log,
        )
        print(
            f"$ juno-conductor started (pid={self._conductor_proc.pid}), logs: {log_path}.")

    def _run(self):
        print("$ Running pytest...")
        log_path = os.path.join(self.log_dir, "pytest.log")
        env = {**os.environ, **self.env}
        with open(log_path, "w") as log_file:
            proc = subprocess.Popen(
                ["pytest", *self.pytest_args, self.pytest_path],
                cwd=REPO_ROOT,
                env=env,
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                text=True,
            )

            def _tee():
                for line in proc.stdout:
                    sys.stdout.write(line)
                    sys.stdout.flush()
                    log_file.write(line)

            tee_thread = threading.Thread(target=_tee, daemon=True)
            tee_thread.start()
            proc.wait()
            tee_thread.join()

        print(f"$ pytest logs saved to {log_path}.")
        if proc.returncode != 0:
            raise SystemExit(proc.returncode)

    def _cleanup(self):
        self._stop_conductor()
        self._stop_lan_agent()
        if self.use_postgres:
            self._stop_postgres()

    def _stop_conductor(self):
        if self._conductor_proc is None:
            return
        print("$ Stopping juno-conductor...")
        self._conductor_proc.send_signal(signal.SIGTERM)
        try:
            self._conductor_proc.wait(timeout=10)
        except subprocess.TimeoutExpired:
            self._conductor_proc.kill()
            self._conductor_proc.wait()
        finally:
            if self._conductor_log:
                self._conductor_log.close()
                self._conductor_log = None
        print("$ juno-conductor stopped.")

    def _stop_lan_agent(self):
        if self._lan_agent_proc is None:
            return
        print("$ Stopping juno-lan-agent...")
        self._lan_agent_proc.send_signal(signal.SIGTERM)
        try:
            self._lan_agent_proc.wait(timeout=10)
        except subprocess.TimeoutExpired:
            self._lan_agent_proc.kill()
            self._lan_agent_proc.wait()
        finally:
            if self._lan_agent_log:
                self._lan_agent_log.close()
                self._lan_agent_log = None
        print("$ juno-lan-agent stopped.")

    def _stop_postgres(self):
        print("$ Stopping postgres via docker compose...")
        if self._postgres_logs_proc is not None:
            self._postgres_logs_proc.terminate()
            self._postgres_logs_proc.wait()
            self._postgres_logs_proc = None
        env = {**os.environ, "ENV_FILE": self.env_file}
        subprocess.run(
            ["docker", "compose", "--env-file", self.env_file, "down", "postgres"],
            cwd=REPO_ROOT,
            env=env,
            stdout=self._postgres_log,
            stderr=self._postgres_log,
            check=False,
        )
        if self._postgres_log is not None:
            self._postgres_log.close()
            self._postgres_log = None


def main():
    assert len(sys.argv) == 2, "Usage: runner.py <test-config>"
    assert os.path.isfile(
        sys.argv[1]), f"Test config file {sys.argv[1]} does not exist"

    config = None
    with open(sys.argv[1], "r") as f:
        config = json.load(f)

    print(f"$ Test runner using config: {config}")
    Runner(config).start()


if __name__ == "__main__":
    main()
