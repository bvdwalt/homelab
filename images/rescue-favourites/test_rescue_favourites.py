import os
import tempfile
import unittest

import rescue_favourites as rf


def touch(path):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    open(path, "w").close()


class MatchKeyTest(unittest.TestCase):
    def test_unicode_dash_matches_ascii(self):
        self.assertEqual(rf.match_key("alt‐J"), rf.match_key("alt-J"))

    def test_en_dash_and_case(self):
        self.assertEqual(rf.match_key("Mercury – Act 1"), rf.match_key("mercury - act 1"))

    def test_symbol_only_name_not_empty(self):
        self.assertTrue(rf.match_key("!!!"))


class DestFilenameTest(unittest.TestCase):
    def test_matches_library_pattern(self):
        self.assertEqual(
            rf.dest_filename("Glass Animals", "Dreamland", 14, "Heat Waves", ".flac"),
            "Glass Animals_Dreamland_14_Heat Waves.flac",
        )

    def test_slash_in_title_is_replaced(self):
        self.assertEqual(rf.dest_filename("AC/DC", "Back", 1, "A/B", ".flac"), "AC-DC_Back_01_A-B.flac")

    def test_missing_track_number(self):
        self.assertEqual(rf.dest_filename("A", "B", None, "T", ".flac"), "A_B_T.flac")


class RelativeExploPathTest(unittest.TestCase):
    def test_plain_relative(self):
        self.assertEqual(
            rf.relative_explo_path("explo/barend/Weekly-Exploration/x.flac", "explo/barend"),
            "explo/barend/Weekly-Exploration/x.flac",
        )

    def test_library_prefixed(self):
        self.assertEqual(
            rf.relative_explo_path("Music/explo/barend/W/x.flac", "explo/barend"), "explo/barend/W/x.flac"
        )

    def test_main_library_path_rejected(self):
        self.assertIsNone(rf.relative_explo_path("Glass Animals/Dreamland/x.flac", "explo/barend"))

    def test_lookalike_dir_rejected(self):
        self.assertIsNone(rf.relative_explo_path("myexplo/barend/x.flac", "explo/barend"))


class PlanMoveTest(unittest.TestCase):
    def setUp(self):
        self._tmp = tempfile.TemporaryDirectory()
        self.root = self._tmp.name
        self.addCleanup(self._tmp.cleanup)
        self.src_rel = "explo/barend/Weekly-Exploration/08 junk name.flac"
        touch(os.path.join(self.root, self.src_rel))

    def song(self, **kw):
        base = {"path": self.src_rel, "artist": "alt-J", "album": "The Dream", "title": "Delta", "track": 10}
        return {**base, **kw}

    def test_reuses_existing_dirs_despite_dash_lookalike(self):
        os.makedirs(os.path.join(self.root, "alt‐J", "The Dream"))
        plan, _ = rf.plan_move(self.song(), self.root, "explo/barend")
        self.assertEqual(plan["dest"], os.path.join(self.root, "alt‐J", "The Dream", "alt-J_The Dream_10_Delta.flac"))

    def test_creates_new_names_when_no_match(self):
        plan, _ = rf.plan_move(self.song(), self.root, "explo/barend")
        self.assertEqual(plan["dest"], os.path.join(self.root, "alt-J", "The Dream", "alt-J_The Dream_10_Delta.flac"))

    def test_prefers_album_artist(self):
        plan, _ = rf.plan_move(self.song(displayAlbumArtist="Various"), self.root, "explo/barend")
        self.assertIn(os.path.join(self.root, "Various"), plan["dest"])

    def test_artist_named_like_reserved_dir_is_skipped(self):
        for name in ("Singles", "explo"):
            plan, reason = rf.plan_move(self.song(artist=name), self.root, "explo/barend")
            self.assertIsNone(plan)
            self.assertIn("reserved", reason)

    def test_missing_source(self):
        plan, reason = rf.plan_move(self.song(path="explo/barend/nope.flac"), self.root, "explo/barend")
        self.assertIsNone(plan)
        self.assertEqual(reason, "source file missing")

    def test_missing_tags(self):
        plan, reason = rf.plan_move(self.song(album=None), self.root, "explo/barend")
        self.assertIsNone(plan)
        self.assertIn("missing", reason)

    def test_symlink_escape_rejected(self):
        outside = tempfile.mkdtemp()
        self.addCleanup(lambda: os.rmdir(outside) if os.path.isdir(outside) else None)
        real = os.path.join(outside, "x.flac")
        touch(real)
        self.addCleanup(lambda: os.remove(real))
        link = os.path.join(self.root, "explo/barend/link.flac")
        os.symlink(real, link)
        plan, reason = rf.plan_move(self.song(path="explo/barend/link.flac"), self.root, "explo/barend")
        self.assertIsNone(plan)
        self.assertEqual(reason, "resolves outside explo dir")


class MoveTest(unittest.TestCase):
    def test_moves_file_and_lrc_sidecar(self):
        with tempfile.TemporaryDirectory() as root:
            src = os.path.join(root, "explo/barend/W/a.flac")
            touch(src)
            touch(os.path.join(root, "explo/barend/W/a.lrc"))
            dest = os.path.join(root, "Artist", "Album", "Artist_Album_01_T.flac")
            rf.move({"src": src, "dest": dest})
            self.assertTrue(os.path.isfile(dest))
            self.assertTrue(os.path.isfile(dest[:-5] + ".lrc"))
            self.assertFalse(os.path.exists(src))


if __name__ == "__main__":
    unittest.main()
