# Actor sheet

Players and monsters are one shape. Both are actors: attributes, a level, and skills, with
derived stats computed by the same function ([derived.md](derived.md)). A monster row is the
starting sheet of an actor, after [composition](composition.md).

```
sheet(monster row, skills to read)
stats   derived(attributes, level)         the same function a player uses
hp      the row's hitPoints, when it has one, in place of the computed value
skill   the row's skills[key], when written
        else skill_start(associatedAbility) = 5 + attribute * 2
```

- **Hit points.** `hitPoints` on a monster is an optional authored value that replaces the
  computed one. Rows written before this rule keep theirs, because computed HP is far above
  the balance bands: level 3 with END 4 computes 59, where `grove_stalker` is written as 30. The example rows of [monsters.md](../modules/monsters.md#generated-monsters) use the computed value.
  [V 2026-11-10]
- **Mana and stamina.** Computed, like every other derived stat. No monster row stores them. [V 2026-11-10]
- **A reduced skill sheet.** A row writes only the skills it fights or casts with. Any other
  skill reads as a player's starting value, from `skill_start` in `progression/core.json`.
  A player's sheet follows the same rule: it writes only the skills the character has raised
  ([progression.md](progression.md#track-only-what-you-raise)), or, on the grouped variant,
  its groups' ranks.
- **A missing attribute reads 0**, as it does in `derived`.

The input is the flat row, so a consumer composes first. The vectors are in
`ruleset/tests/vectors/monster_sheet.json`.
