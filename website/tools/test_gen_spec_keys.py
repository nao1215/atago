"""Unit tests for gen-spec-keys.py's Since resolution.

Run from the repository root: python3 -m unittest website/tools/test_gen_spec_keys.py
(schema_parity_test.go runs it as part of `go test`).
"""
import importlib.util
import pathlib
import unittest

_spec = importlib.util.spec_from_file_location(
    "gen_spec_keys", pathlib.Path(__file__).with_name("gen-spec-keys.py"))
gen = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(gen)


class SinceTest(unittest.TestCase):
    def test_a_tagged_key_takes_its_first_tag(self):
        self.assertEqual(gen.since("run.command", {"run.command": "v0.1.0"}, {"run.command": "v0.9.0"}, "v0.23.0"), "v0.1.0")

    def test_a_release_pr_stamp_survives_until_the_tag_exists(self):
        # The release PR stamped the key; the site is built from the merge
        # before v0.24.0 is tagged, and must not fall back to "unreleased".
        self.assertEqual(gen.since("screenAssert.images", {}, {"screenAssert.images": "v0.24.0"}, "v0.23.0"), "v0.24.0")

    def test_a_stamp_not_newer_than_the_latest_tag_is_dropped(self):
        # v0.23.0 exists and does not carry the key, so the stamp is wrong.
        self.assertEqual(gen.since("x.y", {}, {"x.y": "v0.23.0"}, "v0.23.0"), "unreleased")
        self.assertEqual(gen.since("x.y", {}, {"x.y": "v0.9.0"}, "v0.23.0"), "unreleased")

    def test_versions_compare_numerically(self):
        self.assertEqual(gen.since("x.y", {}, {"x.y": "v0.100.0"}, "v0.99.0"), "v0.100.0")

    def test_a_new_key_is_unreleased(self):
        self.assertEqual(gen.since("x.y", {}, {}, "v0.23.0"), "unreleased")
        self.assertEqual(gen.since("x.y", {}, {"x.y": "unreleased"}, "v0.23.0"), "unreleased")

    def test_without_any_tag_a_stamp_is_kept(self):
        self.assertEqual(gen.since("x.y", {}, {"x.y": "v0.1.0"}, None), "v0.1.0")


if __name__ == "__main__":
    unittest.main()
