"""data_flowcharts.py (M5-63): the perk ladders and formula trees are drawn from the data and
cannot go stale."""

import json
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import data_flowcharts as charts  # noqa: E402


class TestDataFlowcharts(unittest.TestCase):
    def test_the_pages_are_what_the_data_gives(self):
        self.assertEqual(charts.LADDERS_PAGE.read_text(), charts.ladders_page(), "run python ruleset/scripts/data_flowcharts.py --write")
        self.assertEqual(charts.FORMULAS_PAGE.read_text(), charts.formulas_page(), "run python ruleset/scripts/data_flowcharts.py --write")

    def test_every_perk_on_a_ladder_is_a_node_and_every_other_perk_is_listed(self):
        page = charts.ladders_page()
        charted, table = page.split("## Perks on no ladder")
        for perk in charts.perks():
            with self.subTest(perk=perk["key"]):
                linked = perk["prerequisite"].get("perks") or any(
                    charts.held(entry)[0] == perk["key"] for other in charts.perks() for entry in other["prerequisite"].get("perks", []))
                if linked:
                    self.assertIn(f"  {charts.node_id(perk['key'])}[", charted)
                else:
                    self.assertIn(f"| {perk['label']} |", table)

    def test_every_prerequisite_is_an_edge_with_its_rank(self):
        page = charts.ladders_page()
        for perk in charts.perks():
            for entry in perk["prerequisite"].get("perks", []):
                key, rank = charts.held(entry)
                arrow = f" -- rank {rank} --> " if rank > 1 else " --> "
                with self.subTest(perk=perk["key"], needs=key):
                    self.assertIn(f"  {charts.node_id(key)}{arrow}{charts.node_id(perk['key'])}", page)

    def test_every_formula_tree_in_the_data_has_a_chart(self):
        page = charts.formulas_page()
        found = 0
        for file in sorted(charts.DATA.rglob("*.json")):
            if "bases" in file.parts:
                continue
            for path, _ in charts.trees(json.loads(file.read_text())):
                found += 1
                with self.subTest(file=file.name, path=path):
                    self.assertIn(f"### `{'.'.join(path) or '(the file)'}`", page)
        self.assertGreater(found, 0)
        self.assertEqual(page.count("```mermaid"), found)

    def test_a_tree_draws_one_node_per_operation_and_leaf(self):
        tree = {"operation": "add", "arguments": [{"literal": 2}, {"operation": "divide", "arguments": [{"attribute": "agi"}, {"literal": 3}]}]}
        chart = charts.tree_chart(tree)
        self.assertEqual(chart.count("  n"), 5 + 4)  # five nodes, four edges
        self.assertIn('{"+"}', chart)
        self.assertIn('(["AGI"])', chart)


if __name__ == "__main__":
    unittest.main()
