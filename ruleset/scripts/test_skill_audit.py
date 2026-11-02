#!/usr/bin/env python3
"""Tests for skill_audit.py.

Run with:
    python -m unittest ruleset/scripts/test_skill_audit.py
"""

import json
import sys
import tempfile
import unittest
from pathlib import Path

_SCRIPTS = Path(__file__).resolve().parent
sys.path.insert(0, str(_SCRIPTS))

import skill_audit  # noqa: E402
from skill_audit import audit, collect_uses, load_skills, parse_seeds  # noqa: E402

SKILLS = [
    {"key": "blades", "associatedAbility": "str"},
    {"key": "light_armor", "associatedAbility": "agi"},
    {"key": "red_magic", "associatedAbility": "int"},
    {"key": "lockpick", "associatedAbility": "agi"},
    {"key": "athletics", "associatedAbility": "str"},
    {"key": "insight", "associatedAbility": "per"},
    {"key": "repair", "associatedAbility": "int"},
]

DOC = """# Attributes

### Strength (STR)
- **Seeds skills:** `blades`, `athletics`.

### Perception (PER)
- **Seeds skills:** `insight`.

### Intelligence (INT)
- **Seeds skills:** the `<color>_magic` school `red_magic`, plus
  `repair`.

### Agility (AGI)
- **Seeds skills:** `lockpick`, `light_armor`.
- **Feeds:** `lockpick_bonus` is not a seed.

### Luck (LCK)
- **Governs:** fortune; seeds no skill.

## Next section
"""


def _write(root, rel, payload):
    fp = root / rel
    fp.parent.mkdir(parents=True, exist_ok=True)
    fp.write_text(json.dumps(payload))


def _dataset(root, with_check=True):
    _write(root, "skills/core.json", {"$schema": "x", "skills": SKILLS})
    # A base template has no $schema and wraps the weapon block.
    _write(root, "equipment/bases/weapons/sword.json",
           {"key": "sword", "weapon": {"skill": "core:skill/blades"}})
    _write(root, "equipment/bases/armor/vest.json",
           {"key": "vest", "armor": {"category": "light"}})
    _write(root, "spells/core.json",
           {"$schema": "x", "spells": [{"key": "bolt", "color": "red"}]})
    _write(root, "features/core.json", {"$schema": "x", "features": [
        {"key": "tinker", "effects": [
            {"target": "core:effect/fortify_skill", "parameter": "repair"}]}]})
    _write(root, "backgrounds/core.json", {"$schema": "x", "backgrounds": [
        {"skill_bonuses": [{"skill": "core:skill/lockpick", "bonus": 10}]}]})
    checks = [{"key": "climb", "skill": "core:skill/athletics",
               "opposed_by": "core:skill/insight"}]
    if with_check:
        checks.append({"key": "pick_lock", "skill": "core:skill/lockpick"})
    _write(root, "checks/core.json", {"$schema": "x", "checks": checks})


class TestCollectUses(unittest.TestCase):
    def test_finds_each_consumer_kind(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            _dataset(root)
            uses = collect_uses(root)
        self.assertIn("weapon", uses["blades"])
        self.assertIn("armor", uses["light_armor"])
        self.assertIn("spell", uses["red_magic"])
        self.assertIn("perk", uses["repair"])
        self.assertIn("check", uses["athletics"])
        self.assertIn("check", uses["insight"])  # opposed_by counts
        self.assertEqual(set(uses["lockpick"]), {"background", "check"})


class TestAudit(unittest.TestCase):
    def _run(self, with_check=True, doc=DOC):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            _dataset(root, with_check)
            doc_path = root / "attributes.md"
            doc_path.write_text(doc)
            return audit(load_skills(root), collect_uses(root), parse_seeds(doc_path))

    def test_clean_dataset_passes(self):
        _rows, problems = self._run()
        self.assertEqual(problems, [])

    def test_background_only_skill_is_an_orphan(self):
        _rows, problems = self._run(with_check=False)
        self.assertEqual(problems, ["orphan skill lockpick: references only: background"])

    def test_skill_with_no_references_is_an_orphan(self):
        skills = [{"key": "farming", "associatedAbility": "end"}]
        _rows, problems = audit(skills, {}, {"end": {"farming"}})
        self.assertIn("orphan skill farming: no references", problems)

    def test_unknown_skill_ref_is_reported(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            _dataset(root)
            _write(root, "checks/extra.json", {"$schema": "x", "checks": [
                {"key": "x", "skill": "core:skill/no_such_skill"}]})
            doc_path = root / "attributes.md"
            doc_path.write_text(DOC)
            _rows, problems = audit(load_skills(root), collect_uses(root),
                                    parse_seeds(doc_path))
        self.assertEqual(len(problems), 1)
        self.assertIn("unknown skill 'no_such_skill'", problems[0])

    def test_seed_mismatch_is_reported(self):
        doc = DOC.replace("`blades`, `athletics`", "`blades`, `building`")
        _rows, problems = self._run(doc=doc)
        self.assertEqual(problems, [
            "attributes.md seeds for str: missing athletics; not in data building"])

    def test_wrapped_seed_line_is_read(self):
        with tempfile.TemporaryDirectory() as tmp:
            doc_path = Path(tmp) / "attributes.md"
            doc_path.write_text(DOC)
            seeds = parse_seeds(doc_path)
        self.assertEqual(seeds["int"], {"red_magic", "repair"})
        self.assertEqual(seeds["agi"], {"lockpick", "light_armor"})
        self.assertEqual(seeds["lck"], set())


class TestRepository(unittest.TestCase):
    def test_core_data_has_no_orphan_skill(self):
        _rows, problems = audit(load_skills(), collect_uses(), parse_seeds())
        self.assertEqual(problems, [])

    def test_main_exit_code(self):
        self.assertEqual(skill_audit.main(["--quiet"]), 0)


if __name__ == "__main__":
    unittest.main()
