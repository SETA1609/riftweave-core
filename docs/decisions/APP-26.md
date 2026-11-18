# APP-26 The human CLI opens on a main menu, and Start opens the tools

Date: 2026-11-18

- **Decision:** the menu tree is `backend/contract/menu.json`, one entry per row (`key`, `parent`, `label`, and `issue` while that issue is open), embedded by the CLI and for APP-27 to import. The app adds Back to every menu but the main one. Docs: docs/backend/app-layout.md. [V 2026-11-18, `play_test.go`]
- **Pending entries** (Characters APP-33, Spell creation APP-22, Bestiary APP-29, Encounter tracker APP-34) are dimmed with the issue id and say which issue brings them, as the issue proposed [A].
- **Settings screens save at once**, each change checked by `settings.Parse` first; a refused change saves nothing. Export and import ask for a path.
- **Server mode** posts each command to `<server_url>/v1/commands/<name>`. It needs a URL before it can be picked. A test runs a skill resolution against an `httptest` server that calls the core and compares it with the agent CLI. [V 2026-11-18]
- **Skill resolution** takes skill, value, LCK, one target modifier, and a roll; a blank roll is rolled in the CLI and written into the input, so the input replays [A: one modifier field is enough for now].
- **Open:** loading only the checked modules is not part of this issue; core still loads every module in `ruleset/modules/` [U]. The theme only tells Lip Gloss whether the background is light or dark [A].
- **Overlap:** APP-27 also writes docs/backend/app-layout.md; whichever merges second adds to it.
