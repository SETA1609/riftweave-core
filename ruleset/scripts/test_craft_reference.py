"""Run ruleset/modules/crafting/tests/vectors.json against the crafting module reference,
and check the module's data against the rules the schema cannot express."""

import json
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import craft_reference  # noqa: E402

VECTORS = Path(__file__).resolve().parent.parent / "modules" / "crafting" / "tests" / "vectors.json"


class CraftVectorTest(unittest.TestCase):
    def test_every_craft_vector_matches_the_reference(self):
        cases = json.loads(VECTORS.read_text())
        self.assertTrue(cases)
        for case in cases:
            with self.subTest(case["name"]):
                result = craft_reference.run_vector(case)
                for field, expected in case["expect"].items():
                    self.assertEqual(result.get(field), expected, field)


class CraftDataTest(unittest.TestCase):
    def test_the_module_data_obeys_its_rules(self):
        self.assertEqual(craft_reference.check_module(), [])

    def test_the_module_ships_at_least_30_recipes(self):
        self.assertGreaterEqual(len(craft_reference.load_module()["crafting_recipes"]), 30)


if __name__ == "__main__":
    unittest.main()
