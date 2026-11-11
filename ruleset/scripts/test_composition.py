#!/usr/bin/env python3
"""One way to compose a row (docs/core/composition.md)."""

import json
import sys
import tempfile
import unittest
from pathlib import Path

_SCRIPTS = Path(__file__).resolve().parent
sys.path.insert(0, str(_SCRIPTS))

import validate  # noqa: E402

EQUIPMENT = _SCRIPTS.parent / "data" / "equipment"
ITEM_FILES = {"weapons.json": "weapons", "armor.json": "armor", "accessories.json": "accessories"}


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



class TestABaseThatIsAnItem(unittest.TestCase):
    """docs/core/composition.md#a-base-that-is-an-item"""

    def test_no_equipment_row_has_empty_overrides(self):
        for name in ITEM_FILES:
            for row in json.loads((EQUIPMENT / name).read_text())["equipment"]:
                with self.subTest(file=name, key=row["key"]):
                    self.assertNotEqual(row.get("overrides"), {})

    def test_every_equipment_base_states_is_item(self):
        for directory in ITEM_FILES.values():
            for path in sorted((EQUIPMENT / "bases" / directory).glob("*.json")):
                with self.subTest(base=path.name):
                    self.assertIsInstance(json.loads(path.read_text()).get("is_item"), bool)

    def test_an_item_base_is_the_row_of_its_own_key(self):
        rows = {row["key"]: row for row in validate.equipment_rows(EQUIPMENT / "weapons.json")}
        self.assertEqual(rows["dagger"], {"key": "dagger", "base": "core:weapon_base/dagger", "overrides": {}})
        self.assertNotIn("bow", rows)  # a shape only

    def test_resolving_drops_is_item(self):
        data = {"equipment": validate.equipment_rows(EQUIPMENT / "weapons.json")}
        validate._resolve_equipment_bases(data, [], EQUIPMENT / "weapons.json")
        self.assertTrue(all("is_item" not in row for row in data["equipment"]))

    def test_a_row_with_an_item_base_key_is_a_duplicate(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "bases" / "weapons").mkdir(parents=True)
            (root / "bases" / "weapons" / "club.json").write_text(json.dumps({"key": "club", "is_item": True}))
            (root / "weapons.json").write_text(json.dumps({"equipment": [{"key": "club", "base": "core:weapon_base/club", "overrides": {"weight": 2}}]}))
            errors = []
            validate.equipment_rows(root / "weapons.json", errors=errors)
            self.assertTrue(any("club" in error and "duplicate" in error for error in errors), errors)


if __name__ == "__main__":
    unittest.main()
