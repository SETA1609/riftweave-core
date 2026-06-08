# Crafting Module

**Status:** Planned / Design Phase

## Overview

The Crafting module adds systems for characters to create equipment, tools, and other items using recipes, skills, and resources. It turns "I buy a sword" into "I gathered materials and spent downtime to craft it."

## Goals

- Provide a general, data-driven crafting framework.
- Support multiple crafting disciplines (blacksmithing, leatherworking, woodworking, tinkering, etc.).
- Allow recipes to produce core equipment, custom variants, or entirely new item types introduced by the module.
- Integrate with existing features, skills, and proficiencies.

## Likely Data Additions

- `recipes` (or `crafting_recipes`) collection
- Possibly new `equipment.type` values (e.g. `"crafted"`, or more specific subtypes)
- New feature types or prerequisites (e.g. "Tool Proficiency: Smith's Tools")

The **`materials`** collection already exists (`ruleset/data/materials/core.json`,
schema `material.schema.json`): bronze, copper, iron, steel, wood, elven_wood,
silver, gold, obsidian, …. Each material carries a five-phase `phase` element and
`modifiers` (`attack`, `defense`, `effectMagnitude`) it confers on a crafted weapon,
armor, or accessory, plus `tier`, `appliesTo`, and weight/value factors. Recipes will
consume materials and read these modifiers onto the produced item.

## Open Design Questions

- How are recipes represented? (ingredients + quantities + skill checks + time + tools)
- Should "quality" or "masterwork" tiers be modeled as data on the resulting item, or as separate items?
- Can modules introduce new top-level collections, or should everything flow through `features` + `equipment` + effects?
- How do we handle "downtime" or activity systems that many crafting systems need?
- Should there be a generic "activity" or "task" abstraction that multiple modules (crafting, research, training) can use?

## Integration Points

- **Materials** (`data/materials/core.json`): the inputs a recipe consumes; their
  `phase` and `modifiers` shape the crafted item's element, attack/defense, and
  enchantment magnitude.
- **Equipment**: New items or variants produced by recipes.
- **Features**: Crafting-related feats and class features (e.g. "Artisan", "Master Craftsman").
- **Skills**: Existing skills (or new ones) used for crafting checks.
- **Effects**: Possible new effect types like `grant_recipe`, `modify_crafting_cost`, `grant_tool_proficiency`.

## Related Modules

- Alchemy (specialized crafting for consumables)
- Magic Crafting (enchanting and creation of magical items)

## References & Inspiration

- D&D 5e: Crafting rules, Xanathar's downtime activities, tool proficiencies
- Pathfinder 2e: Crafting subsystem
- Various cRPGs (Baldur's Gate 3, Dragon Age, etc.)

---

**Next action:** Flesh out a minimal recipe schema and example data once the core module system shape is clearer.
