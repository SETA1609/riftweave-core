#### BE_S10_03_gm-approval-supabase Rooms: the GM approves steps, and the table sees who is there (backend)
- Status: todo
- Priority: low
- Needs: BE_S10_01_room-log-supabase
- Was: APP-18, split by role on 2026-11-18: this is the backend half (the contract, the Go core, the vectors, and the CLI and TUI with their UX); the browser half is `FE_S10_04_gm-approval-table`
- Do: Owner request, 2026-11-14: at a table the GM rules, so in a room a player's step is a request until the GM accepts it. A player sends a request over the room's Realtime channel; the GM's client shows it with its result previewed by the local core, and accepts it (the GM's client calls `append_step`) or declines it with a note. The GM can undo by appending a correcting step, never by editing the log. Ephemeral things go over Realtime broadcast and presence and are never stored: who is online, whose turn it is, a short chat, and pointing at a spot [A]. Whether players may append their own steps without approval is a room setting, off by default [A]. Turn-based only, as APP-17.
- Keys: the request and decision messages; the room's approval setting
- Files: ~3 (supabase/migrations/, docs/backend/rooms.md, docs/guide/gm.md)
- See: docs/issues/open/BE_S10_01_room-log-supabase.md
- Done when: a player's step lands only after the GM accepts it, and a declined one never does; an undo is a new step in the log; presence shows who is in the room; with approval off, a player's step lands directly; the specs pass.
