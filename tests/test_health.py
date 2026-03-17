import requests


def test_health(base_url):
    r = requests.get(f"{base_url}/health")
    assert r.status_code == 200
    assert r.json()["status"] == "ok"


def test_device_service_health(base_url):
    r = requests.get(f"{base_url}/health/service/device")
    assert r.status_code == 200
    assert r.json()["status"] == "ok"
