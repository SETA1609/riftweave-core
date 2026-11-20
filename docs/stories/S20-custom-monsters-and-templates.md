# S20 Custom monsters and character templates

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S20 Custom monsters and character templates

**Goal:** A GM builds their own monster the way a player builds a character: the same attributes, skills, perks, and derived stats from the same formulas, with none of a player's limits (no point-buy, no attribute cap, no perk prerequisites, any level). And a character can take on a template later in play, such as a lich, a vampire, or a werebeast (wolf, bear, and others), which adds attributes, skills, pools (HP, MP, SP, AP), and perks on top of what the character already is, and can be taken off again.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `PO_S20_01_file-custom-monster-and-template-issues` | PO | File the custom monster and character template issues | — | — |
| `GD_S20_02_custom-monster-rules` | GD | A custom monster computes like a character, without a player's limits | — | — |
| `BE_S20_03_custom-monster-editor-cli` | BE | Build and save a custom monster in the CLI and the TUI (backend) | — | GD_S20_02_custom-monster-rules, BE_S09_07_custom-monster-print |
| `FE_S20_04_custom-monster-editor-page` | FE | Build and save a custom monster in the app (frontend) | — | BE_S20_03_custom-monster-editor-cli, FE_S09_08_custom-monster-page |
| `GD_S20_05_character-template-model` | GD | Character templates: a reusable change that adds attributes, skills, pools, and perks | — | — |
| `GD_S20_06_undead-templates` | GD | Undead templates: lich and vampire | — | GD_S20_05_character-template-model |
| `GD_S20_07_werebeast-templates` | GD | Werebeast templates: werewolf, werebear, and others | — | GD_S20_05_character-template-model |
| `BE_S20_08_templates-on-a-character-cli` | BE | Take on and remove a template on a saved character (backend) | — | GD_S20_05_character-template-model |
| `FE_S20_09_templates-on-the-sheet-page` | FE | A character's templates on the sheet in the app (frontend) | — | BE_S20_08_templates-on-a-character-cli |

**Done when:** every issue above is done; a GM saves a custom monster built past a player's caps and its stat block matches `monster_sheet`; a character takes on the lich and a werewolf template and drops one again, and its sheet matches the `derived` vectors each time.
