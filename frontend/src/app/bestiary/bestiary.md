# bestiary.ts

Modified: 2026-11-27

`rw-bestiary` (FE_S08_06), the bestiary at `/play/bestiary`: the browser half of BE_S08_05, over the `catalog`, `bestiary`, and `monster_sheet` commands ([bestiary.md](../../../../docs/backend/bestiary.md)). It adds no rule; what each viewer sees is the `bestiary` command's `view`.

- **The list** is `catalog list` over `monsters`, run when the page opens (`ready`) and on **List**. The form's fields become catalog filters (`catalogFilters`): a creature type is an item of the row's `tags`, a phase its `phase`, a level range its `level` (a bound of 0 is no bound), and the words a `text` filter on the label and description, as the CLI's `type:`, `phase:`, and `level:` words do.
- **The view** sends the listed refs, in order, to `bestiary view` with the marks and the settings box `bestiary.players_see_only_known`. The viewer is the GM (no `player`), `table` (what the whole table was shown, as the CLI's players' view), or a player named in a mark. A `seen` entry shows `?` for its level, phase, and type.
- **Marks.** The GM marks a monster `unknown`, `seen`, or `known` for the table, or for the player typed in the form. With a game in play (`Saves.current`), a mark is a `bestiary` step of its session and is autosaved, so it replays and travels in a backup; the session fills the step's `knowledge`. With none, the page keeps the marks itself until it closes and says so.
- **A Lore check** needs a game in play and a player: a `bestiary` reveal step with a `check` of the session's character (attributes, skills, level) and a d100 roll; a success marks the monster `known` for that player. The page shows the check's target and whether it met it.
- **An entry** that shows in full reads its row with `catalog get` and its stats with `monster_sheet` (attributes, skills, level, role, perks, and maneuvers, as the row has them); a `seen` one shows only its name, description, and size.
- **Words** are `bestiary.*` locale keys; levels, sizes, and phases show as the data writes them.

References: [frontend.md#conventions](../../../../docs/backend/frontend.md#conventions), [bestiary.md](../../../../docs/backend/bestiary.md), [Angular KeyValuePipe](https://angular.dev/api/common/KeyValuePipe). [U]
