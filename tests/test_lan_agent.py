import json
import socket
import time

import pytest
import requests

from base.fixtures import _find_yeelight, discover_and_wait


def test_lan_agent_health(lan_agent_url):
    resp = requests.get(f"{lan_agent_url}/health")
    assert resp.status_code == 200


def test_lan_agent_discover_missing_addr(lan_agent_url):
    resp = requests.post(
        f"{lan_agent_url}/discover",
        json={"message": "hello", "timeout_sec": 1},
    )
    assert resp.status_code == 400


def test_lan_agent_discover_missing_message(lan_agent_url):
    resp = requests.post(
        f"{lan_agent_url}/discover",
        json={"addr": "127.0.0.1:19099", "timeout_sec": 1},
    )
    assert resp.status_code == 400


def test_lan_agent_discover_invalid_json(lan_agent_url):
    resp = requests.post(
        f"{lan_agent_url}/discover",
        data="not-json",
        headers={"Content-Type": "application/json"},
    )
    assert resp.status_code == 400


def test_lan_agent_discover_wrong_method(lan_agent_url):
    resp = requests.get(f"{lan_agent_url}/discover")
    assert resp.status_code == 405


def test_lan_agent_discover_finds_yeelight(lan_agent_url, mock_yeelight):
    import os
    ssdp_addr = os.environ.get("JUNO_YEELIGHT_SSDP_ADDR", "127.0.0.1:19820")
    ssdp_msg = (
        "M-SEARCH * HTTP/1.1\r\n"
        "HOST: 239.255.255.250:1982\r\n"
        'MAN: "ssdp:discover"\r\n'
        "ST: wifi_bulb\r\n\r\n"
    )

    deadline = time.monotonic() + 8.0
    devices = []
    while time.monotonic() < deadline:
        resp = requests.post(
            f"{lan_agent_url}/discover",
            json={"addr": ssdp_addr, "message": ssdp_msg, "timeout_sec": 2},
            timeout=10,
        )
        assert resp.status_code == 200
        body = resp.json()
        assert "devices" in body
        devices = body["devices"]
        yeelight_devices = [
            d for d in devices
            if "yeelight" in d.get("raw_response", "").lower()
        ]
        if yeelight_devices:
            break
        time.sleep(0.5)

    yeelight_devices = [
        d for d in devices
        if "yeelight" in d.get("raw_response", "").lower()
    ]
    assert len(
        yeelight_devices) >= 1, "No Yeelight device found through lan-agent discover"
    d = yeelight_devices[0]
    assert "ip" in d
    assert "raw_response" in d
    assert d["ip"] != ""


def test_lan_agent_discover_no_responders(lan_agent_url):
    resp = requests.post(
        f"{lan_agent_url}/discover",
        json={"addr": "127.0.0.1:19099", "message": "ping", "timeout_sec": 1},
        timeout=5,
    )
    assert resp.status_code == 200
    body = resp.json()
    assert body["devices"] == []


def _connect_via_proxy(proxy_addr: str, target_addr: str) -> socket.socket:
    host, port = proxy_addr.split("//", 1)[1].rsplit(":", 1)
    sock = socket.create_connection((host, int(port)), timeout=5)

    connect_req = (
        f"CONNECT {target_addr} HTTP/1.1\r\n"
        f"Host: {target_addr}\r\n\r\n"
    )
    sock.sendall(connect_req.encode())

    buf = b""
    while b"\r\n\r\n" not in buf:
        chunk = sock.recv(512)
        if not chunk:
            break
        buf += chunk

    status_line = buf.split(b"\r\n", 1)[0].decode()
    code = int(status_line.split(" ", 2)[1])
    assert code == 200, f"Expected 200 from CONNECT, got: {status_line}"
    return sock


def test_lan_agent_connect_tunnel(lan_agent_url, mock_yeelight):
    target = f"127.0.0.1:{mock_yeelight.tcp_port}"
    sock = _connect_via_proxy(lan_agent_url, target)
    try:
        req = json.dumps({"id": 1, "method": "get_prop",
                         "params": ["power"]}) + "\r\n"
        sock.settimeout(3.0)
        sock.sendall(req.encode())

        data = b""
        deadline = time.monotonic() + 3.0
        while time.monotonic() < deadline:
            try:
                chunk = sock.recv(4096)
                if chunk:
                    data += chunk
                if b"\n" in data:
                    break
            except socket.timeout:
                break

        assert data, "No response received through tunnel"
        response = json.loads(data.strip())
        assert "result" in response or "error" in response
    finally:
        sock.close()


def test_lan_agent_connect_bad_target(lan_agent_url):
    host, port = lan_agent_url.split("//", 1)[1].rsplit(":", 1)
    sock = socket.create_connection((host, int(port)), timeout=5)
    try:
        connect_req = (
            "CONNECT 127.0.0.1:1 HTTP/1.1\r\n"
            "Host: 127.0.0.1:1\r\n\r\n"
        )
        sock.sendall(connect_req.encode())
        sock.settimeout(3.0)
        response = sock.recv(512).decode()
        assert "502" in response
    finally:
        sock.close()


def test_device_discovered_via_lan_agent(base_url, mock_yeelight):
    devices = discover_and_wait(base_url, timeout=12.0)
    yeelight = _find_yeelight(devices)
    assert yeelight is not None, "No Yeelight device found via lan-agent proxied discovery"
    assert yeelight["vendor"] == "Yeelight"

    # Cleanup.
    requests.delete(f"{base_url}/device/id/{yeelight['id']}")
