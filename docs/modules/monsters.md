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
│   ├── ooze.json         # Mindless jellies, high END, almost no AGI or senses
│   ├── plant.json        # Rooted/fungal creatures, vulnerable to fire
│   └── undead.json       # Risen dead, high WIL, immune to poison
├── templates.json        # Element templates a row applies (fire, frost, storm, acid, venom)
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

The general order and merge rules are in [composition.md](../core/composition.md). For a monster:

1. The base is either a creature type, `bases/<base>.json`, or another monster,
   `core:monster/<key>` in the same file (see [Derived monsters](#derived-monsters)).
2. Templates, if the entry lists any, apply in order (see [Templates](#templates)).
3. Fields from `overrides` are shallow-merged on top (override wins).
4. `id` and `key` come from the entry wrapper, never from the base or overrides.
5. After merge, the entry is validated against the flat monster schema.

### Array Merging: `$extend` Convention

By default, override array fields (`abilities`, `tags`) **replace** the base array entirely.
To **append** to the base array instead, use the `$extend` wrapper:

```json
{
  "base": "elemental",
  "overrides": {
    "abilities": {
      "$extend": [
        { "effect": 57, "magnitude": 4, "duration": 10, "target": "touch" }
      ]
    }
  }
}
```

This adds the burn ability to whatever abilities the base template defines.
`$extend` is useful when base templates define shared abilities and individual
monsters add their own on top.

### Attribute Growth / Level Scaling

Base templates can define a `growth` section that describes how attributes
scale with level:

```json
{
  "key": "beast",
  "level": 2,
  "attributes": { "str": 4, "end": 4, "agi": 5, "per": 4, "wil": 2 },
  "hitPoints": 28,
  "growth": {
    "attributes": {
      "str": { "perLevel": 0.5 },
      "end": { "perLevel": 0.4 },
      "agi": { "perLevel": 0.5 },
      "per": { "perLevel": 0.3 },
      "wil": { "perLevel": 0.2 }
    },
    "hitPointsPerLevel": 6
  }
}
```

The growth formula is:
```
computed = min(10, max(1, round(base_attr + perLevel * (level - base_level))))
```

An attribute stays in its 1 to 10 range, as a character's does, so a level-30 monster
built on a level-4 base tops out at 10 instead of passing the schema's cap. [V 2026-11-06]

Individual monster overrides only need to specify attributes that **differ**
from the growth curve. Missing attributes are filled in automatically:

```json
{
  "key": "dire_wolf",
  "base": "beast",
  "overrides": {
    "level": 3,
    "attributes": { "str": 6, "per": 6 }
  }
}
```

This produces: `{str:6, end:4, agi:6, per:6, wil:2}` — str and per overridden,
end/agi/wil computed from growth.

HP follows the same pattern with `hitPointsPerLevel`. If an override specifies
`hitPoints`, that always takes precedence.

### Derived monsters

A monster that is another monster with changes names that monster as its base:
`"base": "core:monster/ooze"`. It inherits the other monster's merged entry (its attacks,
resistances, tags, everything), so a family is written once and its members only say what
differs. The creature type is the root's: a derived monster cannot change it. [V 2026-11-08]

- **Level.** A derived monster always states its `level` in `overrides`, so a reader of one
  row (the loot budget, the Go loader) never walks the chain. When it differs from the
  base monster's, attributes and HP move by the type's growth for the difference:
  `round(base_attr + perLevel × Δlevel)`, within 1 to 10, and `base_hp + hitPointsPerLevel × Δlevel`.
- **Order.** `core.json` lists the base monsters first (those another row derives from), then
  their variants, family by family, then the one-off monsters, each group by level.
- **Chains** may be longer than one step; a cycle, or a base monster that does not exist, is
  a validation error.

### Templates

A template is a reusable change, written once in `monsters/templates.json`, that any monster
row applies with `"templates": ["fire"]`. The five element templates are one per phase:

| Template | Phase | Damage type | Label |
| --- | --- | --- | --- |
| `fire` | fire | fire | Fire … |
| `frost` | water | frost | Frost … |
| `storm` | metal | shock | Storm … |
| `acid` | earth | acid | Acid … |
| `venom` | wood | poison | Venom … |

A template sets the monster's `phase`; turns every attack into its damage type and gives it
the template's phase; adds its `resist` to the monster's, per damage type; appends its
`tags`; and puts its label prefix before the label, unless `overrides` gives a label.
Templates apply after the base and before `overrides`, so a row can still correct any of
it. [V 2026-11-08]

Fire Ooze is an ooze with the fire template:

```json
{ "id": 46, "key": "fire_ooze", "base": "core:monster/ooze", "templates": ["fire"],
  "overrides": { "level": 5, "hitPoints": 32 } }
```

A family member that is not an element change is a derived monster with overrides only:
an elder treant is the treant sapling at level 22.

### Loot

Each creature type declares a default `loot` list in `bases/`: undead and oozes roll
`grave_goods`, humanoids `bandit_purse`, beasts and plants `beast_den`, constructs and
elementals `hoard_minor`, and familiars `grave_goods`. A monster row writes `loot` only when it
differs, a derived monster inherits its base monster's, and a template may add rows. The rules
are in [loot.md](../core/loot.md#drops).

### Summoned creatures

A summon spell names a group in `summons/core.json`, and each group lists monster rows
([magic.md](./magic.md#summon-groups)). Familiars have their own base, `bases/familiar.json`.
The storm, acid, and venom elementals derive from Fire Elemental and apply an element
template. A summoned creature is its row at the caster's level, resolved as a
[derived monster](#derived-monsters).

### Names

Monster names and descriptions come from folklore, myth, or plain description, never from
a game or film franchise. Product identity of other games (Dungeons & Dragons, Final Fantasy,
Pokémon, Warhammer, and the like) is not used, renamed or not: no beholder, mind flayer
(illithid), displacer beast, owlbear, umber hulk, carrion crawler, gelatinous cube,
githyanki, githzerai, kuo-toa, slaad, yuan-ti, flan, tonberry, cactuar, malboro, or moogle.
`test_bestiary.py` fails on any of these names in a key, label, or description. Folklore
names (troll, ghoul, basilisk, hydra, harpy, golem, wraith) are free to use. [V 2026-11-08]

## Attacks, resistances, and role

A monster fights with the same numbers as a character ([core/combat.md](../core/combat.md)).
[V 2026-11-06]

- **`attacks`** — what it hits with: a `key`, a `label`, a `damage` of an `amount` (the
  authored average, the caller's damage roll) and a `type` (a key in
  `combat/damage_types.json`), a `delivery` (`single`, the default, `area`, or `instant`),
  and an optional `phase` that the attack carries into the five-phase cycles. `abilities`
  stay the effects it applies (conditions, drains).
- **`resist`** — a flat number per damage type, the combat pipeline's `defender.resist`.
- **`role`** — `standard` (the default), `boss`, or `minion`: which hits-to-kill band of
  `balance/core.json` the monster's HP is checked against.

The phase is two-sided. A monster's own `phase` is how incoming phased effects land on it
(its resistance side, through the cycles); an attack's `phase` is how its blows land on
others (its attack side). [A]

### The floor

At least 40 monsters span levels 1 to 30; every phase is a monster's phase and an attack's
phase at least once, and every damage type is an attack and a resistance at least once.
Each monster is a base plus overrides. `ruleset/scripts/test_bestiary.py` checks it, and the
balance gate checks each monster's strongest attack and its HP against its tier and role
([core/balance.md](../core/balance.md)). [V 2026-11-06]

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
2. Detects entries with a `base` field, resolves the creature type or the base monster,
   applies the templates, and merges the overrides.
3. Validates the merged entry against the flat monster schema.
4. Runs referential integrity checks (effect IDs exist, etc.).
