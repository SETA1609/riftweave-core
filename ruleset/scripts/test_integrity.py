#!/usr/bin/env python3
"""The four integrity gates of docs/core/integrity.md.

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


class TestReferencesAreTyped(unittest.TestCase):
    def definitions(self):
        return json.loads((SCHEMA_DIR / "schema.json").read_text())["definitions"]

    def test_a_typed_reference_rejects_another_category(self):
        import re
        pattern = re.compile(self.definitions()["monsterRef"]["pattern"])
        for good in ("core:monster/skeleton", "module:m/monster/y"):
            self.assertTrue(pattern.match(good), good)
        for bad in ("core:weapon/dagger", "skeleton", "module:m/weapon/y"):
            self.assertFalse(pattern.match(bad), bad)

    def test_only_a_recipe_input_keeps_the_general_reference(self):
        general = []
        for path in sorted(SCHEMA_DIR.glob("*.schema.json")):
            lines = path.read_text().split("\n")
            general += [path.name for line in lines if "definitions/namespacedRef" in line]
        self.assertEqual(general, ["recipe.schema.json"])

    def test_the_real_parameters_match_their_category(self):
        data_files = list(validate.collect_data_files())
        modules = list(validate.collect_module_data_files())
        self.assertEqual(validate.check_effect_parameters(data_files, modules), [])

    def test_a_parameter_of_the_wrong_kind_is_reported(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            effects = write_json(root / "effects" / "core.json", {"effects": [effect("fortify_skill", parameter_category="skill")]})
            good = {"effect": "core:effect/fortify_skill", "parameter": "core:skill/stealth"}
            bare = {"effect": "core:effect/fortify_skill", "parameter": "stealth"}
            other = write_json(root / "misc" / "x.json", {"things": [good, bare]})
            errors = validate.check_effect_parameters([effects, other])
            self.assertEqual(len(errors), 1)
            self.assertIn("stealth", errors[0])


class TestEveryReferenceResolves(unittest.TestCase):
    def test_no_schema_or_core_data_file_has_a_per_mode_field(self):
        """One rule set (docs/core/actions.md#one-rule-set-two-clocks): no ttrpg or video_game field."""
        def keys(node):
            if isinstance(node, dict):
                for key, value in node.items():
                    yield key
                    yield from keys(value)
            elif isinstance(node, list):
                for value in node:
                    yield from keys(value)
        paths = sorted(validate.SCHEMA_DIR.glob("*.json")) + sorted(validate.DATA_DIR.rglob("*.json"))
        for path in paths:
            with self.subTest(file=path.name):
                found = {key for key in keys(json.loads(path.read_text())) if key in ("ttrpg", "video_game")}
                self.assertFalse(found)

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


def modifier(target, op="add", value=1):
    return {"target": target, "op": op, "value": value, "source": "feature:x"}


def target_row(key, kind="number", **extra):
    return {"key": key, "label": key, "kind": kind, "description": "x", **extra}


class TestModifierTargetsNameOneRow(unittest.TestCase):
    """Gate 4: every modifier target names one row, and the op fits its kind."""

    def tables(self, root, targets=()):
        return [
            write_json(root / "abilities" / "core.json", {"abilities": [{"id": "str", "label": "Strength"}]}),
            write_json(root / "skills" / "core.json", {"skills": [{"key": "persuasion"}, {"key": "block"}, {"key": "piercing"}]}),
            write_json(root / "derived" / "core.json", {"derived": [{"key": "hp"}]}),
            write_json(root / "actions" / "core.json", {"actions": [{"key": "dodge", "cost": 3}, {"key": "block", "cost": 2}]}),
            write_json(root / "conditions" / "core.json", {"conditions": [{"key": "prone"}]}),
            write_json(root / "combat" / "damage_types.json", {"damage_types": [{"key": "fire"}, {"key": "piercing"}]}),
            write_json(root / "modifier_targets" / "core.json", {"modifier_targets": list(targets)}),
        ]

    def check(self, root, modifiers, targets=(), vectors=None):
        files = self.tables(root, targets)
        files.append(write_json(root / "features" / "core.json", {"features": [{"key": "x", "modifiers": modifiers}]}))
        vector_files = [write_json(root / "vectors" / "v.json", vectors)] if vectors is not None else []
        return validate.check_modifier_targets(files, vector_files=vector_files)

    def test_the_real_data_passes(self):
        data_files = list(validate.collect_data_files())
        modules = list(validate.collect_module_data_files())
        self.assertEqual(validate.check_modifier_targets(data_files, modules, validate.collect_vector_files()), [])

    def test_a_target_in_each_table_passes(self):
        with tempfile.TemporaryDirectory() as tmp:
            mods = [modifier("str"), modifier("persuasion"), modifier("hp", "mul", 1.5), modifier("dodge", value=-1),
                    modifier("prone", "grant"), modifier("fire", "resist", 5), modifier("crit_window")]
            self.assertEqual(self.check(Path(tmp), mods, [target_row("crit_window", pending=True)]), [])

    def test_an_unknown_target_is_reported(self):
        with tempfile.TemporaryDirectory() as tmp:
            errors = self.check(Path(tmp), [modifier("persuasoin")])
            self.assertEqual(len(errors), 1, errors)
            self.assertIn("persuasoin", errors[0])

    def test_an_op_that_does_not_fit_the_kind_is_reported(self):
        with tempfile.TemporaryDirectory() as tmp:
            mods = [modifier("persuasion", "grant"), modifier("prone", "add"), modifier("hp", "resist"),
                    modifier("darkvision", "add")]
            errors = self.check(Path(tmp), mods, [target_row("darkvision", "flag", pending=True)])
            self.assertEqual(len(errors), 4, errors)

    def test_a_key_in_two_tables_is_ambiguous_only_where_both_fit_the_op(self):
        with tempfile.TemporaryDirectory() as tmp:
            errors = self.check(Path(tmp), [modifier("block"), modifier("piercing"), modifier("piercing", "resist")])
            self.assertEqual(len(errors), 1, errors)
            self.assertIn("block", errors[0])

    def test_an_unused_target_row_is_reported(self):
        with tempfile.TemporaryDirectory() as tmp:
            errors = self.check(Path(tmp), [], [target_row("idle", pending=True)])
            self.assertEqual(len(errors), 1, errors)
            self.assertIn("idle", errors[0])

    def test_a_target_row_may_not_repeat_another_table(self):
        with tempfile.TemporaryDirectory() as tmp:
            errors = self.check(Path(tmp), [modifier("persuasion")], [target_row("persuasion", pending=True)])
            self.assertEqual(len(errors), 1, errors)
            self.assertIn("skills", errors[0])

    def test_pending_must_match_the_vectors(self):
        with tempfile.TemporaryDirectory() as tmp:
            vectors = [{"modifiers": [modifier("loot_roll")]}]
            mods = [modifier("loot_roll"), modifier("crit_window")]
            errors = self.check(Path(tmp), mods, [target_row("loot_roll", pending=True), target_row("crit_window")], vectors)
            self.assertEqual(len(errors), 2, errors)
            self.assertTrue(any("loot_roll" in e and "vector" in e for e in errors), errors)
            self.assertTrue(any("crit_window" in e and "pending" in e for e in errors), errors)

    def test_a_module_adds_its_own_targets(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            files = self.tables(root)
            module = write_json(
                root / "mod" / "data.json",
                {"modifier_targets": [target_row("guild_member", "flag", pending=True)],
                 "factions": [{"key": "guild", "modifiers": [modifier("guild_member", "grant")]}]},
            )
            self.assertEqual(validate.check_modifier_targets(files, [(module, root / "mod")], []), [])


class TestModifiersNameTheirOperation(unittest.TestCase):
    """M5-46: a modifier in the data spells out `operation`, and `multiply` (CONTRIBUTING.md#naming)."""

    def test_no_data_modifier_uses_op_or_mul(self):
        ruleset = _SCRIPTS.parent
        files = [f for f in (ruleset / "data").rglob("*.json")]
        files += [f for f in (ruleset / "modules").rglob("*.json") if "tests" not in f.parts]
        found = []

        def walk(node, path):
            if isinstance(node, dict):
                if "target" in node and ("op" in node or node.get("operation") == "mul"):
                    found.append(f"{path.relative_to(ruleset)}: {node['target']}")
                for value in node.values():
                    walk(value, path)
            elif isinstance(node, list):
                for value in node:
                    walk(value, path)

        for path in files:
            walk(json.loads(path.read_text()), path)
        self.assertEqual(found, [])


if __name__ == "__main__":
    unittest.main()
