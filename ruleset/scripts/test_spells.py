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

# docs/modules/magic.md#template-spells: one template per target shape.
TEMPLATES = {"bolt": ("fire_bolt", "ranged"), "touch": ("holy_light", "touch"), "burst": ("fire_burst", "area"),
             "ward": ("conjured_barrier", "self"), "mend": ("mend_wounds", "touch")}
# docs/modules/magic.md#uncraftable-effects
UNCRAFTABLE = {"teleport", "mark", "recall", "storage", "bound_weapon", "summon"}
STAY_CRAFTABLE = {"spell_absorption", "life_steal", "mana_leech", "drain_skill", "drain_attribute", "drain_resource",
                  "damage_resource"}


class TestSpellKnowledge(unittest.TestCase):
    def test_the_templates_are_one_per_shape_and_no_other_spell_is_one(self):
        rows = {row["key"]: row for row in spells()}
        self.assertEqual({key for key, row in rows.items() if row.get("template")}, {key for key, _ in TEMPLATES.values()})
        for shape, (key, target) in TEMPLATES.items():
            with self.subTest(shape=shape):
                row = rows[key]
                self.assertEqual([applied["target"] for applied in row["effects"]], [target])
                self.assertNotIn("reagents", row["cost"])
                self.assertFalse(row.get("save") or row.get("concentration"))

    def test_a_template_costs_and_thresholds_what_crafting_it_does(self):
        from casting_reference import run_spellcraft  # noqa: E402
        rows = {row["key"]: row for row in spells()}
        for key, _ in TEMPLATES.values():
            with self.subTest(template=key):
                row = rows[key]
                crafted = run_spellcraft({"operation": "craft", "spell": {"color": row["color"], "effects": row["effects"]},
                                          "caster": {"known": [a["effect"] for a in row["effects"]], "skills": {}}})
                self.assertEqual((crafted["mana"], crafted["cast_threshold"]), (row["cost"]["mana"], row["cost"]["mana"]))

    def test_the_uncraftable_effects_and_the_excluded_mana(self):
        pool = effects()
        self.assertEqual({key for key, effect in pool.items() if effect.get("craftable") is False}, UNCRAFTABLE)
        self.assertEqual(pool["restore_resource"]["uncraftable_parameters"], ["mana"])
        for key in STAY_CRAFTABLE:
            with self.subTest(effect=key):
                self.assertNotEqual(pool[key].get("craftable"), False)
                self.assertNotIn("uncraftable_parameters", pool[key])

    def test_the_authored_only_spells_exist(self):
        rows = {row["key"]: row for row in spells()}
        wanted = {"divine_intervention": ("teleport", "world"), "mark": ("mark", "self"), "recall": ("recall", "world"),
                  "pocket_storage": ("storage", "self"), "bound_blade": ("bound_weapon", "self")}
        for key, (effect, target) in wanted.items():
            with self.subTest(spell=key):
                applied = rows[key]["effects"][0]
                self.assertEqual((applied["effect"], applied["target"]), (f"core:effect/{effect}", target))


class TestCharismaStrengthensSummons(unittest.TestCase):
    """M3-13 (docs/modules/magic.md#summon-groups): a Commanding Summons perk adds the caster's
    CHA - 5, never below 0, to its group of the summoned creature's attributes."""

    REF = "core:monster/fire_elemental"
    MAGIC, PHYSICAL = ("int", "wil"), ("str", "end", "agi")

    def summon(self, cha, *perks):
        import monster_reference

        caster = {"level": 20, "attributes": {"cha": cha}, "perks": [f"core:feature/{p}" for p in perks]}
        return monster_reference.summon(self.REF, caster)

    def plain(self):
        import monster_reference

        return monster_reference.generate(self.REF, level=20)

    def gains(self, summoned):
        before, after = self.plain()["attributes"], summoned["attributes"]
        return {a: after.get(a, 0) - before.get(a, 0) for a in self.MAGIC + self.PHYSICAL}

    def test_the_magic_perk_adds_three_to_int_and_wil_at_cha_8(self):
        gains = self.gains(self.summon(8, "commanding_summons_magic"))
        self.assertEqual(gains, {"int": 3, "wil": 3, "str": 0, "end": 0, "agi": 0})

    def test_the_physical_perk_adds_three_to_str_end_and_agi_at_cha_8(self):
        summoned = self.summon(8, "commanding_summons_physical")
        self.assertEqual(self.gains(summoned), {"int": 0, "wil": 0, "str": 3, "end": 3, "agi": 3})

    def test_an_end_gain_moves_authored_hit_points_by_the_hp_rows_per_point(self):
        plain, summoned = self.plain(), self.summon(8, "commanding_summons_physical")
        if "hitPoints" in plain:
            self.assertEqual(summoned["hitPoints"] - plain["hitPoints"], 3 * validate._hp_per_end())

    def test_both_perks_add_to_both_groups(self):
        gains = self.gains(self.summon(8, "commanding_summons_magic", "commanding_summons_physical"))
        self.assertEqual(set(gains.values()), {3})

    def test_without_the_perk_charisma_changes_nothing(self):
        self.assertEqual(self.summon(10)["attributes"], self.plain()["attributes"])

    def test_charisma_five_or_below_adds_nothing_and_takes_nothing(self):
        for cha in (5, 3, 1):
            with self.subTest(cha=cha):
                summoned = self.summon(cha, "commanding_summons_magic", "commanding_summons_physical")
                self.assertEqual(summoned["attributes"], self.plain()["attributes"])

    def test_each_perk_names_one_bonus_and_the_bonuses_are_the_two_groups(self):
        features = {f["key"]: f for f in json.loads((DATA / "features" / "core.json").read_text())["features"]}
        bonuses = {b["key"]: b for b in json.loads((DATA / "summons" / "bonuses.json").read_text())["summon_bonuses"]}
        self.assertEqual(features["commanding_summons_magic"]["summon_bonus"], "core:summon_bonus/magic")
        self.assertEqual(features["commanding_summons_physical"]["summon_bonus"], "core:summon_bonus/physical")
        self.assertEqual(tuple(bonuses["magic"]["attributes"]), self.MAGIC)
        self.assertEqual(tuple(bonuses["physical"]["attributes"]), self.PHYSICAL)
        self.assertEqual({b["attribute"] for b in bonuses.values()}, {"cha"})


if __name__ == "__main__":
    unittest.main()
