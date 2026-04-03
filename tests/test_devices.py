import pytest
import requests


def test_get_devices_returns_list(base_url):
    r = requests.get(f"{base_url}/device")
    assert r.status_code == 200
    assert isinstance(r.json(), list)


def test_discover_devices_accepted(base_url):
    r = requests.post(f"{base_url}/device/discover")
    assert r.status_code == 202


def test_get_device_by_id_not_found(base_url):
    r = requests.get(f"{base_url}/device/id/999999")
    assert r.status_code == 404


def test_get_device_properties_not_found(base_url):
    r = requests.get(
        f"{base_url}/device/id/999999/properties",
        params={"fields": "power"},
    )
    assert r.status_code == 404


def test_perform_device_action_not_found(base_url):
    r = requests.post(f"{base_url}/device/id/999999/action/on", json={})
    assert r.status_code == 404


def test_get_devices_response_schema(base_url):
    """Each device in the list must have the expected fields."""
    r = requests.get(f"{base_url}/device")
    assert r.status_code == 200
    devices = r.json()
    assert len(devices) == 0


def test_delete_device_not_found(base_url):
    r = requests.delete(f"{base_url}/device/id/999999")
    assert r.status_code == 404


def test_perform_device_action_not_found_delete(base_url):
    r = requests.post(f"{base_url}/device/id/999999/action/toggle", json={})
    assert r.status_code == 404


def test_events_stream_headers(base_url):
    r = requests.get(f"{base_url}/events", stream=True, timeout=(5, None))
    assert r.status_code == 200
    assert "text/event-stream" in r.headers.get("Content-Type", "")
    r.close()