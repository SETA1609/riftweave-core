# Formal Module System

**Branch Status (ft/modules-foundation):** Foundation complete — module directory structure, manifest schema, Python loader with discover/validate/merge, conflict detection, and **validator integration**. The core validator (`validate.py`) now discovers and validates module data automatically. 36 unit tests passing.

---

## Overview

The Riftweave module system is a **data-oriented** extension framework. Modules are self-describing directories under `ruleset/modules/` that can add new entries to existing core collections and introduce entirely new collections.

The design follows three principles:

1. **Declarative** — a module is just a directory containing a `manifest.json` and data files. No registration code needed.
2. **Composable** — modules merge data with core and each other. Conflict detection flags duplicate IDs explicitly.
3. **Backward-compatible** — if no modules exist, `validate.py` behaves exactly as before. The core `ruleset/data/` and `ruleset/schemas/` are untouched.

---

## How to Create a New Module

1. **Create a directory** under `ruleset/modules/<your_module_name>/`.
2. **Add `manifest.json`** with at minimum: `id`, `version`, `description`.
3. **Create `data/`** with JSON files following core data conventions (each file must have a `$schema` field referencing a schema in `ruleset/schemas/`).
4. **List your data files** in `manifest.json` under the `"data"` key.
5. **Run `python ruleset/scripts/validate.py`** — your module data is validated automatically.

Example:

```
ruleset/modules/my-custom-content/
├── manifest.json
└── data/
    └── my_features.json
```

```
// manifest.json
{
  "$schema": "../../schemas/module.manifest.schema.json",
  "id": "my_custom_content",
  "version": "1.0.0",
  "description": "Adds custom features and equipment.",
  "data": ["data/my_features.json"]
}
```

---

## Module Loading Flow

```
┌─────────────────────────────────────────────────────────┐
│                   ruleset/modules/                       │
│                                                         │
│   example-module/          my-module/                   │
│   ├── manifest.json        ├── manifest.json            │
│   ├── data/                ├── data/                    │
│   └── schemas/             └── schemas/                 │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│  1. DISCOVER                                                │
│     validate.py scanns modules/ for subdirs with            │
│     manifest.json. Skip hidden dirs (.name).                │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│  2. VALIDATE MANIFEST                                       │
│     Each manifest validated against                         │
│     module.manifest.schema.json                             │
│     Required: id (snake_case), version (semver),            │
│               description                                   │
│     Invalid manifest → ModuleLoadError, module skipped      │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│  3. LOAD DATA FILES                                         │
│     For each path in manifest.data:                         │
│       Parse JSON, validate $schema against schema store     │
│       Run referential integrity checks                      │
│     Missing files or invalid JSON → validation failure      │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│  4. LOAD SCHEMA FILES (optional)                            │
│     For each path in manifest.schemas:                      │
│       Parse JSON Schema, inject into shared schema_store    │
│       Available for $ref resolution by core + other modules │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
      Core validation complete → results reported
      Module validation complete → results reported
      Combined: X file(s): Y passed, Z failed

      (Merging is a separate step done by consuming engines
       or test tools, not by validate.py)
```

---

## Merging Rules

When a consuming engine loads modules via `module_loader.merge_data()`, the following rules apply:

### Rule 1: Extend Only (No Core Override)
Modules may **add** new entries to core collections. They may **not** override existing core entries. If a module tries to add an entry with an `id` that already exists in the same core collection, a conflict is flagged.

### Rule 2: No Inter-Module Override
If two modules both contribute an entry with the same `id` in the same collection, a conflict is flagged. The first module loaded "wins" in terms of which entry appears in the merged data, but both are flagged.

### Rule 3: Conflicts Are Reported, Not Silently Resolved
The `merge_data()` function returns a list of conflict tuples. The caller decides how to handle them (warn, abort, skip). The default test harness treats any conflict as a failure.

### Conflict Matrix

| Source A | Source B | Outcome |
|----------|----------|---------|
| core (id=1) | module (id=1) | **Conflict** — module cannot override core |
| module_a (id=100) | module_b (id=100) | **Conflict** — duplicate across modules |
| core (id=1) | core (id=1) | Not possible (single source) |
| module (id=500) | — (new) | **OK** — appended |
| module (no id) | — | **OK** — appended (no id to check) |

---

## Extension Points

Modules may extend the following core collections. These represent the stable, intended extension surface:

| Collection | Extensible? | Notes |
|------------|-------------|-------|
| `features` | **Yes** | Perks, traits, racial traits |
| `equipment` | **Yes** | Weapons, armor, consumables |
| `spells` | **Yes** | New spells using existing effects |
| `skills` | **Yes** | New skills (must follow existing ability association pattern) |
| `races` | **Yes** | New ancestries and lineages |
| `materials` | **Yes** | New crafting materials |
| `gems` | **Yes** | New gems |
| `ingredients` | **Yes** | New alchemy ingredients |
| `monsters` | **Yes** | New creature entries |
| `backgrounds` | **Yes** | New character backgrounds |
| `traits` | **Yes** | New racial traits (should pair with race entries) |
| `effects` | **No** | Core registry — tightly coupled to condition/engine |
| `conditions` | **No** | Core registry — tightly coupled to effects |
| `abilities` | **No** | Fixed attribute set (str/per/end/int/wil/agi/cha/lck) |
| `wuxing` | **No** | Fixed five-phase system |
| `tiers` | **No** | Fixed crafting quality ladder |

Modules may also introduce **new top-level collections** (e.g. `recipes`, `crafting_recipes`, `cultural_styles`). These are loaded and pass through the validation pipeline as long as they declare a `$schema` that exists in the store.

---

## Manifest Format

Every module **must** have a `manifest.json` in its root directory.

```json
{
  "$schema": "../../schemas/module.manifest.schema.json",
  "id":       "example_module",
  "version":  "1.0.0",
  "description": "Demonstration module for the Riftweave module system.",

  "author": "Riftweave Core Team",
  "requires": [],

  "extends": {
    "collections": ["features"],
    "types": {
      "equipment": [],
      "feature":   []
    }
  },

  "data":    ["data/example_features.json"],
  "schemas": []
}
```

| Field | Required | Type | Description |
|-------|----------|------|-------------|
| `$schema` | no | string | Relative path to the manifest schema |
| `id` | yes | snake_case string | Unique identifier for the module |
| `version` | yes | semver string | e.g. "1.0.0", "0.3.1" |
| `description` | yes | string | Human-readable summary |
| `author` | no | string | Creator or maintainer |
| `requires` | no | string[] | Module IDs this module depends on (future use) |
| `extends.collections` | no | string[] | Core collections this module extends |
| `extends.types` | no | object | New enum values added to shared vocabularies |
| `data` | no | string[] | Relative paths to JSON data files |
| `schemas` | no | string[] | Relative paths to JSON Schema files |

---

## Data Files

Data files follow the exact same format as core collection files:

```json
{
  "$schema": "../../../schemas/feature.schema.json",
  "features": [
    {
      "id": 1001,
      "key": "module_demo_perk",
      "label": "Module Demo Perk",
      "description": "A perk added by the example module.",
      "type": "perk",
      "category": "combat",
      "prerequisite": { "abilities": { "str": 5 } },
      "ranks": 1,
      "effects": [
        { "type": "bonus", "target": "melee_damage", "value": 0.05 }
      ]
    }
  ]
}
```

Key rules:
- Each data file **must** declare a `$schema` field pointing to a schema in `ruleset/schemas/`.
- The `$schema` is used by `validate.py` for JSON Schema validation and cross-schema `$ref` resolution.
- Module entry `id` values should be in a high range (e.g. 1000+) to avoid accidental collision with core (which uses 1–64). This is a convention, not enforced by the schema.

---

## Validation Integration

The core validator (`ruleset/scripts/validate.py`) now includes automatic module discovery and validation:

1. **Core pass:** validates all files under `ruleset/data/` (same as before).
2. **Module pass:** discovers modules under `ruleset/modules/`, validates each manifest, then validates each data file against its `$schema`.
3. **Combined ID index:** module entry IDs are included in the referential integrity index, so cross-references between core and modules are checked.
4. **Combined report:** results from both passes are reported together.

```
18 file(s): 18 passed, 0 failed
  Modules: 1 passed, 0 failed (of 1 module data files)
19 file(s): 19 passed, 0 failed
```

If no modules exist, the output is identical to the pre-module behavior:

```
18 file(s): 18 passed, 0 failed
```

---

## Example Module

A working example lives at `ruleset/modules/example-module/`. It adds a single combat perk (id 1001) to the features collection. Validate it:

```bash
python ruleset/scripts/validate.py
```

Load it programmatically:

```python
from module_loader import load_all_modules, merge_data

modules = load_all_modules()
core_data = {}  # or load from ruleset/data/
merged, conflicts = merge_data(core_data, modules)
```

---

## Running Tests

```bash
python -m unittest ruleset/scripts/test_module_system.py
```

The test suite covers (36 tests):
- Manifest schema validation (required fields, format checks, extra fields)
- Module discovery (empty dirs, hidden dirs, multiple modules, duplicate IDs)
- Data file loading (success, missing files, bad JSON, no-id entries)
- Schema file loading and store injection
- Data merging (no conflicts, conflict with core, conflict between modules, new collections, empty core, empty modules)
- Negative cases (invalid ids, invalid versions, extra manifest fields)
- Integration (example module loads and merges cleanly)
- Edge cases (nonexistent dir, hidden dirs, no modules dir)

---

## Known Limitations (v1 Foundation)

- **No dependency resolution** — `requires` is declared but not enforced. Modules are loaded in discovery order only.
- **No schema extension** — module schemas are injected into the store but cannot extend core schema enums. Adding new `equipment.type` values requires a schema update.
- **No override mechanism** — duplicates are flagged but not resolved. An explicit `overrides` policy is a future feature.
- **No module-aware CI step** — the example module is validated in the standard `validate.py` run, but there's no separate CI step for module-only testing.
- **No automatic conflict resolution in merge** — `merge_data()` reports conflicts but does not decide which entry to keep.

These limitations are intentional. The foundation is minimal and mergeable. Each can be addressed in follow-up branches.
