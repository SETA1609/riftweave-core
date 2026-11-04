#!/usr/bin/env python3
"""The matrix lists every ordered pair of distinct phases in exactly one cycle.

That is why wuxing_reference.get_interaction has no reverse lookup.

Run with:
    python -m unittest ruleset/scripts/test_wuxing.py
"""

import itertools
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

from kernel_reference import load_data  # noqa: E402

PHASES = ["wood", "fire", "earth", "metal", "water"]


class TestWuxingMatrix(unittest.TestCase):
    def test_every_ordered_pair_is_in_exactly_one_cycle(self):
        cycles = load_data("wuxing/core.json")["cycles"]
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
        cycles = load_data("wuxing/core.json")["cycles"]
        ids = {c["id"] for c in cycles}
        for cycle in cycles:
            if "gate" in cycle:
                self.assertIn(cycle["gate"]["otherwise_cycle"], ids)


if __name__ == "__main__":
    unittest.main()
