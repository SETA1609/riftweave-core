# PO_S20_01_file-custom-monster-and-template-issues File the custom monster and character template issues

Date: 2026-11-20

- **Decision:** a new story, S20 Custom monsters and character templates (owner request, 2026-11-20), in two parts.
  - **Custom monsters:** a GD rule for what a GM's own monster skips and keeps (`GD_S20_02`), then the editor in the CLI and TUI (`BE_S20_03`) and in the app (`FE_S20_04`).
  - **Character templates:** the model (`GD_S20_05`), then the first undead (`GD_S20_06`: lich, vampire) and werebeast templates (`GD_S20_07`: werewolf, werebear, and others), then templates on a saved character (`BE_S20_08`) and on the sheet in the app (`FE_S20_09`).
- **Why a GD issue for custom monsters:** `monster_sheet` already computes a monster with the player formulas, HP included since DE_S02_03, and attributes may pass 10 [V 2026-11-20, docs/core/actor-sheet.md, `test_bestiary.py`]. What isn't written down is which of a player's limits a GM's own monster drops: point-buy, attribute cap, perk prerequisites, level-up pools. An editor needs that rule before it can be built.
- **Builds on S09:** `BE_S09_07_custom-monster-print` already plans the `monster` file kind and the GM's custom monster list, to print them. `BE_S20_03` writes them, so it needs `BE_S09_07`, and `FE_S20_04` needs `FE_S09_08`.
- **Why character templates are their own model:** monster templates exist (`element`, `power`, `nature`), but a monster's creature type never changes, so there is deliberately no `undead` monster template [V 2026-11-20, docs/modules/monsters.md#templates]. A lich or vampire character has to change what the character counts as. A character template is also taken and removed in play, not composed once like a monster row.
- **Shape of a template** (all [A], settled in `GD_S20_05`'s spec step):
  - It has one slot (`undeath`, `lycanthropy`).
  - It adds attributes, skills, pool modifiers (`hp`, `mp`, `sp`, `ap`) and granted perks, plus tags, a creature type, resistances and drawbacks.
  - Everything goes through the modifier pipeline with the `innate` origin, so nothing stacks twice and removal is clean.
- **Open:**
  - Do monster templates move to the same shape, so a vampire monster and a vampire PC share a row?
  - Is a template a level-up choice, a GM's story grant, or both, and does it cost anything?
  - Should a custom monster far outside its role's balance band be refused, or only noted?
  - Are a werebeast's forms part of the template, or a separate condition?
  - Which werebeasts come after werewolf and werebear?
