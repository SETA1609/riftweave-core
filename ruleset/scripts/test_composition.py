#!/usr/bin/env python3
"""One way to compose a row (docs/core/composition.md)."""

import sys
import unittest
from pathlib import Path

_SCRIPTS = Path(__file__).resolve().parent
sys.path.insert(0, str(_SCRIPTS))

import validate  # noqa: E402


class TestComposeRow(unittest.TestCase):
    def test_overrides_land_on_the_base(self):
        row = validate.compose_row({"hitPoints": 10, "size": "small"}, {"hitPoints": 20})
        self.assertEqual(row, {"hitPoints": 20, "size": "small"})

    def test_the_base_is_not_changed(self):
        base = {"tags": ["beast"], "attributes": {"str": 4}}
        validate.compose_row(base, {"attributes": {"str": 9}}, layers=[lambda row: row["tags"].append("x")])
        self.assertEqual(base, {"tags": ["beast"], "attributes": {"str": 4}})

    def test_layers_apply_in_order_before_the_overrides(self):
        seen = []
        layers = [lambda row: seen.append(row["level"]) or row.update(level=2),
                  lambda row: seen.append(row["level"]) or row.update(level=3)]
        row = validate.compose_row({"level": 1}, {"level": 9}, layers=layers)
        self.assertEqual(seen, [1, 2])
        self.assertEqual(row["level"], 9)

    def test_attributes_merge_per_attribute(self):
        row = validate.compose_row({"attributes": {"str": 4, "agi": 5}}, {"attributes": {"str": 9}})
        self.assertEqual(row["attributes"], {"str": 9, "agi": 5})

    def test_a_list_is_replaced_unless_it_says_extend(self):
        base = {"tags": ["beast"]}
        self.assertEqual(validate.compose_row(base, {"tags": ["x"]})["tags"], ["x"])
        self.assertEqual(validate.compose_row(base, {"tags": {"$extend": ["x"]}})["tags"], ["beast", "x"])

    def test_identity_comes_from_the_row_itself(self):
        row = validate.compose_row({"key": "base", "id": 1}, {}, entry={"key": "wolf", "id": 7, "source": {"source": "s"}})
        self.assertEqual((row["key"], row["id"], row["source"]), ("wolf", 7, {"source": "s"}))


if __name__ == "__main__":
    unittest.main()
