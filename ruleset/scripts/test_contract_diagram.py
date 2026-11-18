#!/usr/bin/env python3
"""The command contract diagram is the contract's own (docs/backend/contract-diagram.md, APP-36).

Run with:
    python -m unittest ruleset/scripts/test_contract_diagram.py
"""
import re
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import contract_diagram as diagram  # noqa: E402


class TestContractDiagram(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.text = diagram.PAGE.read_text()

    def test_the_page_is_what_the_contract_gives(self):
        self.assertEqual(self.text, diagram.page(), "run python ruleset/scripts/contract_diagram.py --write")

    def test_every_command_has_its_diagram(self):
        for name in diagram.commands():
            with self.subTest(command=name):
                self.assertIn(f"\n## {name}\n", self.text)
                self.assertIn(f"    class {name}_input {{", self.text)
                self.assertIn(f"    class {name}_result {{", self.text)

    def test_the_session_table_names_the_commands_the_session_adapts(self):
        table = {command for command, _, _ in diagram.session_table()}
        self.assertEqual(table - diagram.NOT_ADAPTED, set(diagram.session_adapters()))

    def test_the_session_flowchart_draws_every_row_of_the_table(self):
        for command, filled, written in diagram.session_table():
            with self.subTest(command=command):
                self.assertIn(f'actor -- "fills: {diagram.label(filled)}" --> {command}', self.text)
                self.assertIn(f'{command} -- "writes back: {diagram.label(written)}" --> actor', self.text)

    def test_a_link_lands_on_a_drawn_shared_type(self):
        drawn = set(re.findall(r"^    class ([a-z_]+) \{", self.text.split("## Session steps")[0], re.M))
        for target in re.findall(r"^    [a-z_]+ --> ([a-z_]+) :", self.text, re.M):
            with self.subTest(target=target):
                self.assertIn(target, drawn)

    def test_each_block_is_a_mermaid_diagram(self):
        blocks = re.findall(r"```mermaid\n(.*?)```", self.text, re.S)
        self.assertEqual(len(blocks), 2 + len(diagram.commands()))
        for block in blocks:
            self.assertRegex(block, r"^(classDiagram|flowchart LR)\n")
            self.assertNotIn("`", block)


if __name__ == "__main__":
    unittest.main()
