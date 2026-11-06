"""Origin rows (docs/core/modifier.md#origin): one default origin per emitter."""

import json
import unittest
from pathlib import Path

ORIGINS = Path(__file__).resolve().parent.parent / "data" / "origins" / "core.json"


class TestOrigins(unittest.TestCase):
    def setUp(self):
        self.rows = json.loads(ORIGINS.read_text())["origins"]

    def test_each_emitter_has_at_most_one_default_origin(self):
        seen = {}
        for row in self.rows:
            for emitter in row.get("defaultFor", []):
                self.assertNotIn(emitter, seen, f"{emitter} defaults to {seen.get(emitter)} and {row['key']}")
                seen[emitter] = row["key"]

    def test_keys_and_ids_are_unique(self):
        self.assertEqual(len({r["key"] for r in self.rows}), len(self.rows))
        self.assertEqual(len({r["id"] for r in self.rows}), len(self.rows))


if __name__ == "__main__":
    unittest.main()
