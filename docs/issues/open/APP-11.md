#### APP-11 Printouts generated from the data
- Status: doing
- Priority: high
- Needs: APP-04, M5-49
- Do: Every printout should match the rules, so make them from the data: a backend command renders a character sheet (only the raised skills, as M5-49 allows, each with its target number printed so a player only checks roll <= number), spell cards, monster stat blocks with to-hit and DR worked out, and condition cards. Output is printable HTML or PDF [A]. A rule change is then a rebuild, not an errata pass.
- Keys: `print` command in `backend/contract/` [A]
- Files: ~8 (backend/, backend/contract/, docs/backend/print.md, docs/guide/quick-start.md)
- See: docs/backend/README.md, docs/issues/open/M5-49.md
- Done when: a sheet, a spell card, a stat block, and a condition card render from data; a test checks that a printed target number equals the backend's computed one.
