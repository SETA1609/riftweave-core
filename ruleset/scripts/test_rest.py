#!/usr/bin/env python3
"""Every rest occasion names a recovery rule, or an explicit 'none', for every resource.

Run with:
    python -m unittest ruleset/scripts/test_rest.py
"""

import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

from kernel_reference import load_data  # noqa: E402

RESOURCES = {"hp", "mp", "stamina", "ap", "conditions"}


class TestRestData(unittest.TestCase):
    def test_every_occasion_covers_every_resource_once(self):
        for occasion in load_data("rest/core.json")["rest"]:
            with self.subTest(occasion=occasion["key"]):
                resources = [row["resource"] for row in occasion["recovery"]]
                self.assertEqual(sorted(resources), sorted(RESOURCES))

    def test_occasion_keys_are_unique(self):
        keys = [o["key"] for o in load_data("rest/core.json")["rest"]]
        self.assertEqual(len(keys), len(set(keys)))


if __name__ == "__main__":
    unittest.main()
