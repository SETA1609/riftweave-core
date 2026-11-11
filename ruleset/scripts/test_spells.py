#!/usr/bin/env python3
"""The core spell list (docs/modules/magic.md#the-core-spell-list).

Run with:
    python -m unittest ruleset/scripts/test_spells.py
"""

import json
import sys
import unittest
from collections import Counter
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

from casting_reference import spell_cost  # noqa: E402
import validate  # noqa: E402  (base + override resolution)

DATA = Path(__file__).resolve().parent.parent / "data"
COLORS = ("red", "orange", "yellow", "green", "blue", "indigo", "violet", "white", "black")
# A color with no spell of a kind names the kind here; magic.md gives the reason.
GAPS = {"red": {"utility"}, "green": {"damage"}}
NEW_DAMAGE = {"damage_radiant": "white", "damage_necrotic": "black", "damage_psychic": "yellow",
              "damage_thunder": "indigo", "damage_acid": "blue"}


def spells():
    return json.loads((DATA / "spells" / "core.json").read_text())["spells"]


def effects():
    return {e["key"]: e for e in json.loads((DATA / "effects" / "core.json").read_text())["effects"]}


def effect_of(applied):
    return effects()[applied["effect"].rsplit("/", 1)[-1]]


def kinds(spell):
    """damage, utility, defensive: what a spell's effects do (magic.md#the-core-spell-list)."""
    found = set()
    for applied in spell["effects"]:
        effect = effect_of(applied)
        if effect["category"] in ("damage", "drain") and effect["polarity"] == "harmful":
            found.add("damage")
        elif effect["category"] in ("resist", "fortify", "restore") or effect["key"] == "spell_absorption":
            found.add("defensive")
        elif effect["category"] in ("utility", "control", "conjure", "cure"):
            found.add("utility")
    return found


def pool(level, intelligence=8):
    """The mana pool, INT x 8 + level x 2 (docs/core/casting.md), for an able caster."""
    return intelligence * 8 + level * 2


class TestSpells(unittest.TestCase):
    def test_at_least_five_per_color_and_forty_five_in_all(self):
        rows = spells()
        self.assertGreaterEqual(len(rows), 45)
        count = Counter(row["color"] for row in rows)
        for color in COLORS:
            with self.subTest(color=color):
                self.assertGreaterEqual(count[color], 5)

    def test_levels_from_1_to_30(self):
        levels = [row["level"] for row in spells()]
        self.assertEqual((min(levels), max(levels)), (1, 30))

    def test_keys_are_unique(self):
        rows = spells()
        self.assertEqual(len({row["key"] for row in rows}), len(rows))

    def test_every_effect_is_in_the_pool_and_allows_spells(self):
        pool = effects()
        for row in spells():
            for applied in row["effects"]:
                key = applied["effect"].rsplit("/", 1)[-1]
                with self.subTest(spell=row["key"], effect=key):
                    self.assertIn(key, pool)
                    self.assertIn("spell", pool[key]["channels"])

    def test_an_authored_spell_stays_on_its_color(self):
        for row in spells():
            for applied in row["effects"]:
                color = effect_of(applied).get("color")
                with self.subTest(spell=row["key"], effect=applied["effect"]):
                    self.assertIn(color, (None, row["color"]))

    def test_every_color_has_damage_utility_and_defense_or_a_reason(self):
        by_color = {color: set() for color in COLORS}
        for row in spells():
            by_color[row["color"]] |= kinds(row)
        for color in COLORS:
            with self.subTest(color=color):
                self.assertEqual({"damage", "utility", "defensive"} - by_color[color], GAPS.get(color, set()))

    def test_every_summon_names_a_group_of_existing_monsters(self):
        monsters = {m["key"] for m in json.loads((DATA / "monsters" / "core.json").read_text())["monsters"]}
        groups = {g["key"]: g for g in json.loads((DATA / "summons" / "core.json").read_text())["summons"]}
        summons = [e for s in spells() for e in s["effects"] if e["effect"] == "core:effect/summon"]
        self.assertTrue(summons)
        for effect in summons:
            with self.subTest(parameter=effect["parameter"]):
                self.assertTrue(effect["parameter"].startswith("core:summon/"))
                self.assertIn(effect["parameter"].removeprefix("core:summon/"), groups)
        for key, group in groups.items():
            for variant in group["variants"]:
                with self.subTest(group=key, variant=variant):
                    self.assertIn(variant.removeprefix("core:monster/"), monsters)

    def test_a_summon_appears_at_its_casters_level(self):
        raw = json.loads((DATA / "monsters" / "core.json").read_text())["monsters"]
        groups = json.loads((DATA / "summons" / "core.json").read_text())["summons"]
        variants = sorted({v.removeprefix("core:monster/") for g in groups for v in g["variants"]})
        caster_level = 20
        summoned = [{"key": f"summoned_{key}", "base": f"core:monster/{key}", "overrides": {"level": caster_level}}
                    for key in variants]
        errors = []
        data = {"monsters": json.loads(json.dumps(raw)) + summoned}
        validate._resolve_monster_bases(data, errors)
        self.assertEqual(errors, [])
        rows = {row["key"]: row for row in data["monsters"]}
        for key in variants:
            with self.subTest(variant=key):
                row, creature = rows[key], rows[f"summoned_{key}"]
                self.assertEqual(creature["level"], caster_level)
                self.assertGreater(creature["hitPoints"], row["hitPoints"])
        # The row still sets the stats: at the same level the zombie keeps its lead over the skeleton.
        self.assertEqual(rows["summoned_zombie"]["hitPoints"] - rows["summoned_skeleton"]["hitPoints"],
                         rows["zombie"]["hitPoints"] - rows["skeleton"]["hitPoints"])

    def offensive(self):
        pool = effects()
        return [s for s in spells() if any(
            pool[e["effect"].rsplit("/", 1)[-1]].get("polarity") == "harmful"
            and pool[e["effect"].rsplit("/", 1)[-1]]["category"] != "conjure"
            and e["target"] != "self" for e in s["effects"])]

    def test_every_offensive_spell_has_a_delivery(self):
        offensive = self.offensive()
        self.assertTrue(offensive)
        for spell in offensive:
            with self.subTest(spell=spell["key"]):
                self.assertIn(spell.get("delivery"), ("single", "area", "instant"))

    def test_a_delivery_matches_the_effects_shape(self):
        for spell in spells():
            if spell.get("delivery") in ("single", "area"):
                with self.subTest(spell=spell["key"]):
                    area = any(e["target"] == "area" for e in spell["effects"])
                    self.assertEqual(spell["delivery"], "area" if area else "single")

    def test_every_instant_spell_names_a_save_and_no_other_does(self):
        for spell in spells():
            with self.subTest(spell=spell["key"]):
                self.assertEqual("save" in spell, spell.get("delivery") == "instant")

    def test_the_hazard_and_lock_spells_exist(self):
        rows = {row["key"]: row for row in spells()}
        expected = {"disarm_trap": "disarm_hazard", "greater_disarm": "disarm_hazard", "unlock": "unlock",
                    "dispel_magic": "dispel", "dispel_ward": "dispel"}
        for key, effect in expected.items():
            with self.subTest(spell=key):
                self.assertIn(f"core:effect/{effect}", [a["effect"] for a in rows[key]["effects"]])
        self.assertEqual(rows["dispel_ward"]["effects"][0]["target"], "area")

    def test_each_new_damage_effect_matches_a_damage_type_and_its_color(self):
        pool = effects()
        types = {t["key"] for t in json.loads((DATA / "combat" / "damage_types.json").read_text())["damage_types"]}
        for key, color in NEW_DAMAGE.items():
            with self.subTest(effect=key):
                self.assertEqual(pool[key]["color"], color)
                self.assertIn(key.removeprefix("damage_"), types)

    def test_every_spell_costs_what_its_effects_cost(self):
        for row in spells():
            with self.subTest(spell=row["key"]):
                self.assertEqual(row["cost"]["mana"], spell_cost(row))

    def test_every_effect_a_spell_may_draw_has_a_base_cost(self):
        for key, effect in effects().items():
            if "spell" in effect["channels"]:
                with self.subTest(effect=key):
                    self.assertGreater(effect.get("baseCost", 0), 0)

    def test_a_spell_is_affordable_at_its_level(self):
        for row in spells():
            with self.subTest(spell=row["key"]):
                self.assertLessEqual(row["cost"]["mana"], pool(row["level"]))

    def test_over_time_costs_its_total(self):
        heal = lambda magnitude, duration: {"effects": [{"effect": "core:effect/restore_resource", "parameter": "health",
                                                         "magnitude": magnitude, "duration": duration, "target": "touch"}]}
        self.assertEqual(spell_cost(heal(10, 10)), spell_cost(heal(100, 1)))

    def test_control_is_priced_by_its_duration_not_its_magnitude(self):
        hold = lambda magnitude, duration: {"effects": [{"effect": "core:effect/paralyze", "magnitude": magnitude,
                                                         "duration": duration, "target": "ranged"}]}
        self.assertEqual(spell_cost(hold(1, 6)), spell_cost(hold(9, 6)))
        self.assertEqual(spell_cost(hold(1, 12)), 2 * spell_cost(hold(1, 6)))

    def test_the_shape_multiplies(self):
        bolt = lambda target: {"effects": [{"effect": "core:effect/damage_fire", "magnitude": 20, "target": target}]}
        self.assertEqual((spell_cost(bolt("touch")), spell_cost(bolt("ranged"))), (10, 15))

if __name__ == "__main__":
    unittest.main()
