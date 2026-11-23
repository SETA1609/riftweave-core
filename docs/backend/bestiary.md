# Bestiary

Modified: 2026-11-23

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
