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
├── templates.json        # Templates a row applies: element, power, and nature
└── core.json             # Concrete monster entries (flat or base+override)
ruleset/data/natural_weapons/
└── core.json             # Bite, claw, slam, and the other natural attacks, with their dice
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
  "base": "core:creature_type/undead",
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
  "base": "core:creature_type/elemental",
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
computed = max(1, round(base_attr + perLevel * (level - base_level)))
```

A monster's attribute has no maximum. 10 is a player's cap at character creation only, and a
player's perks and bonuses pass it too, so a level-30 monster built on a level-4 base keeps
growing. The minimum of 1 stays. [V 2026-11-12, `test_bestiary.py`]

`growth` also holds `damagePerLevel`, 0.25 for every creature type [A]. A monster above its
row's level adds `round(damagePerLevel × Δlevel)` to each attack's damage (`damage.bonus`) and
to the magnitude of each `damage_*` ability, as `hitPointsPerLevel` adds HP. A summoned or
derived creature is where this shows: a fire elemental summoned at level 25 hits 5 harder than
its level-5 row. [V 2026-11-12, `test_bestiary.py`] Whether 0.25 keeps every row in its damage
band at every level is not checked yet. [U]

Individual monster overrides only need to specify attributes that **differ**
from the growth curve. Missing attributes are filled in automatically:

```json
{
  "key": "dire_wolf",
  "base": "core:creature_type/beast",
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
  base monster's, attributes, HP, and damage move by the type's growth for the difference:
  `round(base_attr + perLevel × Δlevel)`, at least 1, `base_hp + hitPointsPerLevel × Δlevel`,
  and `round(damagePerLevel × Δlevel)` on damage.
- **Order.** `core.json` lists the base monsters first (those another row derives from), then
  their variants, family by family, then the one-off monsters, each group by level.
- **Chains** may be longer than one step; a cycle, or a base monster that does not exist, is
  a validation error.

### Templates

A template is a reusable change, written once in `monsters/templates.json`, that any monster
row applies with `"templates": ["fire"]`. Each template has a `slot`, and a row holds at most
one template per slot, its own and its base monster's together. The slots apply in a fixed
order, whatever order the row lists them in: `element`, then `power`, then `nature`.
[V 2026-11-12, `test_bestiary.py`]

| Slot | Templates | What it is |
| --- | --- | --- |
| `element` | `fire`, `frost`, `storm`, `acid`, `venom` | The creature's element: one per phase. |
| `power` | `dire`, `elder`, `ancient` | How far past its kind the creature has grown. |
| `nature` | `demonic`, `angelic`, `otherworldly` | What the creature is touched by. |

A template may carry any of these. Each is optional except `slot`, `label_prefix`, and, for
an element, `phase` and `damage_type`:

- `phase` sets the monster's phase and the phase of every attack.
- `damage_type` sets the damage type of every attack that does not name its own, and turns
  each elemental `damage_*` ability (fire, frost, shock, acid, poison) into the template's
  element. An attack that names its type keeps it, so an elemental's slam stays physical
  under any element. [A]
- `attributes` are added to the monster's, a missing attribute counting 0. An END add moves
  an authored `hitPoints` by the `hp` row's `per_point` (END +1 is +8 HP); mana is computed
  from INT, so an INT add moves it with no step of its own. [A]
- `resist` is added to the monster's, per damage type.
- `spells` are appended to the monster's `spells`, so a creature with a spell list casts as
  a player does, from mana the `mp` row gives it.
- `tags` are appended, skipping any it has, and `loot` rows are added.
- `label_prefix` goes before the label, unless `overrides` gives a label.

Templates apply after the base and before `overrides`, so a row can still correct any of it.
A template never changes the creature type: an undead horse has the `undead` base, not
`beast`, so there is no `undead` template. [V 2026-11-12]

Fire Ooze is an ooze with the fire template:

```json
{ "key": "fire_ooze", "base": "core:monster/ooze", "templates": ["fire"],
  "overrides": { "level": 5, "hitPoints": 32 } }
```

A family member that is not a template change is a derived monster with overrides only:
an elder treant is the treant sapling at level 22.

### Generated monsters

Any row plus a power and a nature is a monster, with no row written for the combination. A
generated monster is named by its row ref and the two template keys,
`{ "monster": "core:monster/ember_drake", "power": "ancient", "nature": "otherworldly" }` [A],
and composes as a derived monster of that row with those templates at the row's level or a
level the caller gives. A slot the row already fills keeps the row's template, so a dire wolf
never becomes dire twice. [A] [V 2026-11-12, `test_bestiary.py`]

Only rows kept as examples are written out. They are worked examples of the player formulas
on a monster: their `hitPoints` is what `derived` gives at average rolls, the `hp` row plus
`(level − 1) × ⌊⌊LCK / 2⌋ / 2⌋` of growth, and their mana is the `mp` row, never stored.

| Row | Built from | Level |
| --- | --- | --- |
| `dog`, `wolf` | the beast type | 3 |
| `dire_wolf` | `wolf` + `dire` | 3 |
| `elder_wolf` | `wolf` + `elder` | 12 |
| `demonic_hound` | `dog` + `demonic` | 6 |
| `undead_horse` | the undead type | 4 |

The player formulas put these rows above the hits-to-kill band of `balance/core.json`, as
[actor-sheet.md](../core/actor-sheet.md) found for every monster; the balance file accepts
them by name until the balance pass. [V 2026-11-12, `balance_report.py`]

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
[derived monster](#derived-monsters), then the caster's power and nature as a
[generated monster](#generated-monsters). The row sets the stats: a summoned dog stays
weaker than a summoned wolf at the same level, whatever goes on top. Summons stay static and
uncraftable: a spell names rows, and only a GM or dev adds a new type.

**Power and nature come from perks.** A perk names what it unlocks in `unlocks`, a list of
refs (`core:monster_template/dire`), because a modifier holds only a number.

- **Power is a ladder.** `summon_dire`, then `summon_elder` (needs `summon_dire`), then
  `summon_ancient` (needs `summon_elder`), by ordinary perk prerequisites. The latest one
  bought replaces the one before, so the caster holds one power.
- **Nature is one at a time.** `summon_demonic`, `summon_angelic`, and `summon_otherworldly`
  need no other nature, and the latest bought replaces the one before; only it is tracked.
  Whether a game locks nature as a one-time choice is left for later. [A]
- **Charisma adds last.** A caster with a summon bonus perk adds `CHA − 5`, never below 0, to the
  attributes of the bonus's group after the power and nature ([magic.md](./magic.md#summon-groups)).

[V 2026-11-12, `test_bestiary.py`]

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

- **`attacks`** — what it hits with. An attack has a `key`, a `label`, and a `weapon`: a
  natural weapon, `core:natural_weapon/<key>`, or, for a creature that carries one, a
  weapon, `core:weapon/<key>`. It rolls that row's dice plus the attribute modifier of M3-12
  (`attribute − 5`, `formulas/attribute_modifier.json`), with no upper limit. `ability` names
  the attribute and defaults to `str`, so a jaguar's claw uses `agi`, as a finesse weapon
  does, and an elephant's stomp uses `str`. `damage.type` is the weapon's unless the attack
  names one; `damage.bonus` is what level growth adds. `delivery` (`single`, the default,
  `area`, or `instant`) and an optional `phase` stay as before. An attack has no `amount`.
  [V 2026-11-12, `test_bestiary.py`]
- **`abilities`** — the effects it applies, from the shared pool. A damaging ability is a
  spell-like ability, below.
- **`spells`** — spells it casts as a player does, `core:spell/<key>`.
- **`resist`** — a flat number per damage type, the combat pipeline's `defender.resist`.
- **`role`** — `standard` (the default), `boss`, or `minion`: which hits-to-kill band of
  `balance/core.json` the monster's HP is checked against.
- **`perks`** — feats from the one `features` table (M3-15): `{ "perk": "core:feature/<key>",
  "rank": 2 }`, the rank 1 when left out and at most the perk's `ranks`. Players and monsters are
  one actor shape, so a monster's perks emit their modifiers at their rank as a player's do (an
  `add` times the rank, a `multiply` of `m` at `1 + (m − 1) × rank`), and `monster_sheet` applies
  them to its derived stats, its authored `hitPoints` included, and its skills. A template's perks
  add to the row's as its spells do, the higher rank kept when both name one; `dire` grants
  Toughness and `elder` Second Wind. Prerequisites are not checked on a monster: an author or a GM
  places a perk on purpose. A perk of the `social` or `utility` category means little on a monster,
  so `validate.py` warns but does not fail. [V 2026-11-14, `monster_sheet` vectors in Python and Go,
  `test_bestiary.py`]
- **`skills`** — only the skills it fights or casts with; any other reads as a player's
  starting value, `5 + attribute × 2`, and an unwritten combat or magic skill grows by 17 per
  level above 1, up to 100, so a level-matched monster hits about as often as a player
  ([actor-sheet.md](../core/actor-sheet.md)). A written skill does not grow.

### Natural weapons

`natural_weapons/core.json` holds each natural attack: `bite`, `claw`, `gore`, `slam`,
`talons`, `sting`, `peck`, `hoof`, `tail`, `tentacle`. A row gives the dice for a medium
creature and a damage type. The dice move along the dice ladder (`combat/dice_ladder.json`)
with the creature's size: tiny −2, small −1, medium 0, large +1, huge +2 steps, stopping at
the ends [A]. A large wolf's bite rolls one step above a medium wolf's. Natural weapons are
not equipment: they cannot be bought, looted, or enchanted, and no loot table names one.
A carried weapon rolls its own row's dice, sized by the weapon's own `size`, not the
creature's [A]. [V 2026-11-12, `test_bestiary.py`]

Every natural weapon is an unarmed attack (M3-16). The file names its skill once,
`skill: core:skill/unarmed`, so a bite, a gore, and a slam roll to hit with `unarmed`, and a
carried weapon rolls with its own row's `skill`. The unarmed perks apply to a natural weapon as to
a fist: `unarmed_damage` adds to the damage, then `unarmed_strike_damage` multiplies it, rounded
down, after the dice, the attribute modifier, and the growth bonus [A]; a carried weapon takes
neither [V 2026-11-14, `test_bestiary.py`]. The table stays because the
attacks differ in dice, damage type, and size; a perk per attack type would be the same table again.
A creature whose row lists no attack and no damaging ability attacks with `core:natural_weapon/slam`
on `str` [A].

A race's natural weapon is the same row: a feature's `natural_weapon` names it,
`core:natural_weapon/<key>` (`catman_claws` names `claw`), and its dice move by the actor's
size on the same steps, so a medium catfolk and a medium jaguar both claw for 1d6 **(breaking)**.
A player's size is its race's `size` [A]. No feature writes its own dice, and `validate.py`
fails a `natural_weapon` ref that does not resolve. [V 2026-11-14, `test_bestiary.py`]

### Spell-like abilities

An attack that is neither natural nor carried, such as a breath, a bolt, or an ooze's
touch, is a spell-like ability: an entry in `abilities` whose effect deals damage. It pays as
a cast does: mana by the spell cost formula on its effect ([magic.md](./magic.md#cost-formula)),
and at a table the `cast` action's AP. With too little mana it is not cast. A creature whose
only attack is a spell-like ability, such as an ooze, pays the AP and no mana, so it never runs
out of attacks [A]. A spell-like ability takes no attribute modifier, since none names an
attribute yet [A].

- **Oozes** have no weapon: an ooze fights with a touch-target acid ability, and the element
  templates turn it to fire or frost.
- **Elementals** have a physical attack, a `slam` that deals bludgeoning damage, plus two
  spell-like abilities of their element: a bolt (target `ranged`) and a touch (target
  `touch`). The elemental type has INT 5, growing 0.4 a level, so it has mana for them. A
  venom elemental's bolt and touch use `damage_poison`, which has the spell channel.
  This holds for the five elementals of the `elemental` summon group. [V 2026-11-12, `test_bestiary.py`]

`balance_report.py` reads a monster's damage as the dice mean plus the modifier plus the bonus,
and a damaging ability's magnitude. [V 2026-11-12]

The phase is two-sided. A monster's own `phase` is how incoming phased effects land on it
(its resistance side, through the cycles); an attack's `phase` is how its blows land on
others (its attack side). [A]

### The floor

At least 40 monsters span levels 1 to 30; every phase is a monster's phase and an attack's
phase at least once, and every damage type is an attack and a resistance at least once.
Every monster has an attack or a damaging ability.
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
