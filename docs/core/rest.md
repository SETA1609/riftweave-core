# Rest and recovery

Rest is a table of rows, one per occasion, in `ruleset/data/rest/core.json`. Each row lists what comes back for every resource in the actor snapshot (`hp`, `mp`, `stamina`, `ap`, `toxicity`), for conditions, and for food effects. A resource with no recovery says `none` instead of being left out, so a missing rule is a data error and not a silent zero. Fallout and Morrowind both leave recovery to scattered prose; this file is the one place for it. [A]

## Occasions

| key | when | time |
| --- | --- | --- |
| `short_rest` | the party stops to catch its breath | 60 minutes |
| `long_rest` | the party sleeps or camps | 480 minutes |
| `end_of_combat` | the last hostile stops acting | 0 minutes |

`end_of_combat` is not a rest the player chooses. A client applies it once when a fight ends. [A]

## Recovery rows

```
resource   hp | mp | stamina | ap | toxicity | conditions | food
mode       full | fraction | none | clear | empty
```

- `full` sets the resource to its maximum.
- `fraction` adds `floor(max * numerator / denominator)` and stops at the maximum.
- `none` changes nothing. It is a rule, not an omission.
- `empty` is for `toxicity` only. It sets toxicity to 0 ([modifier.md](modifier.md#toxicity)).
- `clear` on `conditions` removes every active condition whose modifier duration is in `durations`. Conditions with any other duration (`permanent`, `while_equipped`) stay.
- `clear` on `food` ends every running food effect: each tick and held modifier whose origin is `core:origin/food`. A potion's or a spell's effect stays.

The maximum of `hp`, `mp`, and `ap` is the derived stat of the same key ([derived.md](derived.md)); a modifier on the derived stat moves the maximum and so the recovery. Stamina has no derived row yet (it stays in `docs/modules/character-creation.md`, action combat only), so the caller supplies its maximum. [A]

## Values

| resource | `short_rest` | `long_rest` | `end_of_combat` |
| --- | --- | --- | --- |
| hp | 1/4 of max | full | none |
| mp | 1/4 of max | full | none |
| stamina | full | full | full |
| ap | full | full | full |
| toxicity | empty | empty | none |
| conditions | none | clear `turns`, `uses` | none |
| food | clear | clear | none |

AP is the one value already recorded elsewhere: a short rest refills the pool ([actions.md](actions.md)). Every other number in the table is a first choice, not a measured one. [U] Playtests may move them; the vectors in `ruleset/tests/vectors/rest.json` change with them.

## Food at rest

A rest ends the meal's buff, so camping is not free: a party that rests and wants the buff back
eats during the rest, the same dish again or another one, and pays a ration for it. The `rest`
function takes the effects running on the actor (`running`, the `ticks` and `modifiers` shapes of
`consume`) and an optional `meal`, a consumable with `delivery` `eat` and `type: food`. It
clears food effects by the occasion's `food` row, then eats the meal by the `consume` rules, so
the meal's effects run from the end of the rest. Anything else named as a meal is refused.
[V 2026-11-12, `rest` vectors]

- **End of combat** keeps food effects (`none`): a fight is not a rest, so a buff from a meal
  eaten before it runs until its duration ends or a rest clears it. Food is never eaten in
  combat (M2-07), so every food effect in a fight comes from a meal eaten before it.
- **At a table** the GM may settle what is left of such a buff at once after the fight, as its
  remaining total, when no other encounter is near, or whenever they choose. This is a table
  shortcut, not a rule the vectors test: the vectors keep the per-second rule.
- **Hunger.** A meal eaten at rest also counts toward the survival module's hunger meter, by that
  module's own table, once the module exists.

## Not covered

- Healing by spell, potion, or food. Those are effects with their own rows.
- Time-based recovery out of combat (per hour of travel, per day of downtime). Out of scope until M4. [U]
- Interrupted rest and ambush rules.
