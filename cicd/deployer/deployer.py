import sys
import requests
import argparse
import dataclasses
import enum
import os


class TargetType(enum.StrEnum):
    EDGE = "edge"
    CORE = "core"


def _token():
    return os.getenv("JUNO_DEPLOYMENT_TOKEN")


def _env_template():
    return os.getenv("JUNO_DEPLOYMENT_ENV_TEMPLATE")


TARGET_PATH = "/opt/juno"
LOG_FILE = "/tmp/juno-deployment.log"
JUNO_REPO = "https://github.com/nekoffski/juno"


@dataclasses.dataclass
class Args:
    dry_run: bool
    api_url: str
    target: TargetType
    server_id: str


class Deployer(object):
    def __init__(self, args: Args):
        self._dry_run = args.dry_run
        self._api_url = args.api_url
        self._target = args.target
        self._server_id = args.server_id

    def deploy(self):
        self._preflight_check()

        if self._dry_run:
            print("Dry run mode, skipping deployment")
            return

        self._cmd(
            f"git -C {TARGET_PATH} reset --hard HEAD > {LOG_FILE} 2>&1 "
            f"&& git -C {TARGET_PATH} pull origin main >> {LOG_FILE} 2>&1 "
            f"|| git clone {JUNO_REPO} {TARGET_PATH} > {LOG_FILE} 2>&1"
        )
        print("-- Repo cloned")

        self._cmd(
            f"pushd {TARGET_PATH} && ./cicd/init-deployment.sh '{_env_template()}' {self._target.value} > {LOG_FILE} 2>&1 && popd")

    def _preflight_check(self):
        assert os.getenv(
            "JUNO_DEPLOYMENT_TOKEN") is not None, "JUNO_DEPLOYMENT_TOKEN environment variable must be set"
        assert os.getenv(
            "JUNO_DEPLOYMENT_ENV_TEMPLATE") is not None, "JUNO_DEPLOYMENT_ENV_TEMPLATE environment variable must be set"
        self._health()
        print("-- Preflight check passed")

    def _cmd(self, cmd):
        return self._request(path="exec", args={"cmd": cmd})

    def _health(self):
        res = self._request(path="info")
        assert res['server_id'] == self._server_id, "Server ID mismatch"

    def _request(self, path, args=None):
        if args is None:
            args = {}

        if type(args) != dict:
            raise ValueError("args must be a dict")

        token = _token()

        args.update({
            "srv": self._server_id,
        })

        headers = {
            "Authorization": f"{token}"
        }

        url = f"{self._api_url}/{path}"
        res = requests.post(url, data=args, headers=headers)
        if res.status_code != 200:
            raise RuntimeError(
                f"Request failed with status code {res.status_code}: {res.text}")
        return res.json()


def main():
    try:
        args = _parse_args()
        deployer = Deployer(args=args)
        deployer.deploy()
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


def _parse_args():
    parser = argparse.ArgumentParser(description="Deploy application")
    parser.add_argument("--dry-run", action="store_true",
                        help="Perform a dry run")
    parser.add_argument("--api-url", required=True,
                        help="API URL for deployment")
    parser.add_argument("--type", type=TargetType,
                        required=True, help="Deployment type")
    parser.add_argument("--server-id", required=True,
                        help="ID of the server to deploy to")

    args = parser.parse_args()
    return Args(
        dry_run=args.dry_run,
        api_url=args.api_url,
        target=args.type,
        server_id=args.server_id
    )


if __name__ == "__main__":
    main()
