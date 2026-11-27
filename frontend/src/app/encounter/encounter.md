# encounter.ts

Modified: 2026-11-27

`rw-encounter` (FE_S08_08), the encounter tracker at `/play/encounter_tracker` (GM tools): the browser half of BE_S08_07, over the `encounter` command ([sessions.md#encounters](../../../../docs/backend/sessions.md#encounters)). It adds no rule; the order, whose turn it is, and the round are the command's.

- **Combatants** come three ways. A saved character (the character list's store) takes its initiative and pools from `derived` and its current resources from its stored actor. A monster, by key or ref, takes its row from `catalog get` and its initiative and pools from `monster_sheet`. Anyone else at the table is added by the initiative, AGI, and LCK on their sheet. Each id is unique in the fight (`uniqueId`: a second Wolf is `Wolf 2`).
- **Initiative** is the derived value, never a roll (docs/core/derived.md). AGI and LCK are the attributes as the row or the actor holds them.
- **Start** sends every combatant with one d100 for the roll-off, kept in the step so the order replays. A tie the rolls do not break is refused by the command; the page rolls once more for everyone and tries again, up to `ROLL_OFF_TRIES` (5) times.
- **The fight** shows the round (0 is the surprise round, in which only ambushers act), the initiative order with the current actor and those who acted, whose turn it is and who is next, and the turn's `start`: `surprise`, or `combat` on an actor's first turn, which the `actions` command reads for the partial AP pool. **End turn** is `turn_end`.
- **The game in play.** With one (`Saves.current`), start and every end of turn are `encounter` steps of its session, autosaved, so the fight replays and exports with it; the session fills a `turn_end` with its state. With none, the page runs the command alone and keeps the fight until it closes, and says so.
- **HP, SP, and MP** are numbers the GM keeps for the table. Damage and conditions stay the existing commands; the page does not apply them.
- **Words** are `encounter.*` locale keys; the attribute and pool abbreviations have keys too, so a language may keep or change them.

References: [frontend.md#conventions](../../../../docs/backend/frontend.md#conventions), [derived.md#initiative-ties](../../../../docs/core/derived.md#initiative-ties), [actions.md#pool](../../../../docs/core/actions.md#pool).
