import os
import time

import pytest
import requests

from base.mock_yeelight import MockYeelightDevice


def _find_yeelight(devices: list[dict]) -> dict | None:
    for d in devices:
        if str(d.get("vendor", "")).lower() == "yeelight":
            return d
    return None


@pytest.fixture(scope="session")
def base_url():
    port = os.environ.get("JUNO_REST_PORT", "6000")
    return f"http://localhost:{port}"


@pytest.fixture(scope="session")
def web_url():
    port = os.environ.get("JUNO_WEB_PORT", "6001")
    return f"http://localhost:{port}"


@pytest.fixture(scope="session")
def lan_agent_url():
    port = os.environ.get("JUNO_LAN_AGENT_PORT", "7000")
    return f"http://localhost:{port}"


@pytest.fixture(scope="session", autouse=True)
def clean_devices(base_url):
    """Delete all devices at session start to clear any state left by a previous crashed run."""
    resp = requests.get(f"{base_url}/device")
    if resp.status_code == 200:
        for device in resp.json():
            requests.delete(f"{base_url}/device/id/{device['id']}")


@pytest.fixture()
def mock_yeelight():
    """Start a mock Yeelight device and tear it down after each test."""
    ssdp_addr = os.environ.get("JUNO_YEELIGHT_SSDP_ADDR", "127.0.0.1")
    ssdp_port = os.environ.get("JUNO_YEELIGHT_SSDP_PORT", "19820")
    device = MockYeelightDevice(ssdp_host=ssdp_addr, ssdp_port=int(ssdp_port))
    device.start()
    yield device
    device.stop()


@pytest.fixture()
def discovered_device(base_url, mock_yeelight):
    devices = discover_and_wait(base_url)
    device = _find_yeelight(devices)
    assert device is not None, "Yeelight mock device was not discovered"
    yield device
    requests.delete(f"{base_url}/device/id/{device['id']}")


def discover_and_wait(base_url: str, timeout: float = 10.0) -> list[dict]:
    """POST /device/discover then poll GET /device until at least one device appears."""
    requests.post(f"{base_url}/device/discover")
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        resp = requests.get(f"{base_url}/device")
        if resp.status_code == 200:
            devices = resp.json()
            if devices:
                return devices
        time.sleep(0.3)
    return []


def wait_for_command(mock, method: str, timeout: float = 3.0) -> list[dict]:
    """Poll mock received commands until the given method appears or timeout expires."""
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        cmds = mock.get_received_commands()
        if any(c["method"] == method for c in cmds):
            return cmds
        time.sleep(0.05)
    return mock.get_received_commands()
