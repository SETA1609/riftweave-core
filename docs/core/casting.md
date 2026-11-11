# Casting

Author: SETA1609

One cast is one call of the [resolution function](resolution.md) on the spell's `<color>_magic` skill, wrapped in what the cast spends and what it holds. The coefficients are in `ruleset/data/casting/core.json` (`casting.schema.json`), the spells in `ruleset/data/spells/core.json`, the vectors in `ruleset/tests/vectors/casting.json`. Morrowind resolves a spell as one check and then applies its effect list; this does the same. [V 2026-11-06]

## What is core

| Mechanic | Decision |
| --- | --- |
| Components | Core, as reagents only: a spell's `cost.reagents` are consumed by the cast. Verbal and somatic components are cut; the `silenced` flag blocks every cast. |
| Concentration | Core: a spell with `concentration: true` is held. A caster holds one; a new one ends the old. Damage does not break it: there is no second roll. |
| Counterspell | Cut: there is no reaction slot ([actions.md](actions.md)). Dispel covers undoing a spell. |
| Dispel | Core: the `dispel` effect, magnitude N, ends every standing spell on the target of level N or below. |
| Ritual casting | Not core. Casting outside a turn, over minutes, is a module after 1.0. |
| Spell failure from armor | Core, as a penalty: the heaviest armor category worn adds a negative `add` to the cast target. |

## Procedure

```
spell   = the row named by `spell`; skill = <color>_magic
1  prevented   a caster flag in prevented_by          → outcome prevented, nothing spent
2  mana        mana < the paid cost                    → outcome no_mana, nothing spent
3  reagents    a cost.reagents ref not in the caster's → outcome no_reagents, nothing spent
4  spend       mana -= the paid cost; the reagents are consumed
5  check       target = skill + armor_penalty[armor] + modifiers on the skill   (resolution.md)
6  failure     outcome failed: no effects, concentration unchanged
7  success     outcome cast: the spell's effects, as authored
               concentration spell: hold it; the one held before, if another, ends
               dispel N: every standing spell of level <= N is dispelled
```

- The cost is spent before the roll, so a failed cast still costs its mana and reagents, as `magic.md` says. [A]
- A fumble is only a tag on a failure. It costs nothing more. [A]
- The margin is in the result. The effects are the authored magnitudes: scaling a magnitude by the margin is not core, and a consumer that wants it reads the margin. [A]
- The armor penalty is a modifier with source `armor:<category>`, so it stacks by [modifier.md](modifier.md) like any other. No armor means `none`. [A]
- A dispelled spell's modifiers carry the source `spell:<key>`; the consumer removes them through the removed sources of [modifier.md](modifier.md). The cast returns only the keys. [A]
- The AP cost of a cast is the `cast` row in [actions.md](actions.md); this function does not spend AP.

## Delivery and saves

Every offensive spell names a `delivery`, which decides the defense ([combat.md](combat.md)): `single` for one target and `area` for a radius, and both can be dodged or blocked. An `instant` spell reaches the mind, body, or soul with nothing to dodge, so its defense is a `save`, and it must name one: an `attribute` or a `skill`. [A]

```
save target = the skill's value, or attribute x attribute_multiplier       10
              (a missing skill or attribute reads 0)
save        = one resolution on that target, with the defender's LCK and the defender's roll
failed      the save fails or fumbles: the effects land as authored
partial     the save succeeds with margin < full_margin (30):
            each magnitude and duration x partial_percent (50), rounded down, at least 1
full        margin >= full_margin, or the save is a crit: no effects land
```

The cast must succeed first; a failed cast rolls no save. The `target` input carries the defender's `attributes`, `skills`, and `roll`. An instant spell cast with no `target` returns its effects as authored and a null `save`, so a consumer can roll the save itself. Any other delivery ignores `target`. A crafted spell carries no save ([magic.md](../modules/magic.md)). [V 2026-11-11, `casting` vectors] The values are in `casting/core.json` under `save`. [A]

## Paid cost

What the caster pays depends on skill and modifiers, never on level:

```
discount  = min( <color>_magic / divisor, cap )      divisor 200, cap 0.45
base      = ceil( cost.mana × (1 − discount) )
paid      = max( minimum_paid, base with the modifiers on spell_cost applied )   minimum_paid 1
```

- The discount reads the caster's own skill, before the armor penalty and the modifiers on the skill. [A]
- The modifiers on `spell_cost` (a perk, a robe) stack by [modifier.md](modifier.md), so a `mul` floors. [A]
- The cast and craft thresholds read the cost before the discount ([magic.md](../modules/magic.md#the-casters-side)).

## Result

```
outcome        cast | failed | prevented | no_mana | no_reagents
mana           the caster's mana after the cast, less the paid cost
consumed       the reagent refs the cast used up
check          { target, success, margin, tags }, or null when nothing was rolled
effects        the spell's applied effects on a cast, else []
concentration  the spell key the caster holds after the cast, or null
ended          the concentration spell the cast ended, or null
dispelled      the standing spell keys the cast ended, in input order
save           { target, success, margin, tags, result }, or null when no save was rolled
```
