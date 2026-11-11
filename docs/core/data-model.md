# Data model

What shape Riftweave's information has. For what a data model is, the common types, and why
this one was chosen, see [the theory page](../theory/data-models.md). This page links to the
document that owns each rule and does not repeat it.

## The type

Riftweave is **a relational model stored as JSON documents**, loaded by the backend into flat
in-memory tables. It is a hybrid, and each layer has its own type. [V 2026-11-11]

| Layer | Type | What that means here |
| --- | --- | --- |
| Logical | Relational | Tables of rows, unique keys, references that act as foreign keys, one fact in one place |
| Storage | Document | JSON files under `ruleset/data/`, with some nested lists, such as a spell's `effects` |
| Runtime | Flat tables | `backend/internal/data`: dense slices of rows with a key index, built once at load |

## The parts

| Part | Rule | Owner |
| --- | --- | --- |
| Tables of rows | A file wraps named lists of rows. A row's identity is its `key`; a few vocabulary rows use a string `id`. [V 2026-11-11] | [keys.md](keys.md) |
| Schemas | Every data file names its JSON Schema, except `formulas/` (below), and unknown fields are rejected. [V 2026-11-11] | `ruleset/schemas/` |
| References | `core:<category>/<key>`. A numeric id is not a reference. [V 2026-11-11] | [keys.md](keys.md), [integrity.md](integrity.md) |
| Shared pools | An effect is one row; spells, consumables, and ingredients point at it. [V 2026-11-11] | `ruleset/data/effects/core.json` |
| Composition | A row is a base, plus templates, plus overrides, merged into one flat row. [V 2026-11-11] | [monsters.md](../modules/monsters.md), [overrides.md](overrides.md) |
| Stored and computed | Store the primary facts. Derived stats are functions of them. [V 2026-11-11] | [contract.md](contract.md), [derived.md](derived.md) |
| One way to change a value | A modifier, with one stacking rule. [V 2026-11-11] | [contract.md](contract.md), [modifier.md](modifier.md) |
| Modifier targets | A modifier's `target` is a key naming one row, of the kind its op needs; targets no other table holds are rows of `modifier_targets`. [V 2026-11-11] | [modifier.md](modifier.md#target), [integrity.md](integrity.md) |
| Formulas as data | Math is a tree in JSON, compiled by the backend to a flat program. [V 2026-11-11] | [formula.md](formula.md) |
| Rolls as inputs | The rules never roll. A roll is an input, so a result can be repeated. [V 2026-11-11] | [resolution.md](resolution.md) |
| Vectors | An input and its exact expected result. An implementation that disagrees is wrong. [V 2026-11-11] | [vectors.md](vectors.md) |
| Modules | A module adds rows, and overrides a core row only by listing its key. [V 2026-11-11] | [overrides.md](overrides.md) |

## What it is not

- **Not entity-component (ECS).** A monster is a whole row, not an id with attachable
  components. `monsters.md` describes ECS as a possible later path. [V 2026-11-11]
- **Not object inheritance.** Base, templates, and overrides merge into one flat row. A
  derived monster cannot change its creature type. [V 2026-11-11]
- **Not a database.** There is no query engine or transaction. The files are the source of
  truth, and git is their history. [V 2026-11-11]

## Known exceptions

Places where the data does not yet follow the model:

- **Authored monster HP.** 50 of the 54 monster rows store `hitPoints` instead of computing it. [V 2026-11-11]
- **Authored HP growth.** Each creature type in `monsters/bases/` authors `growth.hitPointsPerLevel`, so a monster's HP does not come from the player formula. [V 2026-11-11]
- **Formulas have no schema.** The four files in `formulas/` name no `$schema`, so `validate.py` skips them with a warning. [V 2026-11-11]
- **Distances in metres.** `safe_fall_meters` and `darkvision_range` still hold metres; M5-33 converts them. [V 2026-11-11]
- **JSON Schema cannot check references.** `validate.py` and tests do that on top of it, and a
  free-text field such as `parameter` is only checked where a test says so. [V 2026-11-11]
- **Module overrides.** `module_loader.merge_data()` does not read `overrides` yet.
  [V 2026-11-11]
