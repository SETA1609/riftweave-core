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
    build_namespaced_ref_index,
    check_equipment_uniqueness,
    check_references,
    equipment_ref_valid,
    resolve_namespaced_ref,
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
                    {"equipment": [{"id": 22, "key": "shortsword", "type": "weapon"}]}
                )
            )
            armor.write_text(
                json.dumps(
                    {"equipment": [{"id": 22, "key": "padded_tunic", "type": "armor"}]}
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
                    {"equipment": [{"id": 45, "key": "shortsword", "type": "weapon"}]}
                )
            )
            armor.write_text(
                json.dumps(
                    {"equipment": [{"id": 22, "key": "padded_tunic", "type": "armor"}]}
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
                {"equipment": [{"id": 45, "key": "shortsword", "type": "weapon"}]}
            )
        )
        self.id_index = {"equipment": {45}}
        self.equipment_index = build_equipment_index([weapons], rel_base=self.root)

    def test_namespaced_key_reference(self):
        self.assertTrue(
            equipment_ref_valid(
                "core:weapon/shortsword", self.id_index, self.equipment_index
            )
        )

    def test_rejects_numeric_reference(self):
        self.assertFalse(equipment_ref_valid(45, self.id_index, self.equipment_index))

    def test_rejects_bare_key_reference(self):
        self.assertFalse(
            equipment_ref_valid("shortsword", self.id_index, self.equipment_index)
        )

    def test_invalid_reference(self):
        self.assertFalse(
            equipment_ref_valid(
                "core:weapon/missing", self.id_index, self.equipment_index
            )
        )


class TestNamespacedRefIndex(unittest.TestCase):
    def test_build_namespaced_ref_index(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            effects = root / "effects.json"
            effects.write_text(
                json.dumps(
                    {
                        "effects": [
                            {"id": 1, "key": "damage_fire"},
                            {"id": 2, "key": "damage_frost"},
                        ]
                    }
                )
            )
            spells = root / "spells.json"
            spells.write_text(
                json.dumps(
                    {
                        "spells": [
                            {"id": 1, "key": "fire_bolt"},
                        ]
                    }
                )
            )
            index = build_namespaced_ref_index([effects, spells])
            self.assertIn("core:effect/damage_fire", index)
            self.assertIn("core:effect/damage_frost", index)
            self.assertIn("core:spell/fire_bolt", index)
            self.assertNotIn("damage_fire", index)
            self.assertNotIn("damage_frost", index)
            self.assertNotIn("fire_bolt", index)
            self.assertEqual(len(index), 3)

    def test_resolve_namespaced_ref_valid(self):
        index = {"core:effect/damage_fire", "core:spell/fire_bolt"}
        self.assertTrue(resolve_namespaced_ref("core:effect/damage_fire", index))
        self.assertTrue(resolve_namespaced_ref("core:spell/fire_bolt", index))
        self.assertFalse(resolve_namespaced_ref("damage_fire", index))
        self.assertFalse(resolve_namespaced_ref("fire_bolt", index))

    def test_resolve_namespaced_ref_invalid(self):
        index = {"core:effect/damage_fire"}
        self.assertFalse(resolve_namespaced_ref("core:effect/missing", index))
        self.assertFalse(resolve_namespaced_ref("core:weapon/shortsword", index))
        self.assertFalse(resolve_namespaced_ref("damage_fire", index))

    def test_resolve_unknown_module_ref(self):
        index = set()
        self.assertTrue(
            resolve_namespaced_ref("module:my-mod/weapon/frost_sword", index)
        )
        self.assertTrue(resolve_namespaced_ref("module:example/effect/custom", index))

    def test_resolve_non_string(self):
        index = set()
        self.assertFalse(resolve_namespaced_ref(42, index))
        self.assertFalse(resolve_namespaced_ref(None, index))


class TestCraftedItemsRefValidation(unittest.TestCase):
    def test_rejects_numeric_effect_in_enchantment(self):
        result = check_references(
            {
                "crafted_items": [
                    {
                        "id": 1,
                        "key": "test_item",
                        "enchantments": [{"effect": 999, "magnitude": 5}],
                        "coatings": [],
                    }
                ]
            },
            "crafted_items",
        )
        self.assertEqual(len(result), 1)
        self.assertIn("numeric refs not allowed", result[0])

    def test_detects_missing_effect_in_coating(self):
        result = check_references(
            {
                "crafted_items": [
                    {
                        "id": 1,
                        "key": "test_item",
                        "enchantments": [],
                        "coatings": [
                            {
                                "effect": "core:effect/missing",
                                "magnitude": 3,
                                "uses_left": 2,
                            }
                        ],
                    }
                ]
            },
            "crafted_items",
            namespaced_ref_index={"core:effect/damage_fire"},
        )
        self.assertEqual(len(result), 1)
        self.assertIn("does not resolve", result[0])

    def test_accepts_valid_namespaced_effect_ref(self):
        result = check_references(
            {
                "crafted_items": [
                    {
                        "id": 1,
                        "key": "test_item",
                        "enchantments": [
                            {"effect": "core:effect/damage_fire", "magnitude": 5}
                        ],
                        "coatings": [],
                    }
                ]
            },
            "crafted_items",
            namespaced_ref_index={"core:effect/damage_fire"},
        )
        self.assertEqual(result, [])

    def test_rejects_invalid_namespaced_effect_ref(self):
        result = check_references(
            {
                "crafted_items": [
                    {
                        "id": 1,
                        "key": "test_item",
                        "enchantments": [
                            {"effect": "core:effect/missing", "magnitude": 5}
                        ],
                        "coatings": [],
                    }
                ]
            },
            "crafted_items",
            namespaced_ref_index={"core:effect/damage_fire"},
        )
        self.assertEqual(len(result), 1)
        self.assertIn("does not resolve", result[0])


if __name__ == "__main__":
    unittest.main()
