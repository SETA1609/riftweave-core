#!/usr/bin/env python3
"""The reference playthrough (docs/core/loop.md#reference-playthrough) is one chain: each step
starts from the state the step before left, and it uses only core rows.

Run with:
    python -m unittest ruleset/scripts/test_playthrough.py
"""

import json
import re
import unittest
from pathlib import Path

CHAIN = Path(__file__).resolve().parent.parent / "tests" / "vectors" / "playthrough.json"


class TestPlaythrough(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.steps = json.loads(CHAIN.read_text())

    def step(self, prefix):
        return next(step for step in self.steps if step["name"].startswith(prefix))

    def test_the_chain_runs_every_stage_in_order(self):
        stages = [step["name"].split(":", 1)[0] for step in self.steps]
        order = ["create", "buy", "fight", "loot", "rest", "level", "cast", "condition", "recover", "perk", "gear"]
        self.assertEqual(list(dict.fromkeys(stages)), order)

    def test_the_second_blow_starts_where_the_first_left_the_zombie(self):
        first = self.step("fight: a hit")["expect"]["hp"]
        self.assertEqual(self.step("fight: a crit")["defender"]["hp"], first)
        self.assertEqual(self.step("fight: a crit")["expect"]["hp"], 0)

    def test_the_character_is_hit_from_its_created_pool_and_rests_from_the_result(self):
        created = self.step("create")["expect"]["hp"]
        slam = self.step("fight: the zombie")
        self.assertEqual(slam["defender"]["hp"], created)
        self.assertEqual(self.step("rest")["current"]["hp"], slam["expect"]["hp"])

    def test_casting_starts_from_the_created_mana_and_recovery_from_what_is_left(self):
        cast = self.step("cast")
        self.assertEqual(cast["caster"]["mana"], self.step("create")["expect"]["mp"])
        self.assertEqual(self.step("recover")["current"]["mp"], cast["expect"]["mana"])

    def test_the_condition_taken_is_the_condition_cleared(self):
        taken = list(self.step("condition")["expect"]["conditions"])
        self.assertEqual([c["key"] for c in self.step("recover")["conditions"]], taken)
        self.assertEqual(self.step("recover")["expect"]["conditions"], [])

    def test_the_chain_names_only_core_rows(self):
        self.assertEqual(re.findall(r"\bmodule:[a-z_/]+", CHAIN.read_text()), [])


if __name__ == "__main__":
    unittest.main()
