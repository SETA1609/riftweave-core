# S11 Localization

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S11 Localization

**Goal:** The frontend and the human CLI speak English, German, and Spanish from one set of locale files.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `FE_S11_01_frontend-i18next` | FE | Localize the frontend with i18next | APP-24 | APP-23 |
| `BE_S11_02_cli-shared-locales` | BE | One set of i18next locale files for the frontend and the human CLI (backend) | APP-28 | FE_S11_01_frontend-i18next, APP-26 |
| `FE_S11_03_frontend-shared-locales` | FE | One set of i18next locale files for the frontend and the human CLI (frontend) | APP-28 | BE_S11_02_cli-shared-locales, FE_S11_01_frontend-i18next, APP-26 |
| `FE_S11_04_frontend-screen-strings` | FE | Localize the play and character screens | FE_S11_01 | FE_S11_01_frontend-i18next |

**GitHub:** story [#268](https://github.com/SETA1609/riftweave-core/issues/268); tasks [`FE_S11_01_frontend-i18next` #328](https://github.com/SETA1609/riftweave-core/issues/328), [`BE_S11_02_cli-shared-locales` #329](https://github.com/SETA1609/riftweave-core/issues/329), [`FE_S11_03_frontend-shared-locales` #330](https://github.com/SETA1609/riftweave-core/issues/330).

**Done when:** every issue above is done, and both apps render the same keys from the same files in en, de, and es.
