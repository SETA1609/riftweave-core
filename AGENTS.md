# AGENTS.md — Riftweave Ruleset

This file provides guidance to AI coding agents (Grok, Claude Code, Cursor, etc.) when working in this repository.

## Project Nature

**Riftweave is a data-driven, engine-agnostic cRPG ruleset.**

- All game content lives as **JSON** under `ruleset/data/`.
- All content is validated against **JSON Schema (draft-07)** under `ruleset/schemas/`.
- There is **no runtime engine** here. The deliverable is the data + schemas + documentation.
- Design ethos: **declarative data first**. New mechanics are expressed as JSON + schema extensions, not code or special-cased behavior.

The same data is intended to drive both computer games and TTRPG/virtual tabletop use.

## Essential Commands

```bash
# One-time setup
pip install -r ruleset/requirements.txt

# Validate the entire ruleset (the primary "build" / test command)
python ruleset/scripts/validate.py

# Docker equivalent (matches CI)
docker build -t riftweave-validate .
docker run --rm riftweave-validate
```

- `validate.py` walks **all** files under `ruleset/data/`, requires a `$schema` field on each, and validates against the referenced schema.
- It exits non-zero on any failure. CI (`.github/workflows/validate-ruleset.yml`) runs this on every push/PR touching `ruleset/**`.
- A change is not ready until validation is green.

## Critical: How Validation Works (Major Footgun)

This is the most important thing to understand before editing schemas or data:

1. `validate.py` loads **every** `ruleset/schemas/*.schema.json` into an in-memory store, keyed by `file://` URI.
2. For each data file, it reads the **`$schema`** field (a **relative path**, e.g. `"../../schemas/race.schema.json"`), resolves it from the data file's directory, and looks up the resulting `file://` URI in the store.
3. Cross-schema `$ref`s (e.g. `"schema.json#/definitions/ability"`) are resolved using a `RefResolver` seeded with the full store.

**Consequences you must follow:**
- Every data file **must** declare a `$schema` field. Files without one are skipped (with a warning) and not validated.
- Schemas may only `$ref` other schema files that physically live in `ruleset/schemas/`.
- The shared vocabulary lives in `ruleset/schemas/schema.json` (see below). Always reference it via `schema.json#/definitions/...` rather than copying definitions.
- If you add a new schema file, it must be in `ruleset/schemas/` and will be auto-discovered on next validation.

## Data & Schema Conventions

- **`ruleset/schemas/schema.json`** is the single source of truth for shared vocabulary:
  - `ability` enum (the 8 core attributes: `str`, `per`, `end`, `int`, `wil`, `agi`, `cha`, `lck`)
  - `diceExpression`, `percentile`, `color` (9 magic schools), `phase` (5 wuxing elements), `qualityGrade`, `appliedEffect`, `sourceRef`.
  - Changing attributes or adding new shared types happens here.

- Every data file is a single wrapper object containing one named array that matches the collection:
  ```json
  { "$schema": "...", "races": [ ... ], "skills": [ ... ], ... }
  ```

- **All schemas set `additionalProperties: false`**. To add any new field to a data entry, you **must** first add it to the corresponding schema, or validation will fail.

- Entries use a numeric `id` (positive integer, unique within the collection) for internal tracking, and a `key` (the familiar lowercase snake_case string, often with prefixes like `damage_`, `fortify_`, for human readability when editing the JSON source) and a `label` (the human-readable display title with spaces and capitalization, e.g. "Restore Resource"). Note: the `effects` collection has removed the numeric `id` field — effects are identified solely by `key`.
- **Namespaced string refs** are the only cross-reference format: `core:<singular-category>/<key>` (e.g. `core:weapon/shortsword`, `core:armor/leather_jerkin`, `core:consumable/potion_minor_healing`, `core:effect/damage_fire`, `core:spell/fire_bolt`). Numeric ids are **not** accepted as cross-references — validation will reject them.
- **Bare keys** (e.g. `shortsword`, `damage_fire`) are also accepted as cross-references, making the data less verbose when the category is unambiguous. Both formats resolve to the same entries.
  - **Deprecation notice**: Bare keys will be removed in a future update. Use namespaced keys (`core:category/key`) for all new content.
- **Namespace rules:**
  - `core:` = Base / official content that ships with the ruleset.
  - Modules can use their own namespace (e.g. `module:examplemod/weapon/frost_sword`).
  - Modules **can override** core entries by using the same `core:category/key` (this is intentional for module flexibility).
  - We use **singular** category names: `weapon`, `armor`, `effect`, `spell`, `background`, `trait`, `condition`, `consumable`, `recipe`, `ingredient`, `material`, `gem`, `skill`, `feature`.
- Use the optional `source: { source, page }` field (defined in `schema.json`) for provenance instead of inventing per-file metadata.

## Game System Highlights (Affects Data Modeling)

- **Classless d100 roll-under**. Skills are 0–100. Attributes are 1–10 and seed base skill values.
- Skills are **point-buy on level-up** (not use-based training). 3 **tag skills** get +2 per point instead of +1.
- **Perks** (in `features/core.json`) are the primary character progression. The single `features` table contains:
  - `perk` — chosen on level-up
  - `creation` — chosen only at character creation (tradeoff "traits")
  - `racial_trait` — automatically granted by ancestry
  - `universal` — always available
- **Magic**: 9 color schools. Spells are compositions of effects from the shared pool, tagged with a `color`. The governing skill is `<color>_magic`.
- **Five-phase (Wuxing) system** (orthogonal to color): `wood` / `fire` / `earth` / `metal` / `water`.
  - Defined in `data/wuxing/core.json` (4 cycles: generating, overcoming, weakening, insulting with multipliers).
  - Races, materials, and (optionally) effects carry a `phase`.
  - Phase drives **interaction** between effects via the cycles; color drives which skill governs the effect.
- **Shared effect registry**: `data/effects/core.json` is the single pool. Spells, consumables, and ingredients reference effects by string key via the `appliedEffect` shape. Respect each effect's `channels`.
- **Races & lineage** (see `docs/modules/race.md` and `race.schema.json`):
  - `lineage.role`: `standalone`, `parent`, `subrace`, `template`, or `kin`.
  - Parent → subrace = inheritance (subrace can override some fields, e.g. phase).
  - Template → kin = shape-only sisterhood (no mechanical inheritance; each kin is fully self-defined).
  - Playable races (standalone/parent/subrace/kin) require `phase`, `speed`, and `size` (with conditional schema rules for abstract parents/templates).
- **Crafting foundations** (in progress): `materials`, `gems`, and `tiers` collections plus the `qualityGrade` vocabulary in `schema.json`. Materials and gems carry `phase` so they participate in the five-phase cycles when enchanted.

There is **no** class schema or data — the system is deliberately classless.

Cross-file references (effect IDs, skill IDs, perk prerequisites, `parentRace`, `color` → `<color>_magic` skill, etc.) are **by string ID only** and are not validated by JSON Schema. Renaming an ID can silently break consumers. (A future referential integrity pass after schema validation is planned.)

## Current Data Collections & Schemas

Each collection has a matching `*.schema.json`:

- `abilities`, `skills`, `effects`, `features`, `spells`, `races`, `monsters`
- `equipment` (split: `weapons.json`, `armor.json`, `consumables.json`)
- `ingredients`, `materials`, `gems`, `tiers`
- `wuxing` (the interaction matrix — single source of truth for cycles)

`features` is the **single** perk table. Do not create separate trait/class/race tables.

## Modules (Planned / Aspirational)

`docs/modules/` contains design documents for optional, composable extensions (crafting, alchemy, magic-crafting, etc.). 

Today the ruleset is monolithic under `ruleset/data/`. The module docs are **design specs** — write data against the contracts described there before implementing new top-level collections or extension points.

See `docs/modules/README.md` for the philosophy and current state.

## Working With This Codebase

1. **Schema first**. Add or extend schema definitions before (or at the same time as) adding data that uses the new fields.
2. Run `python ruleset/scripts/validate.py` after any change. Fix until it is green.
3. Cross-references (effects, skills, races, etc.) must resolve logically even if the schema doesn't enforce it.
4. Match surrounding style: compact one-line entries where the file already uses them; keep description quality high (human + machine readable).
5. Update or add documentation in `docs/modules/` when introducing new concepts or collections.
6. The `source` field is the preferred way to attribute content.

## Licensing

Dual-licensed (see `LICENSING.md`):
- Tooling / schemas / scripts / CI / Dockerfile → Apache-2.0
- Game content (`ruleset/data/`), documentation, prose → CC BY 4.0

Contributions are accepted under the same terms (inbound = outbound).

## Personal / Local Rules

`CLAUDE.md` and `.claude/` are listed in `.gitignore`. This is intentional so individuals can maintain personal overrides without committing them. The canonical, shareable rules for the project live in this `AGENTS.md`.

## References

- `README.md` — high-level project overview
- `CONTRIBUTING.md` — contribution process and rules for changes
- `docs/modules/` — detailed design for core systems and planned modules (start here before adding new mechanics)
- `ruleset/schemas/schema.json` — shared definitions
- `ruleset/scripts/validate.py` — the validator implementation (read if debugging resolution issues)
