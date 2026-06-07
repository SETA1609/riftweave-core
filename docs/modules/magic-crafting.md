# Magic Crafting Module

**Status:** Planned / Design Phase

**Alternative names:** Enchanting, Item Creation, Artifice, Rune Crafting

## Overview

The Magic Crafting module covers the creation and customization of permanent magical items — weapons, armor, wondrous items, rings, staves, and other enchanted gear. It is distinct from both general Crafting (mundane items) and Alchemy (temporary consumables).

## Goals

- Enable data-driven definition of magical item recipes, prerequisites, and costs.
- Support common patterns: enchanting existing items, crafting from scratch, imbuing with spells or runes, awakening sentience, etc.
- Provide clear extension points so different magic traditions (arcane, divine, primal, occult, runic, etc.) can coexist.

## Likely Data Additions

- New or extended `equipment` entries with rich magic properties (many existing equipment types can be targets of enchanting).
- Possibly a new top-level collection such as `magic_items`, `enchantments`, or `item_affixes`.
- `recipes` or `formulas` specific to magic item creation (with much higher requirements).
- New feature types: "Artificer Initiate", "Forge Adept", "Enchanter", etc.
- New effect types for permanent or conditional magical properties.

## Major Design Axes

1. **Item vs. Affix model**
   - Treat every magical item as a distinct `equipment` entry (simple, but high duplication).
   - Or use a base item + applied enchantments/affixes (more data-oriented and flexible).

2. **Spell investment**
   - Many systems require "knowing" or "expending" specific spells during creation (e.g. *magic weapon*, *continual flame*).
   - How do we reference spells from within crafting data?

3. **Charges, attunement, rarity**
   - These are common properties of magical items. Should they live on equipment, or be added by magic crafting data?

4. **Runes / inscriptions**
   - A modular way to add properties (inspired by systems like 13th Age or custom rune systems).

## Open Questions

- Do we extend the core `equipment` schema for magic properties, or keep magic data in a parallel structure that references equipment ids?
- How do we handle "upgrading" an existing item (e.g. +1 longsword → flaming longsword)?
- What is the relationship to class features that grant magic item creation (Artificer, Wizard high-level abilities)?
- Should there be a general "item property" or "tag" system that both mundane crafting and magic crafting can contribute to?

## Integration Points

- **Equipment**: The main target of magical modification.
- **Spells**: Many magic crafting recipes will require specific spells as ingredients or knowledge.
- **Features**: Creation feats, class features, and prerequisites.
- **Effects**: New effect categories for granted abilities, resistances, activated powers, etc.
- **Source**: Critical for tracking which book/tradition a given enchantment comes from.

## Scope Suggestions

Start with a small, well-defined set:
- Basic weapon and armor enchantments (+1, elemental damage, etc.).
- A handful of classic wondrous items with clear creation paths.
- One "signature" subsystem (e.g. runes, or artificer infusions) to prove the model.

## Related Modules

- Alchemy (temporary vs permanent magic)
- Crafting (mundane base items that then get enchanted)

---

This is one of the more complex potential modules. Use this document to capture decisions about the shape of magical item data before writing large amounts of JSON.
