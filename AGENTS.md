# AGENTS.md — Riftweave Ruleset

This file provides guidance to AI coding agents (Grok, Claude Code, Cursor, etc.) when working in this repository.

## Project Nature

**Riftweave is a data-driven, engine-agnostic cRPG ruleset.**

- All game content lives as **JSON** under `ruleset/data/`.
- All content is validated against **JSON Schema (draft-07)** under `ruleset/schemas/`.
- There is **no game engine** here. The deliverable is the data + schemas + documentation. A reference backend in Go (story APP, `docs/stories/07-backend.md`, in `backend/`) serves the ruleset to an agent CLI, an MCP server, an HTTP API, and an Angular frontend; it adds no rules. A new feature is a core command, its MCP tool, and a web page, never a terminal screen: the human CLI (the Charm TUI) is frozen and is being retired (story S23, PO_S23_02).
- Design ethos: **declarative data first**. New mechanics are expressed as JSON + schema extensions, not code or special-cased behavior.

The same data is intended to drive both computer games and TTRPG/virtual tabletop use.

## Essential Commands

```bash
# One-time setup
pip install -r ruleset/requirements.txt

# Validate the entire ruleset (the primary "build" / test command)
python ruleset/scripts/validate.py

# Docker equivalent (matches CI). This image copies the ruleset in; it does not see
# uncommitted edits until you rebuild.
docker build -t riftweave-validate .
docker run --rm riftweave-validate

# Dev toolchain through devops-tools (the checkout is mounted). Host needs uv plus docker or podman.
./scripts/container-image.sh   # image riftweave:python-<version from .toolchain-python-version>
./scripts/validate.sh          # python ruleset/scripts/validate.py
./scripts/test.sh              # unittest discover ruleset/scripts/test_*.py
python ruleset/scripts/run_tests.py   # the same, one process per test file in parallel (what CI runs)
./scripts/format-json.sh       # lay out the listed JSON files; --check fails on drift (CONTRIBUTING.md#json-format)
./scripts/shell.sh             # bash in the toolchain container
./scripts/agent.sh             # devops-tools agent image (Claude Code); see tools/devops/README.md
./scripts/app.sh               # the frontend and its API on the host (Go and Node); docs/backend/frontend.md
./scripts/app-container.sh     # the same, built into one image (docker or podman)
```

- `validate.py` walks **all** files under `ruleset/data/`, requires a `$schema` field on each, and validates against the referenced schema.
- It exits non-zero on any failure. CI (`.github/workflows/validate-ruleset.yml`) runs this on every push/PR touching `ruleset/**`.
- A change is not ready until validation is green.
- The containers come from devops-tools (`devops-tools container`), configured by `[tool.devops-tools.container]` in `tools/devops/pyproject.toml`; `container/Dockerfile` adds only Python on the shared base. Shims live in `scripts/`. Details: `tools/devops/README.md`.
- `./scripts/clean.sh` is end-of-work only. Run it when the user asks. It deletes the `clean-directories` (Python caches) and the `riftweave-uv-cache` volume. It does not touch `cache/logs/`.

## Workflow and pull requests

Read [CONTRIBUTING.md](CONTRIBUTING.md#workflow) before starting an issue. In short:

- Work comes from `docs/issues/open/` (one file per issue), grouped by the stories `S00` to `S24` in `docs/stories/`, one role each (`docs/issues/ROLES.md`). Merged issues are in `docs/issues/done/`, and `close-merged-issue.yml` moves them there, but only when the PR set its issue to `Status: review`; a PO PR that files issues sets its own filing issue to `review` too, or it stays open (PO_S00_10). Format and ids: `docs/stories/README.md`; how an issue closes: `docs/issues/README.md`. Scope and the definition of finished: `ROADMAP.md`.
- One issue, one PR, at most 20 counted changed files (files with 5 or fewer changed lines do not count). Issue ids are `<ROLE>_<STORY>_<ISSUE>_<short-description>` (roles: `docs/issues/ROLES.md`; ids and splitting: `docs/stories/README.md#ids`), and the branch is the full id (e.g. `BE_S07_03_character-list-cli`); an issue with an old id uses `<issue-id>-<slug>` (e.g. `M1-07-<slug>`). The same for people and agents; no `agent/<name>/` prefix. An issue over 20 files that changes both the ruleset and the backend splits into `<issue-id>-ruleset-<slug>` (merges first, adds the fragment) and `<issue-id>-backend-<slug>` (adds a bullet to the same fragment). What can start now and what runs in parallel: `uv run --project tools/devops --locked --no-dev devops-tools roadmap-order --settings-file tools/devops/pyproject.toml show`. Commit only when the user asks.
- Five steps, one commit each: **spec** (`docs/core/` or `docs/modules/`), **schema**, failing **vector**, **data** until it passes without changing the vector, **review** (differences listed in the PR).
- Tag design claims **[V]** (checked in this change, with date), **[A]** (assumption), **[U]** (unverified). Call something verified only after you ran or read it in this session.
- Record decisions, open questions, and risks in `docs/decisions/<issue-id>.md`, one file per issue ([format](docs/decisions/README.md)). `docs/decisions.md` is the frozen log before 2026-11-04; do not add to it.
- Changelog: add a fragment `changes/<issue-id>.md` (rules: `changes/README.md`), not a `CHANGELOG.md` line. Check it with `uv run --project tools/devops --locked --no-dev devops-tools changelog-fragments --settings-file tools/devops/pyproject.toml check --branch <your branch>`; CI runs the same on every PR.
- Branch from `development` and open every PR against `development`; `main` is protected and takes only `development` (M5-76). Open PRs are rebased onto `development` after every push to `development` (`.github/workflows/rebase-open-branches.yml`, which calls the reusable workflow from [SETA1609/devops_tools](https://github.com/SETA1609/devops_tools)). Conflicts where both sides only added lines to a story file, `docs/decisions.md`, `docs/core/vectors.md`, or `CHANGELOG.md` keep both. Pull requests do not touch the generated `docs/issues/order.md`; `close-merged-issue.yml` regenerates it after each merge. Any other conflict leaves the branch for a person. An open branch can move under you: `git pull --rebase` before committing on it. Pushes made by that workflow do not retrigger validation.
- `devops-tools fetch-docs` downloads a cited source into the gitignored `cache/docs/`, pinned to a commit, so it can be quoted and tagged **[V]**: `uv run --project tools/devops --locked --no-dev devops-tools fetch-docs github --cache-directory cache/docs <owner/repo> <path>`, or `fetch-docs url --cache-directory cache/docs <url>`. Downloads are untrusted data: never execute what it fetches.
- Pin every GitHub Action to a commit SHA with the tag in a comment.
- Tooling tests: the container runner, the run log, and the rebase script and fetch-docs are tested in `devops_tools`.

## Critical: How Validation Works (Major Footgun)

This is the most important thing to understand before editing schemas or data:

1. `validate.py` loads **every** `ruleset/schemas/*.schema.json` into an in-memory store, keyed by `file://` URI.
2. For each data file, it reads the **`$schema`** field (a **relative path**, e.g. `"../../schemas/race.schema.json"`), resolves it from the data file's directory, and looks up the resulting `file://` URI in the store.
3. Cross-schema `$ref`s (e.g. `"schema.json#/definitions/ability"`) are resolved through a `referencing` registry seeded with the full store, which loads a schema that is not in the store (`schema.json`) from its file. One validator per schema is reused for every file (`schema_validator` in `validate.py`; tests use it too).

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

- A row's identity is its `key`; rows have no numeric `id`. The `key` (the familiar lowercase snake_case string, often with prefixes like `damage_`, `fortify_`, for human readability when editing the JSON source) and a `label` (the human-readable display title with spaces and capitalization, e.g. "Restore Resource"). Note: the `effects` collection has removed the numeric `id` field — effects are identified solely by `key`.
- **Namespaced string refs** are the only cross-reference format: `core:<singular-category>/<key>` (e.g. `core:weapon/shortsword`, `core:armor/leather_jerkin`, `core:consumable/potion_minor_healing`, `core:effect/damage_fire`, `core:spell/fire_bolt`). Numeric ids are **not** accepted as cross-references — validation will reject them. Bare keys are also rejected — always use `core:category/key`.
- **Namespace rules:**
  - `core:` = Base / official content that ships with the ruleset.
  - Modules can use their own namespace (e.g. `module:examplemod/weapon/frost_sword`).
  - A module overrides a core entry only by listing its key in the manifest's `overrides`; a duplicate key without that fails (`docs/core/overrides.md`).
  - We use **singular** category names: `weapon`, `armor`, `effect`, `spell`, `background`, `trait`, `condition`, `consumable`, `recipe`, `ingredient`, `material`, `gem`, `skill`, `feature`, `crafted_item`.
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
  - Defined in `modules/wuxing/data/wuxing.json` (4 cycles: generating, overcoming, weakening, insulting with multipliers).
  - Races, materials, equipment base templates, and (optionally) effects carry a `phase`.
  - Phase drives **interaction** between effects via the cycles; color drives which skill governs the effect.
  - Equipment base templates assign a default phase per weapon/armor type. Material overrides in the overrides object can change the phase on a per-item basis.
- **Shared effect registry**: `data/effects/core.json` is the single pool. Spells, consumables, and ingredients reference effects by string key via the `appliedEffect` shape. Respect each effect's `channels`.
- **Effect channel enforcement** is active in `validate.py` via `CHANNEL_REQUIREMENTS` — a flat dict mapping context name to required channel (`{"ingredient": "ingredient", "enchantment": "enchantment"}`). Module refs (`module:`) are accepted without channel checks. Add a new entry to the dict to extend enforcement to future channels.
- **Races & lineage** (see `docs/modules/race.md` and `race.schema.json`):
  - `lineage.role`: `standalone`, `parent`, `subrace`, `template`, or `kin`.
  - Parent → subrace = inheritance (subrace can override some fields, e.g. phase).
  - Template → kin = shape-only sisterhood (no mechanical inheritance; each kin is fully self-defined).
  - Playable races (standalone/parent/subrace/kin) require `phase`, `speed`, and `size` (with conditional schema rules for abstract parents/templates).
- **Crafting foundations** (in progress): `materials`, `gems`, and `tiers` collections plus the `qualityGrade` vocabulary in `schema.json`. Materials and gems carry `phase` so they participate in the five-phase cycles when enchanted.

There is **no** class schema or data — the system is deliberately classless.

Cross-file references (effect IDs, skill IDs, perk prerequisites, `parent_race`, `color` → `<color>_magic` skill, etc.) are strings, which JSON Schema cannot check, so `validate.py` fails on any `core:category/key` that does not resolve (`docs/core/integrity.md`). The per-collection checks include:

- Crafted item `base_key`, `material_key`, `engraving_jewel.gem_key`
- Recipe skill, inputs, and output `crafted_item_key`
- Equipment `weapon.skill` (after base resolution)
- Spell effect `parameter` refs (when namespaced)
- Effect `sub_effects`
- All effect refs in spells, consumables, features, ingredients, monsters, conditions
- Perk and skill prerequisites in features
- Race `parent_race` and trait refs
- Background skill bonuses, suggested tag skills, starting spells, and starting equipment
- **Effect channel enforcement**: effects used as `ingredient` must declare `"ingredient"` channel; effects used as `enchantment` (in `crafted_items`) must declare `"enchantment"` channel; effects a maneuver applies must declare `"maneuver"`

Renaming an ID in a data file may cause the validator to report broken references. See `ruleset/scripts/validate.py::check_references()` for the full list of checked fields.

## Current Data Collections & Schemas

Each data file names its schema in `$schema`:

- `abilities`, `skills`, `effects`, `features`, `spells`, `races`, `monsters`, `backgrounds`, `conditions`
- `equipment` (split: `weapons.json`, `armor.json`, `consumables.json`, `accessories.json`, `gear.json`)
- `ingredients`, `materials`, `gems`, `tiers`, `crafting` (recipes and crafted items)
- kernel and world tables: `actions`, `casting`, `checks`, `combat`, `derived`, `formulas`, `progression`, `resolution`, `rest`, `economy`, `loot`, `hazards`, `balance`
- `examples` (worked characters and the play loop)
- `maneuvers` (combat maneuvers: an attack for more or fewer AP, with effects from the shared pool on the `maneuver` channel; `docs/core/actions.md#maneuvers`)
- `skill_groups` (the optional grouped sheet: every skill in one group; `docs/core/progression.md#the-grouped-sheet-variant`)
- `origins` (the kinds of modifier source and their stacking policy, `highest` or `add`)
- `modifier_targets` (the modifier targets no other table holds: a rule's number or a flag; `docs/core/modifier.md#target`)

Equipment (weapons, armors, accessories) and monsters use a **Base + Override** pattern:
- Base templates live in `bases/<category>/` subdirectories (e.g., `equipment/bases/weapons/`, `monsters/bases/`).
- These directories are automatically skipped during file collection; base files are **not** validated standalone.
- A base with `is_item: true` is also the item of its own key (`docs/core/composition.md#a-base-that-is-an-item`); a row in the item file exists only where it changes the base.
- Override entries name a `base` as a typed reference (`core:creature_type/undead`, `core:weapon_base/dagger`, or `core:monster/skeleton`) and provide an `overrides` object with only the differing fields.
- Resolution happens at validation time in `_resolve_equipment_bases()` and `_resolve_monster_bases()`.

`features` is the **single** perk table. Do not create separate trait/class/race tables.

## Modules

`ruleset/modules/` holds optional modules, each a directory with a `manifest.json` (`factions`, `wuxing`, and `example-module` today); `validate.py` checks them along with core data.

`docs/modules/` documents core systems and modules. `docs/core/` wins over it on a conflict ([which document wins](CONTRIBUTING.md#which-document-wins)). For modules not built yet (crafting, alchemy, magic crafting, and the rest), its documents are **design specs** — write data against the contracts described there before implementing new top-level collections or extension points.

See `docs/modules/README.md` for the philosophy and current state.

## Working With This Codebase

1. **Schema first**. Add or extend schema definitions before (or at the same time as) adding data that uses the new fields.
2. Run `python ruleset/scripts/validate.py` after any change. Fix until it is green.
3. Cross-references (effects, skills, races, etc.) must resolve logically even if the schema doesn't enforce it.
4. Match surrounding style: compact one-line entries where the file already uses them; keep description quality high (human + machine readable).
5. Update or add documentation in `docs/modules/` when introducing new concepts or collections.
6. The `source` field is the preferred way to attribute content.
7. **Base + Override pattern**: Weapons, armors, accessories (and monsters) use a base+override data pattern. Base templates live in `ruleset/data/<collection>/bases/<category>/` and define the default stats. Override entries in the main data file reference a `base` and provide only the fields that differ. The files in `bases/` directories are automatically skipped during validation — the `_resolve_equipment_bases()` and `_resolve_monster_bases()` functions merge them at validation time.
   - **Equipment base templates** include `enchantment_slots` (maximum enchantments, negative = cannot enchant), `phase` (Wuxing element), and `default_materials` (the material keys the base is found and sold in, not a limit: any material may be crafted onto any base, as in Elin; odd combinations are player-crafted only).
   - **Material overrides** in override entries can modify `enchantment_slots` (including negative values), change the Wuxing `phase`, and set `effect_magnitude_multiplier` (fractional multiplier on enchantment magnitudes).
   - A negative `enchantment_slots` value means the item cannot receive enchantments (e.g. cursed materials, base items with no magic affinity).
   - Accessory base template files use a `_base` suffix (e.g. `ring_base.json`) to distinguish them from their entries, while weapon and armor bases do not. New base templates should always specify `default_materials`, `enchantment_slots`, and `phase`.
   - **Leather** and **Studded Leather** are materials, not base armor types. They apply to bases like `padded`, `brigandine`, or generic armor bases via the `material` field in overrides.

## Code: dispatch and string literals

Why, with the compilers' own sources: [docs/theory/dispatch.md](docs/theory/dispatch.md).

- **Dispatch on one value is a `switch` or a table, never an `if`/`else if` chain over that value.** Pick the form the language makes fastest:
  - **Go:** `switch` on the value. From 8 dense integer or string cases the compiler emits a jump table (amd64, arm64), below that a binary search. A `switch {}` on conditions compiles as a chain, so do not use it to dispatch. A fixed set of cases is not a `map`; a map is for keys from data.
  - **Python:** a dict from the value to a function or a result, built once at module level (`OPERATIONS[operation](vector)`). `match` and `if`/`elif` try each case in turn.
  - **TypeScript:** `switch`. V8 builds a jump table only for 5 or more compact number-literal cases; a fixed mapping from a string to a value is a lookup on an `as const` object.
  - **State machines:** the state is a small integer (a typed `uint8` in Go, a number in TypeScript), a table indexed by it says what the state holds, and a `switch` on it runs the step.
  - A chain of different conditions (guards, validation, early returns) stays `if`.
- **A string that means something to the code is a named constant**, not a literal repeated where it is compared: an operation, a step, an outcome, a role, a category, a kind.
  - **Go:** a named string type and constants (`type Step string`, `const StepRace Step = "race"`); build JSON from typed structs, not `map[string]any` with string keys.
  - **Python:** module constants (`STEP_RACE = "race"`), and the tuples and dicts built from them.
  - **TypeScript:** an `as const` object with its type derived from it (`export const OUTCOME = { Added: 'added' } as const; type Outcome = (typeof OUTCOME)[keyof typeof OUTCOME]`), not an `enum` or a `const enum`; a template reads it through a component property.
  - **Stays text:** JSON field names a schema or the command contract defines (Go struct tags, result keys), messages, and words shown to a person.

## Frontend

Changes to `frontend/` follow the conventions in [docs/backend/frontend.md](docs/backend/frontend.md#conventions): Angular 22, signals before RxJS, standalone components split into `.ts`, `.html`, and `.css`, Bootstrap's look in our own CSS, small reusable components, Vitest for unit tests, and Playwright for end-to-end tests. Explanations go in a sibling `.md` beside each `.ts` file (`i18n.ts` and `i18n.md`), not in code comments; the file keeps one pointer line. Specs keep their comments and have no sibling.

## Licensing

Dual-licensed (see `LICENSING.md`):
- Tooling / schemas / scripts / CI / Dockerfiles / `tools/devops/` → Apache-2.0
- Game content (`ruleset/data/`), documentation, prose → CC BY 4.0

Contributions are accepted under the same terms (inbound = outbound).

## AI use

AI is welcome as a tool, and a human always reviews before committing ([AI rules](docs/onboarding/ai.md)). An agent commits only when a person asks, after that person has reviewed the change.

## Personal / Local Rules

`.claude/` and any personal `CLAUDE.md` are listed in `.gitignore`, so individuals can keep personal overrides without committing them. The repo's rules for every agent live in this `AGENTS.md`. Claude Code reads it by default, but a `CLAUDE.md` or `CLAUDE.local.md` in the checkout replaces it, so a personal file should start with `@AGENTS.md`.

## References

- `README.md` — high-level project overview
- `docs/onboarding/` — how to get started, and the AI rules
- `ROADMAP.md`, `docs/stories/`, and `docs/issues/` — scope, milestones, and the issues to work on
- `CONTRIBUTING.md` — contribution process and rules for changes
- `docs/core/data-model.md` — the shape of the data; `docs/theory/data-models.md` — why
- `docs/theory/dispatch.md` — why dispatch is a switch or a table and meaningful strings are constants
- `docs/core/` — kernel rules; `docs/modules/` — core systems in detail and modules (start here before adding new mechanics)
- `ruleset/schemas/schema.json` — shared definitions
- `ruleset/scripts/validate.py` — the validator implementation (read if debugging resolution issues)
- `tools/devops/README.md` — the devops-tools pin: toolchain and agent containers, devcontainer
