### PO_S00_10_close-stuck-po-issues Close the four PO issues that merged as todo
- Fixed: `PO_S20_10`, `PO_S21_01`, `PO_S23_01`, and `PO_S24_01` move to `docs/issues/done/`; they merged with `Status: todo`, so the close workflow left them open and the roadmap listed them as ready.
- Added: `test_issue_files.py`, which fails an issue in both folders or a done issue without its merge, and an AGENTS.md line that a filing PR sets its own issue to `review`.
