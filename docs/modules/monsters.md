# Monsters — Base + Override System

## Purpose

Monsters in Riftweave use a **Base + Override** pattern to reduce duplication, enable easy variant creation, and remain architecturally ready for a future ECS/composition refactor.

## Data Files

```
ruleset/data/monsters/
├── bases/
│   ├── beast.json        # Physical creatures, low mental stats
│   ├── construct.json    # Animated objects, high END, no mind
│   ├── elemental.json    # Pure elemental beings, extreme phase stats
│   ├── humanoid.json     # Balanced bipedal sentients
│   ├── plant.json        # Rooted/fungal creatures, vulnerable to fire
│   └── undead.json       # Risen dead, high WIL, immune to poison
└── core.json             # Concrete monster entries (flat or base+override)
```

## How It Works

Each entry in `core.json` can be one of two shapes:

### Flat Entry (standalone)

Contains all required fields directly:

```json
{
  "id": 1,
  "key": "ember_drake",
  "label": "Ember Drake",
  "description": "A lesser dragon wreathed in living flame.",
  "phase": "fire",
  "size": "large",
  "level": 6,
  "attributes": { "str": 7, "end": 6, "agi": 6, "per": 5, "wil": 4 },
  "hitPoints": 78,
  "abilities": [ ... ],
  "tags": ["dragon", "elemental"]
}
```

### Base + Override Entry

References a base template and provides only the fields that differ:

```json
{
  "id": 4,
  "key": "frost_revenant",
  "base": "undead",
  "overrides": {
    "label": "Frost Revenant",
    "description": "The risen dead of those who froze in the high passes...",
    "phase": "water",
    "level": 4,
    "attributes": { "str": 5, "end": 6, "agi": 4, "wil": 6 },
    "hitPoints": 52,
    "abilities": [ { "effect": 2, "magnitude": 14, "target": "ranged" } ]
  }
}
```

### Merge Rules

1. The base template is loaded from `bases/<base>.json`.
2. Fields from `overrides` are shallow-merged onto the base (override wins).
3. Arrays (`abilities`, `tags`) are **replaced entirely**, not concatenated.
4. `id` and `key` come from the entry wrapper, never from the base or overrides.
5. After merge, the entry is validated against the flat monster schema.

## Why Base + Override (Not Full Inheritance)

- **Simple and flat** — no deep chains, no diamond problems, no multiple inheritance.
- **Easy to read** — each entry is self-contained after merge; a tool can always produce the flat view.
- **Future ECS ready** — the override fields map naturally to ECS component slots (`combat`, `abilities`, `traits`, `ai`). Each major section of the data could become a component later without restructuring.

## Future ECS Path

The current data shape is intentionally compatible with a future composition system:

```json
{
  "id": 12,
  "components": {
    "identity": { "key": "zombie", "label": "Zombie" },
    "combat": { "attack": 4, "defense": 2, "hitPoints": 30 },
    "abilities": [ { "effect": 46, "magnitude": 4, "target": "touch" } ],
    "traits": ["undead", "mindless"],
    "ai": { "behavior": "melee_charge" }
  }
}
```

To migrate, the merge logic would stay the same — only the container shape changes.

## Validating

Run `python ruleset/scripts/validate.py` as usual. The validator:

1. Skips `bases/` files (they are templates, not standalone data).
2. Detects entries with a `base` field, loads the referenced base, and merges.
3. Validates the merged entry against the flat monster schema.
4. Runs referential integrity checks (effect IDs exist, etc.).
