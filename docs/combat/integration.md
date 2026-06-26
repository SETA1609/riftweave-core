# Combat Integration — Effects, Perks, Weapons, and Armor

**Status:** Data + schema support added in `feat/combat-resolution`. Runtime resolution lives in the engine or GM adjudication.

This document explains how the various Riftweave systems **compose** during combat. It is the bridge between `docs/modules/combat.md` (the rules) and the actual data files.

---

## 1. Weapons → Combat

Every weapon in `data/equipment/weapons.json` carries:

- **`weapon.skill`** — the governing skill for attack rolls (`blades`, `blunt`, `piercing`, `bows`, `crossbows`, `guns`, `throwing_weapons`)
- **`weapon.damage.dice`** — polyhedral dice expression (e.g. `1d8`, `2d6+3`)
- **`weapon.damage.type`** — damage type key from `data/combat/damage_types.json`
- **`weapon.length`** + **`weapon.attackReach`** — engagement distance in units
- **`weapon.on_hit_effects`** — optional array of `appliedEffect` shapes (from the shared effect registry) that trigger on a successful hit

### Damage Type → Armor Interaction

The damage type's `category` and `tags` determine whether armor DR applies:

| Category | Armor DR Applies? | Examples |
|----------|------------------|----------|
| `physical` | Yes | slashing, piercing, bludgeoning |
| `elemental` | No (bypasses physical armor) | fire, frost, shock, acid, poison |
| `energy` | No | necrotic, radiant, psychic, force, thunder |

Armor pieces may define `resistances` (per-damage-type flat DR bonuses or penalties) to provide elemental protection:

```json
{
  "armor": {
    "slot": "torso",
    "drBase": 5,
    "resistances": { "fire": -3, "frost": 4 }
  }
}
```

This means the piece gives −3 DR vs fire (vulnerability) and +4 DR vs frost (resistance). See the damage type registry for the full list of type keys.

---

## 2. Effects → Combat

All effects in `data/effects/core.json` with the `"spell"`, `"coating"`, `"enchantment"`, or `"innate"` channel can participate in combat.

### How Effects Trigger During Combat

| Source | Trigger | Mechanism |
|--------|---------|-----------|
| **Weapon on_hit_effects** | On successful hit | `on_hit_effects` array on the weapon spec |
| **Perk effects** | Always active or conditional | `features.effects[]` — applied at loadout, removed on un-equip |
| **Spell effects** | On spell landing | `spells[].effects[]` — resolved vs magic defense tree |
| **Consumable effects** | On use (drink/eat/apply/throw) | `consumable.effects[]` — applied to the user or target |
| **Condition effects** | On condition trigger | Conditions reference `appliedBy` effect IDs; the condition applies the effect's consequences |

### Combat-Specific Effects (Added in feat/combat-resolution)

The following effect IDs were added to `data/effects/core.json` specifically for weapon and combat integration:

| ID | Key | Purpose | Typical Source |
|----|-----|---------|----------------|
| 65 | `armor_pierce` | Partially ignores target armor DR | Piercing weapons, AP ammo |
| 66 | `sunder_armor` | Reduces armor DR for a duration | Warhammers, crushing weapons |
| 67 | `rend` | Damage-over-time, more severe than bleed | Serrated blades, beast claws |
| 68 | `cripple` | Reduces movement speed and AGI | Called shots to legs |
| 69 | `weapon_damage_buff` | Flat bonus to weapon damage rolls | Enchantments, coatings |
| 70 | `haste_attack` | Reduces attack recovery time | Haste spells, potions |
| 71 | `fortify_critical` | Expands crit window or increases crit multiplier | Keen edge, crit potions |
| 72 | `fortify_defense` | Bonus to DR or evasion | Stoneskin, barkskin |
| 73 | `life_steal` | Converts % of damage dealt to health | Vampiric weapons |
| 74 | `mana_leech` | Converts % of damage dealt to mana | Mage weapons |
| 75 | `expose_weakness` | Increases damage taken by target | Hunter's mark |
| 76 | `cleave` | Splashes % damage to adjacent target | Great weapons |
| 77 | `interrupt` | Disrupts target's current action on hit | Precision strikes |

---

## 3. Perks → Combat

Perks in `data/features/core.json` affect combat through their `effects` array. Each effect has a `target` and `value`.

### How Perk Effects Map to Combat Stats

| `target` string | Combat Meaning |
|----------------|----------------|
| `melee_damage` | Multiplier to melee weapon damage (e.g. 0.25 = +25%) |
| `ranged_crit_chance` | Flat % added to ranged crit range (e.g. 10 = +10%) |
| `block_amount` | Multiplier to shield block DR (e.g. 0.15 = +15%) |
| `heavy_armor_surcharge` | Reduction to heavy armor action cost multiplier |
| `medium_armor_surcharge` | Reduction to medium armor action cost multiplier |
| `dodge_stamina_cost` | Multiplier to stamina cost of dodging |
| `initiative` | Flat bonus to initiative rolls |
| `red_magic_magnitude` | Multiplier to Red magic effect magnitude |

When `target` is a numeric effect id (e.g. 32 = `fortify_attribute`, 31 = `fortify_skill`), the perk's effect is delivered through the shared effect registry with the `"innate"` channel. This means perks can grant permanent attribute/skill bonuses, damage resistances, and other effects exactly like any other source.

### Example: Power Attack (id 3)

```json
{
  "effects": [
    { "type": "bonus", "description": "+25% melee weapon damage", "target": "melee_damage", "value": 0.25 }
  ]
}
```

At combat resolution time, the engine sums all active `melee_damage` bonuses from perks and applies the multiplier to the weapon's rolled damage.

---

## 4. Armor → Combat

Armor DR (`drBase + material.modifiers.defense`) provides the primary physical damage mitigation. Key integration points:

- **DR is slot-summed**: Add DR from all worn pieces (head + torso + hands + feet).
- **DR is flat subtraction**: Applied after the hit is confirmed and after any block/parry reduction.
- **Resistances modify per-element DR**: The `resistances` object on each armor piece provides flat DR adjustments per damage type key.
- **Evasion penalties** from armor reduce the character's effective evasion skill.
- **Speed penalties** reduce movement speed (units for TTRPG, % for video game).

Material phase on armor (e.g. steel = `metal`) interacts with incoming phased effects through the Wuxing cycles — see `data/wuxing/core.json` and `docs/modules/magic.md`.

---

## 5. Conditions → Combat

Conditions (`data/conditions/core.json`) are the "envelope" for combat status effects. Each condition has an `appliedBy` array of effect IDs and an `effects` object describing its impact in both TTRPG and video game modes.

Combat-relevant conditions include:

| Condition | Effect ID(s) | Combat Impact |
|-----------|-------------|---------------|
| blinded | 47 | -20 PER checks, attacks against have +10 |
| bleeding | 58 | Damage over time (1 HP/round) |
| burning | 57 | Fire DoT (1d4/round) |
| stunned | 55 | No actions, +10 to attacks against |
| staggered | 64 | Lose reaction, -10 on next action |
| prone | 53 | -20 melee, ranged attacks against +10 |
| exposed | 62 | DR halved, incoming damage ×1.25 |

---

## 6. Resolution Flow (End to End)

```
Attacker declares attack with weapon
  → Resolve attack roll (skill + modifiers - defense vs d100)
    → On hit, determine hit quality (critical/solid/glancing)
      → Roll weapon damage dice
        → Check damage type vs armor (physical? → apply DR)
          → Apply on_hit_effects from weapon
            → Apply perk effects that trigger on hit
              → Defender's defense layers (evasion → block → DR → effects)
                → Final damage applied to defender's HP
                  → Conditions applied (if any)
```

Each step is data-driven — the weapon, damage type, armor, effects, and perks all come from JSON files and schemas.

---

## 7. Adding New Combat Content

### New Weapon
1. Add entry to `data/equipment/weapons.json`
2. Set `weapon.damage.type` to an existing key from `data/combat/damage_types.json`
3. Optionally add `on_hit_effects` referencing effect IDs from `data/effects/core.json`

### New Damage Type
1. Add entry to `data/combat/damage_types.json`
2. Set `category` (physical/elemental/energy)
3. Add appropriate `tags` (armor_dr_applies, bypasses_armor, etc.)
4. Armor pieces can now reference this type in their `resistances`

### New Combat Effect
1. Add entry to `data/effects/core.json`
2. Set `channels` to include relevant delivery methods
3. Add `"combat"` tag for discoverability
4. Weapons can now reference it in `on_hit_effects`

### New Combat Perk
1. Add entry to `data/features/core.json` with `category: "combat"`
2. Set `type: "perk"` (or "creation" for tradeoff traits)
3. Add appropriate `prerequisite` (STR/PER/END gates + skill thresholds)
4. Define `effects` that modify combat-relevant targets
