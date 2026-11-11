#!/usr/bin/env python3
"""Composition bases and default materials are typed references the gate resolves (docs/core/keys.md).

Run with:
    python -m unittest ruleset/scripts/test_typed_bases.py
"""

import json
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import validate  # noqa: E402

DATA = Path(__file__).resolve().parent.parent / "data"
EQUIPMENT = {"weapons": "weapon_base", "armor": "armor_base", "accessories": "accessory_base"}


def load(path):
    return json.loads(path.read_text())


class TestTypedBases(unittest.TestCase):
    def test_every_monster_base_is_a_typed_reference(self):
        for row in load(DATA / "monsters" / "core.json")["monsters"]:
            with self.subTest(monster=row["key"]):
                self.assertRegex(row["base"], r"^core:(creature_type|monster)/[a-z0-9_]+$")

    def test_every_equipment_base_is_a_typed_reference(self):
        for name, category in EQUIPMENT.items():
            for row in load(DATA / "equipment" / f"{name}.json")["equipment"]:
                with self.subTest(item=row["key"]):
                    self.assertRegex(row["base"], rf"^core:{category}/[a-z0-9_]+$")

    def test_every_default_material_is_a_typed_reference(self):
        for path in (DATA / "equipment" / "bases").rglob("*.json"):
            for entry in load(path).get("default_materials", []):
                with self.subTest(base=path.stem, entry=entry):
                    self.assertRegex(entry, r"^core:material/[a-z0-9_]+$")

    def test_a_row_that_lists_its_own_default_materials_uses_references(self):
        for name in EQUIPMENT:
            for row in load(DATA / "equipment" / f"{name}.json")["equipment"]:
                for entry in row.get("overrides", {}).get("default_materials", []):
                    with self.subTest(item=row["key"], entry=entry):
                        self.assertRegex(entry, r"^core:material/[a-z0-9_]+$")

    def test_the_gate_knows_every_base_and_refuses_a_missing_one(self):
        known = validate.build_base_refs()
        self.assertIn("core:creature_type/undead", known)
        self.assertIn("core:weapon_base/dagger", known)
        self.assertIn("core:accessory_base/ring_base", known)
        self.assertNotIn("core:creature_type/nobody", known)


if __name__ == "__main__":
    unittest.main()
