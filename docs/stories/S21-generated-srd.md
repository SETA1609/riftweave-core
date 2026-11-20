# S21 Generated SRD

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S21 Generated SRD

**Goal:** A System Reference Document as a PDF, like the D&D 5.1 SRD, built by code from `ruleset/data/` and the rules docs on every release and never kept by hand. Its numbers come from the same core the app runs, its prose from the rules docs without the builder's notes, and a check proves every entry is in it once.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `PO_S21_01_file-srd-book-issues` | PO | File the generated SRD book issues | — | — |
| `GD_S21_02_srd-book-outline` | GD | The SRD book's outline, and which prose a player reads | — | — |
| `GD_S21_03_srd-entry-layouts` | GD | What a stat block, a spell, a perk, and an item show, once for the book and for print | — | — |
| `DE_S21_04_srd-generator` | DE | Generate the SRD's source from the outline and the data | — | GD_S21_02_srd-book-outline, GD_S21_03_srd-entry-layouts |
| `OPS_S21_05_srd-typesetting` | OPS | Typeset the SRD into a PDF | — | DE_S21_04_srd-generator |
| `OPS_S21_06_srd-release` | OPS | Build the SRD PDF in CI and attach it to each release | — | OPS_S21_05_srd-typesetting |
| `QA_S21_07_srd-coverage-check` | QA | Check the generated SRD holds every entry, once, and no builder's notes | — | DE_S21_04_srd-generator |

**Done when:** every issue above is done; a release tag's GitHub release carries `Riftweave_SRD_<version>.pdf`, built in CI from that commit; no PDF is tracked in the repo; the coverage check passes.
