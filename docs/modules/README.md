# Modules

Riftweave is designed to be extended through optional **modules**. Modules allow you to add new mechanics, content, and data without bloating or modifying the core ruleset.

## Goals

- Keep the **core** minimal, stable, and focused on foundational rules (abilities, classes, races, basic equipment, spells, features).
- Allow **optional, composable extensions** (crafting, alchemy, advanced magic item creation, psionics, vehicles, etc.).
- Enable consuming games and tools to load only the modules they care about.
- Support clear ownership: each module owns its data, and can extend or add to shared vocabularies (new effect types, new equipment categories, new feature types, etc.).

## Current State

Today the ruleset is monolithic:

- All data lives under `ruleset/data/` in "core" collections.
- There is a single set of schemas under `ruleset/schemas/`.
- The validation system treats everything as one ruleset.

There is no module system, no dependency declaration, and no composition model yet.

## Philosophy (Data-Oriented)

Modules should follow the same data-oriented principles as the core:

- **Declarative data first** — new mechanics are described in JSON + schema, not hardcoded behavior.
- **Loose but governed extension points** — prefer extending existing structures (`features.effects`, new `equipment.type` values, new feature `type`s) over creating parallel systems when possible.
- **Self-describing** — module data should declare its schemas.
- **Source traceability** — use the `source` field so tools can attribute content to specific modules/books.
- **Composable** — a module should be possible to include or exclude without breaking core validation (future goal).

## Planned / Recommended Structure (Future)

A possible future layout (not yet implemented):

```
ruleset/
  data/                 # core (always loaded)
  schemas/
  modules/
    crafting/
      schemas/          # optional additional schemas or extensions
      data/
        recipes.json
        ...
    alchemy/
      data/
        ...
    magic_crafting/
      ...
```

Modules may:
- Add entirely new top-level collections (e.g. `recipes`).
- Contribute new entries to existing collections (e.g. new `features`, new `equipment` items of custom types).
- Define new effect types or prerequisite kinds (documented in the module's schema/docs).

## Using This Directory

This `docs/modules/` directory is the place to document module design, data shapes, effect vocabularies, and integration points **before** (or while) implementing the actual data.

Create one document per module (or per major subsystem). Use it to:
- Define what new data types or effect types the module introduces.
- Specify any new schemas or extensions to core schemas.
- List dependencies on other modules or core version.
- Capture open questions and design decisions.

## Example Modules

| Module            | Focus                              | Likely Touch Points                     | Status   |
|-------------------|------------------------------------|-----------------------------------------|----------|
| Crafting          | General item creation, recipes     | New equipment types, recipes, features  | Planned  |
| Alchemy           | Potion brewing, extracts, reagents | Equipment (potions), features, skills   | Planned  |
| Magic Crafting    | Enchanting, wondrous items, runes  | Equipment, spells, features, effects    | Planned  |
| Psionics          | Psychic powers                     | New "power" collection or spell school  | Future   |
| Vehicles / Mounts | Mounts, vehicles, ship combat      | New entities + rules                    | Future   |

Start a new document in this directory when you begin designing a module (e.g. `crafting.md`).

## Next Steps (Non-Exhaustive)

- [ ] Decide on module metadata (id, version, dependencies, authors).
- [ ] Design how modules register new `equipment.type` values or `feature.type` values.
- [ ] Define a stable effect type vocabulary that modules can extend.
- [ ] Update the validator (or create a module-aware validator) to handle multiple schema/data roots.
- [ ] Add optional module loading to the README and CI examples.

Contributions to module design are welcome via documentation and proposals in this directory even before any code or data changes.
