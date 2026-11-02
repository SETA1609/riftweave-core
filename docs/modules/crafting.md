# Crafting Module

**Status Update (feat/crafting-blacksmithing):**  
Initial slice complete. Material phase inheritance + jewel engraving (phase overwrite) + permanent vs temporary coatings are implemented. See `ruleset/data/crafting/crafted_items.json` for concrete examples and `basic_recipes.json` for future recipe stubs.

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

## Implemented Slice (feat/crafting-blacksmithing)

The following was added in this branch:

| Artifact | Description |
|----------|-------------|
| `schemas/crafted_item.schema.json` | Schema for composed items: base + material + optional engraving jewel + optional enchantments + optional coatings |
| `schemas/recipe.schema.json` | Stub schema for crafting recipes (data only, no resolution implemented) |
| `data/crafting/crafted_items.json` | 6 example items demonstrating phase inheritance, jewel overwrite, enchantments, and coatings |
| `data/crafting/basic_recipes.json` | 3 recipe stubs showing the input/output structure for future resolution |
| `scripts/crafting_reference.py` | Reference resolver demonstrating phase propagation, Wuxing interaction, and permanent vs temporary tracking |
| `scripts/validate.py` | Crafting key-reference validation: recipe inputs/outputs and crafted_item base_key/material_key/gem_key are checked against their respective collections |

### Composition Model

Base templates live in `equipment/bases/<category>/` (e.g., `equipment/bases/weapons/longsword.json`)
and use a base+override pattern. The equipment data files (`weapons.json`, `armor.json`,
`accessories.json`) reference these bases and may provide overrides for material/quality variants.
The crafted item's `base_key` must resolve to a valid equipment base template key.

The `validate.py` script resolves base+override entries before schema validation, so the
resulting flat entries are what the schemas and cross-reference checks see. Accessories
(rings, necklaces, circlets, earrings) follow the same base+override pattern.

```
Base Template  +  Material  +  [Jewel]  +  [Enchantments]  +  [Coatings]
(longsword)       (steel)      (ruby)       (core:effect/burn)  (paralytic)
     │               │            │               │                │
     ▼               ▼            ▼               ▼                ▼
     └───────────────┴────────────┴───────────────┴────────────────┘
                              │
                              ▼
                     Crafted Item Instance
                     ├── phase: "fire" (jewel overwrites material's "metal")
                     ├── attack_bonus: +5
                     ├── enchantments: [permanent]
                     └── coatings: [temporary, 3 uses]
```

Key rule: **phase defaults to material's phase; engraving jewel overwrites it entirely.** This is the full engraving mechanic for v1.

See `scripts/crafting_reference.py --verbose` for a runnable demonstration.

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

## The module slice (POST-01)

The module `crafting` (`ruleset/modules/crafting/`) turns the data above into a procedure: a recipe, a skill check, time, a station, and a quality grade. It is optional. Core resolves an attack and a level-up with it unloaded ([contract](../core/contract.md)); it reads core rows and emits an item, and nothing calls into it. The older `data/crafting/` stubs and `crafting_reference.py` (phase composition) stay as they are. [A]

### Inspiration and how it differs

- **Morrowind and Oblivion** make an item's quality a function of one skill roll; **Fallout: New Vegas** gates recipes on a workbench **[U]**. Riftweave takes both: one d100 roll-under check on the recipe's skill ([resolution.md](../core/resolution.md)), and a station every recipe names.
- **D&D 5e downtime crafting** counts days against gold **[U]**. Riftweave counts hours and lets the station change them.
- The grade ladder is the shared `qualityGrade` vocabulary in `schema.json` (`petty` to `legendary`, `common` the baseline). No bonus taxonomy: a station's bonus is an `add` on the check.

### Rules

```
target   = skill + station.skill_bonus - recipe.difficulty        no cap (resolution.md)
roll     d100 against target, by the one resolution function
fumble   ruined: no item, all inputs consumed
failure  no item, fail_loss_percent of the inputs consumed
success  margin = target - roll
         grade_index = base_index + floor(margin / 10) + crit_bonus     formula tree
         grade_index = clamp(grade_index, 0, station.max_grade_index)
         the item is made at that grade
time     hours = max(min_hours, floor(recipe.time * station.time_percent / 100))
```

- `base_index` is the index of `common` on the ladder (3). `crit_bonus` is 1 when the roll is tagged `crit`, else 0. A margin of 30 or more on a good station gives `greater`; a station caps what skill alone cannot exceed. [A]
- A station is a module row `module:crafting/station/<key>`. A recipe names the one it needs; a station whose `counts_as` lists it also serves (a master's forge serves as a smithy forge). A recipe with no suitable station cannot be started, and the check never rolls. [A]
- Inputs are core rows (`core:material/...`, `core:ingredient/...`, `core:gem/...`). The output is a core item, with a `core:material/...` for gear. The material must be in the base template's `valid_materials` ([materials.md](../core/materials.md)). The output is an item instance with a grade; this module does not write a `crafted_item` row, and phase composition stays in `crafting_reference.py`.
- Experience: none. Crafting does not award XP ([progression.md](progression.md)).

### Data

| file | holds |
| --- | --- |
| `data/stations.json` | `stations`: skill bonus, highest grade, time percent, `counts_as` |
| `data/quality.json` | `quality`: the grade ladder with value percent, the base grade, the grade formula tree, the time formula tree, the failure loss |
| `data/recipes.json` | `crafting_recipes`: skill, station, difficulty, hours, inputs, output |

Vectors are in `ruleset/modules/crafting/tests/vectors.json`, run by `ruleset/scripts/test_craft_reference.py` (the core vector test checks kernel functions only). The numbers are first choices **[U]**; the check is that the vectors pin them.

