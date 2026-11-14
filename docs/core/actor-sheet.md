# Actor sheet

Players and monsters are one shape. Both are actors: attributes, a level, and skills, with
derived stats computed by the same function ([derived.md](derived.md)). A monster row is the
starting sheet of an actor, after [composition](composition.md).

```
sheet(monster row, skills to read)
stats   derived(attributes, level)         the same function a player uses
hp      the row's hitPoints, when it has one, in place of the computed value
skill   the row's skills[key], when written
        else start = skill_start(associatedAbility) = 5 + attribute * 2
             a combat or magic skill: min(skill_cap, start + per_level * (level - 1))
             any other skill: start
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
- **Fighting skills grow with level.** A player buys rating every level; a monster has no
  pool, so its unwritten fighting skills grow by level instead
  (`monster_skill_growth` in `progression/core.json`). A skill whose `category` is in its
  `categories` (`combat`, `magic`: the weapon skills, `evasion`, `block`, the armor skills, and
  every `<color>_magic`) reads `start + per_level × (level − 1)`, capped at `skill_cap` (100),
  as a bought skill is. `per_level` is 17: an average actor's pool (INT 5, LCK 5:
  `5 + 5 × 2 + 2.5` on an average luck roll) spent on one tag skill at 1 point a rating, rounded
  down, so a monster keeps pace with a player's main skill: within 15 points at every level from 1 to 30. [V 2026-11-14, `monster_sheet` vectors in Python and Go, `test_bestiary.py`] Other skills (stealth, social,
  knowledge, and the rest) stay at `start`: a monster does not train its lockpicking by
  fighting. A written skill is the author's value and does not grow. Before this rule a
  monster fought at `start` at every level, about 25 to 35, while a player's main skill
  reaches 100 by level 5 to 8; the M5-18 agent runs lost none of 240 fights
  ([M5-18](../decisions/M5-18.md)). [A]
- **Perks.** A row's `perks` (M3-15) emit their modifiers at their rank, as a player's do
  ([progression.md](../modules/progression.md#perks-and-traits)), and the sheet applies them: each
  derived stat, the authored `hitPoints` included, takes the modifiers on its key, and each skill
  read takes the modifiers on its key after its start and growth. [A]
- **A missing attribute reads 0**, as it does in `derived`.

The input is the flat row, so a consumer composes first. The vectors are in
`ruleset/tests/vectors/monster_sheet.json`.
