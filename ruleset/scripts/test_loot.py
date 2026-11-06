#!/usr/bin/env python3
"""Consistency of the loot data (docs/core/loot.md): coverage, budget, and drops."""

import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

from loot_reference import (  # noqa: E402
    budget_cp,
    entry_value_cp,
    expected_value_cp,
    load_data,
    monster_expected_value_cp,
    monster_level,
)


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

    def test_every_table_goes_from_least_to_most_valuable(self):
        for table in self.loot["tables"]:
            entries = sorted(table["entries"], key=lambda e: e["min"])
            values = [entry_value_cp(e) for e in entries]
            with self.subTest(table=table["key"]):
                self.assertEqual(values, sorted(values))

    def test_every_core_monster_has_exactly_one_drop_row(self):
        monsters = [f"core:monster/{m['key']}" for m in load_data("monsters/core.json")["monsters"]]
        dropped = [d["monster"] for d in self.loot["drops"]]
        self.assertEqual(sorted(dropped), sorted(monsters))

    def test_every_loot_row_names_a_table_in_the_file(self):
        keys = {t["key"] for t in self.loot["tables"]}
        for drop in self.loot["drops"]:
            for row in drop["loot"]:
                if "table" in row:
                    with self.subTest(monster=drop["monster"], table=row["table"]):
                        self.assertIn(row["table"], keys)

    def test_every_monster_fits_the_budget_at_its_level_on_average(self):
        for drop in self.loot["drops"]:
            level = monster_level(drop["monster"])
            with self.subTest(monster=drop["monster"]):
                self.assertLessEqual(monster_expected_value_cp(self.loot, drop), budget_cp(self.loot, level))


if __name__ == "__main__":
    unittest.main()
