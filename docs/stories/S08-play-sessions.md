# S08 Play sessions in both apps

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S08 Play sessions in both apps

**Goal:** A player and a GM run their part of a session in either app: the player's character, items and crafting, the bestiary, encounters with a menu of what a character can do on their turn and what it costs, and a compendium of every table, with a way back from every screen.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `BE_S08_01_player-session-cli` | BE | A player's own session: level up, perks, spells, and crafting (backend) | APP-31 | APP-27, APP-30 |
| `FE_S08_02_player-session-page` | FE | A player's own session: level up, perks, spells, and crafting (frontend) | APP-31 | BE_S08_01_player-session-cli, APP-27, APP-30 |
| `BE_S08_03_item-lookup-craft-preview` | BE | Look up items, and see what a craft is expected to give (backend) | APP-32 | APP-27, BE_S08_10_compendium-catalog-cli |
| `FE_S08_04_item-lookup-page` | FE | Look up items, and see what a craft is expected to give (frontend) | APP-32 | BE_S08_03_item-lookup-craft-preview, APP-27 |
| `BE_S08_05_bestiary-knowledge-cli` | BE | A bestiary, and the GM decides what the players know (backend) | APP-29 | APP-26, APP-27, BE_S08_10_compendium-catalog-cli |
| `FE_S08_06_bestiary-page` | FE | A bestiary, and the GM decides what the players know (frontend) | APP-29 | BE_S08_05_bestiary-knowledge-cli, APP-26, APP-27 |
| `BE_S08_07_encounter-tracker-cli` | BE | An offline encounter tracker: initiative and whose turn it is (backend) | APP-34 | BE_S07_03_character-list-cli, GD_S19_09_initiative-ties |
| `FE_S08_08_encounter-tracker-page` | FE | An offline encounter tracker: initiative and whose turn it is (frontend) | APP-34 | BE_S08_07_encounter-tracker-cli, FE_S07_04_character-list-page |
| `PO_S08_09_file-compendium-creation-form-issues` | PO | File the compendium, back button, and creation form issues | — | — |
| `BE_S08_10_compendium-catalog-cli` | BE | A compendium of every table to study a build: the catalog command and the CLI screen (backend) | — | APP-27 |
| `FE_S08_11_compendium-tables-page` | FE | A compendium of every table to study a build (frontend) | — | BE_S08_10_compendium-catalog-cli, FE_S08_12_back-on-every-screen |
| `FE_S08_12_back-on-every-screen` | FE | A Back button on every screen | — | — |
| `PO_S08_13_file-action-menu-issues` | PO | File the action menu issues | — | — |
| `GD_S08_14_action-menu-rules` | GD | What a character can do on their turn, and what it costs: the action menu rules | — | GD_S05_01_loadout-attacks-melee-ladders, GD_S05_02_fighting-stances, GD_S05_03_defense-perk-ladders, GD_S05_04_trip-and-disarm |
| `BE_S08_15_action-menu-command-cli` | BE | What a character can do on their turn, and what it costs: the action menu command and the CLI screen (backend) | — | GD_S08_14_action-menu-rules, BE_S08_07_encounter-tracker-cli |
| `FE_S08_16_action-menu-page` | FE | What a character can do on their turn, and what it costs: the action menu in the app (frontend) | — | BE_S08_15_action-menu-command-cli, FE_S08_08_encounter-tracker-page |

**Done when:** every issue above is done, and each screen exists in the human CLI and the frontend with a way back, and its session steps replay the same in Python and Go.
