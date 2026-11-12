# Equipment

What a character can buy, wear, carry, and use up. The rows are in
`ruleset/data/equipment/` (`equipment.schema.json`): weapons, armor, and accessories as a
base in `bases/` plus overrides, consumables and gear as full rows. A base with
`is_item: true` is its own item, so a row is written only where it changes the base
([composition.md](../core/composition.md#a-base-that-is-an-item)). Weapons are in
[weapons.md](weapons.md), armor in [armor.md](armor.md). Every `weight` is in
[weight units](../core/contract.md#weight), 0.5 kg each when read aloud.

## The floor

`ruleset/scripts/test_equipment.py` checks it. [V 2026-11-06]

- **Weapons:** each of the eight weapon skills (`blades`, `blunt`, `piercing`, `unarmed`,
  `bows`, `crossbows`, `guns`, `throwing_weapons`) has at least three weapons, in at least
  two price tiers. A tier is a price band: tier 1 under 5 gp, tier 2 under 25 gp, tier 3
  from 25 gp. Quality tiers (`tiers/core.json`) are a crafting grade on top of this, not
  a second band. [A]
- **Accessories:** at least 12. An accessory's effects are applied effects with the
  `enchantment` channel, so they stack under the `enchantment` origin: two rings of fire
  ward give the larger one.
- **Consumables:** at least 20 potions, poisons, coatings, oils, and food.
- **Gear:** adventuring gear and tools exist as rows (`equipment/gear.json`, refs
  `core:gear/<key>`).

## Consumables

`magnitude` is how much, `duration` is how long, in seconds. Duration 1 is instant and is
the default when omitted: a Potion of Healing restores 40 HP at once. Magnitude is always
per second, so an effect that moves a pool (`restore_resource`, `damage_health`,
`damage_stamina`) totals magnitude times duration: Hearty Stew restores 1 stamina a second
for 60 seconds, 60 in all. An effect that changes a stat (`fortify_attribute`, `resist`)
holds its magnitude for the duration. There is no separate instant and over-time effect:
the same effect takes either duration. [V 2026-11-07, owner decision]

- **Potions** are instant or over time, whichever the row says. Instant healing stays
  near a third of the HP of a character of the level where it drops: minor 20, healing 40,
  greater 80; a mana potion restores 15. [A] Stamina has no formula in
  [derived.md](../core/derived.md), so stamina numbers are provisional. [U]
- **Food** (`type: food`) is always over time; `test_equipment.py` checks that every food
  effect has a duration. [V 2026-11-07] Food can be eaten at any time except in combat, and a
  new dish replaces every effect of the previous one, including effects the new dish does
  not have. Nothing enforces either yet: M2-07 adds the `food` origin with a `replace`
  stacking policy, the tick rule, and the combat check. [U]

### Toxicity

Every `drink` row states `toxicity` ([modifier.md](../core/modifier.md#toxicity)), so a
character cannot drink without limit and food keeps a use: it is preparation before a fight and
cheap recovery after one, and it adds none. [A]

**The formula.** Toxicity has a formula, `formulas/toxicity.json`: the mana the spell cost
formula gives for the drink's effects on the drinker (target `self`), with an over-time pool
restore counting half ([magic.md](magic.md#cost-formula)).

```
toxicity = instant_cost + floor(over_time_cost / 2)
instant_cost     the spell cost of its instant effects and held buffs
over_time_cost   the spell cost of its pool restores with a duration
```

So 40 HP at once costs 20 toxicity, and 2 HP a second for 20 seconds, the same 40, costs 10:
instant is always the more toxic. A crafted potion (the alchemy module) takes the formula's
value. A stock potion (bought, looted, or authored) and a tabletop table keep a static
`toxicity` on the row, at most 5 from the formula (`balance/core.json`), so a table never has
to compute it. [V 2026-11-12, `formula` vectors and `balance_report.py`]

| Strength | Toxicity | Rows |
| --- | --- | --- |
| none | 0 | `antidote`, `potion_cure_disease`: a cure is never blocked by the potions that came before it |
| light | 8 | `potion_mana`: 15 mana by the formula |
| minor | 10 | `potion_minor_healing`, `potion_featherfall` |
| standard | 20 | `potion_healing`, `potion_stamina`, `potion_invisibility`, the resistance potions (50% for 40 seconds), `potion_strength`, `elixir_of_wits` |
| greater | 35 | `potion_greater_healing`, `potion_haste` |

**Stock potions are weaker than magic.** At a magic skill of 50, a Potion of Mana spent on the
best spell for the task restores more per point of toxicity than any stock potion for that
task: 15 mana at 8 toxicity, spent on healing at 2.67 HP a mana (Regeneration, 25% off at skill
50), is 5 HP a point, where a healing potion gives 2. A task no spell does (restoring mana) is
not compared. The alchemy module's crafted potions may beat stock ones. [V 2026-11-12, `balance_report.py`]

### Scrolls

A scroll (`type: scroll`, `delivery: read`) casts one stock spell, named by `consumable.spell`,
once. It costs only gold: it adds no toxicity and spends no mana, so its price is its only
limit. A stock scroll costs at least `scroll_gp_per_mana` (5) gold per mana of its spell.
`scroll_fire_bolt` (6 mana) costs 40 gp. [V 2026-11-12, `balance_report.py`]

At END 5 the limit is 45: two standard potions at once, then about one every four turns as
toxicity falls 5 a turn. At END 1 it is 25, at END 10 it is 70. **Overdosed**
(`core:condition/overdosed`) halves potion magnitude and costs 1 speed while toxicity is above
the limit; it lifts on its own as toxicity falls. [A]

## Gear and tools

A row of `type: tool` or `adventuring_gear` carries a `gear` block. A tool emits contract
modifiers ([modifier.md](../core/modifier.md)) while it is used: `add` on the skill it
helps, source `item:<key>`, origin `core:origin/equipment`. Thieves' tools give +10 to
`lockpick`, a healer's kit +10 to `medicine`. A tool helps only the check it is used for;
the consumer applies its modifiers to that check and no other. [A] `uses` counts the
uses before an item is spent (a healer's kit has 10, a torch 1); with no `uses` it is not
used up.

Adventuring gear (rope, torch, lantern, tent) has no modifiers: what it does is a fact of
the scene, not a number. [A]
