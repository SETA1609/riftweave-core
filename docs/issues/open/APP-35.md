#### APP-35 Online table: a shared battle grid, and the initiative list in sync
- Status: todo
- Priority: low
- Needs: APP-18, APP-34
- Modified: 2026-11-16
- Do: Owner request, 2026-11-16: online is an addition, not a requirement. Rules always resolve on each device's own core (APP-14), offline or not; a room (APP-17, APP-18) adds what a table needs to share when its players are apart. **Battle grid:** a square grid (one cell is one `move` [A]; the unit is in docs/core/actions.md [U]) with a token per character and monster; the GM sets its size and may hide monsters players have not seen (APP-29's marks); a player drags only their own token, and a move is a step the GM approves when the room says so (APP-18). Token positions are session steps, so they replay [A]; pointing at a cell stays an ephemeral broadcast (APP-18). **Initiative and turn:** APP-34's encounter list, current turn, and round are shown to every member from the room's log, and the player whose turn it is is highlighted on every screen. Frontend first; the CLI shows the list and a text grid (coordinates) later [A].
- Keys: grid size and token positions as session steps; the room's encounter view
- Files: ~14 (to split by frontend and backend if over 20)
- See: docs/issues/open/APP-17.md, docs/issues/open/APP-18.md, docs/issues/open/APP-34.md
- Done when: in a room of a GM and four players, each sees the same grid, tokens, initiative order, and current turn; a move appears on every screen once appended; a replay of the log reproduces the grid; the tests pass.
