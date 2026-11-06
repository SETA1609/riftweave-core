#!/usr/bin/env python3
"""The balance report (docs/core/balance.md): each check on small fixtures, and the real data."""

import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

from balance_report import dice_mean, gate_failures, load_all, report  # noqa: E402

BANDS = {
    "tiers": [
        {"key": "early", "levels": [1, 5], "expected_damage": [4, 7], "worn_dr": {"light": [3, 8], "medium": [8, 18], "heavy": [14, 28]}},
        {"key": "mid", "levels": [6, 14], "expected_damage": [7, 12], "worn_dr": {"light": [3, 8], "medium": [8, 18], "heavy": [14, 28]}},
    ],
    "weapon_dice": [3.5, 11],
    "hits_to_kill": {"standard": [4, 6], "boss": [8, 12], "minion": [1, 2]},
    "checks": [
        {"key": "weapon_dice", "mode": "gate", "description": "d"},
        {"key": "armor_piece_dr", "mode": "gate", "description": "d"},
        {"key": "worn_set_dr", "mode": "warn", "description": "d"},
        {"key": "monster_damage", "mode": "warn", "description": "d"},
        {"key": "monster_hits_to_kill", "mode": "warn", "description": "d"},
    ],
    "accepted": [],
}


def weapon(key, dice):
    return {"key": key, "weapon": {"damage": {"dice": dice, "type": "slashing"}}}


def armor(key, category, slot, dr):
    return {"key": key, "armor": {"category": category, "slot": slot, "drBase": dr}}


def monster(key, level, hp, magnitude):
    return {"key": key, "level": level, "hitPoints": hp, "abilities": [{"effect": "core:effect/damage_fire", "magnitude": magnitude}]}


def subjects(findings, check):
    return sorted(f["subject"] for f in findings if f["check"] == check)


class TestDice(unittest.TestCase):
    def test_mean(self):
        self.assertEqual(dice_mean("1d6"), 3.5)
        self.assertEqual(dice_mean("2d10"), 11)
        self.assertEqual(dice_mean("1d8+2"), 6.5)


class TestChecks(unittest.TestCase):
    def test_a_weapon_below_the_dice_band_is_a_finding(self):
        findings = report(BANDS, [weapon("dagger", "1d4"), weapon("sword", "1d8")], [], [])
        self.assertEqual(subjects(findings, "weapon_dice"), ["dagger"])

    def test_an_accepted_finding_is_listed_and_never_fails(self):
        bands = {**BANDS, "accepted": [{"check": "weapon_dice", "subject": "dagger", "reason": "light weapon"}]}
        findings = report(bands, [weapon("dagger", "1d4")], [], [])
        self.assertTrue(findings[0]["accepted"])
        self.assertEqual(gate_failures(findings), [])

    def test_a_gate_finding_fails_and_a_warn_finding_does_not(self):
        findings = report(BANDS, [weapon("dagger", "1d4")], [], [monster("brute", 1, 100, 20)])
        self.assertEqual([f["subject"] for f in gate_failures(findings)], ["dagger"])

    def test_an_armor_piece_may_not_pass_the_highest_set_dr_of_its_category(self):
        findings = report(BANDS, [], [armor("ok", "light", "torso", 8), armor("too_much", "light", "torso", 9)], [])
        self.assertEqual(subjects(findings, "armor_piece_dr"), ["too_much"])

    def test_a_worn_set_takes_the_best_piece_per_slot_and_may_use_lighter_pieces(self):
        pieces = [
            armor("vest", "light", "torso", 2), armor("boots", "light", "feet", 1), armor("cap", "light", "head", 0),
            armor("mail", "medium", "torso", 5), armor("coif", "medium", "head", 1),
            armor("shield", "heavy", "offhand", 9),
        ]
        findings = report(BANDS, [], pieces, [])
        by_subject = {f["subject"]: f for f in findings if f["check"] == "worn_set_dr"}
        self.assertEqual(sorted(by_subject), ["heavy", "medium"])
        self.assertEqual(by_subject["medium"]["value"], 7)  # mail 5 + coif 1 + light boots 1; no shield
        self.assertNotIn("light", by_subject)  # vest 2 + boots 1 + cap 0 = 3

    def test_monster_damage_is_read_against_its_level_tier(self):
        findings = report(BANDS, [], [], [monster("imp", 1, 22, 6), monster("drake", 6, 50, 18)])
        self.assertEqual(subjects(findings, "monster_damage"), ["drake"])

    def test_hits_to_kill_is_hp_over_the_middle_of_the_tier_damage(self):
        findings = report(BANDS, [], [], [monster("imp", 1, 22, 6), monster("rat", 1, 10, 4), monster("golem", 1, 110, 6)])
        self.assertEqual(subjects(findings, "monster_hits_to_kill"), ["golem", "rat"])


class TestRealData(unittest.TestCase):
    def test_no_gate_check_has_an_unaccepted_finding(self):
        bands, weapons, armor_rows, monsters = load_all()
        self.assertEqual(gate_failures(report(bands, weapons, armor_rows, monsters)), [])


if __name__ == "__main__":
    unittest.main()
