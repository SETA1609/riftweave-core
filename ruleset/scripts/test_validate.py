#!/usr/bin/env python3
"""Tests for validate.py integrity checks.

Run with:
    python -m unittest ruleset/scripts/test_validate.py
    # or
    python ruleset/scripts/test_validate.py
"""

import json
import sys
import tempfile
import unittest
from pathlib import Path

_SCRIPTS = Path(__file__).resolve().parent
sys.path.insert(0, str(_SCRIPTS))

from validate import (
    build_equipment_index,
    check_equipment_uniqueness,
    equipment_ref_valid,
)


class TestEquipmentIdUniqueness(unittest.TestCase):
    def test_detects_duplicate_ids_across_equipment_files(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            weapons = root / "data" / "equipment" / "weapons.json"
            armor = root / "data" / "equipment" / "armor.json"
            weapons.parent.mkdir(parents=True, exist_ok=True)

            weapons.write_text(
                json.dumps(
                    {
                        "equipment": [
                            {"id": 22, "key": "shortsword", "type": "weapon"}
                        ]
                    }
                )
            )
            armor.write_text(
                json.dumps(
                    {
                        "equipment": [
                            {"id": 22, "key": "padded_tunic", "type": "armor"}
                        ]
                    }
                )
            )

            errors = check_equipment_uniqueness([weapons, armor], rel_base=root)

            self.assertEqual(len(errors), 1)
            self.assertIn("DUPLICATE equipment id 22", errors[0])
            self.assertIn("shortsword", errors[0])
            self.assertIn("padded_tunic", errors[0])

    def test_passes_when_ids_are_unique(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            weapons = root / "data" / "equipment" / "weapons.json"
            armor = root / "data" / "equipment" / "armor.json"
            weapons.parent.mkdir(parents=True, exist_ok=True)

            weapons.write_text(
                json.dumps(
                    {
                        "equipment": [
                            {"id": 45, "key": "shortsword", "type": "weapon"}
                        ]
                    }
                )
            )
            armor.write_text(
                json.dumps(
                    {
                        "equipment": [
                            {"id": 22, "key": "padded_tunic", "type": "armor"}
                        ]
                    }
                )
            )

            errors = check_equipment_uniqueness([weapons, armor], rel_base=root)
            self.assertEqual(errors, [])


class TestEquipmentKeyUniqueness(unittest.TestCase):
    def test_detects_duplicate_keys_across_equipment_files(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            weapons = root / "data" / "equipment" / "weapons.json"
            armor = root / "data" / "equipment" / "armor.json"
            weapons.parent.mkdir(parents=True, exist_ok=True)

            weapons.write_text(
                json.dumps(
                    {"equipment": [{"id": 1, "key": "shared_key", "type": "weapon"}]}
                )
            )
            armor.write_text(
                json.dumps(
                    {"equipment": [{"id": 2, "key": "shared_key", "type": "armor"}]}
                )
            )

            errors = check_equipment_uniqueness([weapons, armor], rel_base=root)

            self.assertEqual(len(errors), 1)
            self.assertIn("DUPLICATE equipment key 'shared_key'", errors[0])


class TestEquipmentRefResolution(unittest.TestCase):
    def setUp(self):
        self._tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self._tmp.cleanup)
        self.root = Path(self._tmp.name)
        weapons = self.root / "data" / "equipment" / "weapons.json"
        weapons.parent.mkdir(parents=True, exist_ok=True)
        weapons.write_text(
            json.dumps(
                {
                    "equipment": [
                        {"id": 45, "key": "shortsword", "type": "weapon"}
                    ]
                }
            )
        )
        self.id_index = {"equipment": {45}}
        self.equipment_index = build_equipment_index([weapons], rel_base=self.root)

    def test_namespaced_key_reference(self):
        self.assertTrue(
            equipment_ref_valid(
                "core:weapons/shortsword", self.id_index, self.equipment_index
            )
        )

    def test_legacy_numeric_reference(self):
        self.assertTrue(equipment_ref_valid(45, self.id_index, self.equipment_index))

    def test_legacy_bare_key_reference(self):
        self.assertTrue(
            equipment_ref_valid("shortsword", self.id_index, self.equipment_index)
        )

    def test_invalid_reference(self):
        self.assertFalse(
            equipment_ref_valid(
                "core:weapons/missing", self.id_index, self.equipment_index
            )
        )


if __name__ == "__main__":
    unittest.main()