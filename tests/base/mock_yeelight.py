"""Mock Yeelight device for integration tests.

Implements a minimal SSDP responder (UDP) and a Yeelight JSON-RPC server (TCP)
so that the Juno server can discover and control it without real hardware.
"""

import json
import socket
import threading
from collections import deque


def _find_free_port() -> int:
    """Bind to port 0 to let the OS assign a free port, return it."""
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind(("127.0.0.1", 0))
        return s.getsockname()[1]


class MockYeelightDevice:
    """Minimal Yeelight device mock: UDP SSDP + TCP JSON-RPC."""

    def __init__(self, ssdp_host: str, ssdp_port: int):
        self._ssdp_host = ssdp_host
        self._ssdp_port = ssdp_port
        self._tcp_port = _find_free_port()

        # Device state (mirrors real Yeelight properties)
        self._state = {
            "power": "on",
            "bright": "100",
            "rgb": "16711680",  # red packed as int string
        }
        self._state_lock = threading.Lock()

        # Record every command received over TCP for test assertions
        self._received_commands: deque = deque()
        self._commands_lock = threading.Lock()

        self._stop_event = threading.Event()
        self._threads: list[threading.Thread] = []

    # ------------------------------------------------------------------
    # Lifecycle
    # ------------------------------------------------------------------

    def start(self) -> None:
        udp_thread = threading.Thread(target=self._udp_loop, daemon=True)
        tcp_thread = threading.Thread(target=self._tcp_loop, daemon=True)
        udp_thread.start()
        tcp_thread.start()
        self._threads = [udp_thread, tcp_thread]

    def stop(self) -> None:
        self._stop_event.set()
        # Wake up blocking sockets
        try:
            socket.socket(socket.AF_INET, socket.SOCK_DGRAM).sendto(
                b"", (self._ssdp_host, self._ssdp_port)
            )
        except Exception:
            pass
        try:
            socket.create_connection(("127.0.0.1", self._tcp_port), timeout=0.1).close()
        except Exception:
            pass
        for t in self._threads:
            t.join(timeout=2.0)

    # ------------------------------------------------------------------
    # Accessors for test assertions
    # ------------------------------------------------------------------

    def get_state(self) -> dict:
        with self._state_lock:
            return dict(self._state)

    def get_received_commands(self) -> list[dict]:
        with self._commands_lock:
            return list(self._received_commands)

    def clear_received_commands(self) -> None:
        with self._commands_lock:
            self._received_commands.clear()

    @property
    def tcp_port(self) -> int:
        return self._tcp_port

    # ------------------------------------------------------------------
    # UDP – SSDP responder
    # ------------------------------------------------------------------

    def _build_ssdp_response(self) -> bytes:
        # Juno checks: UDP source IP, Location header for port, "yeelight" in body
        lines = [
            "HTTP/1.1 200 OK",
            "Cache-Control: max-age=3600",
            f"Location: yeelight://127.0.0.1:{self._tcp_port}",
            "Service: wifi_bulb",
            "Date: ",
            "Ext: ",
            "Server: POSIX UPnP/1.0 YEELIGHT 1",
            "id: 0x000000000000ffff",
            "model: color",
            "fw_ver: 18",
            "support: get_prop set_default set_power toggle set_bright start_cf stop_cf set_scene cron_add cron_get cron_del set_ct_abx set_rgb",
            "power: on",
            "bright: 100",
            "color_mode: 1",
            "ct: 4000",
            "rgb: 16711680",
            "hue: 0",
            "sat: 100",
            "name: mock_yeelight",
            "",
            "",
        ]
        return "\r\n".join(lines).encode()

    def _udp_loop(self) -> None:
        sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        sock.settimeout(1.0)
        try:
            sock.bind((self._ssdp_host, self._ssdp_port))
        except OSError as e:
            print(f"[mock_yeelight] UDP bind failed on {self._ssdp_host}:{self._ssdp_port}: {e}")
            return

        while not self._stop_event.is_set():
            try:
                data, addr = sock.recvfrom(4096)
            except socket.timeout:
                continue
            except OSError:
                break

            if self._stop_event.is_set():
                break

            response = self._build_ssdp_response()
            try:
                sock.sendto(response, addr)
            except OSError:
                pass

        sock.close()

    # ------------------------------------------------------------------
    # TCP – Yeelight JSON-RPC server
    # ------------------------------------------------------------------

    def _tcp_loop(self) -> None:
        server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        server.settimeout(1.0)
        server.bind(("127.0.0.1", self._tcp_port))
        server.listen(5)

        while not self._stop_event.is_set():
            try:
                conn, _ = server.accept()
            except socket.timeout:
                continue
            except OSError:
                break

            client_thread = threading.Thread(
                target=self._handle_tcp_client, args=(conn,), daemon=True
            )
            client_thread.start()

        server.close()

    def _handle_tcp_client(self, conn: socket.socket) -> None:
        buf = b""
        conn.settimeout(1.0)
        try:
            while not self._stop_event.is_set():
                try:
                    chunk = conn.recv(4096)
                except socket.timeout:
                    continue
                except OSError:
                    break

                if not chunk:
                    break

                buf += chunk
                while b"\r\n" in buf or b"\n" in buf:
                    if b"\r\n" in buf:
                        line, buf = buf.split(b"\r\n", 1)
                    else:
                        line, buf = buf.split(b"\n", 1)

                    line = line.strip()
                    if not line:
                        continue

                    try:
                        msg = json.loads(line)
                    except json.JSONDecodeError:
                        continue

                    response = self._handle_rpc(msg)
                    try:
                        conn.sendall(json.dumps(response).encode() + b"\r\n")
                    except OSError:
                        return
        finally:
            conn.close()

    def _handle_rpc(self, msg: dict) -> dict:
        req_id = msg.get("id", 0)
        method = msg.get("method", "")
        params = msg.get("params", [])

        with self._commands_lock:
            self._received_commands.append({"method": method, "params": params})

        with self._state_lock:
            if method == "get_prop":
                result = [self._state.get(p, "") for p in params]
                return {"id": req_id, "result": result}

            elif method == "set_power":
                # params: ["on"/"off", effect, duration]
                if params:
                    self._state["power"] = params[0]
                return {"id": req_id, "result": ["ok"]}

            elif method == "toggle":
                self._state["power"] = "off" if self._state["power"] == "on" else "on"
                return {"id": req_id, "result": ["ok"]}

            elif method == "on":
                self._state["power"] = "on"
                return {"id": req_id, "result": ["ok"]}

            elif method == "off":
                self._state["power"] = "off"
                return {"id": req_id, "result": ["ok"]}

            elif method == "set_rgb":
                if params:
                    self._state["rgb"] = str(params[0])
                return {"id": req_id, "result": ["ok"]}

            elif method == "bright":
                if params:
                    self._state["bright"] = str(params[0])
                return {"id": req_id, "result": ["ok"]}

            else:
                return {
                    "id": req_id,
                    "error": {"code": -1, "message": f"unknown method: {method}"},
                }
