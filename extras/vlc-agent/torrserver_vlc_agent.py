#!/usr/bin/env python3
"""Small HTTP agent that opens TorrServer streams in VLC on this Linux desktop."""

from __future__ import annotations

import argparse
import hmac
import ipaddress
import json
import logging
import os
import re
import shlex
import signal
import subprocess
import threading
from dataclasses import dataclass
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import PurePath
from typing import Any, Iterable
from urllib.parse import parse_qs, unquote, urlparse

MAX_BODY_BYTES = 16 * 1024
MAX_STREAM_URL_LENGTH = 8192
HASH_PATTERN = re.compile(r"^[0-9a-fA-F]{40}$")
DEFAULT_PLAYER_ARGS = (
    "--no-one-instance",
    "--no-video-title-show",
    "--network-caching=3000",
    "--http-reconnect",
)


@dataclass(frozen=True)
class Config:
    host: str
    port: int
    token: str
    allowed_hosts: frozenset[str]
    player: str
    player_args: tuple[str, ...]
    stop_timeout: float


@dataclass(frozen=True)
class PlayRequest:
    path: str
    hash: str
    index: int
    stream_url: str
    fullscreen: bool


class PlayerManager:
    """Owns one player process and never touches unrelated VLC instances."""

    def __init__(self, player: str, player_args: Iterable[str], stop_timeout: float) -> None:
        self._player = player
        self._player_args = tuple(player_args)
        self._stop_timeout = stop_timeout
        self._process: subprocess.Popen[bytes] | None = None
        self._lock = threading.Lock()

    def launch(self, stream_url: str, fullscreen: bool) -> None:
        with self._lock:
            self._stop_locked()
            window_mode = "--fullscreen" if fullscreen else "--no-fullscreen"
            command = [self._player, *self._player_args, window_mode, stream_url]
            logging.info("Starting player: %s", " ".join(shlex.quote(value) for value in command))
            self._process = subprocess.Popen(
                command,
                stdin=subprocess.DEVNULL,
                stdout=subprocess.DEVNULL,
                stderr=None,
                start_new_session=True,
                close_fds=True,
            )

    def stop(self) -> None:
        with self._lock:
            self._stop_locked()

    def is_running(self) -> bool:
        with self._lock:
            return self._process is not None and self._process.poll() is None

    def _stop_locked(self) -> None:
        process = self._process
        self._process = None
        if process is None or process.poll() is not None:
            return
        logging.info("Stopping previous player process group %s", process.pid)
        try:
            os.killpg(process.pid, signal.SIGTERM)
        except ProcessLookupError:
            return
        try:
            process.wait(timeout=self._stop_timeout)
        except subprocess.TimeoutExpired:
            logging.warning("Player did not stop in %.1fs; killing it", self._stop_timeout)
            os.killpg(process.pid, signal.SIGKILL)
            process.wait(timeout=self._stop_timeout)


class AgentHTTPServer(ThreadingHTTPServer):
    daemon_threads = True
    allow_reuse_address = True

    def __init__(self, address: tuple[str, int], config: Config, player: PlayerManager) -> None:
        super().__init__(address, AgentHandler)
        self.config = config
        self.player = player


class AgentHandler(BaseHTTPRequestHandler):
    server: AgentHTTPServer
    server_version = "TorrServerVLCAgent/1.0"

    def do_GET(self) -> None:  # noqa: N802
        if self.path != "/health":
            self._json(404, {"ok": False, "error": "not found"})
            return
        if not self._authorized():
            self._unauthorized()
            return
        self._json(200, {"ok": True, "player_running": self.server.player.is_running()})

    def do_POST(self) -> None:  # noqa: N802
        if self.path != "/play":
            self._json(404, {"ok": False, "error": "not found"})
            return
        if not self._authorized():
            self._unauthorized()
            return

        try:
            payload = self._read_json()
            request = validate_play_request(payload, self.server.config.allowed_hosts)
            self.server.player.launch(request.stream_url, request.fullscreen)
        except ValueError as error:
            self._json(400, {"ok": False, "error": str(error)})
            return
        except OSError as error:
            logging.exception("Failed to start player")
            self._json(500, {"ok": False, "error": f"failed to start player: {error}"})
            return

        self._json(202, {"ok": True})

    def _authorized(self) -> bool:
        expected = self.server.config.token
        if not expected:
            return True
        header = self.headers.get("Authorization", "")
        prefix = "Bearer "
        if not header.startswith(prefix):
            return False
        return hmac.compare_digest(header[len(prefix) :].encode(), expected.encode())

    def _unauthorized(self) -> None:
        body = json.dumps({"ok": False, "error": "unauthorized"}).encode("utf-8")
        self.send_response(401)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.send_header("WWW-Authenticate", "Bearer")
        self.send_header("Cache-Control", "no-store")
        self.end_headers()
        self.wfile.write(body)

    def _read_json(self) -> dict[str, Any]:
        content_type = self.headers.get("Content-Type", "").split(";", 1)[0].strip().lower()
        if content_type != "application/json":
            raise ValueError("content type must be application/json")
        try:
            length = int(self.headers.get("Content-Length", "0"))
        except ValueError as error:
            raise ValueError("invalid content length") from error
        if length <= 0 or length > MAX_BODY_BYTES:
            raise ValueError("invalid request body size")
        try:
            payload = json.loads(self.rfile.read(length))
        except json.JSONDecodeError as error:
            raise ValueError("invalid JSON") from error
        if not isinstance(payload, dict):
            raise ValueError("request body must be a JSON object")
        return payload

    def _json(self, status: int, payload: dict[str, Any]) -> None:
        body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.send_header("Cache-Control", "no-store")
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, fmt: str, *args: object) -> None:
        logging.info("%s - %s", self.client_address[0], fmt % args)


def validate_play_request(payload: dict[str, Any], allowed_hosts: frozenset[str]) -> PlayRequest:
    raw_path = payload.get("path")
    raw_hash = payload.get("hash")
    raw_index = payload.get("index")
    raw_stream_url = payload.get("stream_url")
    raw_fullscreen = payload.get("fullscreen", False)

    if not isinstance(raw_path, str) or not raw_path.strip():
        raise ValueError("path is required")
    file_name = PurePath(raw_path.replace("\\", "/")).name
    if not file_name or file_name in {".", ".."} or len(file_name) > 1024:
        raise ValueError("invalid path")

    if not isinstance(raw_hash, str) or not HASH_PATTERN.fullmatch(raw_hash):
        raise ValueError("invalid torrent hash")
    normalized_hash = raw_hash.lower()

    if isinstance(raw_index, bool):
        raise ValueError("invalid file index")
    try:
        index = int(raw_index)
    except (TypeError, ValueError) as error:
        raise ValueError("invalid file index") from error
    if index < 0 or index > 100_000:
        raise ValueError("invalid file index")
    if not isinstance(raw_fullscreen, bool):
        raise ValueError("fullscreen must be a boolean")

    if not isinstance(raw_stream_url, str) or not raw_stream_url:
        raise ValueError("stream_url is required")
    if len(raw_stream_url) > MAX_STREAM_URL_LENGTH:
        raise ValueError("stream_url is too long")

    parsed = urlparse(raw_stream_url)
    if parsed.scheme not in {"http", "https"} or not parsed.hostname:
        raise ValueError("stream_url must use http or https")
    if parsed.username is not None or parsed.password is not None or parsed.fragment:
        raise ValueError("stream_url must not contain credentials or a fragment")
    if allowed_hosts and parsed.hostname.lower() not in allowed_hosts:
        raise ValueError("stream_url host is not allowed")

    query = parse_qs(parsed.query, keep_blank_values=True)
    if query.get("link", [""])[0].lower() != normalized_hash:
        raise ValueError("stream_url torrent hash does not match the request")
    if query.get("index", [""])[0] != str(index):
        raise ValueError("stream_url file index does not match the request")
    if "play" not in query:
        raise ValueError("stream_url is missing the play parameter")

    stream_name = unquote(parsed.path.rsplit("/", 1)[-1])
    if stream_name != file_name:
        raise ValueError("stream_url filename does not match the request")

    return PlayRequest(
        path=file_name,
        hash=normalized_hash,
        index=index,
        stream_url=raw_stream_url,
        fullscreen=raw_fullscreen,
    )


def parse_config(argv: list[str] | None = None) -> Config:
    env_player_args = shlex.split(os.getenv("VLC_AGENT_PLAYER_ARGS", ""))
    env_allowed_hosts = split_csv(os.getenv("VLC_AGENT_ALLOWED_HOSTS", ""))

    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--host", default=os.getenv("VLC_AGENT_HOST", "127.0.0.1"))
    parser.add_argument("--port", type=int, default=int(os.getenv("VLC_AGENT_PORT", "8092")))
    parser.add_argument("--token", default=os.getenv("VLC_AGENT_TOKEN", ""))
    parser.add_argument("--allowed-host", action="append", default=[])
    parser.add_argument("--player", default=os.getenv("VLC_AGENT_PLAYER", "vlc"))
    parser.add_argument("--player-arg", action="append", default=[])
    parser.add_argument(
        "--stop-timeout",
        type=float,
        default=float(os.getenv("VLC_AGENT_STOP_TIMEOUT", "3")),
    )
    parser.add_argument(
        "--allow-unauthenticated-network",
        action="store_true",
        help="allow a non-loopback listener without a bearer token",
    )
    args = parser.parse_args(argv)

    if not 1 <= args.port <= 65535:
        parser.error("port must be between 1 and 65535")
    if args.stop_timeout <= 0:
        parser.error("stop timeout must be positive")
    if not args.player.strip():
        parser.error("player command is required")
    if not args.token and not is_loopback_host(args.host) and not args.allow_unauthenticated_network:
        parser.error("a bearer token is required for a non-loopback listener")

    player_args = tuple(args.player_arg or env_player_args or DEFAULT_PLAYER_ARGS)
    allowed_hosts = frozenset(host.lower() for host in [*env_allowed_hosts, *args.allowed_host] if host)
    return Config(
        host=args.host,
        port=args.port,
        token=args.token,
        allowed_hosts=allowed_hosts,
        player=args.player,
        player_args=player_args,
        stop_timeout=args.stop_timeout,
    )


def split_csv(value: str) -> list[str]:
    return [part.strip() for part in value.split(",") if part.strip()]


def is_loopback_host(host: str) -> bool:
    if host.lower() == "localhost":
        return True
    try:
        return ipaddress.ip_address(host).is_loopback
    except ValueError:
        return False


def run(config: Config) -> None:
    player = PlayerManager(config.player, config.player_args, config.stop_timeout)
    server = AgentHTTPServer((config.host, config.port), config, player)

    def request_shutdown(signum: int, _frame: object) -> None:
        logging.info("Received signal %s; shutting down", signum)
        threading.Thread(target=server.shutdown, daemon=True).start()

    signal.signal(signal.SIGINT, request_shutdown)
    signal.signal(signal.SIGTERM, request_shutdown)

    logging.info("TorrServer VLC agent listening on http://%s:%d", config.host, config.port)
    if config.allowed_hosts:
        logging.info("Allowed TorrServer hosts: %s", ", ".join(sorted(config.allowed_hosts)))
    try:
        server.serve_forever(poll_interval=0.25)
    finally:
        server.server_close()
        player.stop()


def main() -> None:
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
    run(parse_config())


if __name__ == "__main__":
    main()
