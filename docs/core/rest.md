# Rest and recovery

Rest is a table of rows, one per occasion, in `ruleset/data/rest/core.json`. Each row lists what comes back for every resource in the actor snapshot (`hp`, `mp`, `stamina`, `ap`) and for conditions. A resource with no recovery says `none` instead of being left out, so a missing rule is a data error and not a silent zero. Fallout and Morrowind both leave recovery to scattered prose; this file is the one place for it. [A]

## Occasions

| key | when | time |
| --- | --- | --- |
| `short_rest` | the party stops to catch its breath | 60 minutes |
| `long_rest` | the party sleeps or camps | 480 minutes |
| `end_of_combat` | the last hostile stops acting | 0 minutes |

`end_of_combat` is not a rest the player chooses. A client applies it once when a fight ends. [A]

## Recovery rows

```
resource   hp | mp | stamina | ap | conditions
mode       full | fraction | none | clear
```

- `full` sets the resource to its maximum.
- `fraction` adds `floor(max * numerator / denominator)` and stops at the maximum.
- `none` changes nothing. It is a rule, not an omission.
- `clear` is for `conditions` only. It removes every active condition whose modifier duration is in `durations`. Conditions with any other duration (`permanent`, `while_equipped`) stay.

The maximum of `hp`, `mp`, and `ap` is the derived stat of the same key ([derived.md](derived.md)); a modifier on the derived stat moves the maximum and so the recovery. Stamina has no derived row yet (it stays in `docs/modules/character-creation.md`, action combat only), so the caller supplies its maximum. [A]

## Values

| resource | `short_rest` | `long_rest` | `end_of_combat` |
| --- | --- | --- | --- |
| hp | 1/4 of max | full | none |
| mp | 1/4 of max | full | none |
| stamina | full | full | full |
| ap | full | full | full |
| conditions | none | clear `turns`, `uses` | none |

AP is the one value already recorded elsewhere: a short rest refills the pool ([actions.md](actions.md)). Every other number in the table is a first choice, not a measured one. [U] Playtests may move them; the vectors in `ruleset/tests/vectors/rest.json` change with them.

## Not covered

- Healing by spell, potion, or food. Those are effects with their own rows.
- Time-based recovery out of combat (per hour of travel, per day of downtime). Out of scope until M4. [U]
- Interrupted rest and ambush rules.
