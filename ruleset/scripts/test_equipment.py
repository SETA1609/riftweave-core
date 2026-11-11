#!/usr/bin/env python3
"""The equipment content floor (docs/modules/equipment.md).

Run with:
    python -m unittest ruleset/scripts/test_equipment.py
"""

import json
import unittest
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from balance_report import sized_dice  # noqa: E402

DATA = Path(__file__).resolve().parent.parent / "data"
EQUIPMENT = DATA / "equipment"
WEAPON_SKILLS = ("blades", "blunt", "piercing", "unarmed", "bows", "crossbows", "guns", "throwing_weapons")
TIER_BOUNDS_GP = (5, 25)  # tier 1 under 5 gp, tier 2 under 25 gp, tier 3 from 25 gp


def rows(name):
    return json.loads((EQUIPMENT / f"{name}.json").read_text())["equipment"]


def to_copper(cost):
    values = {d["unit"]: d["value"] for d in json.loads((DATA / "economy" / "core.json").read_text())["currency"]["denominations"]}
    return cost["quantity"] * values[cost["unit"]]


def resolved_weapons():
    for row in rows("weapons"):
        base = json.loads((EQUIPMENT / "bases" / "weapons" / f"{row['base'].rsplit('/', 1)[-1]}.json").read_text())
        yield row["key"], {**base, **row["overrides"]}


def tier(cost):
    gold = to_copper(cost) / to_copper({"quantity": 1, "unit": "gp"})
    return 1 + sum(gold >= bound for bound in TIER_BOUNDS_GP)


class TestEquipmentFloor(unittest.TestCase):
    def test_each_weapon_skill_has_three_weapons_across_tiers(self):
        by_skill = {}
        for _, weapon in resolved_weapons():
            skill = weapon["weapon"]["skill"].removeprefix("core:skill/")
            by_skill.setdefault(skill, []).append(tier(weapon["cost"]))
        for skill in WEAPON_SKILLS:
            with self.subTest(skill=skill):
                self.assertGreaterEqual(len(by_skill.get(skill, [])), 3)
                self.assertGreaterEqual(len(set(by_skill[skill])), 2, "all in one tier")

    def test_every_ranged_weapon_names_the_attribute_it_uses(self):
        for key, weapon in resolved_weapons():
            if weapon["weapon"]["range"] == "ranged":
                with self.subTest(weapon=key):
                    self.assertIn(weapon["weapon"].get("ability"), ("str", "per", "agi"))

    def test_slow_weapons_read_per_quick_ones_agi_and_heavy_throws_str(self):
        by_key = {key: weapon["weapon"].get("ability") for key, weapon in resolved_weapons()}
        for key, ability in (("arquebus", "per"), ("crossbow", "per"), ("light_crossbow", "per"), ("longbow", "per"),
                             ("flintlock_pistol", "agi"), ("hand_crossbow", "agi"), ("repeating_crossbow", "agi"),
                             ("shortbow", "agi"), ("composite_bow", "agi"),
                             ("javelin", "str"), ("throwing_axe", "str"), ("throwing_hammer", "str"), ("thrown_rock", "str")):
            with self.subTest(weapon=key):
                self.assertEqual(by_key[key], ability)

    def test_twelve_accessories_and_twenty_consumables(self):
        self.assertGreaterEqual(len(rows("accessories")), 12)
        self.assertGreaterEqual(len(rows("consumables")), 20)

    def test_every_weapon_dice_is_on_the_ladder_and_its_size_step_lands(self):
        ladder = json.loads((DATA / "combat" / "dice_ladder.json").read_text())
        rungs = [dice for rung in ladder["dice_ladder"] for dice in rung]
        for key, weapon in resolved_weapons():
            with self.subTest(weapon=key):
                self.assertIn(weapon["weapon"]["damage"]["dice"], rungs)
                self.assertIn(sized_dice(weapon, ladder), rungs)

    def test_size_moves_one_step(self):
        weapons = dict(resolved_weapons())
        self.assertEqual(sized_dice(weapons["shortbow"]), "1d6")
        self.assertEqual(sized_dice({**weapons["shortbow"], "size": "undersized"}), "1d4")
        self.assertEqual(sized_dice({**weapons["longbow"], "size": "oversized"}), "1d10")
        self.assertEqual(sized_dice({**weapons["greatsword"], "size": "oversized"}), "2d8")

    def test_undersized_is_easier_to_use_and_oversized_harder(self):
        sizes = json.loads((DATA / "combat" / "dice_ladder.json").read_text())["sizes"]
        self.assertGreater(sizes["undersized"]["attack"], 0)
        self.assertEqual(sizes["normal"], {"dice_step": 0, "attack": 0, "value_percent": 100})
        self.assertLess(sizes["oversized"]["attack"], 0)

    def test_undersized_is_cheaper_and_oversized_dearer(self):
        sizes = json.loads((DATA / "combat" / "dice_ladder.json").read_text())["sizes"]
        self.assertLess(sizes["undersized"]["value_percent"], 100)
        self.assertGreater(sizes["oversized"]["value_percent"], 100)

    def test_food_is_always_over_time(self):
        for row in rows("consumables"):
            if row["type"] != "food":
                continue
            for effect in row["consumable"]["effects"]:
                with self.subTest(item=row["key"], effect=effect["effect"]):
                    self.assertGreater(effect.get("duration", 0), 0)

    def test_adventuring_gear_and_tools_exist(self):
        types = [row["type"] for row in rows("gear")]
        self.assertIn("adventuring_gear", types)
        self.assertIn("tool", types)

    def test_a_tool_emits_equipment_modifiers_from_itself(self):
        for row in rows("gear"):
            for modifier in row["gear"].get("modifiers", []):
                with self.subTest(item=row["key"]):
                    self.assertEqual(modifier["source"], f"item:{row['key']}")
                    self.assertEqual(modifier["origin"], "core:origin/equipment")


if __name__ == "__main__":
    unittest.main()
