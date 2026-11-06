# Equipment

What a character can buy, wear, carry, and use up. The rows are in
`ruleset/data/equipment/` (`equipment.schema.json`): weapons, armor, and accessories as a
base in `bases/` plus overrides, consumables and gear as full rows. Weapons are in
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
