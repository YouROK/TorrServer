from __future__ import annotations

import io
import json
import sys
import threading
import unittest
from contextlib import redirect_stderr
from unittest.mock import Mock, patch
from urllib.error import HTTPError
from urllib.request import Request, urlopen

from torrserver_vlc_agent import AgentHTTPServer, Config, PlayerManager, parse_config, validate_play_request

HASH = "0" * 40
TOKEN = "[REDACTED_SECRET]"
STREAM_URL = f"http://movies.local:8090/stream/Movie.mkv?link={HASH}&index=1&play="


class RecordingPlayer:
    def __init__(self) -> None:
        self.calls: list[tuple[str, bool]] = []

    def launch(self, stream_url: str, fullscreen: bool) -> None:
        self.calls.append((stream_url, fullscreen))

    def is_running(self) -> bool:
        return bool(self.calls)


class AgentServerTest(unittest.TestCase):
    def setUp(self) -> None:
        self.player = RecordingPlayer()
        self.config = Config(
            host="127.0.0.1",
            port=0,
            token=TOKEN,
            allowed_hosts=frozenset({"movies.local"}),
            player="vlc",
            player_args=(),
            stop_timeout=1,
        )
        self.server = AgentHTTPServer((self.config.host, 0), self.config, self.player)  # type: ignore[arg-type]
        self.thread = threading.Thread(target=self.server.serve_forever, daemon=True)
        self.thread.start()
        self.base_url = f"http://127.0.0.1:{self.server.server_address[1]}"

    def tearDown(self) -> None:
        self.server.shutdown()
        self.server.server_close()
        self.thread.join(timeout=2)

    def request(self, path: str, *, method: str = "GET", payload: dict | None = None, token: str | None = None):
        data = None
        headers = {}
        if payload is not None:
            data = json.dumps(payload).encode("utf-8")
            headers["Content-Type"] = "application/json"
        if token is not None:
            headers["Authorization"] = f"Bearer {token}"
        request = Request(self.base_url + path, data=data, headers=headers, method=method)
        return urlopen(request, timeout=3)

    def test_health_requires_configured_token(self) -> None:
        with self.assertRaises(HTTPError) as unauthorized:
            self.request("/health")
        self.assertEqual(unauthorized.exception.code, 401)

        with self.request("/health", token=TOKEN) as response:
            self.assertEqual(response.status, 200)
            self.assertFalse(json.load(response)["player_running"])

    def test_valid_play_request_starts_player(self) -> None:
        payload = {
            "path": "Movie.mkv",
            "hash": HASH,
            "index": 1,
            "stream_url": STREAM_URL,
            "fullscreen": True,
        }
        with self.request("/play", method="POST", payload=payload, token=TOKEN) as response:
            self.assertEqual(response.status, 202)
        self.assertEqual(self.player.calls, [(STREAM_URL, True)])

    def test_disallowed_stream_host_is_rejected(self) -> None:
        payload = {
            "path": "Movie.mkv",
            "hash": HASH,
            "index": 1,
            "stream_url": STREAM_URL.replace("movies.local", "other.local"),
        }
        with self.assertRaises(HTTPError) as rejected:
            self.request("/play", method="POST", payload=payload, token=TOKEN)
        self.assertEqual(rejected.exception.code, 400)
        self.assertEqual(self.player.calls, [])


class ValidationTest(unittest.TestCase):
    def test_request_fields_must_match_stream_url(self) -> None:
        request = validate_play_request(
            {"path": "folder/Movie.mkv", "hash": HASH.upper(), "index": 1, "stream_url": STREAM_URL},
            frozenset({"movies.local"}),
        )
        self.assertEqual(request.path, "Movie.mkv")
        self.assertEqual(request.hash, HASH)
        self.assertFalse(request.fullscreen)

        fullscreen_request = validate_play_request(
            {
                "path": "Movie.mkv",
                "hash": HASH,
                "index": 1,
                "stream_url": STREAM_URL,
                "fullscreen": True,
            },
            frozenset({"movies.local"}),
        )
        self.assertTrue(fullscreen_request.fullscreen)

        invalid_payloads = [
            {"path": "Movie.mkv", "hash": "bad", "index": 1, "stream_url": STREAM_URL},
            {"path": "Other.mkv", "hash": HASH, "index": 1, "stream_url": STREAM_URL},
            {"path": "Movie.mkv", "hash": HASH, "index": 2, "stream_url": STREAM_URL},
            {"path": "Movie.mkv", "hash": HASH, "index": 1, "stream_url": STREAM_URL, "fullscreen": "yes"},
            {
                "path": "Movie.mkv",
                "hash": HASH,
                "index": 1,
                "stream_url": STREAM_URL.replace("play=", "missing="),
            },
        ]
        for payload in invalid_payloads:
            with self.subTest(payload=payload), self.assertRaises(ValueError):
                validate_play_request(payload, frozenset({"movies.local"}))

    def test_network_listener_requires_token_by_default(self) -> None:
        with redirect_stderr(io.StringIO()), self.assertRaises(SystemExit):
            parse_config(["--host", "0.0.0.0"])

        config = parse_config(["--host", "0.0.0.0", "--token", TOKEN])
        self.assertEqual(config.token, TOKEN)


class PlayerManagerTest(unittest.TestCase):
    @patch("torrserver_vlc_agent.subprocess.Popen")
    def test_manager_appends_requested_window_mode(self, popen: Mock) -> None:
        process = Mock()
        process.poll.return_value = 0
        popen.return_value = process
        manager = PlayerManager("vlc", ("--no-one-instance", "--fullscreen"), stop_timeout=1)

        manager.launch("http://example/windowed", False)
        command = popen.call_args.args[0]
        self.assertEqual(command[-2:], ["--no-fullscreen", "http://example/windowed"])

        manager.launch("http://example/fullscreen", True)
        command = popen.call_args.args[0]
        self.assertEqual(command[-2:], ["--fullscreen", "http://example/fullscreen"])

    def test_manager_replaces_only_its_own_process(self) -> None:
        manager = PlayerManager(
            sys.executable,
            ("-c", "import time; time.sleep(30)"),
            stop_timeout=1,
        )
        try:
            manager.launch("http://example/first", False)
            first = manager._process  # noqa: SLF001 - intentional white-box lifecycle test
            self.assertIsNotNone(first)
            self.assertTrue(manager.is_running())

            manager.launch("http://example/second", True)
            second = manager._process  # noqa: SLF001
            self.assertIsNotNone(second)
            self.assertNotEqual(first.pid, second.pid)
            self.assertIsNotNone(first.poll())
            self.assertTrue(manager.is_running())
        finally:
            manager.stop()
        self.assertFalse(manager.is_running())


if __name__ == "__main__":
    unittest.main()
