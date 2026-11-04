#!/usr/bin/env python3
"""key_diff.py lists the keys removed and added between two versions of the data.

Run with:
    python -m unittest ruleset/scripts/test_key_diff.py
"""

import io
import subprocess
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import key_diff  # noqa: E402

REPO_ROOT = Path(__file__).resolve().parent.parent.parent


def wrapper(array_name, *keys):
    return {"$schema": "x", array_name: [{"id": n, "key": k} for n, k in enumerate(keys, 1)]}


class TestCollectKeys(unittest.TestCase):
    def test_collects_collection_and_key_from_a_wrapper_array(self):
        files = {"skills/core.json": wrapper("skills", "blades", "persuasion")}
        self.assertEqual(key_diff.collect_keys(files), {("skills", "blades"), ("skills", "persuasion")})

    def test_a_string_id_is_a_key_and_a_numeric_id_is_not(self):
        files = {"wuxing/core.json": {"phases": [{"id": "wood"}, {"id": 7}]}}
        self.assertEqual(key_diff.collect_keys(files), {("phases", "wood")})

    def test_a_single_entry_file_uses_its_folder_as_the_collection(self):
        files = {"equipment/bases/weapons/dagger.json": {"key": "dagger", "label": "Dagger"}}
        self.assertEqual(key_diff.collect_keys(files), {("weapons", "dagger")})

    def test_a_file_with_no_keys_adds_nothing(self):
        self.assertEqual(key_diff.collect_keys({"a/b.json": {"pool": {"base": 2}}}), set())

    def test_moving_a_key_to_another_file_keeps_it(self):
        old = {"equipment/weapons.json": wrapper("equipment", "dagger")}
        new = {"equipment/blades.json": wrapper("equipment", "dagger")}
        self.assertEqual(key_diff.collect_keys(old), key_diff.collect_keys(new))


class TestDiff(unittest.TestCase):
    def test_a_rename_is_a_removal_and_an_addition(self):
        old = {("skills", "blades")}
        new = {("skills", "swords")}
        removed, added = key_diff.diff_keys(old, new)
        self.assertEqual(removed, [("skills", "blades")])
        self.assertEqual(added, [("skills", "swords")])

    def test_results_are_sorted(self):
        removed, _ = key_diff.diff_keys({("b", "z"), ("a", "y"), ("a", "x")}, set())
        self.assertEqual(removed, [("a", "x"), ("a", "y"), ("b", "z")])


class TestMain(unittest.TestCase):
    def run_main(self, old, new):
        loaders = {"old": lambda: old, "new": lambda: new}
        out = io.StringIO()
        code = key_diff.main(["old", "new"], out=out, loader=lambda ref: loaders[ref]())
        return code, out.getvalue()

    def test_a_removed_key_exits_1_and_is_listed(self):
        code, text = self.run_main({"s.json": wrapper("skills", "a", "b")}, {"s.json": wrapper("skills", "a")})
        self.assertEqual(code, 1)
        self.assertIn("removed (breaking): 1", text)
        self.assertIn("skills/b", text)

    def test_only_additions_exit_0(self):
        code, text = self.run_main({"s.json": wrapper("skills", "a")}, {"s.json": wrapper("skills", "a", "b")})
        self.assertEqual(code, 0)
        self.assertIn("added: 1", text)
        self.assertIn("removed (breaking): 0", text)


class TestGitRefs(unittest.TestCase):
    def test_head_and_the_worktree_agree_on_a_known_key(self):
        try:
            subprocess.run(["git", "rev-parse", "HEAD"], cwd=REPO_ROOT, check=True, capture_output=True)
        except (OSError, subprocess.CalledProcessError):
            self.skipTest("not a git checkout")
        at_head = key_diff.collect_keys(key_diff.load_ref_files("HEAD"))
        on_disk = key_diff.collect_keys(key_diff.load_ref_files("WORKTREE"))
        self.assertIn(("skills", "persuasion"), at_head)
        self.assertIn(("skills", "persuasion"), on_disk)


if __name__ == "__main__":
    unittest.main()
