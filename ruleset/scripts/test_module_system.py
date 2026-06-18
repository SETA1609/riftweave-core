#!/usr/bin/env python3
"""Tests for the Riftweave module system: manifest validation, loading, merging, and conflict detection.

Run with:
    python -m unittest ruleset/scripts/test_module_system.py
    # or
    python ruleset/scripts/test_module_system.py
"""

import json
import os
import sys
import tempfile
import unittest
from pathlib import Path

# Ensure the scripts dir is on the path so module_loader can be imported
_SCRIPTS = Path(__file__).resolve().parent
sys.path.insert(0, str(_SCRIPTS))

from module_loader import (
    discover_modules,
    load_module,
    load_all_modules,
    merge_data,
    ModuleInfo,
    ModuleLoadError,
    ConflictError,
    load_manifest_schema,
    validate_manifest,
    MODULES_DIR,
)

# --- Helpers ---


def make_manifest(overrides=None):
    """Return a valid manifest dict, optionally overridden."""
    manifest = {
        "id": "test_module",
        "version": "1.0.0",
        "description": "A test module.",
    }
    if overrides:
        manifest.update(overrides)
    return manifest


def make_temp_module(manifest_dict, data_files=None, schema_files=None):
    """Create a temporary module directory with manifest and optional data/schema files.
    Returns the Path to the module directory.
    """
    tmpdir = Path(tempfile.mkdtemp())
    mod_dir = tmpdir / manifest_dict["id"]
    mod_dir.mkdir(parents=True)

    # Write manifest
    manifest_path = mod_dir / "manifest.json"
    manifest_path.write_text(json.dumps(manifest_dict, indent=2))

    # Write data files
    if data_files:
        data_dir = mod_dir / "data"
        data_dir.mkdir(exist_ok=True)
        for rel_path, content in data_files.items():
            fp = data_dir / rel_path
            fp.parent.mkdir(parents=True, exist_ok=True)
            fp.write_text(json.dumps(content, indent=2))

    # Write schema files
    if schema_files:
        schemas_dir = mod_dir / "schemas"
        schemas_dir.mkdir(exist_ok=True)
        for rel_path, content in schema_files.items():
            fp = schemas_dir / rel_path
            fp.parent.mkdir(parents=True, exist_ok=True)
            fp.write_text(json.dumps(content, indent=2))

    return mod_dir


# --- Tests ---


class TestManifestSchema(unittest.TestCase):
    """Test that manifest validation works correctly."""

    def setUp(self):
        self.schema = load_manifest_schema()

    def test_valid_minimal_manifest(self):
        manifest = make_manifest()
        validate_manifest(manifest, self.schema)  # should not raise

    def test_valid_full_manifest(self):
        manifest = make_manifest(
            {
                "author": "Test Author",
                "requires": ["core"],
                "extends": {
                    "collections": ["features", "equipment"],
                    "types": {"equipment": ["tool"]},
                },
                "data": ["data/test.json"],
                "schemas": ["schemas/test.schema.json"],
            }
        )
        validate_manifest(manifest, self.schema)

    def test_missing_id_raises(self):
        manifest = make_manifest({"id": None})
        # Remove id entirely
        del manifest["id"]
        with self.assertRaises(ModuleLoadError):
            validate_manifest(manifest, self.schema)

    def test_missing_version_raises(self):
        manifest = make_manifest({"version": None})
        del manifest["version"]
        with self.assertRaises(ModuleLoadError):
            validate_manifest(manifest, self.schema)

    def test_invalid_version_format(self):
        manifest = make_manifest({"version": "abc"})
        with self.assertRaises(ModuleLoadError):
            validate_manifest(manifest, self.schema)

    def test_invalid_id_format(self):
        manifest = make_manifest({"id": "My Module"})
        with self.assertRaises(ModuleLoadError):
            validate_manifest(manifest, self.schema)

    def test_empty_description_raises(self):
        manifest = make_manifest({"description": None})
        del manifest["description"]
        with self.assertRaises(ModuleLoadError):
            validate_manifest(manifest, self.schema)


class TestModuleDiscovery(unittest.TestCase):
    """Test that module discovery finds valid modules and skips non-modules."""

    def test_discover_empty_directory(self):
        tmpdir = tempfile.mkdtemp()
        results = discover_modules(tmpdir)
        self.assertEqual(results, [])

    def test_discover_one_module(self):
        mod_dir = make_temp_module(make_manifest())
        results = discover_modules(mod_dir.parent)
        self.assertEqual(len(results), 1)
        found_dir, manifest = results[0]
        self.assertEqual(found_dir.name, "test_module")
        self.assertEqual(manifest["id"], "test_module")

    def test_discover_skips_dirs_without_manifest(self):
        tmpdir = Path(tempfile.mkdtemp())
        (tmpdir / "not_a_module").mkdir()
        (tmpdir / "also_not_a_module").mkdir()
        results = discover_modules(tmpdir)
        self.assertEqual(results, [])

    def test_discover_multiple_modules(self):
        tmpdir = Path(tempfile.mkdtemp())
        for name in ["mod_a", "mod_b", "mod_c"]:
            mod_dir = tmpdir / name
            mod_dir.mkdir()
            manifest = make_manifest({"id": name})
            (mod_dir / "manifest.json").write_text(json.dumps(manifest))
        results = discover_modules(tmpdir)
        self.assertEqual(len(results), 3)
        ids = [m["id"] for _, m in results]
        self.assertCountEqual(ids, ["mod_a", "mod_b", "mod_c"])


class TestModuleLoading(unittest.TestCase):
    """Test loading a module: manifest validation, data loading, schema loading."""

    def setUp(self):
        self.schema = load_manifest_schema()

    def test_load_minimal_module(self):
        mod_dir = make_temp_module(make_manifest())
        mod = load_module(mod_dir, make_manifest(), self.schema)
        self.assertIsInstance(mod, ModuleInfo)
        self.assertEqual(mod.id, "test_module")
        self.assertEqual(mod.version, "1.0.0")
        self.assertEqual(mod.data, {})
        self.assertEqual(mod.schemas, {})

    def test_load_module_with_data(self):
        manifest = make_manifest(
            {
                "data": ["data/test_features.json"],
            }
        )
        data_content = {
            "features": [
                {
                    "id": 100,
                    "key": "test_perk",
                    "label": "Test Perk",
                    "description": "A test.",
                    "type": "perk",
                    "category": "combat",
                }
            ]
        }
        mod_dir = make_temp_module(
            manifest, data_files={"test_features.json": data_content}
        )
        mod = load_module(mod_dir, manifest, self.schema)
        self.assertIn("features", mod.data)
        self.assertEqual(len(mod.data["features"]), 1)
        self.assertEqual(mod.data["features"][0]["id"], 100)

    def test_load_module_with_schema(self):
        manifest = make_manifest(
            {
                "schemas": ["schemas/test.schema.json"],
            }
        )
        schema_content = {
            "$schema": "http://json-schema.org/draft-07/schema#",
            "type": "object",
            "properties": {"foo": {"type": "string"}},
        }
        schema_store = {}
        mod_dir = make_temp_module(
            manifest, schema_files={"test.schema.json": schema_content}
        )
        mod = load_module(mod_dir, manifest, self.schema, schema_store)
        self.assertIn("test.schema.json", str(list(mod.schemas.keys())[0]))
        self.assertIn("test.schema.json", str(list(schema_store.keys())[0]))

    def test_load_module_missing_data_file_raises(self):
        manifest = make_manifest(
            {
                "data": ["data/nonexistent.json"],
            }
        )
        mod_dir = make_temp_module(manifest)
        with self.assertRaises(ModuleLoadError):
            load_module(mod_dir, manifest, self.schema)

    def test_load_module_bad_json_raises(self):
        manifest = make_manifest(
            {
                "data": ["data/bad.json"],
            }
        )
        mod_dir = make_temp_module(manifest)
        (mod_dir / "data").mkdir(exist_ok=True)
        (mod_dir / "data" / "bad.json").write_text("{invalid json}")
        with self.assertRaises(ModuleLoadError):
            load_module(mod_dir, manifest, self.schema)


class TestModuleDataMerging(unittest.TestCase):
    """Test merging module data into core."""

    def setUp(self):
        self.schema = load_manifest_schema()

    def _make_core(self):
        return {
            "features": [
                {
                    "id": 1,
                    "key": "toughness",
                    "label": "Toughness",
                    "type": "perk",
                    "category": "general",
                },
                {
                    "id": 2,
                    "key": "alert",
                    "label": "Alert",
                    "type": "perk",
                    "category": "general",
                },
            ],
            "skills": [
                {
                    "id": 1,
                    "key": "blades",
                    "label": "Blades",
                    "associatedAbility": "str",
                    "category": "combat",
                    "taggable": True,
                },
            ],
        }

    def _make_module_data(self):
        return {
            "features": [
                {
                    "id": 100,
                    "key": "module_perk",
                    "label": "Module Perk",
                    "type": "perk",
                    "category": "combat",
                },
            ]
        }

    def test_merge_no_conflicts(self):
        core = self._make_core()
        manifest = make_manifest({"data": ["data/module_features.json"]})
        mod_dir = make_temp_module(
            manifest, data_files={"module_features.json": self._make_module_data()}
        )
        mod = load_module(mod_dir, manifest, self.schema)

        merged, conflicts = merge_data(core, [mod])
        self.assertEqual(len(conflicts), 0)
        self.assertEqual(len(merged["features"]), 3)
        self.assertEqual(len(merged["skills"]), 1)

    def test_merge_detects_conflict_with_core(self):
        core = self._make_core()
        # Module tries to add an entry with id=1 which already exists in core
        module_data = {
            "features": [
                {
                    "id": 1,
                    "key": "toughness_dupe",
                    "label": "Toughness Dupe",
                    "type": "perk",
                    "category": "general",
                },
            ]
        }
        manifest = make_manifest({"data": ["data/conflict.json"]})
        mod_dir = make_temp_module(manifest, data_files={"conflict.json": module_data})
        mod = load_module(mod_dir, manifest, self.schema)

        merged, conflicts = merge_data(core, [mod])
        self.assertEqual(len(conflicts), 1)
        coll, eid, src_a, src_b = conflicts[0]
        self.assertEqual(coll, "features")
        self.assertEqual(eid, 1)
        self.assertEqual(src_a, "core")

    def test_merge_detects_conflict_between_modules(self):
        core = self._make_core()
        # Two modules both try to add id=100
        mod_a_data = {
            "features": [
                {
                    "id": 100,
                    "key": "mod_a_perk",
                    "label": "Mod A",
                    "type": "perk",
                    "category": "combat",
                }
            ]
        }
        mod_b_data = {
            "features": [
                {
                    "id": 100,
                    "key": "mod_b_perk",
                    "label": "Mod B",
                    "type": "perk",
                    "category": "combat",
                }
            ]
        }

        manifest_a = make_manifest({"id": "mod_a", "data": ["data/a.json"]})
        manifest_b = make_manifest({"id": "mod_b", "data": ["data/b.json"]})

        dir_a = make_temp_module(manifest_a, data_files={"a.json": mod_a_data})
        dir_b = make_temp_module(manifest_b, data_files={"b.json": mod_b_data})

        mod_a = load_module(dir_a, manifest_a, self.schema)
        mod_b = load_module(dir_b, manifest_b, self.schema)

        merged, conflicts = merge_data(core, [mod_a, mod_b])
        self.assertEqual(len(conflicts), 1)
        coll, eid, src_a, src_b = conflicts[0]
        self.assertEqual(coll, "features")
        self.assertEqual(eid, 100)
        self.assertEqual(src_a, "mod_a")
        self.assertEqual(src_b, "mod_b")

    def test_merge_adds_unknown_collection(self):
        core = {"features": []}
        module_data = {"new_collection": [{"id": 1, "data": "test"}]}
        manifest = make_manifest({"id": "mod_c", "data": ["data/new.json"]})
        mod_dir = make_temp_module(manifest, data_files={"new.json": module_data})
        mod = load_module(mod_dir, manifest, self.schema)

        merged, conflicts = merge_data(core, [mod])
        self.assertEqual(len(conflicts), 0)
        self.assertIn("new_collection", merged)
        self.assertEqual(len(merged["new_collection"]), 1)


class TestIntegration(unittest.TestCase):
    """Integration tests using the real example module."""

    def test_example_module_loads_cleanly(self):
        example_dir = MODULES_DIR / "example-module"
        self.assertTrue(example_dir.exists(), "example-module directory should exist")
        manifest_path = example_dir / "manifest.json"
        self.assertTrue(manifest_path.exists(), "example-module manifest should exist")

        manifest = json.loads(manifest_path.read_text())
        schema = load_manifest_schema()
        mod = load_module(example_dir, manifest, schema)

        self.assertEqual(mod.id, "example_module")
        self.assertIn("features", mod.data)
        self.assertEqual(len(mod.data["features"]), 1)
        self.assertEqual(mod.data["features"][0]["id"], 1001)

    def test_example_module_merges_without_conflicts(self):
        core = {"features": []}
        example_dir = MODULES_DIR / "example-module"
        manifest = json.loads((example_dir / "manifest.json").read_text())
        schema = load_manifest_schema()
        mod = load_module(example_dir, manifest, schema)

        merged, conflicts = merge_data(core, [mod])
        self.assertEqual(len(conflicts), 0)
        self.assertEqual(len(merged["features"]), 1)

    def test_full_discovery_and_loading(self):
        modules = load_all_modules()
        ids = [m.id for m in modules]
        self.assertIn("example_module", ids)


class TestEdgeCases(unittest.TestCase):
    """Edge cases and robustness."""

    def test_load_all_with_nonexistent_dir(self):
        modules = load_all_modules(modules_dir="/nonexistent/path")
        self.assertEqual(modules, [])

    def test_discover_hidden_directories_skipped(self):
        tmpdir = Path(tempfile.mkdtemp())
        hidden = tmpdir / ".hidden"
        hidden.mkdir()
        manifest = make_manifest({"id": "hidden_mod"})
        (hidden / "manifest.json").write_text(json.dumps(manifest))
        results = discover_modules(tmpdir)
        for d, _ in results:
            self.assertFalse(d.name.startswith("."))

    def test_manifest_with_extra_fields_passed(self):
        manifest = make_manifest({"extra_field": "should still validate"})
        # The schema sets additionalProperties: false, so extra fields should fail
        schema = load_manifest_schema()
        with self.assertRaises(ModuleLoadError):
            validate_manifest(manifest, schema)


if __name__ == "__main__":
    unittest.main()
