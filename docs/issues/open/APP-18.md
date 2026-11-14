#### APP-18 Rooms: the GM approves steps, and the table sees who is there
- Status: todo
- Priority: low
- Needs: APP-17
- Do: Owner request, 2026-11-14: at a table the GM rules, so in a room a player's step is a request until the GM accepts it. A player sends a request over the room's Realtime channel; the GM's client shows it with its result previewed by the local core, and accepts it (the GM's client calls `append_step`) or declines it with a note. The GM can undo by appending a correcting step, never by editing the log. Ephemeral things go over Realtime broadcast and presence and are never stored: who is online, whose turn it is, a short chat, and pointing at a spot [A]. Whether players may append their own steps without approval is a room setting, off by default [A]. Turn-based only, as APP-17.
- Keys: the request and decision messages; the room's approval setting
- Files: ~6 (frontend/src/app/room.ts, frontend/src/app/table.ts, frontend/src/app/table.spec.ts, supabase/migrations/, docs/backend/rooms.md, docs/guide/gm.md)
- See: docs/issues/open/APP-17.md
- Done when: a player's step lands only after the GM accepts it, and a declined one never does; an undo is a new step in the log; presence shows who is in the room; with approval off, a player's step lands directly; the specs pass.
