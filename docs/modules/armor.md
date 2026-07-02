# Armor System

**Status:** Core (implemented as data + schema) · runtime resolution lives in the engine or GM adjudication.

The armor system replaces the legacy D&D-style AC model with a **damage reduction (DR)**
model. Armor is organized by **slot**, **layer**, and **material** — each piece provides
a flat DR value that stacks with other worn pieces.

Data lives in [`ruleset/data/equipment/armor.json`](../../ruleset/data/equipment/armor.json),
governed by `equipment.schema.json`. Materials that modify armor are in
[`ruleset/data/materials/core.json`](../../ruleset/data/materials/core.json).

For the combat resolution system (how DR interacts with evasion and blocking) see
[`combat.md`](./combat.md).

---

## 1. Armor Slots

Armor is divided into these equipment slots. Total DR is the sum of head + torso +
hands + feet (each modified by material):

| Slot | Coverage | Examples | DR Share |
|------|----------|----------|----------|
| **Head** | Helmet, cap, circlet, hood | Helmet, Coif, Crown | 15–20% of total |
| **Torso** | Chest, shoulders | Breastplate, Chainmail, Robe | 50–60% of total |
| **Back** | Cape, cloak, backpack | Cloak, Backpack, Quiver | 0 DR (utility) |
| **Hands** | Gloves, gauntlets, bracers | Leather Gloves, Steel Gauntlets | 5–10% of total |
| **Feet** | Boots, shoes, greaves | Leather Boots, Plate Greaves | 10–15% of total |

Back slot items do not contribute base DR but may carry enchantments or provide
utility (storage, warmth, camouflage). Shields (see §5) occupy the offhand slot
and are not part of the DR sum.

### Accessories

Up to **5 accessories** may be worn simultaneously, but only **one of each type**:

- 1 ring
- 1 amulet / necklace
- 1 trinket / charm
- 1 belt
- 1 miscellaneous (brooch, badge, etc.)

Accessories have 0 base DR and provide no combat protection on their own. They
are magical items that apply effects from the shared effects registry (see
[`data/effects/core.json`](../../ruleset/data/effects/core.json)).

Accessory data uses `type: "accessory"` with an `accessory` object:
```json
{
  "id": 100, "key": "ring_of_protection", "label": "Ring of Protection",
  "type": "accessory",
  "cost": { "quantity": 500, "unit": "gp" }, "weight": 0.1,
  "accessory": {
    "accessoryType": "ring",
    "effects": [
      { "effect": 19, "magnitude": 10, "parameter": "physical" }
    ]
  }
}
```

---

## 2. Armor Layering

Multiple items can occupy the same slot if they belong to different **layers**.
Each layer adds its material-modified DR to the total.

| Layer | Position | Examples | Notes |
|-------|----------|----------|-------|
| **Skin** | Against body | Padded tunic, cloth robe, undercoat | Clothes are always skin layer. Lowest DR contribution per material. |
| **Middle** | Over skin layer | Chainmail, scale shirt, brigandine | Medium DR. **Requires** a skin layer underneath (penalty if missing). |
| **Upper** | Outermost (armor) | Breastplate, plate cuirass, leather jerkin | Highest DR per slot. Can be worn alone or over middle layer. |
| **On Top** | Over everything | Cloak, cape, hooded mantle, surcoat | Utility layer. 0 DR (or minimal). Can hold enchantments. No layer dependency. |

Slot + layer = **16 possible equipment positions** (head/torso/hands/feet × 4 layers)
+ back slot + 5 accessories.

### Layer Dependency Rules

- **Middle layer requires a skin layer** in the same slot. Missing it imposes a
  −1 AGI penalty (chafing, discomfort, poor fit).
- **Upper layer** can be worn alone or over middle layer. No penalty either way.
- **On Top layer** can be worn over any combination. No layer dependencies.
- **Back slot** does not participate in layering. Items sit over the torso's
  on_top layer.
- **Accessories** have no layer — they are independent of the armor layering system.

### Examples

| Character | Equipment | Total DR |
|-----------|-----------|----------|
| Peasant | Cloth tunic (torso/skin) + Cloth hat (head/skin) + Shoes (feet/skin) | 0 |
| Light footman | Padded gambeson (torso/skin, leather) + Chainmail (torso/middle, iron) + Leather cap (head/skin) + Gloves (hands/skin) + Boots (feet/skin) | ~5 |
| Knight | Padded undercoat (torso/skin, cloth) + Chainmail (torso/middle, steel) + Plate cuirass (torso/upper, steel) + Steel helm (head/upper) + Gauntlets (hands/upper) + Greaves (feet/upper) + Wool cloak (back/on_top) | ~14 |

---

## 3. Base Types + Materials

Like weapons, armor entries are **generic base types** that materials modify.
The base entry provides the slot, layer, and base DR. The chosen material adds
its `modifiers.defense` to the base DR and sets the piece's phase and weight.

### Torso

| Base Type | Layer | Category | Base DR | Skill | Properties |
|-----------|-------|----------|---------|-------|------------|
| Padded tunic | Skin | Light | 0 | light_armor | — |
| Cloth robe | Skin | Light | 0 | light_armor | — |
| Leather jerkin | Upper | Light | 1 | light_armor | — |
| Studded leather jerkin | Upper | Light | 2 | light_armor | — |
| Chainmail | Middle | Medium | 2 | medium_armor | Requires skin, evasion −5, speed −1 |
| Scale shirt | Middle | Medium | 3 | medium_armor | Requires skin, evasion −5, speed −1 |
| Brigandine | Middle | Medium | 3 | medium_armor | Requires skin, evasion −5, speed −1 |
| Breastplate | Upper | Heavy | 4 | heavy_armor | Evasion −15, speed −2 |
| Plate cuirass | Upper | Heavy | 5 | heavy_armor | Evasion −15, speed −2, STR 8+ |
| Full plate | Upper | Heavy | 6 | heavy_armor | Evasion −15, speed −2, STR 10+, stealth penalty |

### Head

| Base Type | Layer | Category | Base DR | Skill | Properties |
|-----------|-------|----------|---------|-------|------------|
| Cap/Hood | Skin | Light | 0 | light_armor | — |
| Coif | Middle | Medium | 1 | medium_armor | Requires skin, evasion −5 |
| Helmet | Upper | Heavy | 2 | heavy_armor | Evasion −15 |

### Hands

| Base Type | Layer | Category | Base DR | Skill | Properties |
|-----------|-------|----------|---------|-------|------------|
| Gloves | Skin | Light | 0 | light_armor | — |
| Gauntlets | Upper | Heavy | 1 | heavy_armor | Evasion −15 |

### Feet

| Base Type | Layer | Category | Base DR | Skill | Properties |
|-----------|-------|----------|---------|-------|------------|
| Shoes | Skin | Light | 0 | light_armor | — |
| Boots | Skin | Light | 1 | light_armor | — |
| Greaves | Upper | Heavy | 2 | heavy_armor | Evasion −15, speed −2 |

### Back

| Base Type | Layer | Base DR | Notes |
|-----------|-------|---------|-------|
| Cloak | On Top | 0 | Can hold enchantments, no DR contribution |
| Backpack | On Top | 0 | Utility (storage), no DR contribution |

### Material Modifiers (per piece)

When an armor piece is crafted from a material:

- `material.modifiers.defense` adds to the piece's base DR.
  - e.g. iron chainmail: base DR 2 + iron's defense 3 = **5 DR**
  - steel plate cuirass: base DR 5 + steel's defense 5 = **10 DR**
  - leather jerkin: base DR 1 + leather's defense 1 = **2 DR**
- `material.weightFactor` multiplies the piece's base weight.
- `material.phase` sets the piece's phase for Wuxing interactions.
- Materials that do not include `"armor"` in their `appliesTo` cannot be used for
  armor pieces.

**DR formula for a single piece:**
```
Piece DR = drBase + material.modifiers.defense
```

**Total character DR:**
```
Total DR = sum of all worn armor pieces' Piece DR (head + torso + hands + feet)
```

### Concrete Examples (Material × Base Type)

The table below shows how different materials modify the same base types into meaningfully different armor sets:

| Character | Torso | Head | Hands | Feet | Shield | Material | Total DR |
|-----------|-------|------|-------|------|--------|----------|----------|
| Light skirmisher | Leather jerkin (DR 1) | Cap/hood (DR 0) | Gloves (DR 0) | Boots (DR 1) | Buckler (DR 1) | Leather (+1) | **4** (1+1+0+1+1) + leather ×4 = 1+1+0+1+1 = 4 |
| Iron footman | Chainmail (DR 2) | Coif (DR 1) | Gloves (DR 0) | Boots (DR 1) | Standard shield (DR 2) | Iron (+3) | **13** (5+4+3+4+5) |
| Steel knight | Plate cuirass (DR 5) | Helmet (DR 2) | Gauntlets (DR 1) | Greaves (DR 2) | Standard shield (DR 2) | Steel (+5) | **22** (10+7+6+7+7) |
| Bronze frontliner | Brigandine (DR 3) | Helmet (DR 2) | Gauntlets (DR 1) | Greaves (DR 2) | Tower shield (DR 4) | Bronze (+2) | **17** (5+4+3+4+6) |
| Wood-scavenged | Studded leather (DR 2) | Cap/hood (DR 0) | Gloves (DR 0) | Boots (DR 1) | Buckler (DR 1) | Wood (+0) | **4** (2+0+0+1+1) |
| Elven skirmisher | Chainmail (DR 2) | Coif (DR 1) | Gloves (DR 0) | Boots (DR 1) | — | Elven wood (+3) | **10** (5+4+3+4) |

**Calculation walkthrough (Steel Knight):**
- Steel plate cuirass: drBase 5 + steel defense 5 = **10 DR**
- Steel helmet: drBase 2 + steel defense 5 = **7 DR**
- Steel gauntlets: drBase 1 + steel defense 5 = **6 DR** (errata: steel's attack/defense is 5, see materials/core.json)
- Steel greaves: drBase 2 + steel defense 5 = **7 DR**
- Steel standard shield: drBase 2 + steel defense 5 = **7 DR** (active block only)
- **Total passive DR: 30** (10+7+6+7). With shield block: adds 7 on that hit.
- **Evasion penalty:** −15 (plate) + 0 (cap) −15 (gauntlets) −15 (greaves) −10 (shield) = −55 total. This is why knights rely on block, not evasion.

**Calculation walkthrough (Iron Footman):**
- Iron chainmail: drBase 2 + iron defense 3 = **5 DR**
- Iron coif: drBase 1 + iron defense 3 = **4 DR**
- Iron gloves: drBase 0 + iron defense 3 = **3 DR** (skin layer, but defense applies)
- Iron boots: drBase 1 + iron defense 3 = **4 DR**
- Iron standard shield: drBase 2 + iron defense 3 = **5 DR** (active block only)
- **Total passive DR: 16** (5+4+3+4)
- **Evasion penalty:** −5 (chainmail) + 0 (coif) + 0 (gloves) + 0 (boots) −10 (shield) = −15 total. Manageable for a frontliner.

---

## 4. Governing Skills

Each armor piece's `category` determines which skill governs wearing it effectively:

| Category | Skill | Associated Ability | Penalties for Low Skill |
|----------|-------|-------------------|------------------------|
| Light | light_armor | AGI | Reduced evasion bonus, minor fatigue |
| Medium | medium_armor | END | Higher stamina cost for actions, evasion penalty increased |
| Heavy | heavy_armor | STR | Speed penalty increased, cannot sprint, skill checks at disadvantage |

Unarmored characters rely on the **unarmored** skill (AGI), which provides a bonus
to their evasion layer (see [`combat.md`](./combat.md) §3).

Each piece also carries an `evasionPenalty` that directly reduces the wearer's
evasion skill. Cumulative across all worn pieces.

---

## 5. Shields

Shields are a separate equipment type (`type: "shield"`) governed by the **block**
skill (END). They occupy the **offhand** slot and do not contribute to passive DR.
Instead, their DR value applies only during an **active block** action
(see [`combat.md`](./combat.md) §3 — Layer 2).

### Active Block vs. Passive DR

- **Shield DR is NOT added to the passive DR total.** A shield provides no benefit
  unless the wearer actively blocks.
- **Active block:** A reaction declared when an attack is incoming. The blocker
  makes a block skill check; on success, the shield's DR (base + material modifier)
  is subtracted from incoming damage. On failure, no block benefit.
- **Shield DR does not stack with armor DR during a block** — the shield's DR
  replaces the armor DR contribution for that specific hit (whichever is higher
  applies). Alternatively, if using additive DR: shield DR adds on top of armor DR
  for that hit (GM/engine choice, set once per campaign).

### Shield Types

| Type | Base DR | Properties | Skill |
|------|---------|------------|-------|
| **Buckler** | 1 | Light (0.5 weight). Allows light off-hand item (dagger, wand). −5 evasion penalty. | block |
| **Standard shield** | 2 | Normal weight (1.0). Occupies off-hand fully. −10 evasion penalty. | block |
| **Tower shield** | 4 | Heavy (2.0 weight). Grants half cover to adjacent allies. −20 evasion penalty. Speed −1. STR 8+ required. | block |

Shield DR is further modified by material (`material.modifiers.defense`) just
like armor pieces.

### Parry (Weapon Block)

Some weapons (rapiers, longswords, fencing swords) and the *Duelist* perk allow
a **parry** instead of a shield block:

- **Parry uses the weapon skill** (e.g. blades) instead of block skill.
- **On success:** The attack is negated entirely (not reduced).
- **On failure:** The attack hits with full effect (no DR reduction).
- Parry does not require a shield, but requires a free hand or a light off-hand
  weapon.
- **Two-weapon fighting:** A character wielding two weapons may use the off-hand
  weapon to parry. The off-hand can either attack (bonus action) or be reserved
  to parry — **not both** in the same turn.
- Perks can improve parry windows and success chances.

### Shield Crafting

- Shields can be crafted from materials that include `"armor"` in their `appliesTo`.
- Material phase applies for Wuxing interactions (e.g. a steel shield has metal
  phase, affecting incoming phased attacks via the overcoming cycle).
- Material weight factor multiplies the shield's base weight, affecting encumbrance.

---

## 6. Unarmored Defense

A character wearing **no armor pieces** in any slot uses the **unarmored** skill
(AGI) as their sole physical defense layer (see [`combat.md`](./combat.md) §3).
This provides:

- A bonus to evasion checks (active or passive).
- No DR — every hit that lands deals full damage.
- No evasion, speed, or stealth penalties.
- Full AGI contribution to movement speed and action economy.

Characters may wear clothing (skin layer items with 0 DR) and still be considered
unarmored for skill purposes — the unarmored skill applies as long as no upper,
middle, or heavy pieces are worn.

---

## 7. Schema Reference

### Armor Object (`equipment.schema.json`)

```json
{
  "slot":       "head" | "torso" | "back" | "hands" | "feet" | "offhand",
  "layer":      "skin" | "middle" | "upper" | "on_top",
  "category":   "light" | "medium" | "heavy",          // omit for back/on_top
  "drBase":     integer (base DR before material modifier),
  "evasionPenalty": integer (0 for light, −5 for medium, −15 for heavy),
  "speedPenalty":   integer (0 for light, −1 for medium, −2 for heavy),
  "strengthRequirement": integer (1–10, optional),
  "stealthPenalty": boolean (optional, default false),
  "requiresLayer": ["skin"] | ["middle"] | null         // layer dependency
}
```

### Accessory Object

```json
{
  "accessoryType": "ring" | "amulet" | "trinket" | "belt" | "misc",
  "effects": [ { "effect": "core:effect/<key>", "magnitude": number, "parameter": "..." } ]
}
```

---

## 8. Crafting & Materials Integration

The armor system reuses the existing material system (see
[`materials/core.json`](../../ruleset/data/materials/core.json) and
[`materials.md`](./materials.md)):

- A material's `modifiers.defense` adds to the base DR of any armor piece crafted
  from it.
- A material's `phase` sets the piece's element for Wuxing interactions.
- Armor pieces can only be crafted from materials that include `"armor"` in their
  `appliesTo` array.
- Material tiers, weight factors, and value factors apply identically to armor
  as they do to weapons.

Current armor-applicable materials: leather (wood, DR +1), copper (metal, DR +1),
bronze (metal, DR +2), iron (metal, DR +3), steel (metal, DR +5), wood (wood, DR +0),
elven wood (wood, DR +3). Silver and gold are accessory-only (too soft for armor).
Obsidian is weapon-only (too brittle).

---

## 9. Encumbrance

Total encumbrance = sum of (base weight × material.weightFactor) for all worn items.

| Total Weight (lbs) | Effect |
|--------------------|--------|
| ≤ STR × 10 | Unencumbered — normal speed and evasion |
| > STR × 10 | Encumbered — speed −1, evasion penalty doubled |
| > STR × 15 | Heavily encumbered — speed −2, evasion penalty tripled, cannot sprint or dodge |
| > STR × 20 | Overburdened — cannot move, cannot take actions other than dropping items |

Each armor piece's `evasionPenalty` and `speedPenalty` apply independently of
encumbrance — they stack additively with encumbrance penalties.

### Combat Interaction

See [`combat.md`](./combat.md) §9 for how encumbrance affects combat:
- Speed reduction (units for TTRPG, % for video game).
- Evasion penalty multiplication (doubled at encumbered, tripled at heavily encumbered).
- Action restrictions (Dash and Dodge disabled at heavily encumbered).
- Stamina regeneration penalties (video game mode).
- Detailed worked examples in [`combat.md`](./combat.md) §10.
