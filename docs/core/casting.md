# Casting

Author: SETA1609
Modified: 2026-11-19

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
               summon block: its creatures join caster.summons; past the limit, the oldest are dismissed
               minion block: count creatures join caster.minions, with no limit
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

## Damage to a target

A cast that lands its damage on a creature runs the combat pipeline's steps after the attack
roll ([combat.md](combat.md#order)), so a fire bolt and a sword cut lose HP the same way. It
happens only when `target` carries `hp`; without it the effects come back as authored, for a
consumer that applies them itself.

```
packet    each effect core:effect/damage_<type> whose <type> is a damage type
          (combat/damage_types.json): amount = its magnitude, after the save for an instant spell
attack    the cast check stands in for the attack roll: the same target, roll, and LCK, so the
          same margin, crit, and fumble (a failed cast deals nothing)
steps     combat.md 2 to 8 on target: the defense bands its delivery meets (single and area:
          the evasion band and a blocking stance's band; instant: none), crit, area evade,
          armor and shield DR where the type takes it, resist, then hp
order     the packets in the spell's order; each reads the hp the last one left
result    damage: per packet { effect, type, amount, damage, evaded, blocked }; hp after all
```

- No spell damage type takes armor DR today: every one is tagged `bypasses_armor` or
  `bypasses_physical_armor`, so worn armor never reduces spell damage, as in combat. [V
  2026-11-15, `damage_types.json`]
- A crit cast doubles its damage, the combat crit, since an offensive spell is an attack
  ([combat.md](combat.md)). [A]
- An effect with a `duration` over one second is damage over time and is not applied by the
  cast; no damaging spell has one. [A]
- Orange, green, and violet have no damaging spell. That is intended: orange's attack is its
  summons, green heals and grows, and violet charms and deceives, so a caster of only those
  colors fights through creatures, allies, or control. [A]

## The world target

`world` reaches a place, not a creature: somewhere the caster is sent or something is kept,
anywhere in the world, with no line of sight and no range. Its cost multiplier is 3.0
(`spell_cost.target_multiplier`). Recall (to the caster's mark) and Divine Intervention (to the
nearest temple) use it. Mark and Pocket Storage are `self`: Mark records where the caster
stands, and the pocket opens at hand. Resolving a temple, a mark, or a pocket is engine state
(M4 and the backend), not this function. A crafted spell cannot use `world`: every effect that
reaches it is uncraftable. [A]

## Spellcraft

`spellcraft` checks spell knowledge ([magic.md](../modules/magic.md#spell-knowledge)), by `operation`:

```
craft     caster.known (effect refs), caster.skills, spell { color, effects }
          per effect, the first that fails:
            channel             the effect does not list "spell"
            uncraftable         craftable is false
            excluded_parameter  its parameter is in uncraftable_parameters
            unknown             the ref is not in caster.known
            off_color           the effect has a color other than the spell's
          mana             spell cost of the effects (magic.md#cost-formula), or null if any failed
          cast_threshold   round(mana); craft_threshold round(mana x 1.25), a half rounding up
          skill_ok         <color>_magic >= craft_threshold
          valid            no effect failed and skill_ok
learn     caster.known, spell (a key)
          taught = the spell's craftable effects not known yet, in spell order
          known  = caster.known + taught, sorted
research  caster.known, caster.skills, caster.attributes.lck, effect (a ref), optional color,
          modifiers, roll
          refused  uncraftable (as craft), known, no_color, or skill (skill < the effect's
                   craft threshold at magnitude 1, instant, touch); nothing rolled
          check    one resolution on <color>_magic with the modifiers on it (resolution.md)
          learned  the check succeeds: known gains the effect; failed otherwise
          outcome  learned, failed, or refused (with reason); known after it
          days, cost  research.days and research.cost once rolled, spent whatever the roll;
                   0 and null when refused
```

The color is the effect's own, or `color` when the effect has none. The vectors are in
`ruleset/tests/vectors/spellcraft.json`. [A]

## Paid cost

What the caster pays depends on skill and modifiers, never on level:

```
discount  = min( <color>_magic / divisor, cap )      divisor 200, cap 0.45
base      = ceil( cost.mana × (1 − discount) )
paid      = max( minimum_paid, base with the modifiers on spell_cost applied )   minimum_paid 1
```

- The discount reads the caster's own skill, before the armor penalty and the modifiers on the skill. [A]
- The modifiers on `spell_cost` (a perk, a robe) stack by [modifier.md](modifier.md), so a `multiply` floors. [A]
- The cast and craft thresholds read the cost before the discount ([magic.md](../modules/magic.md#the-casters-side)).

## Summons

A summon spell writes a `summon` block, not a `core:effect/summon` entry in `effects` (owner,
2026-11-14, M3-25): `{ "group": "core:summon/<key>", "duration": <seconds> }`. Summon spells are
custom rows only a GM or dev adds, so they may leave the shared spell shape; `appliedEffect` does
not change, so potions and enchantments are untouched. A spell whose only effect was the summon
has `effects: []`. The block is priced as the summon effect it replaces: `spell_cost` reads it as
`core:effect/summon` with its group, its count (1 for a single summon, the mean of a minion count),
its duration, and target `ranged` ([magic.md](../modules/magic.md#cost-formula)). [A]

The caster's active summons are an input, `caster.summons`, oldest first: one entry per
creature, the summon group ref it came from (`core:summon/elemental`). A cast of a spell with a
`summon` block adds one entry of its group. A caster holds at most
`summon_limit` (`casting/core.json`, 1) plus the `summon_limit` modifiers on it, stacked by
[modifier.md](modifier.md); past the limit the oldest are dismissed, as a new concentration spell
ends the old one. Perks raise the limit: two summons from level 14 and three from level 28
([magic.md](../modules/magic.md#summon-groups)). A failed cast summons and dismisses nothing.
[V 2026-11-14, `casting` vectors in Python and Go: no perk, Retinue, Retinue and Host, the oldest dismissed, a failed cast]

### Minions

Minions are scapegoats: many weak bodies that soak attacks and spread a foe's actions, not
creatures that grow (owner, 2026-11-14). They stay in core for now; later they move with
multiple summons to an optional module, so no other rule here depends on this section, and the
move is the block's minion part, the `minion` template, and the two minion spells.

- **The block's minion part.** `"minion": true` and `count`: a number, or a dice expression such
  as `1d6`. A dice count is rolled when the spell is cast: the caller passes the rolled total as
  `count_roll`, as every die is an input; a dice count without `count_roll`, or with a total its
  dice cannot make, is an error. `count` is allowed only with `minion`. [A]
- **They do not count toward the summon limit.** The cast adds `count` entries of the group to
  `caster.minions` (an input, oldest first, default empty) and dismisses nothing; `caster.summons`
  is untouched. [A]
- **No summoning bonuses.** A minion is its group's row at the caster's level, resolved as a
  derived monster, then the `minion` template ([monsters.md](../modules/monsters.md#minions)): no
  power or nature, no Commanding Summons bonus, no summoner archetype perk. [A]
- **Mana is set for the mean count,** so the cost rule does not change: `1d3+1` is priced as 3
  creatures, the 108 mana the three-creature spells cost before, which a level 22 caster's pool
  holds (a `1d6`, priced as 3.5, would cost 126 and does not). A count from the cast's margin (`1 + ⌊margin / 20⌋`, capped) is a later variant. [A]

Raise Horde (level 22) and Conjured Legion (level 28) are the minion spells, composed from Raise
Thrall and Summon Elemental ([composition.md](composition.md)): each overrides `level`,
`summon.count` (`1d3+1`: two to four), `summon.minion`, `concentration`, and the cost, and keeps its base's group
and duration. [V 2026-11-16, `casting` vectors in Python and Go]

## Bound weapons

A cast whose spell has a `core:effect/bound_weapon` effect also returns the conjured weapon
([magic.md](../modules/magic.md#bound-weapons)):

```
weapon       the effect's parameter: the base it copies (core:weapon/<key>)
skill        the weapon's skill
dice         the weapon's damage dice
material     casting.bound_weapon.material, or the material unlocked (a core:material ref
             in its unlocks) by the held perk with the highest prerequisite level, then
             the higher weapon.attack; the order of caster.perks does not matter
attack       that material's weapon.attack
damage_type  casting.bound_weapon.damage_type (magic), or the type of the coating or the
             enchantment the input names (core:effect/damage_<type>; the enchantment wins)
phase        null, or that coating's or enchantment's phase
effects      every held perk's bound_weapon_effects, in perk order; a parameter-less
             lingering_damage takes the current damage_type
```

The caster's perks are `caster.perks` (feature refs). The coating and the enchantment are the
input's `coating` and `enchantment` (effect refs); one without the matching channel, or that
is not a `damage_<type>` effect, is an error. The weapon's attacks pass `damage_type` as
`damage.type` and `effects` as `attacker.weapon_effects` to the combat pipeline
([combat.md](combat.md#weapon-effects)). Without a bound weapon effect the field is left out.

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
summons        the caster's active summons after the cast, oldest first
dismissed      the summons the cast sent away to stay within the limit, oldest first
minions        the caster's minions after the cast, oldest first
save           { target, success, margin, tags, result }, or null when no save was rolled
bound_weapon   the conjured weapon (Bound weapons above); left out when the spell conjures none
```
