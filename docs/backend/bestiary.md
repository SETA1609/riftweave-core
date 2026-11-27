# Bestiary

Modified: 2026-11-27

The bestiary lists the monsters of core and the enabled modules, and the GM decides what the
players know of each (BE_S08_05, owner request 2026-11-16). The `bestiary` command holds the
knowledge rule; listing and filtering the monsters is the `catalog` command's.

## Knowledge

A **mark** is a monster at a level, `unknown`, `seen`, or `known`, for the whole table (no
`player`) or for one player. What a player knows of a monster is the higher of the table's mark
and their own; with no mark it is `unknown`.

| `operation` | Input | Result |
| --- | --- | --- |
| `reveal` | `knowledge`, `monster`, and a `level`, or a Lore `check` (`actor`, `roll`, optional `modifiers`) by a `player`; optional `player` | `knowledge` after it, the `level` the player (or the table) now has, and the `check` |
| `view` | `knowledge`, `monsters` (refs, in order), optional `player` (absent: the GM), `players_see_only_known` | `entries`: `monster`, `label`, `description`, `size`, `knowledge`, `shows` (`seen` or `full`), and for the GM each player's mark in `players` |

- **A reveal** replaces the holder's earlier mark of that monster in place, or adds one at the
  end, so a replay gives the same list. A player's lower mark never hides what the table knows.
- **A Lore check** is `resolution` with the `lore` skill: the player's `actor`, the `roll`, and
  any `modifiers`. It meets the target, and the monster is `known` for that player; it fails,
  and nothing changes. The `check` result is the resolution result. A check needs a `player`. [A]
- **A view** keeps the order of `monsters` and reads each row through `catalog`, so module
  monsters show as core ones do. The box off (solo play, the default), a player sees every entry
  in full. The box on, a player sees a `seen` monster as its name, description, and size, a
  `known` one in full, and nothing of an `unknown` one. The GM always sees every entry in full,
  with the table's mark and each player's in `players`.
- **The sheet** of a `full` entry is not in the view: the caller reads the row with `catalog`
  `get` and the stats with `monster_sheet`, as the Compendium does.

[V 2026-11-23, `bestiary` vectors in Python and Go]

## Campaign state

The marks are campaign state, not settings: a session keeps them in `bestiary.knowledge`
([sessions.md](sessions.md#campaign-state)). A `bestiary` step is filled with the session's
marks, a `reveal` writes them back, and a `replay` rebuilds them from the log, so they travel in
an exported session and replay the same. A `view` step changes nothing. The settings box
`bestiary.players_see_only_known` ([settings.md](settings.md)) only says what a player's view
shows. [V 2026-11-23, `TestARevealStepKeepsTheMarksAndReplays`]

## The screen

The human CLI's Bestiary (Start, then Bestiary) lists every monster of core and the enabled
modules in the GM's view: level, size, phase, creature type (the first tag), and the table's
mark.

| Key | Does |
| --- | --- |
| `/` | filter: words in the name or description, and `type:<type>`, `level:<n>` or `level:<n>-<m>`, `phase:<phase>`, which become `catalog` filters |
| `s`, `k`, `u` | mark the monster under the cursor `seen`, `known`, or `unknown` for the table, as a `bestiary` step in the CLI's session |
| `p` | the players' view: what the whole table was shown, as the box says; a `seen` monster shows `?` for its level, phase, and type |
| Enter | the entry: description, size, the mark, and the full row when it shows in full |
| Esc | clears the filter, then goes back to Start |

The screen marks for the whole table. A player's own mark and a Lore check go through the
`bestiary` command or a session step; the browser page (FE_S08_06) and a shared room (APP-17)
add them for each player. [A] [V 2026-11-23, `bestiary_test.go`]

## The page

The web app's Bestiary (Start, then Bestiary, `/play/bestiary`; FE_S08_06) lists the same
monsters over the same commands, with a form for the filters (creature type, phase, a level
range, and words), a **view as** choice (the GM, `table`, or a player named in a mark), and the
settings box. The GM marks a monster for the table, or for the player typed in the form, and
with a game in play runs a Lore check for that player with the game's character. With a game in
play every mark is a `bestiary` step of its session and is autosaved, so it replays; with none,
the marks last until the page closes, and the page says so. An entry that shows in full opens
with its row and its `monster_sheet` stats. [V 2026-11-27, `app.spec.ts` and
`e2e/bestiary.spec.ts` over both transports]
