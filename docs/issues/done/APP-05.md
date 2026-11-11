#### APP-05 Human CLI
- Status: done
- Merged: #125, 2026-11-11
- Priority: low
- Needs: APP-03
- Do: A readable, interactive mode of the same binary, built as a terminal UI with the Charm libraries (Bubble Tea for the program loop, Bubbles for inputs and tables, Lip Gloss for styling): prompts for inputs, tables for results, a short log of the rolls. It calls the same core as the agent CLI; it holds no rule. The agent CLI path imports none of it, so its start stays fast, and `backend/core` and `backend/internal/` stay free of Charm.
- Files: ~4 (backend/cmd/riftweave/, docs/backend/cli.md)
- See: docs/backend/cli.md
- Done when: a person can create a character and resolve an attack interactively; the result matches the agent CLI for the same inputs and rolls.
