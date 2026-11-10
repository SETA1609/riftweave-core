#!/usr/bin/env python3
"""The matrix lists every ordered pair of distinct phases in exactly one cycle.

That is why wuxing_reference.get_interaction has no reverse lookup.

Run with:
    python -m unittest ruleset/scripts/test_wuxing.py
"""

import itertools
import json
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

from wuxing_reference import load_wuxing, run_wuxing  # noqa: E402
import validate  # noqa: E402

RULESET = Path(__file__).resolve().parent.parent
MODULE_VECTORS = RULESET / "modules" / "wuxing" / "tests" / "vectors.json"
PHASES = ["wood", "fire", "earth", "metal", "water"]


MODULE_ON = MODULE_VECTORS.exists()


@unittest.skipUnless(MODULE_ON, "the wuxing module is off")
class TestWuxingMatrix(unittest.TestCase):
    def test_every_ordered_pair_is_in_exactly_one_cycle(self):
        cycles = load_wuxing()["cycles"]
        for actor, recipient in itertools.permutations(PHASES, 2):
            with self.subTest(pair=f"{actor}->{recipient}"):
                found = [
                    c["id"]
                    for c in cycles
                    for e in c["edges"]
                    if e["from"] == actor and e["to"] == recipient
                ]
                self.assertEqual(len(found), 1, found)

    def test_a_gate_names_a_cycle_that_exists(self):
        cycles = load_wuxing()["cycles"]
        ids = {c["id"] for c in cycles}
        for cycle in cycles:
            if "gate" in cycle:
                self.assertIn(cycle["gate"]["otherwise_cycle"], ids)


@unittest.skipUnless(MODULE_ON, "the wuxing module is off")
class TestModuleVectors(unittest.TestCase):
    def test_every_vector_matches_the_module_data(self):
        for vector in json.loads(MODULE_VECTORS.read_text()):
            with self.subTest(vector=vector["name"]):
                result = run_wuxing(vector)
                for field, expected in vector["expect"].items():
                    self.assertEqual(result[field], expected, field)


class TestBestiaryPhases(unittest.TestCase):
    def test_every_phase_defends_and_attacks(self):
        data = {"monsters": json.loads((RULESET / "data" / "monsters" / "core.json").read_text())["monsters"]}
        validate._resolve_monster_bases(data, [])
        rows = data["monsters"]
        defending = {row["phase"] for row in rows}
        attacking = {attack["phase"] for row in rows for attack in row.get("attacks", []) if "phase" in attack}
        for phase in PHASES:
            with self.subTest(phase=phase):
                self.assertIn(phase, defending, "no monster has this phase")
                self.assertIn(phase, attacking, "no attack carries this phase")


if __name__ == "__main__":
    unittest.main()
