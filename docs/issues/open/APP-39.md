#### APP-39 Print the current character from both apps
- Status: todo
- Priority: medium
- Needs: APP-33
- Modified: 2026-11-17
- Do: Owner request, 2026-11-17: the `print` command (APP-11) renders a sheet from a stored actor, but no app calls it yet. Add a **Print** action to the character sheet and the character list (APP-33) in both apps. It prints the character as it is now: the current level, perks, spells, gear, and the session's state (HP, mana, and stamina spent, held conditions and modifiers), all read from the session, so every printed number is the one the rules compute. The Angular app opens the page and the browser's print dialog (PDF through the browser, offline through WebAssembly, APP-14); the human CLI writes the HTML to a file and prints its path. A choice between **current** (the session's state) and **fresh** (full HP, mana, and stamina, no conditions) covers both a mid-campaign printout and a sheet for a new table [A].
- Keys: `print` input `state: current | fresh` [A]; the print action in both apps
- Files: ~10 (backend/contract/print.json, backend/core/print.go and its test, backend/cmd/riftweave/play/print.go, play_test.go, frontend/src/app/sheet.ts, sheet.html, frontend/src/app/app.spec.ts, e2e/print.spec.ts, docs/backend/print.md, changes/APP-39.md)
- See: docs/backend/print.md, docs/issues/done/APP-11.md, docs/issues/open/APP-33.md
- Done when: both apps print a saved character; a test checks that a printed HP and a skill target equal the session's numbers for a wounded character, and that `fresh` prints full resources; the tests pass.
