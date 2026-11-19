"""End-to-end run of main() against a fake Navidrome.

Like the real server, the Subsonic API reports virtual tag-based paths while the native
REST API reports the real on-disk path.
"""

import contextlib
import io
import json
import os
import tempfile
import threading
import unittest
from http.server import BaseHTTPRequestHandler, HTTPServer
from unittest import mock
from urllib.parse import parse_qs, urlparse

import rescue_favourites as rf

SRC_REL = "explo/barend/Weekly-Exploration/sombr_I Barely Know Her_02_12 to 12.flac"
DEST_REL = "Sombr/I Barely Know Her/Sombr_I Barely Know Her_02_12 to 12.flac"


class FakeNavidrome(BaseHTTPRequestHandler):
    state = None

    def log_message(self, *a):
        pass

    def _send(self, body, status=200):
        data = json.dumps(body).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(data)

    def do_POST(self):
        if self.path == "/auth/login":
            self.rfile.read(int(self.headers.get("Content-Length", 0)))
            return self._send({"token": "tok"})
        self._send({}, 404)

    def do_GET(self):
        url = urlparse(self.path)
        q = parse_qs(url.query)
        st = type(self).state
        if url.path == "/api/song":
            if self.headers.get("x-nd-authorization") != "Bearer tok":
                return self._send({}, 401)
            songs = st["native"]
            if q.get("starred") == ["true"]:
                songs = [s for s in songs if s["starred"]]
            start = int(q.get("_start", ["0"])[0])
            end = int(q.get("_end", ["1000"])[0])
            return self._send(songs[start:end])
        if url.path.startswith("/rest/"):
            ep = url.path.split("/")[-1]
            ok = {"subsonic-response": {"status": "ok"}}
            if ep == "getStarred2":
                ok["subsonic-response"]["starred2"] = {
                    "song": [{"id": "1", "path": "Sombr/I Barely Know Her/02 - 12 to 12.flac", "title": "12 to 12"}]
                }
            elif ep == "getScanStatus":
                ok["subsonic-response"]["scanStatus"] = {"scanning": False}
            elif ep == "startScan":
                st["scans"] += 1
                st["native"].append(
                    {"id": "new1", "path": DEST_REL, "title": "12 to 12", "starred": False}
                )
            elif ep == "star":
                st["starred_ids"].append(q["id"][0])
            return self._send(ok)
        self._send({}, 404)


class MainEndToEndTest(unittest.TestCase):
    def setUp(self):
        self._tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self._tmp.cleanup)
        self.root = self._tmp.name
        src = os.path.join(self.root, SRC_REL)
        os.makedirs(os.path.dirname(src))
        open(src, "w").close()

        FakeNavidrome.state = {
            "scans": 0,
            "starred_ids": [],
            "native": [
                {
                    "id": "1",
                    "path": SRC_REL,
                    "title": "12 to 12",
                    "album": "I Barely Know Her",
                    "albumArtist": "Sombr",
                    "artist": "Sombr",
                    "trackNumber": 2,
                    "starred": True,
                },
                {"id": "2", "path": "Singles/X/y.flac", "title": "y", "starred": True},
            ],
        }
        self.server = HTTPServer(("127.0.0.1", 0), FakeNavidrome)
        threading.Thread(target=self.server.serve_forever, daemon=True).start()
        self.addCleanup(self.server.server_close)
        self.addCleanup(self.server.shutdown)

    def run_main(self, dry_run):
        env = {
            "NAVIDROME_URL": f"http://127.0.0.1:{self.server.server_port}",
            "SYSTEM_USERNAME": "u",
            "SYSTEM_PASSWORD": "p",
            "MUSIC_ROOT": self.root,
            "EXPLO_SUBDIR": "explo/barend",
            "DRY_RUN": "true" if dry_run else "false",
        }
        out = io.StringIO()
        with mock.patch.dict(os.environ, env), mock.patch.object(rf.time, "sleep"), contextlib.redirect_stdout(out):
            code = rf.main()
        return code, out.getvalue()

    def test_dry_run_plans_move_from_real_path(self):
        code, out = self.run_main(dry_run=True)
        self.assertEqual(code, 0)
        self.assertIn(f"WOULD MOVE {SRC_REL} -> {DEST_REL}", out)
        self.assertTrue(os.path.exists(os.path.join(self.root, SRC_REL)))
        self.assertEqual(FakeNavidrome.state["scans"], 0)

    def test_live_run_moves_rescans_and_restars(self):
        code, out = self.run_main(dry_run=False)
        self.assertEqual(code, 0, out)
        self.assertTrue(os.path.isfile(os.path.join(self.root, DEST_REL)))
        self.assertFalse(os.path.exists(os.path.join(self.root, SRC_REL)))
        self.assertEqual(FakeNavidrome.state["scans"], 1)
        self.assertEqual(FakeNavidrome.state["starred_ids"], ["new1"])


if __name__ == "__main__":
    unittest.main()
