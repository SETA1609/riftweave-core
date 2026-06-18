# Formal Module System

**Branch Status (ft/modules-foundation):** Foundation implemented — module directory structure, manifest format with JSON Schema, Python loader with discovery/validation/merging/conflict detection, 26 unit tests passing, example module included.

---

## Overview

The Riftweave module system is a **data-oriented** extension framework. Modules are self-describing directories under `ruleset/modules/` that can add new entries to existing core collections, introduce entirely new collections, and provide additional JSON schemas.

The design follows three principles:

1. **Declarative** — a module is just a directory containing a `manifest.json` and data files. No registration code needed.
2. **Composable** — modules merge data with core and each other. Conflict detection flags duplicate IDs explicitly.
3. **Backward-compatible** — the core `ruleset/data/` and `ruleset/schemas/` remain untouched. The validator continues to validate core in isolation.

---

## Module Loading Flow

```
┌─────────────────────────────────────────────────────────┐
│                   ruleset/modules/                       │
│                                                         │
│   example-module/          future-module/               │
│   ├── manifest.json        ├── manifest.json            │
│   ├── data/                ├── data/                    │
│   └── schemas/             └── schemas/                 │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│  1. DISCOVER                                            │
│     Scan modules/ for subdirs containing manifest.json  │
│     Skip hidden dirs (.name) and dirs without manifest  │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│  2. VALIDATE                                            │
│     Each manifest validated against                     │
│     module.manifest.schema.json                         │
│     Required: id (snake_case), version (semver),        │
│               description                               │
│     Optional: author, requires[], extends{},             │
│               data[], schemas[]                         │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│  3. LOAD DATA FILES                                     │
│     For each path in manifest.data:                     │
│       Parse JSON, extract collection arrays             │
│       Store in ModuleInfo.data[coll_name]                │
│     Missing files or invalid JSON → ModuleLoadError     │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│  4. LOAD SCHEMA FILES (optional)                        │
│     For each path in manifest.schemas:                  │
│       Parse JSON Schema                                 │
│       Insert into shared schema_store by file:// URI    │
│       Available for $ref resolution                     │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│  5. MERGE & CONFLICT DETECTION                          │
│     For each module, for each collection:               │
│       Append entries to core collection array           │
│       Track owner (core/module_id) per (coll, id)       │
│       If same (coll, id) appears twice → ConflictError  │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
            ┌─────────────────────┐
            │   Merged data ready │
            │   + conflict report │
            └─────────────────────┘
```

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
      "equipment": ["tool"],
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
| `extends.collections` | no | string[] | Core collections extended (e.g. "features", "equipment") |
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
- The schema is used for validation and cross-schema `$ref` resolution.
- Multiple data files per module are allowed; their entries are merged into the appropriate collections.

---

## Conflict Detection

The merge step tracks every entry ID by collection and source:

| Source A | Source B | Outcome |
|----------|----------|---------|
| core (id=1) | module (id=1) | **Conflict** returned in conflict list |
| module_a (id=100) | module_b (id=100) | **Conflict** returned in conflict list |
| core (id=1) | core (id=1) | Not possible (single source) |
| module (id=500) | — (new) | **OK** — appended |

Conflicts are **not automatically resolved**. The merge function returns a conflict list, and the caller decides how to handle them (skip, warn, abort).

---

## Example Module

A working example lives at `ruleset/modules/example-module/`. It adds a single combat perk (id 1001) to the features collection. Load it with:

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
# or
python ruleset/scripts/test_module_system.py
```

The test suite covers:
- Manifest schema validation (required fields, format checks)
- Module discovery (empty dirs, hidden dirs, multiple modules)
- Data file loading (success, missing files, bad JSON)
- Schema file loading and store injection
- Data merging without conflicts
- Conflict detection (module vs core, module vs module)
- New collection creation from modules
- Edge cases (nonexistent module dir, hidden directories, extra fields)

---

## Known Limitations (v1 Foundation)

- **No dependency resolution** — `requires` is declared but not enforced. Modules are loaded in discovery order only.
- **No schema extension** — module schemas are loaded into the store but cannot extend core schema enums. Adding new `equipment.type` values requires a schema update.
- **No module-level validation** — the core `validate.py` validates core data only. Module data validation requires running the loader + validator manually.
- **No override mechanism** — duplicates are flagged but not resolved. An explicit `overrides` policy is a future feature.
- **No CI integration** — the example module is not validated in CI. The next step is to add a module-aware validation step to CI.

These limitations are intentional — the foundation is meant to be minimal and mergeable. Each can be addressed in follow-up branches.
