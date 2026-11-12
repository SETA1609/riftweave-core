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


def features():
    return json.loads(FEATURES.read_text())["features"]


def creation_traits():
    return [f for f in features() if f["type"] == "creation"]


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


class TestOneStatementOfAFeature(unittest.TestCase):
    """M5-35: a feature states its mechanics once, in `modifiers` (docs/modules/progression.md#what-a-perk-emits)."""

    def test_no_feature_has_an_effects_list(self):
        for feature in features():
            with self.subTest(feature=feature["key"]):
                self.assertNotIn("effects", feature)

    def test_every_creation_trait_emits_trained_modifiers_from_itself(self):
        self.assertGreaterEqual(len(creation_traits()), 10)
        for trait in creation_traits():
            with self.subTest(trait=trait["key"]):
                self.assertTrue(trait.get("modifiers"), "a creation trait emits modifiers")
                for modifier in trait["modifiers"]:
                    self.assertEqual(modifier["source"], f"feature:{trait['key']}")
                    self.assertEqual(modifier["origin"], "core:origin/trained")

    def test_every_creation_trait_has_an_upside_and_a_downside(self):
        """A trait is a tradeoff: at least one modifier helps and one hinders or restricts."""
        for trait in creation_traits():
            with self.subTest(trait=trait["key"]):
                self.assertGreaterEqual(len(trait["modifiers"]), 2)

    def test_gifted_adds_one_to_every_attribute(self):
        gifted = next(f for f in features() if f["key"] == "gifted")
        adds = {m["target"]: m["value"] for m in gifted["modifiers"] if m["op"] == "add"}
        self.assertEqual({ability: adds.get(ability) for ability in ABILITIES}, dict.fromkeys(ABILITIES, 1))

    def test_a_phobia_grants_a_flag_that_frightens(self):
        for key, flag in (("undead_phobia", "afraid_of_undead"), ("beast_phobia", "afraid_of_beasts"),
                          ("arcane_phobia", "afraid_of_constructs")):
            trait = next(f for f in features() if f["key"] == key)
            with self.subTest(trait=key):
                self.assertIn((flag, "grant"), emitted(trait))


if __name__ == "__main__":
    unittest.main()
