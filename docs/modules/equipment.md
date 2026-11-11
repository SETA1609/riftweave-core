# Equipment

What a character can buy, wear, carry, and use up. The rows are in
`ruleset/data/equipment/` (`equipment.schema.json`): weapons, armor, and accessories as a
base in `bases/` plus overrides, consumables and gear as full rows. A base with
`is_item: true` is its own item, so a row is written only where it changes the base
([composition.md](../core/composition.md#a-base-that-is-an-item)). Weapons are in
[weapons.md](weapons.md), armor in [armor.md](armor.md).

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
