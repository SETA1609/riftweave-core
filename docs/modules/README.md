# Modules

Riftweave is designed to be extended through optional **modules**. Modules allow you to add new mechanics, content, and data without bloating or modifying the core ruleset.

## Goals

- Keep the **core** minimal, stable, and focused on foundational rules (abilities, classes, races, basic equipment, spells, features).
- Allow **optional, composable extensions** (crafting, alchemy, advanced magic item creation, psionics, vehicles, etc.).
- Enable consuming games and tools to load only the modules they care about.
- Support clear ownership: each module owns its data, and can extend or add to shared vocabularies (new effect types, new equipment categories, new feature types, etc.).

## Current State

The module system foundation is **implemented** as of the `ft/modules-foundation` branch:

- A formal module directory structure exists under `ruleset/modules/`.
- Each module is self-describing via `manifest.json` (validated by `module.manifest.schema.json`).
- The Python module loader (`ruleset/scripts/module_loader.py`) handles discovery, validation, data loading, schema loading, and conflict detection.
- An `example-module` demonstrates the pattern.
- 26 unit tests cover manifest validation, discovery, loading, merging, and edge cases.

The validator (`validate.py`) validates core data only — module data is merged at load time by consuming engines or test tools. See `docs/modules/module-system.md` for the full specification.

## Philosophy (Data-Oriented)

Modules should follow the same data-oriented principles as the core:

- **Declarative data first** — new mechanics are described in JSON + schema, not hardcoded behavior.
- **Loose but governed extension points** — prefer extending existing structures (`features.effects`, new `equipment.type` values, new feature `type`s) over creating parallel systems when possible.
- **Self-describing** — module data should declare its schemas.
- **Source traceability** — use the `source` field so tools can attribute content to specific modules/books.
- **Composable** — a module should be possible to include or exclude without breaking core validation (future goal).

## Module Structure

See `docs/modules/module-system.md` for the full specification.

Current layout:

```
ruleset/
  data/                 # core (always loaded)
  schemas/
    module.manifest.schema.json
  modules/
    example-module/
      manifest.json
      data/
  scripts/
    module_loader.py    # Python module loader
    test_module_system.py
```

Modules may:
- Add entirely new top-level collections (e.g. `recipes`).
- Contribute new entries to existing collections (e.g. new `features`, new `equipment` items).
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
| Advancement       | XP level-up + item-boosted breakthroughs (cores, elixirs, Wuxing) | progression, alchemy, ingredients, effects, features | In design |
| Psionics          | Psychic powers                     | New "power" collection or spell school  | Future   |
| Vehicles / Mounts | Mounts, vehicles, ship combat      | New entities + rules                    | Future   |

Start a new document in this directory when you begin designing a module (e.g. `crafting.md`).

## Next Steps (After Foundation Merge)

- [x] Decide on module metadata (id, version, dependencies, authors).
- [x] Design module manifest format and JSON Schema.
- [x] Implement Python module loader with discovery, validation, and merging.
- [x] Basic conflict detection (duplicate IDs across modules/core).
- [ ] Register new `equipment.type` values or `feature.type` values from module schemas.
- [ ] Define a stable effect type vocabulary that modules can extend.
- [ ] Module-aware validator that validates core + selected modules together.
- [ ] Add optional module loading to the README and CI examples.
- [ ] Dependency resolution (module A requires module B).

Contributions to module design are welcome via documentation and proposals in this directory even before any code or data changes.
