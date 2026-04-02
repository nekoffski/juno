import pytest
import requests

from base.fixtures import discover_and_wait


def _find_yeelight(devices: list[dict]) -> dict | None:
    for d in devices:
        if str(d.get("vendor", "")).lower() == "yeelight":
            return d
    return None


@pytest.fixture()
def discovered_device(base_url, mock_yeelight):
    devices = discover_and_wait(base_url)
    device = _find_yeelight(devices)
    assert device is not None, "Yeelight mock device was not discovered"
    yield device
    requests.delete(f"{base_url}/device/id/{device['id']}")


# After triggering discovery the mock device should appear in the device list
# with the correct vendor and all expected capabilities.
def test_yeelight_discovery(base_url, mock_yeelight):
    devices = discover_and_wait(base_url)
    device = _find_yeelight(devices)
    assert device is not None, "No Yeelight device found after discovery"

    assert device["vendor"] == "Yeelight"
    assert isinstance(device["capabilities"], list)
    assert "on" in device["capabilities"]
    assert "off" in device["capabilities"]
    assert "toggle" in device["capabilities"]
    assert "rgb" in device["capabilities"]

    requests.delete(f"{base_url}/device/id/{device['id']}")


# Fetching a discovered device by its ID should return 200 with all required fields.
def test_yeelight_get_by_id(base_url, discovered_device):
    device_id = discovered_device["id"]
    resp = requests.get(f"{base_url}/device/id/{device_id}")
    assert resp.status_code == 200

    body = resp.json()
    assert body["id"] == device_id
    assert body["vendor"] == "Yeelight"
    assert "name" in body
    assert "status" in body
    assert "capabilities" in body
    assert "properties" in body


# Requesting specific property fields should return 200 with those fields present.
def test_yeelight_get_properties(base_url, discovered_device):
    device_id = discovered_device["id"]
    resp = requests.get(
        f"{base_url}/device/id/{device_id}/properties",
        params={"fields": "power,brightness"},
    )
    assert resp.status_code == 200
    body = resp.json()
    assert "power" in body
    assert "brightness" in body


# Sending the "on" action should return 200 and the mock must receive the command.
def test_yeelight_action_on(base_url, discovered_device, mock_yeelight):
    device_id = discovered_device["id"]
    mock_yeelight.clear_received_commands()

    resp = requests.post(f"{base_url}/device/id/{device_id}/action/on", json={})
    assert resp.status_code == 200

    methods = [c["method"] for c in mock_yeelight.get_received_commands()]
    assert "on" in methods


# Sending the "off" action should return 200 and the mock must receive the command.
def test_yeelight_action_off(base_url, discovered_device, mock_yeelight):
    device_id = discovered_device["id"]
    mock_yeelight.clear_received_commands()

    resp = requests.post(f"{base_url}/device/id/{device_id}/action/off", json={})
    assert resp.status_code == 200

    methods = [c["method"] for c in mock_yeelight.get_received_commands()]
    assert "off" in methods


# Sending the "toggle" action should return 200 and the mock must receive the command.
def test_yeelight_action_toggle(base_url, discovered_device, mock_yeelight):
    device_id = discovered_device["id"]
    mock_yeelight.clear_received_commands()

    resp = requests.post(f"{base_url}/device/id/{device_id}/action/toggle", json={})
    assert resp.status_code == 200

    methods = [c["method"] for c in mock_yeelight.get_received_commands()]
    assert "toggle" in methods


# Sending an RGB action should return 200 and the mock must receive set_rgb with the
# correctly packed integer value.
def test_yeelight_action_rgb(base_url, discovered_device, mock_yeelight):
    device_id = discovered_device["id"]
    mock_yeelight.clear_received_commands()

    resp = requests.post(
        f"{base_url}/device/id/{device_id}/action/rgb",
        json={"params": {"r": 0, "g": 128, "b": 255}},
    )
    assert resp.status_code == 200

    cmds = mock_yeelight.get_received_commands()
    assert "set_rgb" in [c["method"] for c in cmds]

    set_rgb = next(c for c in cmds if c["method"] == "set_rgb")
    # (0 << 16) | (128 << 8) | 255 = 32895
    assert set_rgb["params"][0] == 32895


# Deleting a device should return 200 and subsequent GET by ID should return 404.
def test_yeelight_delete_device(base_url, mock_yeelight):
    devices = discover_and_wait(base_url)
    device = _find_yeelight(devices)
    assert device is not None

    device_id = device["id"]
    assert requests.delete(f"{base_url}/device/id/{device_id}").status_code == 200
    assert requests.get(f"{base_url}/device/id/{device_id}").status_code == 404


# Deleting a device ID that does not exist should return 404.
def test_yeelight_delete_nonexistent(base_url):
    resp = requests.delete(f"{base_url}/device/id/999999")
    assert resp.status_code == 404
