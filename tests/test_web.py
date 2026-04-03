import requests


def test_web_dashboard(web_url):
    r = requests.get(f"{web_url}/")
    assert r.status_code == 200
    assert "text/html" in r.headers.get("Content-Type", "")
    assert "<html" in r.text.lower()


def test_web_devices_tab(web_url):
    r = requests.get(f"{web_url}/tabs/devices")
    assert r.status_code == 200
    assert "text/html" in r.headers.get("Content-Type", "")


def test_web_metrics_tab(web_url):
    r = requests.get(f"{web_url}/tabs/metrics")
    assert r.status_code == 200
    assert "text/html" in r.headers.get("Content-Type", "")


def test_web_events_tab(web_url):
    r = requests.get(f"{web_url}/tabs/events")
    assert r.status_code == 200
    assert "text/html" in r.headers.get("Content-Type", "")


def test_web_sse_stream_headers(web_url):
    r = requests.get(f"{web_url}/sse", stream=True, timeout=(5, None))
    assert r.status_code == 200
    assert "text/event-stream" in r.headers.get("Content-Type", "")
    r.close()


def test_web_perform_action_invalid_brightness(web_url):
    r = requests.post(
        f"{web_url}/device/1/action/brightness",
        data={"brightness": "notanumber"},
    )
    assert r.status_code == 400


def test_web_perform_action_toggle(web_url, discovered_device):
    device_id = discovered_device["id"]
    r = requests.post(
        f"{web_url}/device/{device_id}/action/toggle",
        data={},
    )
    assert r.status_code == 200
    assert "text/html" in r.headers.get("Content-Type", "")


def test_web_perform_action_rgb(web_url, discovered_device):
    device_id = discovered_device["id"]
    r = requests.post(
        f"{web_url}/device/{device_id}/action/rgb",
        data={"color": "#ff8000"},
    )
    assert r.status_code == 200
    assert "text/html" in r.headers.get("Content-Type", "")
