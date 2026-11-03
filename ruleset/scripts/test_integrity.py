#!/usr/bin/env python3
"""The three integrity gates of docs/core/integrity.md.

Run with:
    python -m unittest ruleset/scripts/test_integrity.py
"""

import json
import sys
import tempfile
import unittest
from pathlib import Path

_SCRIPTS = Path(__file__).resolve().parent
sys.path.insert(0, str(_SCRIPTS))

import validate  # noqa: E402

SCHEMA_DIR = _SCRIPTS.parent / "schemas"

# Property names that hold a namespaced reference in the data today. A schema that defines one
# of them as a plain string accepts a bare string where a key is required.
REFERENCE_PROPERTIES = {
    "appliedBy", "removedBy", "effect", "effects", "sub_effects", "reagents", "perks",
    "parentRace", "starting_spells", "suggested_tag_skills", "base_key", "material_key",
    "gem_key", "crafted_item_key",
}
# A name that is a reference in the data but free text in this one schema.
REFERENCE_FIELD_EXCEPTIONS = {
    ("module.manifest.schema.json", "effect"),  # extends.types: names of new effect categories
}


def write_json(path, data):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(data))
    return path


def effect(key, **extra):
    return {"id": 1, "key": key, "label": key, "description": "x", "category": "c",
            "polarity": "harm", "channels": ["spell"], **extra}


def is_plain_string(definition):
    """True for a schema node that is a string (or array of strings) with no pattern, enum, or $ref."""
    if not isinstance(definition, dict):
        return False
    if any(k in definition for k in ("pattern", "enum", "const", "$ref", "oneOf", "anyOf")):
        return False
    if definition.get("type") == "string":
        return True
    if definition.get("type") == "array":
        return is_plain_string(definition.get("items"))
    return False


def walk_properties(node, found, schema_name):
    if isinstance(node, dict):
        props = node.get("properties")
        if isinstance(props, dict):
            for name, definition in props.items():
                if name in REFERENCE_PROPERTIES and (schema_name, name) not in REFERENCE_FIELD_EXCEPTIONS:
                    found.append((schema_name, name, definition))
        for value in node.values():
            walk_properties(value, found, schema_name)
    elif isinstance(node, list):
        for value in node:
            walk_properties(value, found, schema_name)


class TestSchemasAcceptNoBareStringAsReference(unittest.TestCase):
    def test_reference_properties_are_patterns_not_plain_strings(self):
        found = []
        for path in sorted(SCHEMA_DIR.glob("*.schema.json")):
            walk_properties(json.loads(path.read_text()), found, path.name)
        self.assertTrue(found, "no reference property found; the property list is stale")
        for schema_name, name, definition in found:
            with self.subTest(schema=schema_name, property=name):
                self.assertFalse(is_plain_string(definition), f"{schema_name}: {name} is a plain string")

    def test_namespaced_ref_rejects_a_bare_string(self):
        schema = json.loads((SCHEMA_DIR / "schema.json").read_text())["definitions"]["namespacedRef"]
        import re
        pattern = re.compile(schema["pattern"])
        for good in ("core:weapon/dagger", "core:effect/damage_fire", "module:factions/standing"):
            self.assertTrue(pattern.match(good), good)
        for bad in ("dagger", "core:dagger", "core:weapon/Dagger", "weapon/dagger", "", "core:weapon/"):
            self.assertFalse(pattern.match(bad), bad)


class TestEveryReferenceResolves(unittest.TestCase):
    def test_the_real_data_has_no_dangling_reference(self):
        data_files = list(validate.collect_data_files())
        modules = list(validate.collect_module_data_files())
        self.assertEqual(validate.check_all_references(data_files, modules), [])

    def test_a_dangling_reference_in_any_field_is_reported(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            effects = write_json(root / "effects" / "core.json", {"effects": [effect("burn")]})
            other = write_json(root / "misc" / "x.json", {"things": [{"note": "core:effect/burn", "oops": "core:effect/bunr"}]})
            errors = validate.check_all_references([effects, other], rel_base=root)
            self.assertEqual(len(errors), 1)
            self.assertIn("core:effect/bunr", errors[0])
            self.assertIn("misc/x.json", errors[0])

    def test_an_unknown_category_is_reported(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            other = write_json(root / "misc" / "x.json", {"a": "core:widget/anything"})
            errors = validate.check_all_references([other], rel_base=root)
            self.assertEqual(len(errors), 1)

    def test_a_module_reference_is_accepted_without_a_lookup(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            other = write_json(root / "misc" / "x.json", {"a": "module:factions/standing"})
            self.assertEqual(validate.check_all_references([other], rel_base=root), [])


class TestEffectsAreUsedOrReserved(unittest.TestCase):
    def test_the_real_registry_passes(self):
        data_files = list(validate.collect_data_files())
        modules = list(validate.collect_module_data_files())
        self.assertEqual(validate.check_effect_usage(data_files, modules), [])

    def test_an_unreferenced_effect_that_is_not_reserved_fails(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            effects = write_json(root / "effects" / "core.json", {"effects": [effect("burn"), effect("idle")]})
            spells = write_json(root / "spells" / "core.json", {"spells": [{"effects": ["core:effect/burn"]}]})
            errors = validate.check_effect_usage([effects, spells])
            self.assertEqual(len(errors), 1)
            self.assertIn("idle", errors[0])

    def test_a_reserved_effect_passes_while_unreferenced(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            effects = write_json(root / "effects" / "core.json", {"effects": [effect("idle", reserved=True)]})
            self.assertEqual(validate.check_effect_usage([effects]), [])

    def test_a_reserved_effect_that_something_references_fails(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            effects = write_json(root / "effects" / "core.json", {"effects": [effect("idle", reserved=True)]})
            spells = write_json(root / "spells" / "core.json", {"spells": [{"effects": ["core:effect/idle"]}]})
            errors = validate.check_effect_usage([effects, spells])
            self.assertEqual(len(errors), 1)
            self.assertIn("reserved", errors[0])

    def test_a_reference_from_another_effect_does_not_count_as_use(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            effects = write_json(
                root / "effects" / "core.json",
                {"effects": [effect("combo", sub_effects=["core:effect/part"]), effect("part")]},
            )
            errors = validate.check_effect_usage([effects])
            self.assertEqual(len(errors), 2, errors)


if __name__ == "__main__":
    unittest.main()
