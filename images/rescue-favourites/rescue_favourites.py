"""Move favourited Explo tracks into the main library before Explo's weekly cleanup."""

import hashlib
import json
import os
import secrets
import sys
import time
import unicodedata
import urllib.parse
import urllib.request

RESERVED_DIRS = ("Singles", "explo")
SIDECAR_EXTS = (".lrc",)
SCAN_TIMEOUT_S = 600


def log(msg):
    print(msg, flush=True)


def match_key(name):
    # Library folders use Unicode lookalikes (alt‐J, en dashes); compare on letters/digits only
    s = unicodedata.normalize("NFKC", name).casefold()
    key = "".join(c for c in s if c.isalnum())
    return key or s


def sanitize(name):
    name = name.replace("/", "-").replace("\x00", "")
    return name.strip().rstrip(".") or "Unknown"


def find_existing_dir(parent, wanted):
    key = match_key(wanted)
    try:
        entries = sorted(os.listdir(parent))
    except FileNotFoundError:
        return None
    for entry in entries:
        if os.path.isdir(os.path.join(parent, entry)) and match_key(entry) == key:
            return entry
    return None


def dest_filename(artist, album, track, title, ext):
    parts = [artist, album]
    if track:
        parts.append(f"{int(track):02d}")
    parts.append(title)
    return sanitize("_".join(sanitize(p) for p in parts)) + ext


def relative_explo_path(song_path, explo_subdir):
    # Navidrome may or may not prefix the library path; anchor on the explo subdir
    norm = unicodedata.normalize("NFC", song_path).replace("\\", "/")
    marker = explo_subdir.strip("/") + "/"
    idx = norm.find(marker)
    if idx == -1 or (idx > 0 and norm[idx - 1] != "/"):
        return None
    return norm[idx:]


def plan_move(song, music_root, explo_subdir):
    """Return a move plan dict, or (None, reason)."""
    rel = relative_explo_path(song.get("path", ""), explo_subdir)
    if rel is None:
        return None, "not under explo dir"
    src = os.path.join(music_root, rel)
    explo_root = os.path.realpath(os.path.join(music_root, explo_subdir))
    if not os.path.realpath(src).startswith(explo_root + os.sep):
        return None, "resolves outside explo dir"
    if not os.path.isfile(src):
        return None, "source file missing"

    artist = song.get("displayAlbumArtist") or song.get("albumArtist") or song.get("artist")
    album = song.get("album")
    title = song.get("title")
    if not (artist and album and title):
        return None, "missing artist/album/title tag"

    if match_key(artist) in {match_key(d) for d in RESERVED_DIRS}:
        return None, "artist name collides with a reserved library folder"
    artist_dir = find_existing_dir(music_root, artist) or sanitize(artist)
    album_dir = find_existing_dir(os.path.join(music_root, artist_dir), album) or sanitize(album)

    ext = os.path.splitext(src)[1]
    name = dest_filename(artist, album, song.get("track"), title, ext)
    dest = os.path.join(music_root, artist_dir, album_dir, name)
    return {"src": src, "dest": dest, "title": title}, None


def sidecars(src, dest):
    stem_src, stem_dest = os.path.splitext(src)[0], os.path.splitext(dest)[0]
    for ext in SIDECAR_EXTS:
        if os.path.isfile(stem_src + ext):
            yield stem_src + ext, stem_dest + ext


def move(plan):
    os.makedirs(os.path.dirname(plan["dest"]), mode=0o755, exist_ok=True)
    os.rename(plan["src"], plan["dest"])
    for s, d in sidecars(plan["src"], plan["dest"]):
        if not os.path.exists(d):
            os.rename(s, d)


class Subsonic:
    def __init__(self, base, user, password):
        self.base, self.user, self.password = base.rstrip("/"), user, password

    def call(self, endpoint, **params):
        salt = secrets.token_hex(8)
        token = hashlib.md5((self.password + salt).encode()).hexdigest()
        query = urllib.parse.urlencode(
            {"u": self.user, "t": token, "s": salt, "v": "1.16.1", "c": "rescue-favourites", "f": "json", **params}
        )
        with urllib.request.urlopen(f"{self.base}/rest/{endpoint}?{query}", timeout=30) as resp:
            body = json.load(resp)["subsonic-response"]
        if body.get("status") != "ok":
            raise RuntimeError(f"{endpoint} failed: {body.get('error')}")
        return body

    def starred_songs(self):
        return self.call("getStarred2").get("starred2", {}).get("song", [])

    def scan_and_wait(self):
        self.call("startScan")
        deadline = time.time() + SCAN_TIMEOUT_S
        while time.time() < deadline:
            time.sleep(3)
            if not self.call("getScanStatus")["scanStatus"].get("scanning"):
                return
        raise RuntimeError("library scan did not finish in time")

    def find_song_by_path(self, title, rel_path):
        want = unicodedata.normalize("NFC", rel_path)
        for s in self.call("search3", query=title, songCount=50, artistCount=0, albumCount=0).get(
            "searchResult3", {}
        ).get("song", []):
            if unicodedata.normalize("NFC", s.get("path", "")).endswith(want):
                return s
        return None

    def star(self, song_id):
        self.call("star", id=song_id)


def main():
    dry_run = os.environ.get("DRY_RUN", "true").lower() != "false"
    music_root = os.environ.get("MUSIC_ROOT", "/music")
    explo_subdir = os.environ.get("EXPLO_SUBDIR", "explo/barend")
    api = Subsonic(os.environ["NAVIDROME_URL"], os.environ["SYSTEM_USERNAME"], os.environ["SYSTEM_PASSWORD"])

    log(f"rescue-favourites start dry_run={dry_run}")
    songs = api.starred_songs()
    log(f"starred songs: {len(songs)}")

    plans, errors = [], 0
    for song in songs:
        plan, reason = plan_move(song, music_root, explo_subdir)
        if plan is None:
            if reason != "not under explo dir":
                log(f"SKIP {song.get('path')}: {reason}")
            continue
        plans.append(plan)

    to_restar = []
    for plan in plans:
        rel_dest = os.path.relpath(plan["dest"], music_root)
        rel_src = os.path.relpath(plan["src"], music_root)
        if os.path.exists(plan["dest"]):
            log(f"EXISTS {rel_src} -> {rel_dest}: already in library, not overwriting")
            to_restar.append((plan["title"], rel_dest))
            continue
        log(f"{'WOULD MOVE' if dry_run else 'MOVE'} {rel_src} -> {rel_dest}")
        if dry_run:
            continue
        try:
            move(plan)
            to_restar.append((plan["title"], rel_dest))
        except OSError as e:
            errors += 1
            log(f"ERROR moving {rel_src}: {e}")

    if to_restar and not dry_run:
        api.scan_and_wait()
        for title, rel_dest in to_restar:
            song = api.find_song_by_path(title, rel_dest)
            if song is None:
                errors += 1
                log(f"ERROR could not find {rel_dest} in Navidrome after scan")
            elif not song.get("starred"):
                api.star(song["id"])
                log(f"RESTAR {rel_dest}")

    log(f"done: planned={len(plans)} restar={len(to_restar)} errors={errors}")
    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main())
