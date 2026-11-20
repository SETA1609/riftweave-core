# Character creation

Modified: 2026-11-21

Character creation is one command, `creation`, so a form in either app holds no rule
(BE_S07_08). Its contract is [`backend/contract/creation.json`](../../backend/contract/creation.json);
the draft it reads is `creation_draft` in [`common.json`](../../backend/contract/common.json).
The Python reference is `ruleset/scripts/creation_reference.py` and the Go core
`backend/internal/creation`; both run the vectors in `ruleset/tests/vectors/creation.json`.
[V 2026-11-21]

## Operations

| Operation | Input | Result |
| --- | --- | --- |
| `rules` | none | `creation` (progression's numbers and `tag_skills`), `abilities`, `races` (every race but templates, with its ability modifiers summed with its parent's and its traits), `backgrounds`, `creation_perks`, `skills` |
| `check` | `draft` | `complete`, `problems` (`step`, `message`, in step order), `points_left`, `attributes`, `skills`, `derived` |
| `finish` | `draft` | `name` and `actor`, a [stored actor](../core/state.md#the-stored-actor), or a rule error that lists the check's messages |

## The draft

Any steps, in any order: `name`, `race`, `points`, `background`, `creation_perks`, and
`tag_skills`. `points` is the points spent per attribute, not a total, so a form only adds and
removes: an attribute is `attribute_start` (4) plus its points, from `attribute_min` (1) to
`attribute_max` (10), and the points spent may not pass `attribute_points` (10). A negative
entry lowers an attribute under the start and frees points (GD_S07_07). All the numbers are
progression's `creation` block ([progression.md](../core/progression.md#character-creation)).

The steps a check names, in order:

| Step | A problem when |
| --- | --- |
| `name` | it is missing or blank |
| `race` | it is missing, not defined, or a template |
| `attributes` | an attribute is over the maximum or under the minimum; more points are spent than there are; or points are left |
| `background` | it is missing or not defined |
| `creation_perks` | one is not defined, not a creation perk, or chosen twice; or there are more than `creation_perks_max` |
| `tag_skills` | one is not defined or chosen twice; or there are not exactly `tag_skills` |

The limits read the starting modifiers: a race trait or a creation perk that `add`s to
`tag_skills` or `creation_perks` raises it, so a human picks 4 tag skills and up to 3
creation perks. [V 2026-11-21, `creation` vectors]

## What a check shows, and what a finish stores

`attributes` is the sheet: the points, the race's ability modifiers, then the race traits'
and creation perks' modifiers at rank 1. `skills` is each skill's start from those attributes
plus the background's bonuses. `derived` is the derived stats at the start level with the same
modifiers.

The stored actor keeps the attributes before the perks' modifiers, because it stores the race
traits and creation perks in `perks` (rank 1) and a session emits them again
([sessions.md](sessions.md)). Its `skills` are the check's, `tags` the tag skill keys, and its
`resources` the full pools of `derived`. Spells, equipment, and spending the level-1 skill
points are not steps yet. [A]

## The human CLI

Start, then Character creation, opens a screen with one row per step. Left and right pick a
race or a background, + and − add or remove a point, the perk and tag skill rows take typed
keys (`small_frame`, `survival stealth bows`), and Finish sends the draft to `finish`. Every
change runs `check`, so the points left, the sheet, and the problems are always the core's.
The created character is the one a dev session and the attack form start from.
