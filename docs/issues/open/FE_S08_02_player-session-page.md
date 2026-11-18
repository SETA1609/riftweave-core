#### FE_S08_02_player-session-page A player's own session: level up, perks, spells, and crafting (frontend)
- Status: todo
- Priority: low
- Needs: BE_S08_01_player-session-cli, APP-27, APP-30
- Was: APP-31
- Do: Frontend half, split by role on 2026-11-18 (PO_S00_01_issue-splitting-by-role): the Angular screens of `BE_S08_01_player-session-cli` and their UX, over the same commands and session steps, built from the shared components (docs/backend/frontend.md#conventions). The behaviour, the keys, and the rules are specified in `BE_S08_01_player-session-cli`; this issue adds no contract or core change.
- Keys: the frontend screen and its route; no new contract keys
- Files: to split when started (frontend/src/app/play/character/)
- See: docs/issues/open/BE_S08_01_player-session-cli.md
- Done when: split into issues by area (level up, perks, spells, crafting), each with its own done line.
