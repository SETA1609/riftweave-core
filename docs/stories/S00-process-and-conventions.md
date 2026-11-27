# S00 Process and conventions

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S00 Process and conventions

**Goal:** Issues are easy to file, split, and find: one role per issue, a story for each goal, and tools that read the ids.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `PO_S00_01_issue-splitting-by-role` | PO | Better documentation on how issues should be split, with a table reference | — | M5-76 |
| `OPS_S00_02_format-data-files` | OPS | Format the data files | M5-52 | M5-51 |
| `OPS_S00_03_doc-last-changed-stamp` | OPS | Every doc shows when it last changed | M5-53 | — |
| `OPS_S00_04_refresh-stale-story-docs` | OPS | Refresh the story and issue docs after S05 and S06 landed | — | — |
| `OPS_S00_06_faster-python-checks` | OPS | Faster Python checks: reuse validators, cache loot values, run test files side by side | — | — |
| `OPS_S00_07_dispatch-and-string-constants` | OPS | Dispatch by switch or table, and named constants for meaningful strings | — | — |
| `FE_S00_09_frontend-unit-test-gaps` | FE | Unit specs for the frontend code that has none | — | — |
| `PO_S00_10_close-stuck-po-issues` | PO | Close the four PO issues that merged as todo | — | — |
| `PO_S00_11_json-schema-theory` | PO | A theory page on JSON Schema: how it works, why, and for what | — | — |

| `FE_S00_08_sibling-docs-for-code` | FE | Move the frontend's code comments into sibling Markdown files | — | FE_S11_01_frontend-i18next |

| `FE_S00_05_regenerate-openapi-types` | FE | Regenerate the frontend's API types from the current contract | — | — |

**GitHub:** story [#258](https://github.com/SETA1609/riftweave-core/issues/258); tasks [`OPS_S00_02_format-data-files` #276](https://github.com/SETA1609/riftweave-core/issues/276), [`OPS_S00_03_doc-last-changed-stamp` #277](https://github.com/SETA1609/riftweave-core/issues/277).

**Done when:** every issue above is done, and every open issue has a role id and a story; the tooling reads the new ids; the data files and docs carry their format and modified date.
