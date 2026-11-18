#!/usr/bin/env python3
"""The data model diagram is the schemas' own (docs/core/data-model-diagram.md, M5-57), and so is
the composition diagram (docs/core/composition-diagram.md, M5-64).

Run with:
    python -m unittest ruleset/scripts/test_data_model_diagram.py
"""
import re
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import data_model_diagram as diagram  # noqa: E402


class TestDataModelDiagram(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.model = diagram.collections(diagram.load_schemas())

    def test_the_page_is_what_the_schemas_give(self):
        self.assertEqual(diagram.PAGE.read_text(), diagram.page(self.model),
                         "run python ruleset/scripts/data_model_diagram.py --write")

    def test_every_reference_lands_on_a_drawn_collection(self):
        for collection, edges in self.model.items():
            for field, target in edges:
                with self.subTest(collection=collection, field=field):
                    self.assertIn(target, self.model)

    def test_the_references_validate_py_checks_are_drawn(self):
        pairs = {(c, t) for c, edges in self.model.items() for _, t in edges}
        for pair in [("spells", "effects"), ("features", "features"), ("monsters", "monsters"),
                     ("equipment", "equipment"), ("races", "features"), ("recipes", "crafted_items"),
                     ("crafted_items", "materials"), ("summons", "monsters")]:
            with self.subTest(pair=pair):
                self.assertIn(pair, pairs)

    def test_every_area_names_only_collections_that_exist(self):
        for title, members in diagram.AREAS:
            for member in members:
                with self.subTest(area=title, collection=member):
                    self.assertIn(member, self.model)

    def test_each_diagram_is_a_mermaid_er_diagram_with_safe_names(self):
        blocks = re.findall(r"```mermaid\n(.*?)```", diagram.PAGE.read_text(), re.S)
        self.assertEqual(len(blocks), 1 + len(diagram.AREAS))
        for block in blocks:
            self.assertTrue(block.startswith("erDiagram\n"))
            for name in re.findall(r"^    ([A-Z_]+) [{}]", block, re.M):
                self.assertRegex(name, r"^[A-Z][A-Z0-9_]*$")


class TestCompositionDiagram(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.schemas = diagram.load_schemas()
        cls.text = diagram.COMPOSITION_PAGE.read_text()

    def test_the_page_is_what_the_schemas_and_data_give(self):
        self.assertEqual(self.text, diagram.composition_page(self.schemas),
                         "run python ruleset/scripts/data_model_diagram.py --write")

    def test_every_base_and_overrides_family_is_found(self):
        found = {collection: categories for collection, _, _, _, categories in diagram.families(self.schemas)}
        self.assertEqual(found, {"equipment": ["weapon_base", "armor_base", "accessory_base"],
                                 "monsters": ["creature_type", "monster"], "spells": ["spell"]})

    def test_each_family_has_its_shape_and_its_instances(self):
        for collection in ("Equipment", "Monsters", "Spells"):
            with self.subTest(family=collection):
                section = self.text.split(f"## {collection}\n", 1)[1].split("\n## ", 1)[0]
                self.assertEqual(section.count("```mermaid\nclassDiagram\n"), 2)
                self.assertIn("_WITH_BASE : base", section)

    def test_the_composed_row_lists_only_base_overrides_and_identity(self):
        for collection, _, full, composed, _ in diagram.families(self.schemas):
            with self.subTest(family=collection):
                self.assertIn("base", composed)
                self.assertIn("overrides", composed)
                self.assertNotIn("base", full)

    def test_the_perk_kinds_say_they_wait_for_m5_58_until_it_merges(self):
        self.assertEqual("Pending M5-58" in self.text, diagram.perk_kinds_pending(self.schemas))

    def test_each_block_is_a_class_diagram_with_safe_names(self):
        blocks = re.findall(r"```mermaid\n(.*?)```", self.text, re.S)
        self.assertEqual(len(blocks), 7)
        for block in blocks:
            self.assertTrue(block.startswith("classDiagram\n"))
            for name in re.findall(r"^    class ([^ \[{]+)", block, re.M):
                self.assertRegex(name, r"^[A-Za-z][A-Za-z0-9_]*$")


if __name__ == "__main__":
    unittest.main()
