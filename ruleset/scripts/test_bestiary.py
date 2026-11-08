#!/usr/bin/env python3
"""The bestiary content floor (docs/modules/monsters.md#the-floor).

Run with:
    python -m unittest ruleset/scripts/test_bestiary.py
"""

import json
import re
import sys
import unittest
from pathlib import Path

_SCRIPTS = Path(__file__).resolve().parent
sys.path.insert(0, str(_SCRIPTS))

import validate  # noqa: E402  (base + override resolution)

DATA = _SCRIPTS.parent / "data"
PHASES = ("wood", "fire", "earth", "metal", "water")
# Product identity of other games (docs/modules/monsters.md#names): never a monster name.
PROTECTED_NAMES = (
    "beholder", "mind flayer", "illithid", "displacer", "owlbear", "umber hulk", "carrion crawler",
    "gelatinous cube", "githyanki", "githzerai", "kuo-toa", "slaad", "yuan-ti", "flan", "tonberry",
    "cactuar", "malboro", "moogle", "chocobo",
)


def raw_monsters():
    return json.loads((DATA / "monsters" / "core.json").read_text())["monsters"]


def monsters():
    data = {"monsters": raw_monsters()}
    validate._resolve_monster_bases(data, [])
    return data["monsters"]


def resolve(rows):
    """Resolve a list of rows as validate.py does; fail the test on any error."""
    errors = []
    data = {"monsters": rows}
    validate._resolve_monster_bases(data, errors)
    if errors:
        raise AssertionError(errors)
    return {row["key"]: row for row in data["monsters"]}


def damage_types():
    return [t["key"] for t in json.loads((DATA / "combat" / "damage_types.json").read_text())["damage_types"]]


def tiers():
    return json.loads((DATA / "balance" / "core.json").read_text())["tiers"]


class TestBestiary(unittest.TestCase):
    def test_at_least_forty_monsters_from_level_1_to_30(self):
        rows = monsters()
        self.assertGreaterEqual(len(rows), 40)
        levels = [row["level"] for row in rows]
        self.assertEqual((min(levels), max(levels)), (1, 30))

    def test_every_balance_tier_has_monsters(self):
        levels = [row["level"] for row in monsters()]
        for tier in tiers():
            low, high = tier["levels"]
            with self.subTest(tier=tier["key"]):
                self.assertGreaterEqual(sum(low <= level <= high for level in levels), 5)

    def test_every_phase_defends_and_attacks(self):
        rows = monsters()
        defending = {row["phase"] for row in rows}
        attacking = {attack["phase"] for row in rows for attack in row.get("attacks", []) if "phase" in attack}
        for phase in PHASES:
            with self.subTest(phase=phase):
                self.assertIn(phase, defending, "no monster has this phase")
                self.assertIn(phase, attacking, "no attack carries this phase")

    def test_every_damage_type_is_an_attack_and_a_resistance(self):
        rows = monsters()
        attacks = {attack["damage"]["type"] for row in rows for attack in row.get("attacks", [])}
        resisted = {kind for row in rows for kind in row.get("resist", {})}
        for kind in damage_types():
            with self.subTest(damage_type=kind):
                self.assertIn(kind, attacks, "no monster attacks with it")
                self.assertIn(kind, resisted, "no monster resists it")

    def test_every_monster_is_a_base_plus_overrides(self):
        for row in raw_monsters():
            with self.subTest(monster=row["key"]):
                self.assertIn("base", row)
                self.assertIn("overrides", row)

    def test_every_monster_attacks(self):
        for row in monsters():
            with self.subTest(monster=row["key"]):
                self.assertTrue(row.get("attacks"))

    def test_families_derive_from_a_base_monster(self):
        rows = {row["key"]: row for row in raw_monsters()}
        families = {
            "ooze": ["acid_ooze", "fire_ooze", "frost_ooze"],
            "giant": ["frost_giant", "storm_giant_youth"],
            "fire_elemental": ["magma_titan"],
            "stone_golem": ["bronze_guardian", "iron_colossus"],
            "treant_sapling": ["elder_treant"],
            "bandit": ["bandit_captain"],
            "ember_drake": ["ancient_dragon"],
            "skeleton": ["bone_colossus"],
            "dire_wolf": ["shadow_hound"],
            "storm_hawk": ["thunderbird"],
        }
        for parent, children in families.items():
            for child in children:
                with self.subTest(monster=child):
                    self.assertEqual(rows[child]["base"], f"core:monster/{parent}")

    def test_element_variants_use_a_template(self):
        rows = {row["key"]: row for row in raw_monsters()}
        expected = {"acid_ooze": ["acid"], "fire_ooze": ["fire"], "frost_ooze": ["frost"],
                    "frost_giant": ["frost"], "storm_giant_youth": ["storm"]}
        for key, templates in expected.items():
            with self.subTest(monster=key):
                self.assertEqual(rows[key].get("templates"), templates)

    def test_every_derived_monster_states_its_level(self):
        for row in raw_monsters():
            if row["base"].startswith("core:monster/"):
                with self.subTest(monster=row["key"]):
                    self.assertIn("level", row["overrides"])

    def test_a_template_sets_phase_attacks_resist_tags_and_label(self):
        merged = resolve([
            {"id": 1, "key": "blob", "base": "ooze", "overrides": {
                "label": "Blob", "level": 2, "resist": {"fire": 2, "acid": 5}, "tags": ["ooze"],
                "attacks": [{"key": "slam", "label": "Slam", "damage": {"amount": 5, "type": "bludgeoning"}}]}},
            {"id": 2, "key": "fire_blob", "base": "core:monster/blob", "templates": ["fire"], "overrides": {"level": 2}},
        ])["fire_blob"]
        self.assertEqual(merged["label"], "Fire Blob")
        self.assertEqual(merged["phase"], "fire")
        self.assertEqual(merged["attacks"][0]["damage"]["type"], "fire")
        self.assertEqual(merged["attacks"][0]["phase"], "fire")
        self.assertEqual(merged["resist"]["acid"], 5)
        self.assertGreater(merged["resist"]["fire"], 2)
        self.assertEqual(merged["tags"].count("ooze"), 1)
        self.assertIn("fire", merged["tags"])

    def test_a_derived_monster_moves_by_its_types_growth(self):
        rows = resolve([
            {"id": 1, "key": "blob", "base": "ooze", "overrides": {"level": 2}},
            {"id": 2, "key": "big_blob", "base": "core:monster/blob", "overrides": {"level": 6, "label": "Big Blob"}},
        ])
        base = json.loads((DATA / "monsters" / "bases" / "ooze.json").read_text())
        growth = base["growth"]
        self.assertEqual(rows["big_blob"]["hitPoints"], rows["blob"]["hitPoints"] + 4 * growth["hitPointsPerLevel"])
        for ability, rule in growth["attributes"].items():
            with self.subTest(ability=ability):
                expected = min(10, max(1, round(rows["blob"]["attributes"][ability] + 4 * rule["perLevel"])))
                self.assertEqual(rows["big_blob"]["attributes"][ability], expected)

    def test_a_missing_or_cyclic_base_monster_or_template_fails(self):
        cases = {
            "missing": [{"id": 1, "key": "a", "base": "core:monster/nobody", "overrides": {"level": 1}}],
            "cycle": [{"id": 1, "key": "a", "base": "core:monster/b", "overrides": {"level": 1}},
                      {"id": 2, "key": "b", "base": "core:monster/a", "overrides": {"level": 1}}],
            "template": [{"id": 1, "key": "a", "base": "ooze", "templates": ["glitter"], "overrides": {}}],
            "no level": [{"id": 1, "key": "a", "base": "ooze", "overrides": {}},
                         {"id": 2, "key": "b", "base": "core:monster/a", "overrides": {}}],
        }
        for name, rows in cases.items():
            with self.subTest(case=name):
                errors = []
                validate._resolve_monster_bases({"monsters": rows}, errors)
                self.assertTrue(errors)

    def test_bases_then_variants_then_one_off_monsters(self):
        rows = raw_monsters()
        parent = {row["key"]: row["base"].removeprefix("core:monster/") for row in rows if row["base"].startswith("core:monster/")}
        group = [0 if row["key"] in parent.values() else 1 if row["key"] in parent else 2 for row in rows]
        self.assertEqual(group, sorted(group))

    def test_no_name_from_another_franchise(self):
        for row in monsters():
            text = " ".join([row["key"].replace("_", " "), row["label"], row["description"]]).lower()
            for name in PROTECTED_NAMES:
                with self.subTest(monster=row["key"], name=name):
                    self.assertIsNone(re.search(rf"\b{re.escape(name)}\b", text))


if __name__ == "__main__":
    unittest.main()
