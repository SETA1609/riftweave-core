#### APP-34 An offline encounter tracker: initiative and whose turn it is
- Status: todo
- Priority: medium
- Needs: APP-33
- Modified: 2026-11-16
- Do: Owner request, 2026-11-16: the GM runs a fight on one device with no connection. A GM tool in both apps holds the encounter: the party (APP-33) and the monsters the GM adds from the bestiary (APP-29 when merged, else by key), each with initiative, the derived `initiative` value only, no roll (docs/core/derived.md) (owner, 2026-11-16), HP, SP, mana, AP left, and conditions; the **initiative list** in order, the **current turn**, next turn, and the round count. Every change is a session step (`encounter_start`, `turn_end`, damage and conditions through the existing commands), so the encounter replays, exports with the session (APP-30), and is the same state the online table syncs (APP-35) [A]. Ties break by AGI, then LCK, then the GM's order [A]. No grid offline: positions are what the table says [A].
- Keys: encounter session steps; the initiative order
- Files: ~12 (backend/contract/session.json, the session package, ruleset/tests/vectors/session.json, ruleset/scripts/kernel_reference.py, backend/cmd/riftweave/play/encounter.go, play_test.go, frontend/src/app/encounter/, app.spec.ts, e2e/encounter.spec.ts, docs/backend/sessions.md, docs/core/derived.md, changes/APP-34.md)
- See: docs/core/derived.md, docs/modules/combat.md, docs/issues/open/APP-29.md
- Done when: a GM runs an encounter of four characters and two monsters offline in each app; the initiative order and turns replay the same in Python and Go; the tests pass.
