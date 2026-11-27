# S11 Localization

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S11 Localization

**Goal:** The frontend and the printouts speak English, German, and Spanish, from one set of catalogs the Go core and the frontend both read (owner, 2026-11-27). Sharing the locale files with the terminal app was dropped with that app (BE_S11_02 and FE_S11_03, PO_S23_02).

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `FE_S11_01_frontend-i18next` | FE | Localize the frontend with i18next | APP-24 | APP-23 |
| `BE_S11_02_cli-shared-locales` | BE | One set of i18next locale files for the frontend and the human CLI (backend) | APP-28 | FE_S11_01_frontend-i18next, APP-26 |
| `FE_S11_03_frontend-shared-locales` | FE | One set of i18next locale files for the frontend and the human CLI (frontend) | APP-28 | BE_S11_02_cli-shared-locales, FE_S11_01_frontend-i18next, APP-26 |
| `FE_S11_04_frontend-screen-strings` | FE | Localize the play and character screens | FE_S11_01 | FE_S11_01_frontend-i18next |
| `BE_S11_05_shared-locale-catalogs` | BE | One set of locale catalogs for the Go core and the frontend, with a Go reader (backend) | — | FE_S11_04_frontend-screen-strings, FE_S11_08_split-locale-files |
| `FE_S11_06_frontend-own-i18n-reader` | FE | The frontend's own catalog reader in place of i18next, with plurals | — | BE_S11_05_shared-locale-catalogs |
| `BE_S11_07_localized-printouts` | BE | Printouts in the reader's language: a locale for the print command | — | BE_S11_05_shared-locale-catalogs |
| `FE_S11_08_split-locale-files` | FE | One locale folder per area, so a new page adds its own words without touching shared files | — | FE_S11_04_frontend-screen-strings |

**GitHub:** story [#268](https://github.com/SETA1609/riftweave-core/issues/268); tasks [`FE_S11_01_frontend-i18next` #328](https://github.com/SETA1609/riftweave-core/issues/328), [`BE_S11_02_cli-shared-locales` #329](https://github.com/SETA1609/riftweave-core/issues/329), [`FE_S11_03_frontend-shared-locales` #330](https://github.com/SETA1609/riftweave-core/issues/330).

**Done when:** every issue above is done; the frontend and the `print` pages render their keys in en, de, and es from `backend/contract/locales/`, read by one Go reader and one TypeScript reader that give the same answers for the shared vectors.
