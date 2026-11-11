#!/usr/bin/env python3
"""Economy data invariants the vectors do not reach (docs/core/economy.md).

Run with:
    python -m unittest ruleset/scripts/test_economy.py
"""

import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

from kernel_reference import DATA_DIR, load_data  # noqa: E402
from validate import equipment_rows  # noqa: E402

ITEM_FILES = {
    "weapon": "equipment/weapons.json",
    "armor": "equipment/armor.json",
    "consumable": "equipment/consumables.json",
    "accessory": "equipment/accessories.json",
}


class TestEconomyData(unittest.TestCase):
    def setUp(self):
        self.economy = load_data("economy/core.json")

    def test_base_unit_is_worth_one_and_keys_are_unique(self):
        values = {d["unit"]: d["value"] for d in self.economy["currency"]["denominations"]}
        self.assertEqual(values[self.economy["currency"]["base"]], 1)
        self.assertEqual(len(values), len(self.economy["currency"]["denominations"]))

    def test_barter_never_beats_the_items_own_price(self):
        max_shift = self.economy["barter"]["max_shift"]
        for merchant in self.economy["merchants"]:
            with self.subTest(merchant=merchant["key"]):
                self.assertGreaterEqual(merchant["buy_percent"] - max_shift, 100)
                self.assertLessEqual(merchant["sell_percent"] + max_shift, 100)

    def test_every_stock_ref_resolves(self):
        for merchant in self.economy["merchants"]:
            for entry in merchant["stock"]:
                with self.subTest(item=entry["item"]):
                    namespace, rest = entry["item"].split(":", 1)
                    category, key = rest.split("/", 1)
                    self.assertEqual(namespace, "core")
                    keys = {e["key"] for e in equipment_rows(DATA_DIR / ITEM_FILES[category])}
                    self.assertIn(key, keys)

    def test_barter_skill_exists(self):
        skills = {s["key"] for s in load_data("skills/core.json")["skills"]}
        self.assertIn(self.economy["barter"]["skill"], skills)


if __name__ == "__main__":
    unittest.main()
