#!/usr/bin/env python3
"""Character templates (GD_S20_05, docs/core/templates.md): the gate on the rows.

Run with:
    python -m unittest ruleset/scripts/test_character_templates.py
"""

import json
import unittest
from pathlib import Path

import validate

DATA = Path(__file__).resolve().parent.parent / "data"


class TestCharacterTemplates(unittest.TestCase):
    def test_every_modifier_is_innate_and_every_skill_resolves(self):
        rows = [{"key": "odd", "skills": {"core:skill/nope": 5}, "modifiers": [{"target": "hp", "operation": "add", "value": 5, "source": "template:odd", "origin": "core:origin/trained"}]}]
        self.assertEqual(
            validate.character_template_problems(rows, {"core:skill/intimidation"}),
            [
                "character template odd: skill core:skill/nope does not resolve",
                "character template odd: modifier on hp is core:origin/trained, not core:origin/innate",
            ],
        )

    def test_the_core_templates_pass(self):
        self.assertEqual(validate.check_character_templates(), [])
        rows = json.loads((DATA / "character_templates" / "core.json").read_text())["character_templates"]
        self.assertEqual({row["key"]: row["slot"] for row in rows}, {"revenant": "undeath", "ghoul": "undeath"})


if __name__ == "__main__":
    unittest.main()
