#### APP-29 A bestiary, and the GM decides what the players know
- Status: todo
- Priority: medium
- Needs: APP-26, APP-27
- Modified: 2026-11-16
- Do: Owner request, 2026-11-16: a bestiary in both apps lists the monsters (core and enabled modules) with their sheet from the `monster_sheet` command, filterable by creature type, level, and phase. Its settings (in `settings.bestiary`, APP-25) hold one check box, **"Players see only what they know"**. Off (solo play, the default): every entry shows in full. On (playing with friends): the GM marks each monster's knowledge for the table as `unknown` (not listed), `seen` (name, picture or description, size), or `known` (the full sheet) [A], and may instead let a check decide: a Lore check (`resolution`, the `lore` skill) by the player that meets the monster reveals it to `known` for that player [A]. The marks are campaign state, not app settings: they are steps in the session log (`reveal`, with monster, level, and player or table) so they replay, travel in an exported session (APP-16, APP-30), and reach players through a room later (APP-17) [A]. The GM's view always shows everything, with each entry's mark.
- Keys: `settings.bestiary.players_see_only_known`; a `reveal` session step; knowledge `unknown` | `seen` | `known`
- Files: ~14 (backend/contract/session.json, backend/internal/... the session step, ruleset/tests/vectors/session.json, ruleset/scripts/kernel_reference.py, backend/cmd/riftweave/play/bestiary.go, play_test.go, frontend/src/app/bestiary/, frontend/src/app/app.spec.ts, e2e/bestiary.spec.ts, docs/backend/bestiary.md, docs/backend/sessions.md, changes/APP-29.md)
- See: backend/contract/monster_sheet.json, docs/issues/open/APP-17.md, docs/issues/open/APP-18.md
- Done when: both apps list and filter the bestiary; with the box off every sheet shows; with it on a player sees only `seen` and `known` entries at their level; a `reveal` step and a Lore check reveal replay the same in Python and Go; the tests pass.
