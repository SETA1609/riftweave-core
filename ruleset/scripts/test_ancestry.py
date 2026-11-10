#!/usr/bin/env python3
"""Races, traits, and backgrounds emit contract modifiers (docs/modules/race.md#5a-what-a-race-emits),
and every lineage role keeps its schema rules."""

import json
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

from jsonschema import Draft7Validator, RefResolver  # noqa: E402

from ancestry_reference import ancestry_modifiers, load_data  # noqa: E402
from kernel_reference import apply_modifiers  # noqa: E402
from validate import load_store  # noqa: E402

SCHEMAS = Path(__file__).resolve().parent.parent / "schemas"


def race(key, role, **fields):
    row = {"key": key, "label": key.title(), "description": "d", "lineage": {"role": role}}
    if "parent" in fields:
        row["lineage"]["parentRace"] = fields.pop("parent")
    row.update(fields)
    return row


FULL = {"phase": "earth", "speed": 6, "size": "medium"}


class TestLineageSchema(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        store = load_store()
        path = SCHEMAS / "race.schema.json"
        schema = store[path.as_uri()]
        cls.validator = Draft7Validator(schema, resolver=RefResolver(path.as_uri(), schema, store=store))

    def valid(self, row):
        return not list(self.validator.iter_errors({"races": [row]}))

    def test_standalone_needs_phase_speed_and_size(self):
        self.assertTrue(self.valid(race("human", "standalone", **FULL)))
        self.assertFalse(self.valid(race("human", "standalone", phase="earth", size="medium")))

    def test_parent_needs_phase_speed_and_size(self):
        self.assertTrue(self.valid(race("elf", "parent", **FULL)))
        self.assertFalse(self.valid(race("elf", "parent", phase="wood", speed=30)))

    def test_subrace_needs_a_parent_and_may_inherit_speed_and_size(self):
        self.assertTrue(self.valid(race("green_elf", "subrace", parent="core:race/elf", phase="wood")))
        self.assertFalse(self.valid(race("green_elf", "subrace", phase="wood")))
        self.assertTrue(self.valid(race("green_elf", "subrace", parent="core:race/elf")))

    def test_template_needs_none_of_phase_speed_and_size(self):
        self.assertTrue(self.valid(race("beastman", "template")))

    def test_kin_needs_its_template_and_phase_speed_and_size(self):
        self.assertTrue(self.valid(race("catman", "kin", parent="core:race/beastman", **FULL)))
        self.assertFalse(self.valid(race("catman", "kin", **FULL)))
        self.assertFalse(self.valid(race("catman", "kin", parent="core:race/beastman", phase="wood")))


def sources(modifiers):
    return sorted({m["source"] for m in modifiers})


def find(modifiers, source, target):
    return [m for m in modifiers if m["source"] == source and m["target"] == target]


class TestAncestryModifiers(unittest.TestCase):
    def test_ability_modifiers_are_innate_adds_with_the_race_as_source(self):
        mods = ancestry_modifiers("human")
        self.assertEqual(find(mods, "race:human", "cha"), [{"target": "cha", "op": "add", "value": 1, "source": "race:human", "origin": "core:origin/innate"}])

    def test_a_subrace_stacks_its_parents_modifiers_and_traits_with_its_own(self):
        mods = ancestry_modifiers("green_elf")
        self.assertEqual(find(mods, "race:elf", "agi")[0]["value"], 1)
        self.assertEqual(find(mods, "race:green_elf", "per")[0]["value"], 1)
        self.assertIn("feature:elf_darkvision", sources(mods))
        self.assertIn("feature:green_elf_woodland_stride", sources(mods))

    def test_a_subrace_with_no_fields_of_its_own_is_its_parent(self):
        own = {m["source"] for m in ancestry_modifiers("common_elf")} - {"race:common_elf"}
        self.assertEqual(own, {m["source"] for m in ancestry_modifiers("elf")})

    def test_a_kin_inherits_nothing_from_its_template(self):
        self.assertNotIn("feature:beastfolk_kinship", sources(ancestry_modifiers("catman")))
        self.assertNotIn("race:beastman", sources(ancestry_modifiers("catman")))

    def test_a_template_is_never_played(self):
        with self.assertRaises(ValueError):
            ancestry_modifiers("beastman")

    def test_trait_modifiers_carry_the_feature_source_and_the_innate_origin(self):
        mods = find(ancestry_modifiers("dwarf"), "feature:dwarf_stoneworker", "repair")
        self.assertEqual(mods, [{"target": "repair", "op": "add", "value": 10, "source": "feature:dwarf_stoneworker", "origin": "core:origin/innate"}])

    def test_background_skill_bonuses_are_trained_adds(self):
        mods = find(ancestry_modifiers("human", "scout"), "background:scout", "survival")
        self.assertEqual(mods, [{"target": "survival", "op": "add", "value": 12, "source": "background:scout", "origin": "core:origin/trained"}])

    def test_race_and_background_stack_on_one_skill(self):
        mods = [m for m in ancestry_modifiers("black_elf", "street_urchin") if m["target"] == "stealth"]
        self.assertEqual(apply_modifiers(30, mods), 42)


class TestLiveData(unittest.TestCase):
    def test_every_racial_trait_emits_a_modifier_or_a_natural_weapon(self):
        for row in load_data("features/core.json")["features"]:
            if row["type"] == "racial_trait":
                with self.subTest(trait=row["key"]):
                    self.assertTrue(row.get("modifiers") or row.get("natural_weapon"))

    def test_every_trait_modifier_names_its_own_feature_as_source(self):
        for row in load_data("features/core.json")["features"]:
            for m in row.get("modifiers", []):
                with self.subTest(trait=row["key"]):
                    self.assertEqual(m["source"], f"feature:{row['key']}")

    def test_every_playable_race_and_background_reads(self):
        races = [r["key"] for r in load_data("races/core.json")["races"] if r["lineage"]["role"] != "template"]
        for key in races:
            for background in [b["key"] for b in load_data("backgrounds/core.json")["backgrounds"]]:
                with self.subTest(race=key, background=background):
                    self.assertTrue(ancestry_modifiers(key, background))


if __name__ == "__main__":
    unittest.main()
