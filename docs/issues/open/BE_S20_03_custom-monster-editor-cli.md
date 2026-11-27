#### BE_S20_03_custom-monster-editor-cli Build and save a custom monster in the CLI and the TUI (backend)
- Status: todo
- Priority: high
- Needs: GD_S20_02_custom-monster-rules, BE_S09_07_custom-monster-print
- Modified: 2026-11-23
- Do: Owner request, 2026-11-20: the backend half of building a custom monster (`GD_S20_02_custom-monster-rules`). BE_S09_07 adds the `monster` file kind and the GM's custom monster list, to print them; this issue writes them.
  - **The editor** (the web page is FE_S20_04; an agent uses the commands and their MCP tools, and nothing is built for the retired terminal app, PO_S23_02): a GM starts from a blank monster, a creature type, or a core monster, and sets:
    - the level, role, and attributes, with no point-buy and no cap;
    - skills, perks at any rank, attacks, and resistances;
    - templates, one per slot.

    Each change shows the stat block `monster_sheet` gives, so the GM sees HP, to-hit, and evasion move as they edit.
  - **Saving:** the monster goes into the custom monster list as a `monster` file (the envelope of docs/backend/files.md, with `rules_version` and modules), which exports and imports like a character.
  - **The agent CLI:** `riftweave monster new|edit|list|export|import` with JSON in and out.
  - **Refusals:** a file the GD rules refuse (an unknown perk, two templates in one slot) is refused with its problems; a monster outside its role's balance band is saved with a note [A].
- Keys: `riftweave monster` commands; the editor screen
- Files: ~10 (backend/cmd/riftweave/monsters.go and its test, the monster MCP tools and their test, backend/cmd/riftweave/main.go, backend/contract/file.json, docs/backend/cli.md, docs/backend/files.md, docs/decisions/BE_S20_03_custom-monster-editor-cli.md, changes/BE_S20_03_custom-monster-editor-cli.md)
- See: docs/issues/open/BE_S09_07_custom-monster-print.md, docs/backend/files.md, docs/core/actor-sheet.md
- Done when: an MCP test builds a monster past every player limit, saves it, lists it, and prints a stat block equal to `monster_sheet`'s; an invalid one is refused with its problems; the agent CLI round-trips a monster file.
