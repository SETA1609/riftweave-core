#!/usr/bin/env python3
"""The bestiary content floor (docs/modules/monsters.md#the-floor).

Run with:
    python -m unittest ruleset/scripts/test_bestiary.py
"""

import json
import sys
import unittest
from pathlib import Path

_SCRIPTS = Path(__file__).resolve().parent
sys.path.insert(0, str(_SCRIPTS))

import validate  # noqa: E402  (base + override resolution)

DATA = _SCRIPTS.parent / "data"
PHASES = ("wood", "fire", "earth", "metal", "water")


def raw_monsters():
    return json.loads((DATA / "monsters" / "core.json").read_text())["monsters"]


def monsters():
    data = {"monsters": raw_monsters()}
    validate._resolve_monster_bases(data, [])
    return data["monsters"]


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


if __name__ == "__main__":
    unittest.main()
