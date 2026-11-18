### PO_S00_01_issue-splitting-by-role Better documentation on how issues should be split, with a table reference
- Added: `docs/issues/ROLES.md` with the roles PO, GD, DE, BE, FE, QA, and OPS, and stories `S00` to `S16`, one file each in `docs/stories/`.
- Changed: new issues are named `<ROLE>_<STORY>_<ISSUE>_<short-description>`; the open issues are renamed with a `Was:` line, and full-stack issues are split into a backend and a frontend issue. Old ids resolve through the table in `docs/stories/README.md`.
- Removed: the prefix proposal of M5-77 (dropped).
