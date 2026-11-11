# Modifier

Author: SETA1609

One payload. Conditions, features, and applied effects emit this row. They do not invent a second shape.

```
target     a key naming one row: attribute, skill, derived stat, action, condition, damage type, or modifier target (see Target)
op         add | mul | set | grant | resist
value      number
source     perk:quick, condition:blinded
origin     core:origin/alchemy (optional; the kind of source, see below)
duration   turns | uses | while_equipped | while_applied | permanent
turns      how many turns a `turns` modifier lasts
stack      none | add
```

The stacking rule is written once, in [contract.md](contract.md); this file fixes its order so a second implementation gets the same number. The `modifiers` and `conditions` vectors pin every step. [V 2026-11-04]

## Origin

The source says *which* thing made a modifier (`effect:fortify_strength`). The origin says *what kind* of thing it is (`core:origin/alchemy`). Origins are rows in `ruleset/data/origins/core.json`, and each row has a stacking policy:

- **`highest`**: modifiers of this origin on one target and op do not stack with each other. The strongest bonus counts, and the strongest penalty counts. A Potion of Fortify Strength (+10) and a Potion of Giant Strength (+5) give +10, in any order.
- **`add`**: modifiers of this origin stack, one per source, as if they had no origin. Two perks, two armor pieces, or two conditions add.
- **`replace`**: a new source of this origin removes every running modifier and tick of the origin first, even effects the new source lacks. Fries, then a hot dog: only the hot dog counts.

Different origins always stack: a Fortify Strength potion (alchemy, +10) and a Fortify Strength spell (arcane, +10) give +20, though both carry the same effect.

| Origin | Policy | What carries it | Default for |
| --- | --- | --- | --- |
| `arcane` | highest | spells of the nine color schools | spell |
| `divine` | highest | prayers, blessings, shrines, relics | none: a row names it |
| `primal` | highest | nature and spirit magic, totems, shamanic rites | none: a row names it |
| `alchemy` | highest | potions, elixirs, poisons, oils, raw ingredients | consumable, ingredient |
| `enchantment` | highest | magic effects on worn or wielded items, gems | an applied effect on an item, gem |
| `martial` | highest | stances, battle cries, combat techniques | none: a row names it |
| `morale` | highest | inspiration, leadership, fear, despair | none: a row names it |
| `environment` | highest | terrain, weather, light, cover | none: a row names it |
| `equipment` | add | an item's own numbers: armor DR, weapon damage | weapon, armor, accessory stats |
| `trained` | add | perks, features, backgrounds | feature, background |
| `innate` | add | ancestry and natural traits | race, monster |
| `condition` | add | status conditions | condition |
| `food` | replace | dishes and rations | food |

The emitter stamps the origin when it turns a row into modifiers: the row's own `origin` if it names one (an `appliedEffect` may), else the default for its category. A modifier with no origin stacks by source alone, the same as an `add` origin. [A] The White school is not mapped to `divine`; a white spell that should be divine names the origin itself. [U]

## Effects over time

An applied effect's `duration` is seconds, and 1 (the default) is instant. One effect serves both: there is no separate instant and over-time effect. [V 2026-11-07, owner decision, M2-05]

- A **pool effect** (`restore_resource`) applies `magnitude` at once when instant. Over time it adds `magnitude` for each second that passes, so its total is magnitude times duration. A pool stays between 0 and its maximum.
- Any **other effect** holds its `magnitude` for `duration` seconds as a modifier: `resist` on its parameter, `add` on its parameter or its own key, and a `grant` when it has no magnitude.
- **Turns.** One turn is `seconds_per_turn` seconds (5, `actions/core.json`). A turn advances effects by that many seconds, so a tick covers magnitude times the seconds that remain, and the total is the same in both modes. [V 2026-11-11, `consume` vectors]
- **Same source.** Using the same item again replaces its own running effects, so the effect restarts and does not stack. [A]
- **What can be used.** Only an item with `delivery` `drink` or `eat`. Food (`type: food`) is refused in combat, and it is always over time. A dish is one source, `food:<key>`; any other item is `item:<key>`. A dish's effects are the combined effects of its ingredients, written on the row; cooking is crafting.

### Toxicity

Drinking adds the item's `toxicity` to a stored `toxicity` on the actor. It falls by
`toxicity_decay.per_second` (1, `rest/core.json`) for each second that passes, by the same tick
as a pool effect, and never below 0. A long rest clears it. The limit is the derived row
`toxicity_limit`, `20 + END × 5` ([derived.md](derived.md)), passed in as `max.toxicity`. A drink
that takes toxicity above the limit is not refused: its effects land, and the actor gains
`core:condition/overdosed` while toxicity stays above the limit. Food adds none and does not
speed the fall. A real-time client and a turn-based table read the same numbers, since a turn
is 5 seconds of decay. [A]

`burn` and `poison_health_drain` stay effects that run under this rule, with `duration` as their length. They are not conditions; the `burning` condition is its own row. [A]

## Target

A modifier's `target` is a key, and the key names one row. Where the row lives says what the target is:

| Row in | Kind | What a modifier changes |
| --- | --- | --- |
| `abilities/core.json` | number | an attribute (`str`) |
| `skills/core.json` | number | a skill (`persuasion`) |
| `derived/core.json` | number | a derived stat (`hp`, `ap`) |
| `actions/core.json` | number | the action's AP cost (`dodge`, [actions.md](actions.md)) |
| `conditions/core.json` | flag | the condition's flag (`prone`) |
| `combat/damage_types.json` | damage type | a packet of that type, for `resist` (`fire`) |
| `modifier_targets/core.json` | its `kind` | anything else: a rule's number (`crit_window`, `loot_roll`) or a flag (`darkvision`) |

- **The op fits the kind.** `grant` names a flag. `add`, `mul`, and `set` name a number. `resist` names a damage type. [V 2026-11-10]
- **One row per target.** A `modifier_targets` row may not reuse a key from the other tables; it would hide that row. A key in two of the other tables is ambiguous for an op that both fit: `block` is a skill and an action, so a numeric modifier on it is refused until one is renamed (M5-37). `piercing` is a skill and a damage type, which is not ambiguous: `add` reads the skill and `resist` the damage type. [V 2026-11-10]
- **A modifier_targets row is used.** A row no modifier names is dead and fails, as an unreferenced effect does. A module adds its own rows in its own `modifier_targets` file, for the targets only it names (a faction's membership flags).
- **`pending`.** A row whose target no golden vector applies a modifier to is marked `pending: true`: data may name it, but no implementation is yet held to a rule for it. The flag is checked against `ruleset/tests/vectors/` and each module's `tests/` in both directions, so it cannot go stale: a pending target that a vector uses fails, and so does a target without the flag that no vector uses. [V 2026-11-10]
- **What this does not settle.** The table records what the data names today, including names that break other rules. The target `attack` is the attack action's AP cost here, but `combat/situations.json` uses it for a bonus on the attack roll; M5-37 splits them. [V 2026-11-10]

## Order for one numeric target

```
1. expire     drop a modifier whose source was removed,
              and a `turns` modifier whose turns have elapsed (elapsed >= turns)
2. dedupe     same origin, source, target, and op: the strongest wins
3. origin     for a `highest` origin, same origin, target, and op: the strongest bonus
              and the strongest penalty win
4. add        value = value + sum(add)
5. mul        value = floor(value * product(mul))
6. set        if any set: value = the latest set in list order
```

- **Strongest.** For `add` and `resist`, a bonus is a value above 0 and the strongest is the largest; a penalty is below 0 and the strongest is the most negative. For `mul`, a bonus is above 1 and a penalty below 1, the same way. One bonus and one penalty from the same group both count: an alchemy +10 and an alchemy −3 on one target give +7. A ±0 `add` or a ×1 `mul` changes nothing and is dropped. Ties keep the first in list order. `set` is not ranked: the latest wins. [A] Bonuses and penalties are ranked apart so a poison cannot cancel a potion of the same origin.
- **Same source does not stack.** Two `add` from `perk:tough` on `hp` count once; the larger one wins, in any order. A source that touches two targets, or uses two ops on one target, keeps each: the key is origin, source, target, and op. [V 2026-11-06]
- **Different sources add** unless they share a `highest` origin, and two `mul` from different sources multiply. [V 2026-11-06]
- **`mul` after `add`**, whatever the list order. The product is floored, so `mul 1.5` on 5 is 7. [A] Numbers in the snapshot are integers; the floor keeps them so.
- **`set` is last.** It replaces the value whatever the `add` and `mul`, so a paralyzed actor with an AP bonus still has 0 AP. Two `set` on one target: the latest in list order (load order) wins. [A] The contract says `set` replaces the value "for that source's duration"; expiry (step 1) is what ends it.
- **`grant`** adds the flag named by `target`. It adds no number, and its `value` is ignored. A flag stays while any live source grants it: removing one of two sources that grant `prone` leaves the actor prone. [V 2026-11-04]
- **`resist`** subtracts from a packet whose tag equals `target`: `amount = max(amount - sum(resist), 0)`. Same-source, origin, and different-source rules are the ones above: two enchanted rings of fire resistance give the larger one. It is the same op the combat pipeline's resist step reads ([combat.md](combat.md)). [V 2026-11-04]
- **Expiry.** `permanent` never expires. `while_applied` and `while_equipped` end when their source is removed. `turns` ends when the elapsed turns reach `turns`. `uses` is counted by the caller, who removes the source when the uses run out; the vectors do not cover it. [A]

## Conditions

A condition is a row in `ruleset/data/conditions/core.json`: a named bundle of modifiers, all with source `condition:<key>`, plus `appliedBy` and `removedBy` effect lists.

1. **An effect event** applies every condition whose `appliedBy` lists it, and removes every condition whose `removedBy` lists it. One effect may do several: `blight` applies `exhaustion` and `diseased`; `cure` removes every condition that lists it.
2. **Apply** adds the condition, or one stack if it is already active and has `maxStacks`. Stacks never pass `maxStacks`. A condition without `maxStacks` applied twice is still one condition: same source does not stack. [V 2026-11-04]
3. **Remove** drops the condition and every modifier with its source, all stacks at once. [A]
4. **A remove event** (`remove: <key>`) ends a condition with no removing effect: `prone` ends when the actor stands up, an action the caller resolves. [A]
5. The actor's values are then the base values with every live condition's modifiers, in the order above.

The `combat` and `effects` fields on a condition row (defense bonus, TTRPG and video-game prose) are not modifiers yet; they are read by nobody and are left to the issues that turn them into rows (M1-01, M3-03). [U]
