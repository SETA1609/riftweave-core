# FE_S08_08_encounter-tracker-page An offline encounter tracker: initiative and whose turn it is (frontend)

Date: 2026-11-27

- **Decision:** the tracker lives at `/play/encounter_tracker` under GM tools. In the web app the entry is no longer pending; it stays pending in the terminal app (`issue_apps: ["cli"]`, as Spell creation does), which S23 retires. It holds no rule: the order and the turns are the `encounter` command's [V 2026-11-27, e2e against the Go command].
- **Combatants:** saved characters (initiative and pools from `derived`, current resources from the stored actor), monsters by key (`catalog get`, then `monster_sheet`), and anyone at the table by the numbers on their sheet, since a table may play characters the app does not hold [A].
- **The roll-off** is made at the start, one d100 each, and kept in the step. On a tie the rolls do not break, the page rolls again for everyone, up to five times; the command decides whether the tie stands [A].
- **Where the fight lives:** with a game in play, in its session as `encounter` steps, autosaved; the session fills a `turn_end`. With none, in the page until it closes, with a notice [A]. Reopening the page does not yet rebuild the combatant list from `session.encounter`, which keeps only the ids [U].
- **Pools:** HP, SP, and MP are numbers the GM edits for the table. Applying damage, AP, and conditions through `combat`, `actions`, and `conditions` per actor is left to the action menu (FE_S08_16) [A].
- **Fixed on the way:** adding a monster cleared the key field when its answer came back, even after the GM had typed the next one; it now clears only the key it added [V 2026-11-27, e2e over HTTP].
