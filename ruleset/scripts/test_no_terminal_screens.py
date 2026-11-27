#!/usr/bin/env python3
"""PO_S23_02 (docs/stories/S23-retire-the-terminal-app.md): the terminal app is retired, so no open
issue outside S23 asks for a terminal screen. A feature is a core command, an MCP tool, and a web
page. An issue's title is its frozen name and is not read; S23's own issues name the app they retire,
and a dropped issue builds nothing.
"""
import re
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent.parent
OPEN = ROOT / "docs" / "issues" / "open"
DROPPED = re.compile(r"^- Status: dropped$", re.M)
SCREEN = re.compile(r"\bTUI\b|human CLIs?\b|terminal screen|\bCLI screen|both apps|cmd/riftweave/play/|\bplay_test\.go")


def screen_lines(text):
    """The lines of an issue file, title left out, that ask for a terminal screen; none for a dropped
    issue."""
    if DROPPED.search(text):
        return []
    return [line for line in text.splitlines()[1:] if SCREEN.search(line)]


class TestNoTerminalScreens(unittest.TestCase):
    def test_no_open_issue_outside_s23_builds_a_terminal_screen(self):
        for path in sorted(OPEN.glob("*.md")):
            if "_S23_" in path.name:
                continue
            with self.subTest(issue=path.stem):
                self.assertEqual(screen_lines(path.read_text(encoding="utf-8")), [])

    def test_the_check_reads_a_screen_step_and_not_a_title(self):
        self.assertEqual(screen_lines("#### X_S08_01 A page in both apps\n- Do: a web page and an MCP tool."), [])
        self.assertEqual(len(screen_lines("#### X\n- Do: a TUI screen.\n- Files: backend/cmd/riftweave/play/x.go")), 2)
        self.assertEqual(screen_lines("#### X\n- Status: dropped\n- Do: a TUI screen."), [])


if __name__ == "__main__":
    unittest.main()
