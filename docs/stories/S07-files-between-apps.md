# S07 Files and characters between apps

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S07 Files and characters between apps

**Goal:** A file, a character, or a build made in one app opens in the other, and a character is created in either app through one form that the core checks step by step.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `APP-30` | BE, FE | A file made in one app opens in the other | — | APP-16, APP-26 |
| `BE_S07_03_character-list-cli` | BE | A character list, and characters travel between players and the GM as files (backend) | APP-33 | APP-30 |
| `FE_S07_04_character-list-page` | FE | A character list, and characters travel between players and the GM as files (frontend) | APP-33 | BE_S07_03_character-list-cli, APP-30 |
| `BE_S07_05_character-from-build-cli` | BE | A new character from a suggested build, in both apps (backend) | APP-38 | M5-47, BE_S08_01_player-session-cli, BE_S07_03_character-list-cli, BE_S07_08_character-creation-command |
| `FE_S07_06_character-from-build-page` | FE | A new character from a suggested build, in both apps (frontend) | APP-38 | BE_S07_05_character-from-build-cli, M5-47, FE_S08_02_player-session-page, FE_S07_04_character-list-page, FE_S07_09_character-creation-cv-form |
| `GD_S07_07_creation-rules-as-data` | GD | The character creation numbers as data | — | — |
| `BE_S07_08_character-creation-command` | BE | Character creation as one command that checks a draft step by step (backend) | — | GD_S07_07_creation-rules-as-data |
| `FE_S07_09_character-creation-cv-form` | FE | Character creation as a form laid out like a CV (frontend) | — | BE_S07_08_character-creation-command, FE_S08_12_back-on-every-screen |

**GitHub:** story [#264](https://github.com/SETA1609/riftweave-core/issues/264); tasks [`BE_S07_03_character-list-cli` #295](https://github.com/SETA1609/riftweave-core/issues/295), [`FE_S07_04_character-list-page` #296](https://github.com/SETA1609/riftweave-core/issues/296), [`BE_S07_05_character-from-build-cli` #297](https://github.com/SETA1609/riftweave-core/issues/297), [`FE_S07_06_character-from-build-page` #298](https://github.com/SETA1609/riftweave-core/issues/298), [`GD_S07_07_creation-rules-as-data` #299](https://github.com/SETA1609/riftweave-core/issues/299), [`BE_S07_08_character-creation-command` #300](https://github.com/SETA1609/riftweave-core/issues/300), [`FE_S07_09_character-creation-cv-form` #301](https://github.com/SETA1609/riftweave-core/issues/301).

**Done when:** every issue above is done, and a character made from a build in the frontend opens in the agent CLI and back, as files (the human CLI is retired, PO_S23_02).
