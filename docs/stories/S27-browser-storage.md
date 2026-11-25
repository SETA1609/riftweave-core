# S27 Browser storage: IndexedDB laid out for how it is read

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S27 Browser storage: IndexedDB laid out for how it is read

**Goal:** The web app keeps characters, saved games, the profile, and drafts in IndexedDB the data-oriented way ([docs/theory/indexeddb.md](../theory/indexeddb.md)): small records split by who reads them, an index for each filter and sort a page has, one record per game step instead of the whole game rewritten, one transaction per user action, versioned upgrades tested from every old version, and persistent storage asked for when it matters. The core gets a column index for its lists only where a benchmark shows it pays.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `PO_S27_01_indexeddb-theory-and-storage-issues` | PO | A theory page on IndexedDB, and the storage issues | — | — |
| `FE_S27_02_versioned-storage-upgrades` | FE | Versioned IndexedDB upgrades, tabs, failures, and persistent storage (frontend) | — | — |
| `FE_S27_03_append-only-save-log` | FE | An append-only save log: one record per step (frontend) | — | FE_S27_02_versioned-storage-upgrades |
| `FE_S27_04_character-summary-store` | FE | The roster's summary store and its indexes (frontend) | — | FE_S27_02_versioned-storage-upgrades |
| `BE_S27_05_catalog-column-index` | BE | A measured column index for catalog lists (backend) | — | — |

**GitHub:** story [#499](https://github.com/SETA1609/riftweave-core/issues/499); tasks [`PO_S27_01_indexeddb-theory-and-storage-issues` #494](https://github.com/SETA1609/riftweave-core/issues/494), [`FE_S27_02_versioned-storage-upgrades` #495](https://github.com/SETA1609/riftweave-core/issues/495), [`FE_S27_03_append-only-save-log` #496](https://github.com/SETA1609/riftweave-core/issues/496), [`FE_S27_04_character-summary-store` #497](https://github.com/SETA1609/riftweave-core/issues/497), [`BE_S27_05_catalog-column-index` #498](https://github.com/SETA1609/riftweave-core/issues/498).

**Done when:** every issue above is done, a 200-step game autosaves one step record per step, the roster filters 500 characters through an index, and both databases upgrade from version 1 with their records intact.
