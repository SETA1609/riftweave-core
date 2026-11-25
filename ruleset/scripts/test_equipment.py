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
from validate import compose_row, equipment_rows  # noqa: E402
import validate  # noqa: E402  (the ammunition gates)

DATA = Path(__file__).resolve().parent.parent / "data"
EQUIPMENT = DATA / "equipment"
WEAPON_SKILLS = ("blades", "blunt", "piercing", "unarmed", "bows", "crossbows", "guns", "throwing_weapons")
TIER_BOUNDS_GP = (5, 25)  # tier 1 under 5 gp, tier 2 under 25 gp, tier 3 from 25 gp


def rows(name):
    return equipment_rows(EQUIPMENT / f"{name}.json")


def to_copper(cost):
    values = {d["unit"]: d["value"] for d in json.loads((DATA / "economy" / "core.json").read_text())["currency"]["denominations"]}
    return cost["quantity"] * values[cost["unit"]]


def resolved_weapons():
    for row in rows("weapons"):
        base = json.loads((EQUIPMENT / "bases" / "weapons" / f"{row['base'].rsplit('/', 1)[-1]}.json").read_text())
        yield row["key"], compose_row(base, row["overrides"])


def tier(cost):
    gold = to_copper(cost) / to_copper({"quantity": 1, "unit": "gold"})
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


# docs/modules/equipment.md#toxicity
TOXICITY = {"antidote": 0, "potion_cure_disease": 0, "potion_minor_healing": 10, "potion_featherfall": 10,
            "potion_healing": 20, "potion_mana": 8, "potion_stamina": 20, "potion_invisibility": 20,
            "potion_fire_resistance": 20, "potion_frost_resistance": 20, "potion_strength": 20,
            "elixir_of_wits": 20, "potion_greater_healing": 35, "potion_haste": 35}


class TestToxicity(unittest.TestCase):
    def test_every_drink_states_its_toxicity_and_nothing_else_does(self):
        for row in rows("consumables"):
            with self.subTest(item=row["key"]):
                consumable = row["consumable"]
                if consumable["delivery"] == "drink":
                    self.assertEqual(consumable.get("toxicity"), TOXICITY[row["key"]])
                else:
                    self.assertNotIn("toxicity", consumable)

    def test_overdosed_is_a_condition(self):
        conditions = {c["key"]: c for c in json.loads((DATA / "conditions" / "core.json").read_text())["conditions"]}
        self.assertIn("overdosed", conditions)


if __name__ == "__main__":
    unittest.main()


class TestAmmunition(unittest.TestCase):
    """GD_S17_02 (docs/core/combat.md#ammunition): every weapon that fires names its kind, and every
    ammunition row fires from some weapon."""

    def test_the_data_has_no_ammunition_problem(self):
        self.assertEqual(validate.check_ammunition(), [])

    def test_a_base_with_the_ammunition_property_and_no_kind_fails(self):
        bases = [{"key": "bow", "weapon": {"properties": ["ammunition", "two_handed"]}}]
        self.assertEqual(validate.ammunition_problems(bases, []), ["weapon base bow has the ammunition property and no weapon.ammo"])

    def test_a_row_whose_kind_no_weapon_loads_fails(self):
        bases = [{"key": "bow", "weapon": {"properties": ["ammunition"], "ammo": "arrow"}}]
        rows = [{"key": "sling_stones", "ammunition": {"kind": "stone", "count": 20}}]
        self.assertEqual(validate.ammunition_problems(bases, rows), ["ammunition sling_stones is stone, which no weapon loads"])

    def test_each_kind_has_a_bundle_of_the_size_the_spec_names(self):
        rows = json.loads((DATA / "equipment" / "ammunition.json").read_text())["equipment"]
        counts = {row["ammunition"]["kind"]: row["ammunition"]["count"] for row in rows if "damage_type" not in row["ammunition"] and "effects" not in row["ammunition"]}
        self.assertEqual(counts, {"arrow": 20, "bolt": 20, "bullet": 10, "scatter": 10, "stone": 20})


class TestGrenades(unittest.TestCase):
    """GD_S17_03 (docs/core/combat.md#grenades): every grenade's riders carry the grenade channel,
    and the five grenades of the spec are in the data."""

    def test_the_data_has_no_grenade_problem(self):
        self.assertEqual(validate.check_grenades(), [])

    def test_a_grenade_whose_rider_lacks_the_grenade_channel_fails(self):
        rows = [{"key": "fire_flask", "type": "grenade", "consumable": {"delivery": "throw", "effects": [{"effect": "core:effect/burn"}]}}]
        channels = {"core:effect/burn": {"spell", "coating"}}
        self.assertEqual(
            validate.grenade_problems(rows, channels),
            ["grenade fire_flask: effect core:effect/burn lacks the grenade channel"],
        )

    def test_a_grenade_must_be_thrown(self):
        rows = [{"key": "fire_flask", "type": "grenade", "consumable": {"delivery": "drink", "effects": []}}]
        self.assertEqual(validate.grenade_problems(rows, {}), ["grenade fire_flask is not thrown (delivery throw)"])

    def test_the_five_grenades_of_the_spec(self):
        rows = json.loads((DATA / "equipment" / "consumables.json").read_text())["equipment"]
        grenades = {
            row["key"]: (row["consumable"].get("damage"), row["consumable"]["radius"])
            for row in rows
            if row["type"] == "grenade"
        }
        self.assertEqual(
            grenades,
            {
                "fire_flask": ({"dice": "1d6", "type": "fire"}, 1),
                "frost_flask": ({"dice": "1d6", "type": "frost"}, 1),
                "thunder_flask": ({"dice": "1d6", "type": "thunder"}, 2),
                "acid_flask": ({"dice": "1d6", "type": "acid"}, 1),
                "smoke_pot": (None, 2),
            },
        )

    def test_a_bursting_round_needs_grenade_riders_and_a_bolt_or_a_bullet(self):
        rows = [
            {"key": "fire_arrows", "type": "ammunition", "ammunition": {"kind": "arrow", "count": 5, "damage_type": "fire", "burst_radius": 1, "effects": []}},
            {"key": "fire_bolts", "type": "ammunition", "ammunition": {"kind": "bolt", "count": 5, "damage_type": "fire", "burst_radius": 1, "effects": [{"effect": "core:effect/burn"}]}},
        ]
        self.assertEqual(
            validate.burst_problems(rows, {"core:effect/burn": {"spell"}}),
            [
                "ammunition fire_arrows bursts, but only bolts and bullets do (docs/core/combat.md#elemental-rounds)",
                "ammunition fire_bolts: effect core:effect/burn lacks the grenade channel",
            ],
        )

    def test_the_eight_elemental_rounds_of_the_spec(self):
        rows = json.loads((DATA / "equipment" / "ammunition.json").read_text())["equipment"]
        rounds = {
            row["key"]: (row["ammunition"]["kind"], row["ammunition"]["damage_type"], row["ammunition"]["burst_radius"], row["ammunition"]["count"])
            for row in rows
            if "burst_radius" in row["ammunition"]
        }
        expected = {}
        for kind in ("bolt", "bullet"):
            for element, radius in (("fire", 1), ("frost", 1), ("thunder", 2), ("acid", 1)):
                expected[f"{element}_{kind}s"] = (kind, element, radius, 5)
        self.assertEqual(rounds, expected)
        self.assertEqual(validate.check_grenades(), [])
