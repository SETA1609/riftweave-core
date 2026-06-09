# Alchemy Module

**Status:** Core data shapes in place; brewing/quality rules still in design.

## Overview

Alchemy produces consumables — potions, poisons, weapon coatings, oils, and food —
by combining **ingredients**. It is built entirely on the **shared effect registry**
(`ruleset/data/effects/core.json`): ingredients expose effect ids, and brewing turns
matching effects into a finished consumable that carries those effects.

See [progression.md § Effect registry](progression.md#effect-registry--the-shared-pool)
for the pool itself.

## The pieces (already in the ruleset)

- **Effects** (`data/effects/core.json`) — atomic effects with a `polarity`
  (beneficial / harmful / neutral) and a `channels` list. Alchemy only draws
  effects whose channels include `ingredient`, `potion`, `poison`, `coating`, or
  `food`.
- **Ingredients** (`data/ingredients/core.json`, schema `ingredient.schema.json`) —
  each lists the effect ids it can impart (magnitude is *not* stored; it is derived
  at brew time). Example: `nightshade` → `damage_health`, `damage_stamina`.
- **Consumables** (`data/equipment/consumables.json`) — equipment of type
  `potion` / `poison` / `coating` / `oil` / `food` with a `consumable` block
  (`delivery` + `effects[]` referencing the pool). These are the brewing *output*
  and can also be authored directly as loot/shop stock.

## Brewing model (design)

1. **Combine ingredients** that share at least one effect id. The shared effect(s)
   become the brew's effect(s) — only overlapping effects carry into the product.
2. **Polarity decides the product:** beneficial shared effects → a **potion**
   (drink); harmful shared effects → a **poison** / **coating** (apply to weapon).
3. **Quality / magnitude** is derived from `f(Alchemy skill, ingredient qualities,
   station tier)` — exact formula TBD (mirrors the engine crafting quality rule).
   Higher Alchemy skill yields larger magnitude and longer duration.
4. **Discovery:** which effects an ingredient holds is revealed by gameplay
   (experimentation, being taught, finding notes) rather than known up front.

## Skills & perks

- Governed by the **Alchemy** skill (`utility` category, `int`-based) — point-buy
  like any skill; brewing does not train it.
- Alchemy perks live in `features/core.json` (`type: "perk"`, `category: "utility"`),
  gated by `prerequisite.skills.alchemy`.

## Open questions

- The quality/magnitude formula and ingredient "purity/quality" attribute.
- Station/tool gating (alchemy lab tier) — shared with the Crafting module.
- Multi-effect brews: do all shared effects carry over, or only the strongest?
- Shelf life / instability as data, if modeled at all.

## Related modules

- **Crafting** — general synthesis; alchemy is the consumables-focused sibling.
- **Magic Crafting** — permanent enchantments (also draws from the effect pool via
  the `enchantment` channel).
- **Advancement / Breakthroughs** (`docs/modules/advancement.md`) — Alchemy will
  gain a "refining" or "pill crafting" path that turns raw monster cores into
  higher-grade or harmonized essence elixirs for use in Wuxia-style level-up
  breakthroughs (in addition to normal potion/poison brewing).
