# S09 Custom content and printing

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S09 Custom content and printing

**Goal:** Players and GMs author spells and monsters and print characters, spell cards, and stat blocks from both apps.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `BE_S09_01_spell-author-wasm` | BE | A spell editor page over the author command (backend) | APP-22 | APP-10, APP-14 |
| `FE_S09_02_spell-editor-page` | FE | A spell editor page over the author command (frontend) | APP-22 | BE_S09_01_spell-author-wasm, APP-10, APP-14 |
| `BE_S09_03_print-character-cli` | BE | Print the current character from both apps (backend) | APP-39 | BE_S07_03_character-list-cli |
| `FE_S09_04_print-character-page` | FE | Print the current character from both apps (frontend) | APP-39 | BE_S09_03_print-character-cli, FE_S07_04_character-list-page |
| `BE_S09_05_spell-cards-print` | BE | Spell cards, and custom spells printed as cards (backend) | APP-40 | APP-10, BE_S09_01_spell-author-wasm, BE_S09_03_print-character-cli |
| `FE_S09_06_spell-cards-page` | FE | Spell cards, and custom spells printed as cards (frontend) | APP-40 | BE_S09_05_spell-cards-print, APP-10, FE_S09_02_spell-editor-page, FE_S09_04_print-character-page |
| `BE_S09_07_custom-monster-print` | BE | Custom NPCs and monsters, printed (backend) | APP-41 | BE_S08_05_bestiary-knowledge-cli, BE_S09_03_print-character-cli |
| `FE_S09_08_custom-monster-page` | FE | Custom NPCs and monsters, printed (frontend) | APP-41 | BE_S09_07_custom-monster-print, FE_S08_06_bestiary-page, FE_S09_04_print-character-page |

**GitHub:** story [#266](https://github.com/SETA1609/riftweave-core/issues/266); tasks [`BE_S09_01_spell-author-wasm` #314](https://github.com/SETA1609/riftweave-core/issues/314), [`FE_S09_02_spell-editor-page` #315](https://github.com/SETA1609/riftweave-core/issues/315), [`BE_S09_03_print-character-cli` #316](https://github.com/SETA1609/riftweave-core/issues/316), [`FE_S09_04_print-character-page` #317](https://github.com/SETA1609/riftweave-core/issues/317), [`BE_S09_05_spell-cards-print` #318](https://github.com/SETA1609/riftweave-core/issues/318), [`FE_S09_06_spell-cards-page` #319](https://github.com/SETA1609/riftweave-core/issues/319), [`BE_S09_07_custom-monster-print` #320](https://github.com/SETA1609/riftweave-core/issues/320), [`FE_S09_08_custom-monster-page` #321](https://github.com/SETA1609/riftweave-core/issues/321).

**Done when:** every issue above is done, and both apps author a spell, print a character, a spell card sheet, and a custom monster, and the printed numbers equal the commands'.
