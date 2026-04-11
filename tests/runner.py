import sys
import json
import os
import signal
import subprocess
import threading
import time
import urllib.request
import urllib.error


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
        self._cfg = config
        self._use_postgres = config.get("postgres-enabled", False)
        self._conductor_config = os.path.join(
            REPO_ROOT, config["conductor-config"])
        self._conductor_cmd = os.path.join(REPO_ROOT, config["conductor-cmd"])
        env_file = config.get(
            "env-file", os.path.join(REPO_ROOT, "conf", ".env.example"))
        self._env_file = os.path.abspath(env_file)
        self._env = _load_env_file(self._env_file)
        self._pytest_args = config.get("pytest-args", [])
        self._pytest_path = os.path.join(
            REPO_ROOT, config.get("pytest-path", "tests"))
        self._log_dir = os.path.join(REPO_ROOT, "logs")
        gocoverdir = config.get("gocoverdir", "")
        self._gocoverdir = os.path.abspath(gocoverdir) if gocoverdir else ""
        lan = config.get("lan-agent-cmd", "")
        self._lan_agent_cmd = os.path.join(REPO_ROOT, lan) if lan else ""
        self._lan_agent_proc = None
        self._lan_agent_log = None
        self._conductor_proc = None
        self._conductor_log = None
        self._postgres_log = None
        self._postgres_logs_proc = None
        self._skip_init = config.get("skip-init", False)
        self._skip_cleanup = config.get("skip-cleanup", False)
        self._api_url = config.get("api-url", None)

        assert os.path.isfile(
            self._env_file), f"Env file {self._env_file} does not exist"

        if not os.path.exists(self._log_dir):
            os.makedirs(self._log_dir)

        for key, value in self._env.items():
            print(f"  {key}={value}")

    def start(self):
        try:
            if not self._skip_init:
                self._init()
            self._run()
        finally:
            if not self._skip_cleanup:
                self._cleanup()

    def _init(self):
        os.makedirs(self._log_dir, exist_ok=True)
        if self._use_postgres:
            self._start_postgres()
        if self._lan_agent_cmd:
            self._start_lan_agent()
        self._start_conductor()

    def _start_postgres(self):
        print("$ Starting postgres...")
        log_path = os.path.join(self._log_dir, "postgres.log")
        self._postgres_log = open(log_path, "w")

        subprocess.run(
            [os.path.join(REPO_ROOT, "cicd", "run-postgres.sh")],
            cwd=REPO_ROOT,
            stdout=self._postgres_log,
            stderr=self._postgres_log,
            check=True,
        )

        for _ in range(30):
            result = subprocess.run(
                ["docker", "inspect",
                    "--format={{.State.Health.Status}}", "postgres"],
                capture_output=True,
                text=True,
            )
            if result.stdout.strip() == "healthy":
                break
            time.sleep(1)
        else:
            raise RuntimeError("Postgres did not become healthy in time")

        # Stream live container logs into the same file
        self._postgres_logs_proc = subprocess.Popen(
            ["docker", "logs", "-f", "postgres"],
            stdout=self._postgres_log,
            stderr=self._postgres_log,
        )
        print(f"$ Postgres is ready, logs: {log_path}.")

    def _start_lan_agent(self):
        print(f"$ Starting juno-lan-agent ({self._lan_agent_cmd})...")
        env = {**os.environ, **self._env}
        if self._gocoverdir:
            env["GOCOVERDIR"] = self._gocoverdir
        log_path = os.path.join(self._log_dir, "lan-agent.log")
        self._lan_agent_log = open(log_path, "w")
        self._lan_agent_proc = subprocess.Popen(
            [self._lan_agent_cmd],
            cwd=REPO_ROOT,
            env=env,
            stdout=self._lan_agent_log,
            stderr=self._lan_agent_log,
        )
        print(
            f"$ juno-lan-agent started (pid={self._lan_agent_proc.pid}), logs: {log_path}.")

    def _start_conductor(self):
        print(f"$ Starting juno-conductor ({self._conductor_cmd})...")
        env = {**os.environ, **self._env}
        if self._gocoverdir:
            os.makedirs(self._gocoverdir, exist_ok=True)
            env["GOCOVERDIR"] = self._gocoverdir
        log_path = os.path.join(self._log_dir, "conductor.log")
        self._conductor_log = open(log_path, "w")
        self._conductor_proc = subprocess.Popen(
            [self._conductor_cmd, "-config", self._conductor_config],
            cwd=REPO_ROOT,
            env=env,
            stdout=self._conductor_log,
            stderr=self._conductor_log,
        )
        print(
            f"$ juno-conductor started (pid={self._conductor_proc.pid}), logs: {log_path}.")
        self._wait_for_server()

    def _wait_for_server(self):
        rest_port = self._env.get("JUNO_REST_PORT", "6001")
        url = f"http://localhost:{rest_port}/health"
        print(f"$ Waiting for juno-server to be ready at {url}...")
        for _ in range(30):
            try:
                urllib.request.urlopen(url, timeout=1)
                print("$ juno-server is ready.")
                return
            except (urllib.error.URLError, OSError):
                time.sleep(1)
        raise RuntimeError("juno-server did not become ready in time")

    def _run(self):
        print("$ Running pytest...")
        log_path = os.path.join(self._log_dir, "pytest.log")
        env = {**os.environ, **self._env}

        if self._api_url:
            env["TEST_API_URL"] = self._api_url

        with open(log_path, "w") as log_file:
            proc = subprocess.Popen(
                ["pytest", *self._pytest_args, self._pytest_path],
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
        if self._use_postgres:
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
        print("$ Stopping postgres...")
        if self._postgres_logs_proc is not None:
            self._postgres_logs_proc.terminate()
            self._postgres_logs_proc.wait()
            self._postgres_logs_proc = None
        subprocess.run(
            [os.path.join(REPO_ROOT, "cicd", "stop-postgres.sh")],
            cwd=REPO_ROOT,
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
