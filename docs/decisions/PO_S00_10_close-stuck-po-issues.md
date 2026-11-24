# PO_S00_10_close-stuck-po-issues Close the four PO issues that merged as todo

Date: 2026-11-24

- **Cause:** `close-merged-issue.yml` moves an issue only when its PR set `Status: review`. Four PO filing PRs did not (#414, #424, #455, #473), so their issues stayed in `open/` as `todo` and `roadmap-order show` listed them as ready. [V 2026-11-24, the merged PR branches against `docs/issues/open/`]
- **Not the workflow:** it worked as written; PO_S20_01 and PO_S23_02, whose PRs set `review`, were moved. It reads only the merged PR's own branch, so it will not come back to these four.
- **Repair:** the four move to `done/` with `Status: done` and their `Merged:` line, noted as moved by this issue. That breaks the rule that a PR does not move an issue file; it is the only way left, and once. `order.md` is left to the workflow, which regenerates it when this PR merges.
- **Kept open:** `FE_S00_09_frontend-unit-test-gaps` also matches a merged branch (#440), but that PR only filed the issue; its work is not done.
- **The guard:** `test_issue_files.py` fails an issue in both folders or a done issue without its status and `Merged:` line; AGENTS.md says a filing PR sets its own issue to `review`.
