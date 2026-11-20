#### BE_S20_08_templates-on-a-character-cli Take on and remove a template on a saved character (backend)
- Status: todo
- Priority: medium
- Needs: GD_S20_05_character-template-model
- Modified: 2026-11-20
- Do: Owner request, 2026-11-20: the backend half of character templates (`GD_S20_05_character-template-model`).
  - **The stored actor:** gains `templates`, a list of template keys (and a werebeast's current form).
  - **The core:** `derived`, `combat`, `casting`, and the other commands read a held template's modifiers, perks, and tags as they read a perk's. The contract and the Go core run the GD vectors.
  - **The CLI and the TUI:**
    - **Take on a template:** the GM's grant in a session (a story or game event). A player cannot take one by themselves, unless the take is marked **outside the rules** (`outside_rules: true`), the app's checkbox for building a supposed character. Either way it is refused with a reason if the slot is taken; the stored template keeps `granted` or `outside_rules`.
    - **Remove it.** The sheet goes back to what it was.
    - **Change form** (a werebeast).

    The character file keeps its templates, so they sync, export, and import like the rest of the character (`BE_S07_03_character-list-cli`).
- Keys: stored actor `templates`; `take_template` and `remove_template` (or a `template` operation) in the contract
- Files: ~10 (backend/contract/common.json, backend/contract/progression.json, backend/internal/derived/derived.go, backend/internal/data/data.go, backend/core/session.go, backend/cmd/riftweave/play/characters.go, play_test.go, docs/backend/sessions.md, docs/decisions/BE_S20_08_templates-on-a-character-cli.md, changes/BE_S20_08_templates-on-a-character-cli.md)
- See: docs/issues/open/GD_S20_05_character-template-model.md, docs/backend/files.md
- Done when: every template vector passes in Go through every transport; a player's take without `outside_rules` is refused, and with it is stored as `outside_rules`; a TUI test takes on the lich template, checks the sheet, removes it, and the sheet matches the one before.
