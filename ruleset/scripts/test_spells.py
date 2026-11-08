#!/usr/bin/env python3
"""The core spell list (docs/modules/magic.md#the-core-spell-list).

Run with:
    python -m unittest ruleset/scripts/test_spells.py
"""

import json
import math
import unittest
from collections import Counter
from pathlib import Path

DATA = Path(__file__).resolve().parent.parent / "data"
COLORS = ("red", "orange", "yellow", "green", "blue", "indigo", "violet", "white", "black")
# A color with no spell of a kind names the kind here; magic.md gives the reason.
GAPS = {"red": {"utility"}, "green": {"damage"}}
# The spells from before M3-02 keep their costs: the casting vectors read them.
KEPT_COSTS = {"fire_bolt": 5, "force_dart": 8, "mend_wounds": 10, "blessing": 12, "ward": 8,
              "fire_burst": 25, "raise_thrall": 30, "dispel_magic": 12}
UNIT_WEIGHT = {"points": 1.0, "percent": 0.5, "dice": 1.0, "count": 8.0, "seconds": 0.25, "level": 6.0}
TARGET_MULT = {"self": 1.0, "touch": 1.0, "ranged": 1.25, "world": 3.0}
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


def mana(spell):
    """The authored cost: §5's complexity at a skill of 50 (discount 0.25), rounded up."""
    total = 0.0
    for applied in spell["effects"]:
        effect = effect_of(applied)
        cost = applied.get("magnitude", 0) * UNIT_WEIGHT.get(effect.get("magnitudeUnit"), 1.0)
        cost += applied.get("duration", 1) * 0.1 if applied.get("duration", 1) > 1 else 0
        target = applied.get("target", "self")
        cost *= 1.5 + 0.15 * applied.get("radius", 0) if target == "area" else TARGET_MULT[target]
        total += cost
    return math.ceil(round(total * 0.75, 6))


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

    def test_keys_and_ids_are_unique(self):
        rows = spells()
        self.assertEqual(len({row["key"] for row in rows}), len(rows))
        self.assertEqual(len({row["id"] for row in rows}), len(rows))

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

    def test_older_spells_keep_their_costs_and_new_ones_follow_the_formula(self):
        for row in spells():
            with self.subTest(spell=row["key"]):
                expected = KEPT_COSTS.get(row["key"], mana(row))
                self.assertEqual(row["cost"]["mana"], expected)


if __name__ == "__main__":
    unittest.main()
