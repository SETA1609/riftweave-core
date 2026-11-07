#!/usr/bin/env python3
"""The perk catalogue's content floor (docs/modules/progression.md#perks-and-traits).

Run with:
    python -m unittest ruleset/scripts/test_perks.py
"""

import json
import unittest
from pathlib import Path

FEATURES = Path(__file__).resolve().parent.parent / "data" / "features" / "core.json"
ABILITIES = ("str", "per", "end", "int", "wil", "agi", "cha", "lck")


def perks():
    return [f for f in json.loads(FEATURES.read_text())["features"] if f["type"] == "perk"]


def level(perk):
    return perk.get("prerequisite", {}).get("level", 1)


def emitted(perk):
    """The (target, op) pairs a perk's modifiers name."""
    return frozenset((m["target"], m["op"]) for m in perk.get("modifiers", []))


class TestPerkCatalogue(unittest.TestCase):
    def test_at_least_sixty_perks(self):
        self.assertGreaterEqual(len(perks()), 60)

    def test_every_perk_emits_trained_modifiers_from_itself(self):
        for perk in perks():
            with self.subTest(perk=perk["key"]):
                self.assertTrue(perk.get("modifiers"), "a perk emits modifiers")
                for modifier in perk["modifiers"]:
                    self.assertEqual(modifier["source"], f"feature:{perk['key']}")
                    self.assertEqual(modifier["origin"], "core:origin/trained")

    def test_prerequisite_perks_are_keys_of_perks(self):
        keys = {perk["key"] for perk in perks()}
        for perk in perks():
            for ref in perk.get("prerequisite", {}).get("perks", []):
                with self.subTest(perk=perk["key"], needs=ref):
                    self.assertTrue(ref.startswith("core:feature/"))
                    self.assertIn(ref.removeprefix("core:feature/"), keys)

    def test_a_prerequisite_perk_is_reachable_at_a_lower_or_equal_level(self):
        by_key = {perk["key"]: perk for perk in perks()}
        for perk in perks():
            for ref in perk.get("prerequisite", {}).get("perks", []):
                with self.subTest(perk=perk["key"], needs=ref):
                    self.assertLessEqual(level(by_key[ref.removeprefix("core:feature/")]), level(perk))

    def test_every_attribute_gates_a_perk(self):
        gated = {a for perk in perks() for a in perk.get("prerequisite", {}).get("abilities", {})}
        self.assertEqual(gated, set(ABILITIES))

    def test_combat_magic_and_social_each_have_perks(self):
        counts = {}
        for perk in perks():
            counts[perk["category"]] = counts.get(perk["category"], 0) + 1
        for category in ("combat", "magic", "social"):
            with self.subTest(category=category):
                self.assertGreaterEqual(counts.get(category, 0), 8)

    def test_no_two_perks_at_one_level_emit_the_same_targets(self):
        """A perk is dominated when another at its level does the same thing; the catalogue
        avoids it by giving each perk at one level its own set of targets and ops."""
        seen = {}
        for perk in perks():
            key = (level(perk), emitted(perk))
            with self.subTest(perk=perk["key"]):
                self.assertNotIn(key, seen, f"same level and targets as {seen.get(key)}")
            seen[key] = perk["key"]


if __name__ == "__main__":
    unittest.main()
