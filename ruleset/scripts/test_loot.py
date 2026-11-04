#!/usr/bin/env python3
"""Consistency of the loot data (docs/core/loot.md): coverage, budget, and drops."""

import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

from loot_reference import budget_cp, expected_value_cp, load_data  # noqa: E402


class TestLootData(unittest.TestCase):
    def setUp(self):
        self.loot = load_data("loot/core.json")

    def test_every_table_covers_1_to_100_exactly_once(self):
        for table in self.loot["tables"]:
            covered = [n for e in table["entries"] for n in range(e["min"], e["max"] + 1)]
            with self.subTest(table=table["key"]):
                self.assertEqual(sorted(covered), list(range(1, 101)))

    def test_every_table_fits_the_budget_at_its_level_on_average(self):
        for table in self.loot["tables"]:
            with self.subTest(table=table["key"]):
                self.assertLessEqual(expected_value_cp(table), budget_cp(self.loot, table["level"]))

    def test_every_core_monster_drops_exactly_one_table(self):
        monsters = [f"core:monster/{m['key']}" for m in load_data("monsters/core.json")["monsters"]]
        dropped = [d["monster"] for d in self.loot["drops"]]
        self.assertEqual(sorted(dropped), sorted(monsters))

    def test_every_drop_names_a_table_in_the_file(self):
        keys = {t["key"] for t in self.loot["tables"]}
        for drop in self.loot["drops"]:
            with self.subTest(monster=drop["monster"]):
                self.assertIn(drop["table"], keys)


if __name__ == "__main__":
    unittest.main()
