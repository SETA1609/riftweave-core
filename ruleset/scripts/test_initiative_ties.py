#!/usr/bin/env python3
"""One rule for initiative ties (GD_S19_09): written once in docs/core/derived.md, and every
other place that orders initiative links to it instead of stating its own.

Run with:
    python -m unittest ruleset/scripts/test_initiative_ties.py
"""

import re
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent.parent
DERIVED = REPO_ROOT / "docs" / "core" / "derived.md"
ANCHOR = "derived.md#initiative-ties"
# The places that order initiative and must link to the rule.
LINKERS = (
    REPO_ROOT / "docs" / "modules" / "combat.md",
    REPO_ROOT / "docs" / "backend" / "sessions.md",
)
# Docs a player or a programmer reads for the rules; issues and decisions quote the history.
RULE_DOCS = ("docs/core", "docs/modules", "docs/guide", "docs/backend")
TIE_STATEMENT = re.compile(r"\bties? (go|goes|break|breaks)\b", re.IGNORECASE)


def section(text, heading):
    _, found, rest = text.partition(f"\n## {heading}\n")
    return rest.split("\n## ", 1)[0] if found else ""


class TestInitiativeTies(unittest.TestCase):
    def test_derived_md_states_the_tie_rule(self):
        rule = section(DERIVED.read_text(), "Initiative ties")
        self.assertTrue(rule, "docs/core/derived.md needs a '## Initiative ties' section")
        order = [rule.find(word) for word in ("AGI", "LCK", "d100")]
        self.assertNotIn(-1, order, "the rule names AGI, then LCK, then a d100 roll-off")
        self.assertEqual(order, sorted(order), "AGI breaks a tie first, then LCK, then the roll-off")

    def test_the_other_places_link_to_it(self):
        for path in LINKERS:
            with self.subTest(file=path.name):
                self.assertTrue(ANCHOR in path.read_text(), f"{path.name} links to {ANCHOR}")

    def test_no_other_rule_doc_states_its_own_tie_rule(self):
        for folder in RULE_DOCS:
            for path in sorted((REPO_ROOT / folder).rglob("*.md")):
                if path == DERIVED:
                    continue
                for number, line in enumerate(path.read_text().splitlines(), 1):
                    if "initiative" in line.lower() and TIE_STATEMENT.search(line):
                        with self.subTest(file=str(path.relative_to(REPO_ROOT)), line=number):
                            self.assertIn(ANCHOR, line, "state the tie rule only in derived.md and link to it")


if __name__ == "__main__":
    unittest.main()
