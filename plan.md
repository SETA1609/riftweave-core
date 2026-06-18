# Plan: Implement Full Combat Resolution System

**Branch Status (ft/combat):** Core combat resolution system complete. See `docs/modules/combat.md` for the full specification. Validation passes. New combat perks added to data. This plan document is historical — the combat system has been implemented.

**Priority:** High — blocks both TTRPG and game engine delivery.
**Labels:** core-mechanics, combat, ttrpg, engine

---

## Current State Audit

- **Resolution:** d100 roll-under, margin of success, crit on 01–05, fumble on 96–00
- **Derived stats:** HP (END×8), MP (INT×8), AP (AGI, turn-based) or Stamina (END, action-combat), initiative=PER, move speed from race
- **Weapons:** damage.type→skill, length/reach, base 3.5 dice + material attack bonuses
- **Materials:** `modifiers.attack` (0–+6 range) and `modifiers.defense` for armor; `appliesTo` restricts to weapon/armor/accessory
- **Effects registry:** 46 effects including `resist` (id 19), `cure` (id 20), `spell_absorption` (id 21), `fortify_*`, `drain_*`, `damage_*`, poisons, diseases
- **Armor:** legacy AC model (`acBase`, `acDexBonus`) — progression.md explicitly calls out "armor is still closer to legacy shape"
- **Skills (combat):** `light_armor` (agi), `medium_armor` (end), `heavy_armor` (str), `unarmored` (agi), `block` (end), `evasion` (agi) — **already split**; weapons skills unchanged
- **No action economy, no conditions, no cover/flanking/terrain, no formal defense model**

### Attribute Skill Seeds That Need Updating

The newly split armor/defense skills are not reflected in `attributes.md`:

| Attribute | Currently Lists | Should Also List |
|-----------|----------------|------------------|
| STR | blades, blunt, piercing, athletics | **heavy_armor** |
| END | block | **medium_armor** |
| AGI | unarmed, stealth, lockpick, sleight_of_hand | **light_armor, unarmored, evasion** |

---

## Design Principle: Dual Resolution

Every subsystem specifies **two resolution modes**:

- **TTRPG (turn-based):** GM-facing rules, dice rolls, adjudication, AP-based action economy
- **Video game (action-combat):** Engine-facing rules, player input drives all actions, real-time, Stamina-based

Where a mechanic is identical in both modes, say so explicitly. For **cRPG** (party-based, tactical): lean closer to TTRPG resolution with automation.

---

## Deliverables

### 1. `docs/modules/combat.md` — Core Combat Module

#### 1a. Action Economy

**TTRPG (turn-based):**
- Structure: Start of turn → Action → Bonus action → Movement (split) → Reaction (once per round)
- Actions: Attack, Cast Spell, Dash (×2 movement), Disengage (no OA), Dodge (evasion bonus), Hide (stealth), Ready (triggered action), Use Object, Help
- Bonus actions: off-hand light weapon attack, certain perks/spells
- Reaction: one per round; opportunity attack, Parry spell, triggered perks
- Movement: speed in units, split before/after actions
- **AP pool variant:** AGI-based AP (max 10, carry over between turns) can replace discrete action types for groups that prefer a resource-budget model

**Video game (action-combat):**
- No discrete turn structure. **Player input drives every action** — there is no automated backswing or auto-trigger.
- Attacks, blocks, special moves: mapped to buttons/inputs. Each drains Stamina (END pool, regen via WIL).
- Movement: speed rating, always active. Sprinting/dodging costs Stamina.
- No "action types" — the player decides what to do moment-to-moment.
- Cooldowns and Stamina costs gate ability use instead of "bonus action" / "reaction" slots.
- **cRPG note:** If implementing a party-based tactical cRPG, use TTRPG-style discrete actions with AI-controlled companions.

#### 1b. Opportunity Attacks / Reactive Strikes

**Universal trigger:** leaving an opponent's reach without using Disengage.

**TTRPG:**
- Provoking movement: exiting threatened squares/units (reach weapon = 2 unit threat).
- Resolve: single melee attack as reaction, resolved as a normal attack roll.
- Disengage prevents OA. Some perks grant OA on other triggers (casting, standing up).

**Video game (action RPG):**
- **No auto-triggered backswing.** The player controls hit timing entirely.
- When an enemy disengages while in melee range, the player's next attack against that target within a short window gets a **bonus** (+10 attack or +25% damage) — rewarding player awareness, not automating it.
- Perks can widen the bonus window or add additional effects.
- **cRPG note:** For party-based cRPGs, use TTRPG OA rules (discrete, reaction-based) with optional auto-pause on trigger.

#### 1c. Defense Model — Two Trees

Both trees consider **active effects** on the player (from the effects registry, e.g. `resist` id 19, `spell_absorption` id 21) as modifiers at their respective layers.

```
PHYSICAL DEFENSE TREE

Attack hits (d100 ≤ weapon skill)
  → Layer 1: Evasion (evasion skill + AGI) — chance to avoid entirely
    → Layer 2: Block (block skill + shield DR) — reaction, reduces/negates on success
      → Layer 3: Armor DR (sum of all worn armor pieces, modified by materials)
        → Layer 4: Active effects (resist physical, stoneskin, fortify_endurance, etc.)
          → Final physical damage

MAGIC DEFENSE TREE

Spell lands (d100 ≤ cast check)
  → Layer 1: Evasion* — evasion can apply to magic **only with a perk** (not baseline)
  → Layer 2: Spell Absorption (spell_absorption effect id 21, WIL-based %) — % chance to absorb into mana
    → Layer 3: Magic Resistance (resist effect id 19 with parameter "magic" + WIL) — % reduction or resist roll
      → Layer 4: Elemental/Phase Resistance (resist effect id 19 with parameter = phase/type) — per-element DR
        → Layer 5: Active effects (ward spells, aura effects, fortify_wil)
          → Final magical damage
```

*Baseline evasion only applies to physical attacks. A perk (e.g. "Spell Dodger" or similar) unlocks Layer 1 in the magic tree, allowing evasion skill to contribute against targeted spells.

**TTRPG:**
- Evasion: passive penalty to attacker's roll, or active dodge as reaction. Magic evasion only with perk.
- Block: declared reaction before attack roll; skill check vs attacker's margin.
- Magic: each layer is a separate roll or % reduction applied sequentially.
- Armor DR: flat subtraction from incoming physical damage.

**Video game:**
- Evasion: % dodge chance from evasion skill + AGI; auto-rolled per attack. Magic evasion only with perk.
- Block: input-timed block/parry window; skill determines window length and DR %. Parry = higher risk/reward (reflect or stun) — **optional, not enforced**.
- Magic: % mitigation computed from effects and stats; applied automatically.
- Armor DR: flat reduction per hit from all worn armor.
- Active effects: visible buff icons, auto-applied to combat formulas.

#### 1d. Critical Hits & Fumbles

**Universal base:** critical success on 01–05. Critical confirmation roll required.

**Confirmation system:**
- On a natural 01–05 (the **critical window**), roll again.
- If the confirmation roll also succeeds (≤ modified target number), it is a confirmed critical.
- If the confirmation roll fails, it is a normal hit.
- **Luck** helps: `+floor(LCK / 2)` to the confirmation roll, and every 2 points of LCK expands the critical window by +1 (e.g. LCK 10 → window is 01–10).

**TTRPG:**
- Confirmed crit: roll damage dice twice, sum both (or max + roll).
- Optional special effects table (disarm, bleed, stagger, sever limb) — GM discretion.
- **Fumble:** only on a natural **100** (not 96–00). Flavor-only for TTRPG (drop weapon, stumble). No mechanical penalty enforced.
- Luck does not affect fumble range.

**Video game:**
- Confirmed crit: fixed multiplier (×1.5 or ×2) applied after DR.
- **No fumble mechanic.** A roll of 100 is simply a miss (or a "glancing blow" that deals 0).
- Luck feeds into crit chance formula via expanded window and confirmation bonus.

#### 1e. Cover, Flanking, Difficult Terrain

**TTRPG:**
- Cover: none / half (−10 attack) / three-quarters (−20) / full (immune).
- Flanking: +10 attack when allies on opposite sides of a melee target.
- Difficult terrain: costs double movement; cannot Disengage while in it.
- Prone: −20 melee attack, +10 vs ranged, half move to stand.

**Video game (action RPG):**
- **No formal cover system.** Environment has LoS blockers (walls, pillars, terrain) that block projectiles naturally. No to-hit bonus/penalty from cover.
- Flanking: damage % bonus when attacking from behind or opposite an ally (e.g. +25%). Calculation is fine; automation works.
- Difficult terrain: slow effect (% speed reduction) while traversing. Auto-applied.
- Prone: knockdown state with get-up animation duration.
- **cRPG note:** If implementing a tactical cRPG with grid/positioning, use TTRPG cover and prone rules throughout.

#### 1f. Initiative

**TTRPG:** `d20 + PER + modifiers` (or `d10 + floor(PER/2)` for tighter spread). Rolled each combat or static per encounter.
**Video game:** Not needed (real-time). Aggro/threat or proximity system handles engagement.

#### 1g. Mode Comparison Table

| Aspect | TTRPG (turn-based) | Video game (action-combat) |
|--------|-------------------|---------------------------|
| Actions | Discrete Action/Bonus/Reaction | Player input per button; Stamina cost |
| Movement | Speed in units per action | Speed rating, free + sprint costs Stamina |
| Reactions | 1/round, declared | Timed block/parry input (optional); cooldowns |
| OA | Auto-trigger on disengage | Bonus on next attack vs disengaging enemy |
| Initiative | d20 + PER | Not needed (aggro/threat) |
| Defense | Evasion check, Block reaction | Auto-dodge %, timed block/parry |
| Magic defense | Sequential rolls per layer (evasion only with perk) | Auto-mitigation from stats + effects (evasion only with perk) |
| Crit | Confirmation roll, double dice | Confirmation roll, ×1.5–2 multiplier |
| Fumble | On 100 only; flavor table | None — 100 is a miss |
| Cover | −10/−20 to-hit | LoS blockers only (no to-hit modifier) |
| Flanking | +10 to-hit | +25% damage |
| Terrain | Double movement cost | % speed slow |
| Conditions | GM tracks duration | Automated timers, visual indicators |

---

### 2. Conditions System

#### 2a. Design: Effects-Driven Conditions

Conditions **must reference effects from the shared effects registry** (`data/effects/core.json`) rather than describing their own mechanics inline. Each condition is a binary state container that is applied/removed by specific effects.

**How it works:**
- A condition entry declares which **effect(s)** apply it (e.g. `paralyze` effect id 11 applies the `paralyzed` condition).
- The same `cure` effect (id 20) that cures poisons/diseases also removes conditions via `parameter`.
- The condition's mechanical effects (penalties, immunities) are described in the doc and the condition data entry — but the **application** always flows through an effect.

**Condition data entry shape:**
```json
{
  "id": 1,
  "key": "blinded",
  "label": "Blinded",
  "description": "Cannot see. Vision-dependent checks auto-fail.",
  "effects": {
    "ttrpg": "−20 on PER-based checks, auto-fail sight- dependent rolls. Attacks against you have +10.",
    "video_game": "−20% accuracy, enemy attacks +10% crit chance."
  },
  "appliedBy": [ 8, 46 ],         // effect ids from effects registry (e.g. damage effects, poison)
  "removedBy": [ 20 ],             // cure effect id 20, with parameter matching
  "stacking": false,
  "maxStacks": 1
}
```

#### 2b. Core Conditions (15)

| # | Key | appliedBy (effect ids) | removedBy |
|---|-----|----------------------|-----------|
| 1 | blinded | 8 (invisibility inverse?), poison effects | 20 (cure with parameter "blinded") |
| 2 | charmed | yellow magic control effects | 20 (cure "charmed"), damage breaks |
| 3 | deafened | damage effects, environmental | 20 (cure "deafened") |
| 4 | frightened | phobia 22, control effects | 20 (cure "frightened") |
| 5 | grappled | innate/monster effects | 20 (cure "grappled"), athletics escape |
| 6 | incapacitated | damage when HP=0, control effects | 20 (cure "incapacitated"), healing |
| 7 | invisible | 8 (invisibility) | 20 (cure "invisible"), attacking breaks |
| 8 | paralyzed | 11 (paralyze), 43 (paralytic poison) | 20 (cure "paralyzed") |
| 9 | petrified | petrify spell/effect | 20 (cure "petrified"), greater magic |
| 10 | poisoned | 43–46 (poison effects) | 20 (cure "poison") |
| 11 | prone | knockdown effects, monster attacks | stand action |
| 12 | restrained | control effects, web/entangle | 20 (cure "restrained"), STR check |
| 13 | stunned | shock damage 3, control effects | 20 (cure "stunned") |
| 14 | unconscious | HP=0, sleep effects | 20 (cure "unconscious"), healing |
| 15 | exhaustion | 40 (blight), fatigue effects, stacking | rest, 20 (cure "exhaustion") |

#### 2c. Resolution

**TTRPG:** Conditions are binary states with explicit duration tracked by the GM. Applied when the corresponding effect successfully lands on a target. Cured via rest, specific spells, or `cure` effects (id 20).

**Video game:** Conditions are timed debuff states. Applied automatically when an effect procs. Duration tracked by engine. Visual/audio indicators. Stacking conditions have intensity counters.

#### 2d. Schema: `ruleset/schemas/condition.schema.json`

```
{
  condition: {
    id, key, label, description,
    effects: { ttrpg: string, video_game: string },
    appliedBy: [entryId],   // effect ids that apply this condition
    removedBy: [entryId],   // effect ids that remove it (typically cure id 20)
    stacking: boolean,
    maxStacks: integer
  }
}
```

#### 2e. Doc: `docs/modules/conditions.md`

Design reference: effects-driven condition model, application/removal flow, TTRPG vs video game resolution, stacking rules, immunity, interaction with Wuxing phases.

---

### 3. Armor: Replace AC Model with DR Model + Generic Armors + Slots + Layering

#### 3a. Armor Slots

Replace single "armor" entry with these equipment slots, each with its own base DR contribution:

| Slot | Coverage | Examples | DR Share |
|------|----------|----------|----------|
| **Head** | helmet, cap, circlet, hood | Helmet, Coif, Crown | 15–20% of total |
| **Torso** | chest, shoulders | Breastplate, Chainmail, Robe | 50–60% of total |
| **Back** | cape, cloak, backpack | Cloak, Backpack, Quiver | 0 DR (utility) |
| **Hands** | gloves, gauntlets, bracers | Leather Gloves, Steel Gauntlets | 5–10% of total |
| **Feet** | boots, shoes, greaves | Leather Boots, Plate Greaves | 10–15% of total |
| **Accessories** | rings, amulets, trinkets, belts | Ring of Protection, Amulet of Health | 0 DR (magic items) |

Total DR = sum of head + torso + hands + feet (each modified by material). Back and accessories do not contribute base DR but may provide magical effects or utility.

**Accessories rule:** Up to 5 accessories may be worn simultaneously, but **only one of each type**:
- 1 ring
- 1 amulet / necklace
- 1 trinket / charm
- 1 belt
- 1 miscellaneous (brooch, badge, etc.)

#### 3b. Armor Layering

Multiple items can occupy the same slot if they belong to different **layers**:

| Layer | Position | Examples | Notes |
|-------|----------|----------|-------|
| **Skin** | Against body | Padded tunic, cloth robe, undercoat | Clothes are always skin layer. Lowest DR contribution per material. |
| **Middle** | Over skin layer | Chainmail, scale shirt, brigandine | Medium DR. Requires skin layer underneath (penalty if missing). |
| **Upper** | Outermost (armor) | Breastplate, plate cuirass, leather jerkin | Highest DR per slot. Can be worn alone or over middle layer. |
| **On Top** | Over everything | Cloak, cape, hooded mantle, surcoat | Utility layer. 0 DR (or minimal). Can hold enchantments. No layer dependency. |

**Back slot** uses a single layer (no layering). Holds capes, backpacks, quivers. **Accessories** have no layer — they are separate.

Slot + layer = 16 possible equipment positions (head/torso/hands/feet × 4 layers) + back + 5 accessories.

**Examples:**
- Peasant: Cloth tunic (torso/skin) + Cloth hat (head/skin) + Cloth shoes (feet/skin) — DR 0–1 total
- Light footman: Padded gambeson (torso/skin, leather) + Chainmail (torso/middle, iron) + Leather cap (head/skin) + Leather gloves (hands/skin) + Leather boots (feet/skin) — DR ~5 total
- Knight: Padded undercoat (torso/skin, cloth) + Chainmail (torso/middle, steel) + Plate cuirass (torso/upper, steel) + Steel helm (head/upper) + Steel gauntlets (hands/upper) + Plate greaves (feet/upper) + Wool cloak (back, on_top) — DR ~14 total

**Rules:**
- Middle layer cannot be worn without a skin layer in the same slot (penalty: −1 AGI, chafing).
- Upper layer can be worn alone or over middle layer (no penalty).
- On Top layer can be worn over any combination. No layer dependencies.
- Each layer adds its material-modified DR to the total (On Top usually adds 0).
- Encumbrance/weight is sum of all items × material weightFactor.
- Back slot items do not layer with torso armor — they sit over it.

#### 3c. Generic Armor Base Types + Materials

Like weapons, armor entries become **base types** that materials modify:

| Base Type | Slot | Layer | Base DR | Governing Skill | Properties |
|-----------|------|-------|---------|-----------------|------------|
| Padded tunic | Torso | Skin | 0 | light_armor | — |
| Cloth robe | Torso | Skin | 0 | light_armor | — |
| Leather jerkin | Torso | Upper | 1 | light_armor | — |
| Chainmail | Torso | Middle | 2 | medium_armor | Requires skin layer |
| Scale shirt | Torso | Middle | 3 | medium_armor | Requires skin layer |
| Brigandine | Torso | Middle | 3 | medium_armor | Requires skin layer |
| Breastplate | Torso | Upper | 4 | heavy_armor | — |
| Plate cuirass | Torso | Upper | 5 | heavy_armor | Requires 8+ STR |
| Full plate | Torso | Upper | 6 | heavy_armor | Requires 10+ STR, includes attached upper legs |
| Cap/Hood | Head | Skin | 0 | light_armor | — |
| Coif | Head | Middle | 1 | medium_armor | Requires skin layer |
| Helmet | Head | Upper | 2 | heavy_armor | — |
| Gloves | Hands | Skin | 0 | light_armor | — |
| Gauntlets | Hands | Upper | 1 | heavy_armor | — |
| Shoes | Feet | Skin | 0 | light_armor | — |
| Boots | Feet | Skin | 1 | light_armor | — |
| Greaves | Feet | Upper | 2 | heavy_armor | — |
| Cloak | Back | — | 0 | none | on_top layer equivalent |
| Backpack | Back | — | 0 | none | Utility, no DR |

**Material modifiers applied per piece:**
- `material.modifiers.defense` adds to the piece's base DR.
- `material.weightFactor` multiplies the piece's base weight.
- `material.phase` sets the piece's phase for Wuxing interactions.
- Materials that do not `appliesTo` "armor" cannot be used for armor pieces.

**Shields** remain a separate equipment type (`type: "shield"`) with their own DR value (base 2, modified by material). Governed by `block` skill. Not part of any slot — held in hand.

**Unarmored:** No armor pieces worn. Evasion skill + AGI is the sole physical defense (Layer 1 only). Unarmored skill provides a bonus to evasion.

#### 3d. Schema Change: `equipment.schema.json`

**Replace** the current `armor` object:

```
New armor object:
{
  slot:       "head" | "torso" | "back" | "hands" | "feet",
  layer:      "skin" | "middle" | "upper" | "on_top" | null,
  category:   "light" | "medium" | "heavy" | null,     // governs which skill applies; null for back/on_top
  drBase:     integer (base DR before material modifier),
  evasionPenalty: integer (penalty to evasion skill, 0/ −5/ −15),
  speedPenalty: integer (movement units reduction, 0/ −1/ −2),
  strengthRequirement: integer (optional, range 1–10),
  stealthPenalty: boolean,
  requiresLayer: ["skin"] | ["middle"] | null    // layer dependency
}
```

New accessory type:
```
{
  accessory: {
    type: "ring" | "amulet" | "trinket" | "belt" | "misc",
    effects: [appliedEffect]   // magical effects from the registry
  }
}
```

Add `type: "shield"` to the armor slot/layer system: `slot: "offhand"`, `layer: null`, `drBase` applies only during active block.

Add `layer` and `slot` constraints to `oneOf` — an equipment entry must provide either `weapon`, `armor`, `consumable`, or `shield`.

#### 3e. Data Update: `armor.json` → `equipment/armor/*.json`

Restructure. Each slot can have its own file or keep a single `armor.json`. Schema allows the same root `equipment` array.

#### 3f. Material Schema Update

`material.schema.json` already has `modifiers.defense` — this is correct. No change needed to the material schema itself. Materials already have `appliesTo: ["armor", "accessory"]` where appropriate. Currently `leather` has `armor`, `copper`/`elven_wood`/`wood` have `armor`+`accessory`, `gold`/`silver` have `accessory` only. Need to add `armor` to `copper`, `bronze`, `iron`, `steel`, `elven_wood` (if not present), and `wood`. `silver` and `gold` remain accessory-only (too soft for armor).

---

### 4. Updated `attributes.md` — Skill Seeds

| Attribute | Currently Seeds | Add |
|-----------|----------------|-----|
| STR | blades, blunt, piercing, athletics | **heavy_armor** |
| END | block | **medium_armor** |
| AGI | unarmed, stealth, lockpick, sleight_of_hand | **light_armor, unarmored, evasion** |

Also update the "How attributes feed the system" table to point derived combat stats to the new `combat.md`.

---

### 5. Inventory of Required Changes

| File | Action |
|------|--------|
| `ruleset/schemas/condition.schema.json` | **Create** — condition schema referencing effect ids |
| `ruleset/data/conditions/core.json` | **Create** — 15 condition entries |
| `docs/modules/conditions.md` | **Create** — conditions design, effects-driven flow, dual resolution |
| `docs/modules/combat.md` | **Create** — full combat doc with dual resolution throughout |
| `ruleset/schemas/equipment.schema.json` | **Update** — replace AC with DR, add slot/layer/drBase/penalties |
| `ruleset/data/equipment/armor.json` | **Rewrite** — generic base types by slot+layer, remove hardcoded items |
| `ruleset/schemas/material.schema.json` | **Review** — `modifiers.defense` already exists, ensure `appliesTo` handles armor slots |
| `ruleset/data/materials/core.json` | **Update** — add `armor` to `appliesTo` for applicable materials |
| `docs/modules/attributes.md` | **Update** — skill seeds (STR/END/AGI) + links to combat.md |
| `docs/modules/progression.md` | **Update** — derived stats table, open items point to new combat doc |
| `docs/modules/weapons.md` | **Update** — armor mention removed (now in combat.md) |

---

## Implementation Order

1. Create `condition.schema.json` + `conditions/core.json` → `validate.py`
2. Update `equipment.schema.json` (DR + slot/layer) → `validate.py`
3. Rewrite `armor.json` (generic base types by slot+layer) → `validate.py`
4. Update `materials/core.json` (add armor to `appliesTo`) → `validate.py`
5. Create `docs/modules/conditions.md`
6. Create `docs/modules/combat.md`
7. Update `attributes.md` (skill seeds + derived stat links)
8. Update `progression.md` (point to new docs, close open items)
9. Final `validate.py` pass

---

## Open Questions

- **Slot+Layer complexity:** 16+ positions is a lot of equipment management. For TTRPG this is fine (paper sheet). For video game, consider simplifying: hide layering behind an "undergarment" auto-slot, or limit to one item per slot in the UI.
- **Shield as offhand slot:** Shield uses the offhand position; two-handed weapons occupy both hand slots implicitly.
- **Condition → effect linking:** Some conditions don't have a dedicated applying effect yet (e.g. `blinded`, `charmed`). These may need new effects added to the registry, or reuse existing ones with appropriate `parameter` values. Deferred to implementation — if a condition lacks an `appliedBy` effect, add one.
- **Deferred combat sections:** mounted combat, underwater combat, vehicle combat, mass combat — note as future sections in combat.md.

---

## Success Criteria

- `validate.py` passes on all new/updated schemas and data
- `combat.md` covers all 7 sections with dual resolution (TTRPG + video game) on every mechanic
- `conditions.md` + data covers 15 conditions with `appliedBy`/`removedBy` referencing real effect ids
- `attributes.md` lists all new armor/defense skills under correct attributes
- Armor data uses **slot + layer + drBase** model with material modifiers (no legacy AC)
- Equipment schema has DR fields, not `acBase`/`acDexBonus`/`acMaxDex`
- Materials updated so armor-relevant materials allow `armor` in `appliesTo`
- Every mechanic specifies TTRPG and video game resolution where they differ
- The defense model has two trees (physical and magic), both integrating active effects; magic evasion is perk-gated
- Critical system uses confirmation roll + Luck-expanded window; fumble reduced to 100-only TTRPG flavor
- OA in video game mode: no auto-trigger, player-input-driven with bonus on disengage
- Armor has 5 slots (head/torso/back/hands/feet) + 5 accessories (1 per type), with 4 layers (skin/middle/upper/on_top)
- Strength requirements on armor are achievable within the 1–10 attribute range
- Material schema reused for armor crafting; existing `modifiers.defense` drives material DR contribution
