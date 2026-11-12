#!/usr/bin/env python3
"""A percentage is a `mul`, a distance is in units, and a weight is in weight units
(docs/core/modifier.md, docs/core/contract.md#distance, docs/core/contract.md#weight).

Run with:
    python -m unittest ruleset/scripts/test_modifier_units.py
"""

import json
import re
import unittest
from pathlib import Path

DATA = Path(__file__).resolve().parent.parent / "data"
SCHEMAS = DATA.parent / "schemas"
POUNDS = r"(?i)\b(pounds?|lbs?)\b"

# Perk: (new target, the percent per rank it gave as an `add`, its listed ranks).
SCALED = {
    "power_attack": ("melee_damage", 25, 1), "shield_wall": ("block_reduction", 15, 1),
    "pyromancer": ("red_magic_magnitude", 15, 2), "ghost": ("detection_chance", -25, 1),
    "field_medic": ("medicine_healing", 50, 1), "heavy_armor_expert": ("heavy_armor_surcharge", -50, 1),
    "acrobat": ("dodge_stamina", -25, 1), "iron_fist": ("unarmed_strike_damage", 50, 1),
    "sniper": ("ranged_range_penalty", -50, 1), "berserker": ("damage_below_half_hp", 20, 1),
    "executioner": ("crit_damage", 50, 1), "mana_flow": ("mp_regen", 25, 1),
    "battle_mage": ("armor_cast_penalty", -50, 1), "healer": ("green_magic_magnitude", 15, 2),
    "illusionist": ("yellow_magic_duration", 25, 1), "conjurer": ("summon_duration", 50, 1),
    "transmuter": ("blue_magic_duration", 25, 1), "telekinetic": ("indigo_magic_range", 50, 1),
    "zealot": ("white_magic_magnitude", 15, 2), "archmage": ("spell_cost", -20, 1),
    "backstab": ("sneak_attack_damage", 50, 1), "master_alchemist": ("potion_magnitude", 25, 1),
    "meditative": ("rest_mp", 50, 1),
}
# Perk: (target, value). These have nothing to scale, so the value stays in the unit their row states.
POINTS = {
    "silver_tongue": ("barter_shift", 10), "haggler": ("barter_shift", 5), "streetwise": ("fence_shift", 10),
    "iron_stomach": ("poison_resistance", 25), "dwarf_resilience": ("poison_resistance", 25),
}
UNITS = {
    "night_eyes": ("darkvision_range", 6), "elf_darkvision": ("darkvision_range", 12),
    "black_elf_sunless_sight": ("darkvision_range", 12), "dwarf_darkvision": ("darkvision_range", 12),
    "catman_landing": ("safe_fall", 4),
}


def features():
    return {row["key"]: row for row in json.loads((DATA / "features" / "core.json").read_text())["features"]}


def modifier(feature, target):
    return next(m for m in feature["modifiers"] if m["target"] == target)


class TestModifierUnits(unittest.TestCase):
    def test_no_target_names_a_percent_or_a_unit_other_than_units(self):
        for row in json.loads((DATA / "modifier_targets" / "core.json").read_text())["modifier_targets"]:
            with self.subTest(target=row["key"]):
                self.assertNotIn("percent", row["key"])
                if row["key"] != "lands_on_feet":  # a flag about landing, not a distance in feet
                    self.assertNotRegex(row["key"], r"(_meters|_metres|_feet|_ft)$")

    def test_each_rewritten_perk_gives_the_same_number_at_its_listed_rank(self):
        rows = features()
        for key, (target, percent, ranks) in SCALED.items():
            for rank in range(1, ranks + 1):
                with self.subTest(perk=key, rank=rank):
                    row = modifier(rows[key], target)
                    self.assertEqual(row["op"], "mul")
                    self.assertAlmostEqual(1 + (row["value"] - 1) * rank, 1 + percent * rank / 100)

    def test_a_target_with_nothing_to_scale_keeps_its_points(self):
        rows = features()
        for key, (target, value) in POINTS.items():
            with self.subTest(perk=key):
                self.assertEqual((modifier(rows[key], target)["op"], modifier(rows[key], target)["value"]), ("add", value))

    def test_distances_are_in_units(self):
        rows = features()
        for key, (target, value) in UNITS.items():
            with self.subTest(perk=key):
                self.assertEqual(modifier(rows[key], target)["value"], value)

    def test_carry_is_in_weight_units(self):
        rows = {row["key"]: row for row in json.loads((DATA / "derived" / "core.json").read_text())["derived"]}
        self.assertEqual(rows["carry"]["unit"], "weight_units")

    def test_no_schema_names_pounds(self):
        for path in sorted(SCHEMAS.glob("*.json")):
            with self.subTest(schema=path.name):
                self.assertIsNone(re.search(POUNDS, path.read_text()))

    def test_no_data_row_names_pounds(self):
        for path in sorted(DATA.rglob("*.json")):
            with self.subTest(file=str(path.relative_to(DATA))):
                self.assertIsNone(re.search(POUNDS, path.read_text()))


if __name__ == "__main__":
    unittest.main()
