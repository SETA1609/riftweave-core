#!/usr/bin/env python3
"""The grouped sheet's groups (docs/core/progression.md#the-grouped-sheet-variant).

Run with:
    python -m unittest ruleset/scripts/test_skills.py
"""

import json
import unittest
from pathlib import Path

DATA = Path(__file__).resolve().parent.parent / "data"


def load(rel, collection):
    return json.loads((DATA / rel).read_text())[collection]


def key_of(ref):
    return ref.rsplit("/", 1)[-1]


class TestSkillGroups(unittest.TestCase):
    def setUp(self):
        self.skills = {row["key"] for row in load("skills/core.json", "skills")}
        self.groups = load("skill_groups/core.json", "skill_groups")

    def test_every_skill_is_in_exactly_one_group(self):
        seen = {}
        for group in self.groups:
            for ref in group["skills"]:
                with self.subTest(skill=ref):
                    self.assertNotIn(key_of(ref), seen, f"also in {seen.get(key_of(ref))}")
                seen[key_of(ref)] = group["key"]
        self.assertEqual(set(seen), self.skills)

    def test_about_a_dozen_groups_besides_the_magic_colors(self):
        """Each magic color is its own group, so a table plays only the colors it uses."""
        magic = [g for g in self.groups if all(key_of(r).endswith("_magic") for r in g["skills"])]
        self.assertEqual(len(magic), 9)
        self.assertTrue(all(len(g["skills"]) == 1 for g in magic))
        self.assertTrue(10 <= len(self.groups) - len(magic) <= 14)

    def test_only_the_armor_skills_are_untracked(self):
        untracked = {key_of(r) for g in self.groups if not g["is_tracked"] for r in g["skills"]}
        self.assertEqual(untracked, {"light_armor", "medium_armor", "heavy_armor", "unarmored"})

    def test_the_group_cost_is_in_data(self):
        rules = json.loads((DATA / "progression" / "core.json").read_text())["progression"]
        self.assertGreaterEqual(rules["skill_group_cost"]["members_per_step"], 1)


if __name__ == "__main__":
    unittest.main()
