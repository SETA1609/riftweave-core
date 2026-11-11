#!/usr/bin/env python3
"""Each combat number lives in one table (docs/core/resolution.md, docs/core/combat.md).

Run with:
    python -m unittest ruleset/scripts/test_one_combat_table.py
"""

import json
import unittest
from pathlib import Path

RULESET = Path(__file__).resolve().parent.parent
DATA = RULESET / "data"


class TestOneCombatTable(unittest.TestCase):
    def test_the_legacy_table_and_its_simulator_are_gone(self):
        self.assertFalse((DATA / "combat" / "base_resolution.json").exists())
        self.assertFalse((RULESET / "scripts" / "combat_reference.py").exists())

    def test_no_data_or_schema_has_a_video_game_block(self):
        for path in [*DATA.rglob("*.json"), *(RULESET / "schemas").glob("*.json")]:
            with self.subTest(file=str(path.relative_to(RULESET))):
                self.assertNotIn('"videoGame"', path.read_text())

    def test_each_number_is_in_its_kernel_table(self):
        resolution = json.loads((DATA / "resolution" / "core.json").read_text())["resolution"]
        self.assertIn("fumble_at", resolution)
        self.assertIn("crit_damage", json.loads((DATA / "combat" / "pipeline.json").read_text())["pipeline"])
        self.assertIn("pool", json.loads((DATA / "actions" / "core.json").read_text()))


if __name__ == "__main__":
    unittest.main()
