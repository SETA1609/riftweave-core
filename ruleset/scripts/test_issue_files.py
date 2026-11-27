#!/usr/bin/env python3
"""PO_S00_10 (docs/issues/README.md#how-an-issue-closes): a merged issue is in done/ with Status
done and its Merged line, and an issue is in only one of open/ and done/. The four PO issues that
merged with Status todo (#414, #424, #455, #473) were left in open/ by close-merged-issue.yml."""
import re
import unittest
from pathlib import Path

ISSUES = Path(__file__).resolve().parent.parent.parent / "docs" / "issues"
MERGED_STUCK = {
    "PO_S20_10_template-grant-answer": "#414, 2026-11-20",
    "PO_S21_01_file-srd-book-issues": "#424, 2026-11-20",
    "PO_S23_01_file-tui-retirement-issues": "#455, 2026-11-23",
    "PO_S24_01_file-go-refactor-issues": "#473, 2026-11-24",
}


class TestIssueFiles(unittest.TestCase):
    def test_an_issue_is_in_one_folder(self):
        opened = {path.stem for path in (ISSUES / "open").glob("*.md")}
        done = {path.stem for path in (ISSUES / "done").glob("*.md")}
        self.assertEqual(opened & done, set())

    def test_every_done_issue_is_done_with_its_merge(self):
        for path in sorted((ISSUES / "done").glob("*.md")):
            text = path.read_text(encoding="utf-8")
            with self.subTest(issue=path.stem):
                self.assertRegex(text, re.compile(r"^- Status: (done|dropped)$", re.M))
                if "- Status: done" in text:
                    self.assertRegex(text, re.compile(r"^- Merged: #\d+(, #\d+)*, \d{4}-\d{2}-\d{2}", re.M))

    def test_the_stuck_po_issues_are_done(self):
        for issue, merged in MERGED_STUCK.items():
            with self.subTest(issue=issue):
                path = ISSUES / "done" / f"{issue}.md"
                self.assertTrue(path.exists(), f"{issue} is still open")
                self.assertIn(f"- Merged: {merged}", path.read_text(encoding="utf-8"))


if __name__ == "__main__":
    unittest.main()
