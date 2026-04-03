import pytest
import requests

from base.fixtures import discover_and_wait, wait_for_command


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

    cmds = wait_for_command(mock_yeelight, "on")
    assert "on" in [c["method"] for c in cmds]


# Sending the "off" action should return 200 and the mock must receive the command.
def test_yeelight_action_off(base_url, discovered_device, mock_yeelight):
    device_id = discovered_device["id"]
    mock_yeelight.clear_received_commands()

    resp = requests.post(f"{base_url}/device/id/{device_id}/action/off", json={})
    assert resp.status_code == 200

    cmds = wait_for_command(mock_yeelight, "off")
    assert "off" in [c["method"] for c in cmds]


# Sending the "toggle" action should return 200 and the mock must receive the command.
def test_yeelight_action_toggle(base_url, discovered_device, mock_yeelight):
    device_id = discovered_device["id"]
    mock_yeelight.clear_received_commands()

    resp = requests.post(f"{base_url}/device/id/{device_id}/action/toggle", json={})
    assert resp.status_code == 200

    cmds = wait_for_command(mock_yeelight, "toggle")
    assert "toggle" in [c["method"] for c in cmds]


# Sending an RGB action should return 200 and the mock must receive set_rgb with the
# correctly packed integer value.
def test_yeelight_action_rgb(base_url, discovered_device, mock_yeelight):
    device_id = discovered_device["id"]
    mock_yeelight.clear_received_commands()

    resp = requests.post(
        f"{base_url}/device/id/{device_id}/action/rgb",
        json={"params": {"color": {"r": 0, "g": 128, "b": 255}}},
    )
    assert resp.status_code == 200

    cmds = wait_for_command(mock_yeelight, "set_rgb")
    assert "set_rgb" in [c["method"] for c in cmds]

    set_rgb = next(c for c in cmds if c["method"] == "set_rgb")
    # (0 << 16) | (128 << 8) | 255 = 33023
    assert set_rgb["params"][0] == 33023


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


# Device list entry has all expected fields when a device is present.
def test_yeelight_devices_list_schema(base_url, discovered_device):
    resp = requests.get(f"{base_url}/device")
    assert resp.status_code == 200
    devices = resp.json()
    device = next((d for d in devices if d["id"] == discovered_device["id"]), None)
    assert device is not None
    assert isinstance(device["id"], int)
    assert isinstance(device["name"], str)
    assert device["vendor"] == "Yeelight"
    assert isinstance(device["capabilities"], list)
    assert len(device["capabilities"]) > 0
    assert isinstance(device["properties"], dict)
    assert "status" in device


# Fetching properties without a fields filter returns 200 with an object.
def test_yeelight_get_properties_no_filter(base_url, discovered_device):
    device_id = discovered_device["id"]
    resp = requests.get(f"{base_url}/device/id/{device_id}/properties")
    assert resp.status_code == 200
    assert isinstance(resp.json(), dict)


# Mock device state is updated correctly after on/off actions.
def test_yeelight_state_after_on_and_off(base_url, discovered_device, mock_yeelight):
    device_id = discovered_device["id"]

    mock_yeelight.clear_received_commands()
    resp = requests.post(f"{base_url}/device/id/{device_id}/action/off", json={})
    assert resp.status_code == 200
    wait_for_command(mock_yeelight, "off")
    assert mock_yeelight.get_state()["power"] == "off"

    mock_yeelight.clear_received_commands()
    resp = requests.post(f"{base_url}/device/id/{device_id}/action/on", json={})
    assert resp.status_code == 200
    wait_for_command(mock_yeelight, "on")
    assert mock_yeelight.get_state()["power"] == "on"


# Mock device state reflects correct packed RGB integer after set_rgb action.
def test_yeelight_rgb_state_verification(base_url, discovered_device, mock_yeelight):
    device_id = discovered_device["id"]
    mock_yeelight.clear_received_commands()

    resp = requests.post(
        f"{base_url}/device/id/{device_id}/action/rgb",
        json={"params": {"color": {"r": 255, "g": 0, "b": 0}}},
    )
    assert resp.status_code == 200
    wait_for_command(mock_yeelight, "set_rgb")
    # (255 << 16) | (0 << 8) | 0 = 16711680
    assert mock_yeelight.get_state()["rgb"] == "16711680"


# Triggering discovery a second time while a device is already registered must
# not create a duplicate entry (exercises the exists() check in device service).
def test_yeelight_no_duplicate_on_rediscovery(base_url, mock_yeelight):
    devices_before = discover_and_wait(base_url)
    count_before = len([d for d in devices_before if d.get("vendor") == "Yeelight"])
    assert count_before >= 1

    requests.post(f"{base_url}/device/discover")
    import time; time.sleep(1.0)

    resp = requests.get(f"{base_url}/device")
    assert resp.status_code == 200
    count_after = len([d for d in resp.json() if d.get("vendor") == "Yeelight"])
    assert count_after == count_before

    for d in resp.json():
        if d.get("vendor") == "Yeelight":
            requests.delete(f"{base_url}/device/id/{d['id']}")


# RGB action with malformed params (missing color object) must be rejected.
def test_yeelight_action_rgb_invalid_params(base_url, discovered_device):
    device_id = discovered_device["id"]
    resp = requests.post(
        f"{base_url}/device/id/{device_id}/action/rgb",
        json={"params": {"wrong_key": 123}},
    )
    assert resp.status_code >= 400
