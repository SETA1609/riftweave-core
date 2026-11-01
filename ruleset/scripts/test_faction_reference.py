"""Run ruleset/modules/factions/tests/vectors.json against the factions reference."""

import json
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import faction_reference  # noqa: E402

VECTORS = Path(__file__).resolve().parent.parent / "modules" / "factions" / "tests" / "vectors.json"


class FactionVectorTest(unittest.TestCase):
    def test_every_faction_vector_matches_the_reference(self):
        cases = json.loads(VECTORS.read_text())
        self.assertTrue(cases)
        for case in cases:
            with self.subTest(case["name"]):
                self.assertEqual(faction_reference.run_vector(case), case["expect"])


if __name__ == "__main__":
    unittest.main()
