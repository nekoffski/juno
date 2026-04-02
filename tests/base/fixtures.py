import os
import time

import pytest
import requests

from base.mock_yeelight import MockYeelightDevice


@pytest.fixture(scope="session")
def base_url():
    port = os.environ.get("JUNO_REST_PORT", "6000")
    return f"http://localhost:{port}"


@pytest.fixture()
def mock_yeelight():
    """Start a mock Yeelight device and tear it down after each test."""
    ssdp_addr = os.environ.get("JUNO_YEELIGHT_SSDP_ADDR", "127.0.0.1:19820")
    host, port_str = ssdp_addr.rsplit(":", 1)
    device = MockYeelightDevice(ssdp_host=host, ssdp_port=int(port_str))
    device.start()
    yield device
    device.stop()


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
