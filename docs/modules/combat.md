# Combat Resolution

**Status:** Core (implemented as data + schema) · runtime resolution lives in the engine or GM adjudication.

## How to Use This Module

This document is the **single source of truth** for combat resolution in Riftweave. Consuming engines (Zig, Godot, custom TTRPG sheet, etc.) should treat each section as a contract:

- **Data files**: `ruleset/data/equipment/weapons.json`, `ruleset/data/equipment/armor.json`, `ruleset/data/conditions/core.json`, `ruleset/data/effects/core.json` — these supply the building blocks (weapon stats, armor DR, conditions, effects). The schemas in `ruleset/schemas/` enforce shape; this doc explains how to use them at runtime.
- **Skills & attributes**: See `progression.md` and `attributes.md` for seeding formulas. Combat-relevant skills are `blades`, `blunt`, `piercing`, `bows`, `crossbows`, `guns`, `throwing_weapons`, `block`, `evasion`, `light_armor`, `medium_armor`, `heavy_armor`, `unarmored`, plus all nine `<color>_magic` schools.
- **Perks**: Many combat perks live in `ruleset/data/features/core.json`. See §16 (Combat Perks & Traits) below for integration guidance.
- **GM (TTRPG)**: Roll d100, adjudicate reactions, track conditions manually. Use the TTRPG resolution blocks.
- **Engine developer**: Implement the Video Game resolution blocks. Use the comparison table (§8) as a quick reference for which mode applies to your game type.
- **cRPG (party-based tactical)**: Lean toward TTRPG resolution with automation. Notes are marked throughout.

Key relationships: `combat.md` → [`armor.md`](./armor.md) (slots, layers, DR stacking) → [`conditions.md`](./conditions.md) (status effects) → [`progression.md`](./progression.md) (skill/attribute formulas). Effects from the shared registry (`effects/core.json`) are the atomic units that spells, conditions, and consumables apply.

---

This document defines the full combat resolution system for Riftweave: action economy,
opportunity attacks, defense, critical hits, cover and terrain, and initiative. Every
mechanic specifies **two resolution modes**:

- **TTRPG (turn-based):** GM-facing rules, dice rolls, adjudication, AP-based action economy.
- **Video game (action-combat):** Engine-facing rules, player input drives all actions, real-time, Stamina-based.

Where a mechanic is identical in both modes, it is stated explicitly. For **cRPG**
(party-based, tactical): lean closer to TTRPG resolution with automation.

For the character-progression view (skills, attributes, derived stats) see
[`progression.md`](./progression.md). For conditions see
[`conditions.md`](./conditions.md). For the armor system (slots, layers, DR)
see [`armor.md`](./armor.md).

---

## 1. Action Economy

### TTRPG (turn-based)

**Turn structure:**
```
Start of turn → Action → Bonus action → Movement (split) → Reaction → End of turn
```

- **Actions** (one per turn): Attack, Cast Spell, Dash (×2 movement), Disengage (no OA),
  Dodge (evasion bonus), Hide (stealth), Ready (triggered action), Use Object, Help.
- **Bonus actions** (one per turn): Off-hand light weapon attack, certain perks and
  spells (e.g. Quick Cast, racial abilities).
- **Reaction** (one per round): Opportunity attack, Parry spell, triggered perks.
  Refreshes at the start of the character's turn.
- **Movement:** Speed in units (from race, modified by AGI and encumbrance). Can be
  split before and after the action. e.g. move 3 → attack → move 2 (total 5).
  - **Encumbrance effects:** See [`armor.md`](./armor.md) §9. Carrying more than
    `STR × 10` lbs reduces speed by 1 unit and doubles evasion penalties. At
    `STR × 15`, speed drops by 2 and sprinting/dodging is prevented.

**AP pool variant (TTRPG):** AGI-based Action Points can replace the discrete
action/bonus action model for groups that prefer a resource-budget approach.

| Rule | Value |
|------|-------|
| Starting pool per turn | `2 + floor(AGI / 3)` |
| Maximum pool | 10 (cannot exceed) |
| Carry-over | Unspent AP carry over between turns (capped at max) |
| Recovery after combat | Full pool restored after a short rest |

**AP costs for common actions:**

| Action | AP Cost |
|--------|---------|
| Standard attack (melee or ranged) | 3 |
| Cast spell | 2–5 (varies by spell level) |
| Move (per unit of speed) | 1 |
| Dash (double movement for the turn) | 2 |
| Disengage | 2 |
| Dodge | 3 |
| Block / Parry (reaction) | 2 (reserved from pool) |
| Use object / interact | 1 |
| Stand from prone | 2 |
| Aimed shot (see §8) | +2 (added to attack cost) |
| Ready action | 2 |

**Reactions** reserve AP from the pool ahead of time (declared at turn start or
when the trigger occurs, consuming the cost immediately on use). Unused reaction
AP is refunded at end of round.

This variant works well for groups that want granular resource management over
set action types.

### Video game (action-combat)

- **No discrete turn structure.** Player input drives every action — there is no
  automated backswing or auto-trigger.
- Attacks, blocks, and special moves are mapped to buttons/inputs. Each drains
  **Stamina** (END pool `15 + END × 5 + level × 2`, regen scales with WIL).
- Movement: speed rating, always active. Sprinting and dodging cost Stamina.
- No "action types" — the player decides what to do moment-to-moment.
- Cooldowns and Stamina costs gate ability use instead of "bonus action" or "reaction" slots.

**cRPG note:** If implementing a party-based tactical cRPG, use TTRPG-style discrete
actions with AI-controlled companions and optional auto-pause on events.

---

## 2. Opportunity Attacks / Reactive Strikes

**Universal trigger:** Leaving an opponent's reach without using the Disengage action.

### TTRPG

- **Provoking movement:** Exiting a threatened square or unit of length. Reach weapons
  (attackReach 2) threaten 2 units; normal and short weapons threaten 1 unit.
- **Resolve:** A single melee attack as a reaction, resolved as a normal attack roll.
  The attacker uses their weapon skill; the defender's armor DR applies normally.
- **Modifiers:** The Disengage action prevents OA. Some perks grant OA on other
  triggers (casting a spell while within melee range, standing from prone, drinking a potion).

### Video game (action RPG)

- **No auto-triggered backswing.** The player controls hit timing entirely.
- When an enemy disengages while in melee range, the player's next attack against
  that target within a short window (e.g. 2 seconds) gets a **bonus** (+10 attack
  or +25% damage) — rewarding player awareness, not automating it.
- Perks can widen the bonus window or add additional effects (slow, interrupt).

**cRPG note:** For party-based cRPGs, use TTRPG OA rules (discrete, reaction-based)
with optional auto-pause on OA trigger.

---

## 3. Defense Model — Two Trees

Both trees consider **active effects** on the player (from the effects registry,
e.g. `resist` id 19, `spell_absorption` id 21) as modifiers at their respective
layers. See [`conditions.md`](./conditions.md) and [`armor.md`](./armor.md) for details
on conditions and armor DR.

### Physical Defense Tree

```
Attack hits (d100 ≤ weapon skill)
  → Layer 1: Evasion (evasion skill + AGI) — chance to avoid entirely
    → Layer 2: Block (block skill + shield DR) — reaction, reduces/negates on success
      → Layer 3: Armor DR (sum of all worn armor pieces, modified by materials)
        → Layer 4: Active effects (resist physical, stoneskin, fortify_endurance, etc.)
          → Final physical damage
```

### Magic Defense Tree

```
Spell lands (d100 ≤ cast check)
  → Layer 1: Evasion* — evasion can apply to magic only with a perk (not baseline)
  → Layer 2: Spell Absorption (spell_absorption id 21, WIL-based %) — % chance to absorb into mana
    → Layer 3: Magic Resistance (resist id 19 with parameter "magic" + WIL) — % reduction or resist roll
      → Layer 4: Elemental/Phase Resistance (resist id 19 with parameter = phase/type) — per-element DR
        → Layer 5: Active effects (ward spells, aura effects, fortify_wil)
          → Final magical damage
```

\*Baseline evasion only applies to physical attacks. A perk (e.g. "Spell Dodger")
unlocks Layer 1 in the magic tree, allowing evasion skill to contribute against
targeted spells.

### TTRPG Resolution

- **Evasion:** Passive penalty to the attacker's roll (reduces their effective skill),
  or an active dodge as a reaction (declared before the attack roll, contested vs
  the attacker's margin). Magic evasion only with the perk.
- **Block (shield):** Declared reaction before the attack roll. The blocker makes a
  block skill check against the attacker's margin of success. On success, the shield's
  DR is subtracted from incoming damage. On failure, no block benefit.
  - **Shield types:** Buckler (DR +1, light off-hand item allowed), Standard shield
    (DR +2), Tower shield (DR +4, grants cover to adjacent allies, heavy).
  - **Parry (weapon):** Some weapons (rapiers, swords) and the *Duelist* perk allow
    a parry instead of a block. The parry uses the weapon skill instead of block skill.
    On success, the attack is negated (not reduced). On failure, the attack hits with
    no DR reduction. Parry does not require a shield.
  - **Two-weapon fighting:** An off-hand weapon can parry (use the off-hand weapon
    skill) but cannot block. A character with two weapons may choose to attack with
    the off-hand as a bonus action OR reserve it to parry — not both in the same turn.
  - **Unarmed block:** The *Deflect Arrows* perk allows blocking missiles with
    unarmored skill. Melee unarmed block is not possible without specific perks.
- **Armor DR:** Flat subtraction from incoming physical damage after the hit is confirmed.
  Sum of all worn armor pieces (see [`armor.md`](./armor.md) for slot+layer rules).
- **Magic:** Each layer is a separate roll or percentage reduction applied sequentially.
  The GM resolves each layer (absorption check → resistance roll → elemental DR).

### Video Game Resolution

- **Evasion:** Percentage dodge chance computed from evasion skill + AGI. Auto-rolled
  on each incoming attack. Magic evasion only with perk. Visually: a "miss" or "dodge"
  animation plays.
- **Block:** Input-timed block/parry window. Pressing block just before impact triggers
  a successful block; skill determines window length and DR %.
  **Parry** (higher risk/reward): pressing block within a narrower window reflects
  the attack or stuns the enemy — **optional, not enforced**.
- **Armor DR:** Flat reduction per hit from all worn armor, computed and applied
  automatically. Displayed in the damage popup (`Damage: 12 - 4 (DR) = 8`).
- **Magic:** Percentage mitigation computed from stats and active effects. Applied
  automatically when a spell lands. No per-layer rolls; the engine aggregates all
  magic defense layers into a single %.
- **Active effects:** Visible buff/debuff icons on the HUD. Auto-applied to combat
  formulas by the engine.

---

## 4. Critical Hits & Fumbles

### Universal Base

A natural **01–05** on the d100 attack roll is a **critical window**. A critical
**confirmation roll** is then required.

### Confirmation System

1. On a natural 01–05, roll again.
2. If the confirmation roll also succeeds (≤ modified target number), the hit is a
   **confirmed critical**.
3. If the confirmation roll fails, the hit is a normal hit (no special effect).
4. **Luck** helps:
   - `+floor(LCK / 2)` bonus to the confirmation roll.
   - Every 2 points of LCK expands the critical window by +1.
     - e.g. LCK 4 → window is 01–07. LCK 10 → window is 01–10.

### Luck & Critical Hit Examples

#### Critical Window by Luck

| LCK | Natural Crit Range | Effective Crit Chance | Confirmation Bonus |
|-----|-------------------|----------------------|-------------------|
| 1   | 01–05             | 5%                   | `+0`              |
| 2   | 01–06             | 6%                   | `+1`              |
| 4   | 01–07             | 7%                   | `+2`              |
| 6   | 01–08             | 8%                   | `+3`              |
| 8   | 01–09             | 9%                   | `+4`              |
| 10  | 01–10             | 10%                  | `+5`              |

The confirmation roll uses the **same modified target number** as the original
attack — the attacker's effective skill after all modifiers. If the original
attack roll benefited from a situational bonus (e.g. flanking +10), the
confirmation roll uses the same total. If the roll has advantage or multiple
dice (video game), the confirmation is a single flat roll against the modified TN.

#### Example: Confirmation with LCK

> A character with Blades **65**, LCK **6** attacks a bandit. They roll a natural
> **04** — inside the crit window (01–08 at LCK 6). The confirmation roll target
> is their effective skill of 65, plus `+floor(LCK/2) = +3`, giving an effective
> target of **68**. They roll a **57** — success! The crit is confirmed. Had they
> rolled **72**, it would have been a normal hit (no doubled damage).

#### Example: LCK Expanding the Window

> With LCK **10** (window 01–10), a character rolls **09** on their attack. Without
> LCK this would be a normal hit. With LCK, it falls inside the expanded crit
> window. The confirmation roll gets `+floor(10/2) = +5`, making the confirmation
> much more likely. A character with LCK 10 and high skill can reliably convert
> the top 10% of their rolls into critical hits.

### TTRPG

- **Confirmed crit:** Roll damage dice twice and sum both (alternatively: max dice + roll).
- **Roll on the Critical Effects Table** to determine bonus effects. Apply the result
  immediately.

#### Critical Effects Table (d100)

| d100 | Effect | Description |
|------|--------|-------------|
| 1–15 | Bludgeoning Blow | Extra damage only (max damage dice + roll again) |
| 16–30 | Deep Wound | Target bleeds: takes `damage_health` (effect id 5, magnitude 2) at start of each turn until healed or DC 10 Medicine check |
| 31–40 | Stagger | Target loses next reaction and takes −10 on next action for 1 round. See `staggered` condition |
| 41–50 | Disarm | Weapon knocked from grip. STR save (DC 12) or weapon lands 1d4 units away |
| 51–60 | Crippled Arm | Dominant arm crippled: −20 to attack rolls, −50% damage. Two-handed weapons unusable. Heals after combat or via `cure` effect |
| 61–70 | Crippled Leg | Leg crippled: speed −50%, cannot dodge/evade. Heals after combat or via `cure` effect |
| 71–80 | Knockdown | Target knocked prone (see `prone` condition). Loses next action standing. OA provoked |
| 81–85 | Armor Pierce | Ignore all armor DR on this hit + extra damage (max dice) |
| 86–92 | Stun | Target stunned for 1 full turn (no actions, reactions, movement). See `stunned` condition |
| 93–97 | Grievous Wound | Max ×2 damage + Deep Wound (permanent until long rest) + crippled limb (random) |
| 98–100 | Overkill | Damage ×3 + Deep Wound + Stun 2 turns + permanent injury (GM discretion) |

- **Luck modifier:** The player adds `+floor(LCK / 2)` to the d100 roll (shifts toward
  better effects). If LCK ≥ 8, they may re-roll once and keep the better result.
- **Interaction with Called Shots:** When a called shot (see §6) scores a confirmed crit,
  the called-shot location effect is **guaranteed** (no save). Roll on this table for an
  additional effect.
- **Interaction with Conditions:** Effects on this table (bleed, stagger, knockdown, stun)
  correspond to entries in `conditions/core.json` and the effects they reference. Engines
  apply them via the shared effect system.
- **Fumble:** Only on a natural **100** (not 96–00). Flavor-only for TTRPG:
  drop weapon, stumble, hit an ally by accident. No mechanical penalty enforced.
  High Luck does not affect fumble range.

### Video Game

- **Confirmed crit:** Fixed multiplier (×1.5 or ×2) applied after DR.
  `Final damage = (base damage × crit multiplier) - armor DR`
- **Automatic effects:** Critical hits automatically apply a minor status effect based on
  weapon type (blade → bleed, blunt → stagger, piercing → armor impair). Drawn from the
  effects registry, lasting 1–2 ticks.
- **No fumble mechanic.** A roll of 100 is simply a miss (or a glancing blow
  that deals 0 damage).
- Luck feeds into crit chance via expanded window and confirmation bonus.

---

## 5. Cover, Flanking, Difficult Terrain

### TTRPG

| Feature | Effect |
|---------|--------|
| **Cover (half)** | −10 to attack roll; +2 to DEX saves (or equivalent AGI-based check) |
| **Cover (three-quarters)** | −20 to attack roll; +5 to DEX saves |
| **Cover (full)** | Immune to ranged attacks (cannot be targeted) |
| **Flanking** | +10 to melee attack when two allies are on opposite sides of the target |
| **Difficult terrain** | Movement costs double. Cannot use Disengage while in difficult terrain |
| **Prone** | −20 to melee attack rolls. Ranged attacks against you have +10. Spend half your movement to stand |

### Video Game (action RPG)

- **No formal cover system.** The environment has line-of-sight blockers (walls,
  pillars, terrain) that block projectiles naturally. No to-hit bonus or penalty
  from cover.
- **Flanking:** Damage percentage bonus when attacking from behind or opposite an
  ally (e.g. +25%). Calculation is automated; collision or angle checks drive it.
- **Difficult terrain:** Slow effect (movement speed %) while traversing. Applied
  automatically when entering the area.
- **Prone:** Knockdown state with a get-up animation duration. Cannot attack while
  prone. Ranged attacks deal +X% damage against prone targets.

**cRPG note:** For grid/positioning-based tactical cRPGs, use TTRPG cover and prone
rules throughout.

---

## 6. Called / Targeted Shots

A character may declare a **targeted shot** before making an attack roll, aiming at
a specific body part instead of making a general attack. This imposes a penalty to
the attack roll but grants bonus effects on a hit.

### Shot Locations

| Location | Attack Penalty | Damage Multiplier | Special Effect on Hit |
|----------|---------------|-------------------|----------------------|
| **Torso** | 0 (standard) | ×1.0 | No special effect — the default |
| **Head** | −20 | ×1.5 (or max dice) | + Stagger (enemy loses next reaction). Video game: + bonus crit %, screen shake |
| **Arms** | −15 | ×0.75 | + Disarm chance (STR check or drop weapon). Video game: −X% enemy attack for 5s |
| **Legs** | −15 | ×0.75 | + Knockdown or slow (speed −50% for 2 rounds). Video game: slow + stumble animation |
| **Groin / Eyes** | −30 | ×2.0 | + Stagger + possible stun (END save or stunned 1 round). Video game: extended stagger, screen flash |

**Ranged note:** Called shots with ranged weapons beyond short range take an
additional −10 penalty. At long range, called shots are not possible.

### TTRPG Resolution

1. Declare the target location before rolling.
2. Apply the attack penalty to the d100 roll-under check.
3. On a hit (≤ modified skill), the damage multiplier applies and the special effect
   takes hold (GM determines duration or makes a save for the target).
4. On a miss, the action is wasted (no grazing or partial effect).
5. Perks can reduce penalties: *Marksman* (halve ranged penalty), *Street Samurai*
   (halve called shot penalties in general), *Surgeon* (remove damage penalty on arm/leg shots).

Rolling a **critical hit** on a called shot:
- Crit confirmation bonuses apply as normal.
- On a confirmed crit, the special effect is guaranteed (no save).
- The damage multiplier stacks with the crit multiplier (×1.5 ×2 = ×3 damage).

### Video Game Resolution

- Called shots are **manual aim** — the player aims at a body part using a targeting
  mechanic (V.A.T.S.-style pause-and-target or free-aim reticle shift).
- Hit probability is computed from the player's weapon skill + distance + target
  movement, displayed as a % in the targeting UI.
- On hit, the location-specific effect fires automatically (slow on legs, disarm
  on arms, stagger on head).
- Damage multipliers apply before DR.
- Perks reduce the penalty or unlock new targeting options.

**cRPG note:** For tactical cRPGs, use TTRPG-style called shot resolution with
% hit chance displayed per body part.

---

## 7. Initiative

### TTRPG

`d20 + PER + modifiers`

- Rolled once per combat encounter (or each round for variable initiative).
- Ties broken by higher PER, then higher AGI, then reroll.
- The *Alert* perk adds +10.
- Alternative tighter spread: `d10 + floor(PER / 2)`.

### Video Game

**Not needed.** Real-time action combat does not use initiative. Engagement order
is determined by:
- **Aggro/threat system:** Enemies target the highest-threat character.
- **Proximity:** Enemies engage the nearest visible target.
- **Perks/abilities:** Taunts force retargeting; stealth bypasses engagement.

---

## 8. Mode Comparison Table

| Aspect | TTRPG (turn-based) | Video game (action-combat) |
|--------|-------------------|---------------------------|
| **Actions** | Discrete Action/Bonus/Reaction | Player input per button; Stamina cost |
| **Movement** | Speed in units per action | Speed rating, free + sprint costs Stamina |
| **Reactions** | 1/round, declared | Timed block/parry input (optional); cooldowns |
| **OA** | Auto-trigger on disengage | Bonus on next attack vs disengaging enemy |
| **Initiative** | d20 + PER | Aggro/threat or proximity |
| **Physical defense** | Evasion check, Block reaction, flat DR | Auto-dodge %, timed block/parry, flat DR |
| **Magic defense** | Sequential rolls per layer (evasion only with perk) | Auto-mitigation from stats + effects (evasion only with perk) |
| **Critical hits** | Confirmation roll, double dice + d100 effects table | Confirmation roll, ×1.5–2 multiplier, auto-status effect |
| **Fumble** | On 100 only; flavor table | None — 100 is a miss |
| **Critical effects** | d100 table (bleed, stagger, disarm, cripple, stun, etc.) | Weapon-type auto-status (blade→bleed, blunt→stagger) |
| **Block / Parry** | Shield block (block skill) or weapon parry (weapon skill) | Timed input block; parry optional (narrow window) |
| **Shields** | Passive DR only during active block; tower shields grant cover | Block button; shield type affects block window size |
| **Two-weapon** | Off-hand attacks OR parry (not both) | Dual-wield attack chain; no block while dual-wielding |
| **Encumbrance** | Speed penalty, doubled evasion penalties, no sprint at heavy | Speed % reduction, dodge penalty, stamina drain |
| **Cover** | −10/−20 to-hit | LoS blockers only (no to-hit modifier) |
| **Flanking** | +10 to-hit | +25% damage |
| **Terrain** | Double movement cost | % speed slow |
| **Conditions** | GM tracks duration | Automated timers, visual indicators |

---

## 9. Encumbrance in Combat

Encumbrance affects combat performance directly. See [`armor.md`](./armor.md) §9
for the full encumbrance table (weight thresholds by STR). This section describes
how those thresholds translate into combat mechanics.

### TTRPG

| Encumbrance Level | Speed Penalty | Evasion Penalty | Action Cost | Other |
|-------------------|---------------|-----------------|-------------|-------|
| Unencumbered (≤ STR×10) | None | None | Normal | — |
| Encumbered (> STR×10) | −1 unit | Doubled (sum evasion penalties ×2) | Dash costs +1 AP | Cannot jump full distance |
| Heavily encumbered (> STR×15) | −2 units | Tripled | Dash and Dodge disabled | Cannot sprint, −20 athletics |
| Overburdened (> STR×20) | 0 (cannot move) | Auto-fail evasion | Only drop items and talk | Requires STR check to lift anything |

- Evasion penalties from armor (per piece) are doubled or tripled at encumbered
  and heavily encumbered levels respectively.
- Encumbrance is calculated from all carried items (worn + inventory), not just armor.

### Video Game

| Encumbrance Level | Speed | Dodge | Stamina | Other |
|-------------------|-------|-------|---------|-------|
| Unencumbered | 100% | Normal | Normal regen | — |
| Encumbered | −20% speed | −10% dodge chance | Stamina regen −25% | Jump height reduced |
| Heavily encumbered | −40% speed | −25% dodge chance | Stamina regen −50%, sprint drains ×2 | Cannot climb |
| Overburdened | Cannot move | Cannot dodge | No regen, actions cost ×2 | Only drop items |

### cRPG Note

For party-based tactical cRPGs, use TTRPG-style encumbrance with automated
calculation. Display encumbrance level on the character sheet and warn when
approaching thresholds. Consider a "container" system (inventory weight ≠ worn
weight) for realism vs. convenience.

---

## 9a. Armor Weight & Action Costs

Armor category (Light / Medium / Heavy) imposes a **stamina or AP surcharge** on
demanding actions. This is separate from encumbrance (total carried weight) —
it applies based on the **heaviest armor category worn in any slot**. A character
wearing a steel breastplate (heavy torso) with leather boots (light feet) is
considered Heavy for action cost purposes (the heaviest category wins).

**Design rationale:** The penalty only hits aggressive/mobile actions, not passive
existence. Heavy armor tanks can stand and trade blows without extra cost — they
pay when they dodge, sprint, or wind up a power attack. This reinforces role
identity (tanks hold ground, skirmishers move) without making heavy armor feel
punishing just for being equipped.

### Category Determination

| Armor Category | Heaviest Piece Worn | Label |
|----------------|---------------------|-------|
| Unarmored | No armor (skin-layer clothing only) | Unarmored |
| Light | All worn armor pieces are Light or skin-layer | Light |
| Mixed Light/Medium | At least one Medium piece, no Heavy pieces | Medium |
| Any Heavy | At least one Heavy piece in any slot | Heavy |

- Shields contribute their `category` (light for buckler, medium for standard,
  heavy for tower shield) to the determination.
- Back slot and accessory items do not affect the category.

### Video Game (Action-Combat) — Stamina Surcharges

The base stamina costs for demanding actions are multiplied by the character's
armor category factor:

| Action | Base Stamina Cost | Light Multiplier | Medium Multiplier | Heavy Multiplier |
|--------|------------------|-----------------|-------------------|------------------|
| Dodge roll | 15 | ×1.0 | ×1.5 (22.5 → 23) | ×2.0 (30) |
| Sprint (per second) | 5/s | ×1.0 | ×1.4 (7/s) | ×1.8 (9/s) |
| Power attack / heavy attack | 20 | ×1.0 | ×1.3 (26) | ×1.6 (32) |
| Jump | 8 | ×1.0 | ×1.25 (10) | ×1.5 (12) |
| Clamber / vault | 10 | ×1.0 | ×1.5 (15) | ×2.0 (20) |
| Block (shield) | 8 | ×1.0 | ×1.25 (10) | ×1.5 (12) |

**Key principles:**
- **Light armor:** No surcharge on any action. Default costs. This is the mobility
  advantage of light armor.
- **Medium armor:** Moderate surcharge (×1.25 to ×1.5). Penalties are noticeable but
  manageable with good stamina management. A frontliner in chainmail can dodge
  occasionally but not repeatedly.
- **Heavy armor:** Significant surcharge (×1.5 to ×2.0). Dodge rolls are very
  expensive — a knight should block or tank hits, not dodge out of the way. Sprint
  cost is nearly doubled, encouraging measured positioning over hit-and-run.

**Actions that are NOT penalized:**
- Light attack / basic attack (flat stamina cost, independent of armor)
- Walking / normal movement (no stamina cost)
- Standing still / idle
- Using items / interacting
- Casting spells (magic has its own stamina cost rules)

**Stamina regen interaction:**
- Armor action surcharges do not affect stamina regen rate (that's handled by
  encumbrance, see §9 above).
- However, because the surcharges make actions more expensive, the effective
  stamina budget per fight is tighter for heavier armor — the same pool must
  cover fewer high-cost actions.

**UI feedback:**
- Tooltip on the stamina bar or action button: "Dodge: 15 Stamina (×2.0 Heavy Armor)"
- Color code the stamina cost indicator: white (light), yellow (medium), red (heavy).
- Consider an optional HUD element showing "Effective Action Cost" when hovering
  over the armor weight indicator.

### TTRPG (Turn-Based) — AP Surcharges

In TTRPG mode, the same concept translates to additional AP costs on the AP pool
variant (see §1):

| Action | Base AP Cost | Light Surcharge | Medium Surcharge | Heavy Surcharge |
|--------|-------------|-----------------|------------------|-----------------|
| Dodge | 3 | — | +1 (4) | +2 (5) |
| Dash (double movement) | 2 | — | +1 (3) | +1 (3) |
| Stand from prone | 2 | — | +1 (3) | +2 (4) |
| Sprint / Run (discrete move) | 1 per extra unit | — | — | +1 per unit beyond normal speed |
| Power attack / aimed shot | +2 | — | +1 (+3 total) | +2 (+4 total) |
| Jump (chasm, obstacle) | 2 | — | +1 (3) | +2 (4) |

**Discrete action model note:** In the standard action/bonus action/reaction model
(not the AP pool variant), the surcharge manifests as a **restriction** rather than
a numeric cost:
- **Heavy armor:** Cannot take the Dodge action. Dash costs both the action and
  bonus action (effectively preventing any other action that turn).
- **Medium armor:** Dodge and Dash are available but cost both the action and bonus
  action (no attack or spell on a dodge turn).
- **Light armor:** No additional restrictions.

This mirrors the video game design: heavy armor trades mobility for protection,
and the tradeoff is expressed as opportunity cost rather than a flat penalty.

### Interaction with Encumbrance

The armor category surcharge and the encumbrance system stack **additively** on
affected actions:

- An encumbered character in heavy armor pays: base AP cost + heavy surcharge +
  encumbrance AP penalty (Dash costs +1 AP at encumbered).
- Example: A heavily encumbered knight Dashing pays: Dash base 2 + heavy surcharge
  +1 + encumbrance penalty +1 = **4 AP** for a single Dash.
- This is intentional — being overburdened AND wearing heavy armor should be
  prohibitive for mobility. Players are incentivised to manage both their carried
  weight and their equipped armor strategically.

### Perks That Modify Armor Action Costs

Several perks can reduce or eliminate armor action surcharges:

| Perk | Effect | Category |
|------|--------|----------|
| **Heavy Armor Expert** (new) | Halve all heavy armor stamina/AP surcharges (×1.5 instead of ×2.0 for dodge, etc.). Requires heavy_armor 60+ and STR 8. | perk |
| **Light Foot** (new) | Ignore medium armor surcharges entirely. Requires light_armor 50+ and AGI 7. | perk |
| **Brutal Charge** (new) | Power attacks in heavy armor cost stamina equal to a light attack (no heavy surcharge). Requires STR 9 and Power Attack perk. | perk |
| **Acrobat** (existing, reimagined) | Reduce dodge roll stamina cost by 25% (stacks multiplicatively with armor surcharges — +1.5× dodge cost becomes +1.125× effective). | perk |

These perks provide build-specific relief without removing the baseline tradeoff.

### Worked Example: Light vs Heavy Armor Dodge

**Video game comparison:**

| Stat | Light Skirmisher (leather) | Steel Knight (full plate) |
|------|---------------------------|--------------------------|
| Armor category | Light | Heavy |
| Base dodge stamina cost | 15 | 15 |
| Armor surcharge | ×1.0 (none) | ×2.0 |
| **Actual dodge cost** | **15 stamina** | **30 stamina** |
| Dodge chance (auto-dodge %) | ~45% (evasion 60, AGI 9) | ~5% (evasion 15, AGI 6, −55 penalties) |
| Dodges per full stamina pool (60 stamina) | 4 dodges | 2 dodges |
| Recommended defense strategy | Dodge + parry | Block + tank hits |

The steel knight pays double stamina for a dodge that only succeeds 5% of the time
— he should never be dodging. The light skirmisher pays 15 stamina for a 45%-chance
dodge, making it a viable (but not spammable) defense option.

**TTRPG comparison (AP pool variant):**

| Stat | Light Skirmisher | Steel Knight |
|------|-----------------|--------------|
| Armor category | Light | Heavy |
| Base dodge AP cost | 3 | 3 |
| Armor surcharge | — | +2 |
| **Actual dodge cost** | **3 AP** | **5 AP** |
| AP pool per turn | 5 (AGI 9) | 4 (AGI 6) |
| Maximum dodges per turn | 1 (leaves 2 AP for other actions) | 0 (cannot afford; 5 > 4 pool) |

The knight cannot dodge in the standard AP model — his pool of 4 is less than the
5 AP a dodge costs in heavy armor. This is intentional: the knight's defense comes
from Block (2 AP) and raw DR, not evasion. The light skirmisher can dodge once
and still act.

### Open Questions

- **Shield weight interaction:** Should a heavy shield (tower shield) increase the
  armor category for action costs even if the torso armor is light? Current design
  says yes — shields contribute their category to the determination. A character
  wearing leather + tower shield is treated as Heavy for action costs.
- **Per-tier refinement:** The current model uses a single multiplier per category.
  Future refinement could add per-piece granularity (e.g., "each heavy piece adds
  +10% to dodge cost" vs. "any heavy piece = ×2.0"). The current binary model is
  simpler and recommended for initial implementation.
- **Skill mitigation:** Should a high heavy_armor skill reduce the stamina surcharge?
  This is modeled through the proposed Heavy Armor Expert perk for now, but a skill
  scaling formula (e.g., `effectiveMultiplier = max(1.0, baseMultiplier − skill/100)`)
  could be explored for deeper simulation systems.
- **Stamina cost display:** In the Video Game UI, should the displayed stamina cost
  show the pre-multiplier or post-multiplier value? Recommended: show post-multiplier
  with a breakdown tooltip. Example: "Dodge: 30 Stamina (15 × 2.0 Heavy Armor)".

---

## 10. Worked Combat Examples

### Example 1: TTRPG Turn-Based — Knight vs. Bandit

**Setup:**
- **Sir Aldric** (human knight): Blades 72, Block 55, Evasion 30, heavy_armor 65.
  Wears iron plate cuirass (torso/upper, DR 8 = base 5 + iron defense 3) + iron
  helm (head/upper, DR 5 = base 2 + iron defense 3) + iron gauntlets (hands/upper,
  DR 4 = base 1 + iron defense 3) + iron greaves (feet/upper, DR 5 = base 2 + iron
  defense 3) = **22 total DR**. Carries a longsword (1d10 slashing) and a standard
  shield (DR +2 on block). Attributes: STR 8, AGI 6, END 7, LCK 4.
- **Bandit** (human): Blades 45, Evasion 25. Wears leather jerkin (torso/upper, DR 2)
  + leather cap (head/skin, DR 0) = **2 total DR**. Wields a shortsword (1d6).
  No shield.

**Round 1:**

1. **Initiative:** Sir Aldric rolls d20 + PER (4) = 18. Bandit rolls 11. Aldric acts first.

2. **Aldric's turn:**
   - **Action:** Attacks bandit with longsword. Roll Blades 72 → d100 = **47** (success).
   - **Bandit's reaction (evasion):** Bandit attempts active dodge: Evasion 25 → d100 = **63** (fail). Hit lands.
   - **Damage:** Longsword 1d10 = **8** slashing. Bandit's armor DR = 2. Net damage = 8 − 2 = **6**.
   - Bandit HP: 30 → **24**.
   - **Movement:** Aldric moves 2 units north to block the path (speed 6, unencumbered).

3. **Bandit's turn:**
   - **Action:** Attacks Aldric with shortsword. Roll Blades 45 → d100 = **32** (success).
   - **Aldric's reaction (block):** Aldric declares block. Roll Block 55 vs. attacker's
     margin (45 − 32 = 13). Block check: d100 = **41** ≤ 55 (success, margin = 14 ≥ 13).
     Block negates full shield DR. Shield DR = 2 (base) + 0 (steel material) = **2**.
   - **Damage:** Shortsword 1d6 = **5**. Shield blocks 2 → 3 penetrates. Aldric's
     armor DR 22 absorbs all 3. **0 damage** to Aldric.
   - **Movement:** Bandit attempts to disengage (move east). Aldric's OA triggers.
   - **OA:** Aldric's reaction. Roll Blades 72 → d100 = **61** (success).
     Damage: 1d10 = **7** − bandit DR 2 = **5** damage. Bandit HP: 24 → **19**.

**Round 1 result:** Bandit took 11 total damage (6 + 5). Aldric unscathed.

---

### Example 2: TTRPG Turn-Based — Critical Hit with Effects Table

Continuing from Example 1, Round 2:

1. **Aldric's turn (Round 2):**
   - **Action:** Attacks bandit again.
   - Roll Blades 72 → d100 = **natural 04** (inside crit window, LCK 4 → window 01–07).
   - **Confirmation:** Roll d100 ≤ effective skill (72) + `floor(LCK/2) = +2` → target **74**.
     Roll d100 = **52** (confirmed crit!).
   - **Damage dice:** 1d10 × 2 = (8 × 2) = **16** slashing.
   - **Critical effects table:** d100 = **44** → **Disarm**. Bandit's weapon flies
     1d4 = **3** units away.
   - **Armor DR:** 16 − bandit DR 2 = **14** damage.
   - Bandit HP: 19 → **5**. Bandit is now disarmed and at critical HP.

2. **Bandit's turn (Round 2):**
   - **No weapon.** Bandit uses movement to retrieve shortsword (3 units away).
   - **No action left** to attack. Ends turn.

3. **Aldric's OA:** None (bandit did not leave reach).

**Round 2 result:** Bandit nearly dead, disarmed for 1 turn. Aldric can finish
next round.

---

### Example 3: cRPG Party Combat — Tactical Engagement

**Setup:**
- **Valeria** (elven archer): Bows 78, Evasion 55, light_armor 60. Wears studded
  leather (torso/upper, DR 3) + leather boots (feet/skin, DR 1) = **4 DR**.
  Longbow (1d8 piercing). LCK 6.
- **Dorn** (dwarven tank): Blunt 68, Block 70, heavy_armor 72. Wears full plate
  (DR 17 total across all slots) + tower shield (DR +4 on block). Warhammer (1d10
  bludgeoning). STR 10.
- **Lyra** (human mage): Red_magic 65, Evasion 20, light_armor 15. Wears cloth
  robe (DR 0). Unarmored defense. INT 9.
- **Enemies:** 3 goblins (Blades 35, Evasion 20, DR 1 each, HP 15).

**Round 1:**

1. **Initiative:** Valeria 22, Dorn 14, Lyra 9, Goblins 7. Valeria acts first.

2. **Valeria's turn:**
   - **Called shot (head):** −20 penalty. Effective Bows = 78 − 20 = **58**.
     Roll d100 = **44** (success).
   - **Damage:** Longbow 1d8 = **5** × 1.5 (head multiplier) = **7** (rounded up).
   - Goblin 1 DR 1 → net **6 damage**. Goblin 1 HP: 15 → **9**.
   - **Special effect:** Head shot staggers. Goblin 1 loses its reaction.
   - **Movement:** Valeria repositions to high ground (+10 to-hit next round).

3. **Dorn's turn:**
   - **Action:** Charges Goblin 2. Roll Blunt 68 → d100 = **23** (success).
     Damage: 1d10 = **9** − goblin DR 1 = **8**. Goblin 2 HP: 15 → **7**.
   - **Bonus action:** Dorn raises tower shield (readies block reaction).
   - **Movement:** Blocks the corridor, preventing goblins from reaching Lyra.

4. **Lyra's turn:**
   - **Action:** Casts *Fire Bolt* (red_magic) at Goblin 3. Roll red_magic 65 →
     d100 = **51** (success).
   - **Damage:** Effect magnitude 4 (fire). Goblin 3 has no fire resistance.
     Net **4 damage**. Goblin 3 HP: 15 → **11**.
   - **Movement:** None (maintains position behind Dorn).

5. **Goblins' turn:**
   - **Goblin 1** (staggered, no reaction): Attacks Valeria with shortbow.
     Roll Blades 35 → d100 = **67** (miss).
   - **Goblin 2:** Attacks Dorn. Roll 35 → d100 = **28** (success).
     Dorn's block reaction: Block 70 → d100 = **22** (success). Tower shield
     negates 4 DR. Warhammer damage 1d6 = **3** − 4 = **0**.
   - **Goblin 3:** Attempts to run past Dorn to reach Lyra. OA from Dorn:
     Roll Blunt 68 → d100 = **31** (success). Damage: 1d10 = **6** − goblin DR 1
     = **5**. Goblin 3 HP: 11 → **6**. Goblin 3 stops (cannot afford the OA).

**Round 1 result:** Goblins heavily wounded (9, 7, 6 HP). Party at full health.
Goblins bottled up by Dorn's position.

---

### Example 4: Video Game Action-Combat — Rogue vs. Guard Captain

**Setup:**
- **Kestrel** (shadow elf rogue): Piercing 68, Evasion 60, light_armor 55. Wears
  studded leather jerkin (torso/upper, DR 2, leather) + leather boots (feet/skin, DR 1)
  + leather gloves (hands/skin, DR 0) + cloth hood (head/skin, DR 0) = **3 total DR**.
  Wields a rapier (1d6 piercing, finesse, parry-capable) and a dagger (off-hand).
  Attributes: STR 5, AGI 9, END 5, LCK 7. Stamina pool: `15 + 5×5 + 10×2 = 60`.
  Has the *Ghost* perk (−25% detection).
- **Captain Voss** (human guard): Blades 62, Block 45, heavy_armor 60. Wears
  breastplate (torso/upper, DR 4, iron, evasion −15, speed −2) + iron helmet
  (head/upper, DR 2) + leather gauntlets (hands/upper, DR 1) + leather boots
  (feet/skin, DR 1) = **8 total DR**. Wields a longsword (1d8 slashing). Standard
  shield (DR +2 on block). No shield equipped (two-handed longsword). Attributes:
  STR 8, END 7, AGI 4.

**Encounter (Video Game / Action-Combat Mode):**

1. **Engagement:** Kestrel enters melee range. Voss detects her (PER check failed
   due to Ghost perk, but proximity triggers aggro). Combat HUD activates.

2. **Kestrel opens:**
   - **Input:** Presses light attack (rapier) → drains **8 Stamina** (52 remaining).
   - **Auto-hit check:** Voss's auto-dodge: base 10% + (evasion 25 × 0.15 = 3.75%)
     + (AGI 4 × 2 = 8%) ≈ **22% dodge chance**. Roll d100 = **73** (does not dodge).
   - **Hit confirmed.** Rapier 1d6 = **5** piercing − Voss's DR 8 = **0 net damage**
     (fully absorbed by armor). Gray "0" damage popup.
   - **Kestrel recovers** (0.3s), circles.

3. **Voss counter-attacks:**
   - **AI attack:** Voss executes a heavy swing (longsword). Drains 12 Stamina.
   - **Kestrel's player input:** Taps block (parry attempt). Parry window:
     `piercing 68 × 0.005 = 0.34s`. Input timing succeeds → **parry!**
   - **Parry result:** Attack negated. Voss is staggered (1s stun, cannot act).
   - **Kestrel's follow-up:** During Voss's stagger window, Kestrel inputs
     light attack + off-hand dagger strike (dual-wield chain). Drains **14 Stamina**
     total (38 remaining).
   - **First hit (rapier):** 1d6 = **4** − DR 8 = 0 (absorbed).
   - **Second hit (dagger):** 1d4 = **3** piercing − DR 8 = 0 (absorbed).
   - Both hits show gray damage (0 net). Kestrel's weapons cannot penetrate Voss's
     armor with chip damage alone.

4. **Kestrel adapts:**
   - **Input:** Dodges backward (dodge roll, drains 10 Stamina, 28 remaining).
     Invincibility frames active for 0.3s.
   - **Draws and applies** a paralytic poison coating (consumable, uses 1 charge).
     Next successful hit will apply `paralyze` effect (id 11) → `paralyzed` condition.
   - **Input:** Lunging attack (heavy, drains 15 Stamina, 13 remaining).
   - **Auto-hit check:** Voss's dodge rolls **87** (miss). Hit confirmed.
   - **Damage:** Rapier 1d6 = **6** − DR 8 = 0 (armor still holds).
   - **Poison triggers:** Paralyze effect (id 11) bypasses armor. Voss must resist:
     END save vs DC 12. Roll d100 = **34** (fail). Voss is **paralyzed** for 4 seconds.
   - Condition icon appears on Voss's HUD. He is immobilized and defenseless.

5. **Kestrel's finishing sequence:**
   - **Paralyzed target:** Melee attacks against paralyzed targets auto-crit (TTRPG)
     or deal ×2 damage (video game).
   - **Rapier crit:** 1d6 × 2 = **12** − DR 8 = **4 net damage** (orange popup).
     Voss HP: 40 → **36**.
   - **Dagger crit:** 1d4 × 2 = **6** − DR 8 = 0 (gray).
   - Voss remains paralyzed. Kestrel can continue attacking until paralysis ends.

**Result:** Kestrel cannot brute-force through Voss's armor (DR 8 exceeds her
base damage). By using a consumable (paralytic poison) to bypass the physical
defense tree and trigger a condition, she lands meaningful damage. The parry
(stagger) created the opening to apply the poison. Video game rhythm: parry →
stagger → apply poison → capitalized damage during paralysis.

---

### Example 5: TTRPG — Magic Defense Tree, Condition Spread, and Active Effects

**Setup:**
- **Lyra** (human mage, from Example 3): Red_magic 65, Evasion 20, WIL 8, INT 9.
  Cloth robe (DR 0). Mana pool: `9×8 + 10×2 = 92`. No armor evasion penalty.
- **Cinder** (fire salamander, caster enemy): Red_magic 55, Evasion 30, WIL 6.
  Natural scales provide DR 3. Has `spell_absorption` (effect id 21, 15% chance).
  Knows *Fire Bolt* (damage_fire, magnitude 6) and *Blinding Flash* (blind effect
  id 47, 4s duration).
- **Terrain:** Open field, no cover. Distance 8 units.

**Round 1:**

1. **Initiative:** Lyra rolls d20 + PER (5) = **17**. Cinder rolls **12**. Lyra acts first.

2. **Lyra's action (cast *Fire Bolt*):**
   - **Cast check:** Roll red_magic 65 → d100 = **38** (success, margin 27).
   - **Magic Defense Tree Resolution:**
     - **Layer 1 (Evasion):** No perk. Baseline evasion **does not apply** to magic.
     - **Layer 2 (Spell Absorption):** Cinder has 15% spell absorption. Roll d100 =
       **73** (no absorption).
     - **Layer 3 (Magic Resistance):** Cinder's WIL-based resistance. WIL 6 gives
       base 24%. Roll d100 = **61** (no resist).
     - **Layer 4 (Elemental/Phase Resistance):** Fire vs. fire. Cinder is a fire
       salamander — innate fire resistance (resist effect id 19, parameter "fire",
       magnitude 50%). **50% reduction**.
     - **Layer 5 (Active Effects):** None active.
   - **Final damage:** Fire Bolt magnitude 6 × 50% (fire resistance) = **3 fire damage**.
     Cinder's natural DR 3 does not apply to magic (physical DR only).
     Cinder HP: 40 → **37**.

3. **Cinder's action:**
   - **Cast check:** Roll red_magic 55 → d100 = **29** (success, margin 26).
   - **Chooses *Blinding Flash* (blind effect id 47) instead of direct damage.**
   - **Magic Defense Tree Resolution:**
     - **Layer 1:** No perk — evasion does not apply.
     - **Layer 2:** Lyra has no spell absorption.
     - **Layer 3:** Lyra's WIL 8 → base MR 32%. Roll d100 = **44** (fail, no resist).
     - **Layer 4:** Blind has no phase (non-elemental). No elemental resistance.
     - **Layer 5:** No active effects.
   - **Blind lands.** Lyra is now **blinded** (condition id 1, applied by effect id 47).
     - **TTRPG effect:** −20 on PER-based checks, auto-fail sight-dependent rolls.
       Attacks against Lyra have +10.
     - Duration: 4 rounds (Cinder's base duration modified by margin of success).

4. **Lyra's recovery (Round 2):**
   - **Blinded penalty active.** Lyra attempts to cast *Cure* (effect id 20,
     parameter "blinded") on herself — but sight-dependent targeting is required.
     **Auto-fail.** Lyra cannot target herself with a visual spell.
   - **Alternative:** Lyra uses a pre-prepared potion (consumable, cure effect id 20
     with parameter "blinded").
   - **Drink potion** (use object, 1 AP or input action). Blind removed.
   - **Back to full effectiveness.** Lyra re-engages.

**Result:** The magic defense tree absorbed most of Lyra's Fire Bolt (50% fire
resist). Cinder's Blinding Flash bypassed armor entirely — conditions are a
powerful tool against casters who lack physical protection. Lyra's prepared
consumable (a potion) saved her from multiple turns of ineffectiveness.

---

## 11. Combat Flow Checklist

### TTRPG (Turn-Based) — Full Round Flow

```
┌─ START OF ROUND ─────────────────────────────────────────┐
│ 1. Roll initiative (d20 + PER) if first round             │
│ 2. Refresh reactions for all participants                 │
│ 3. Refill AP pools (AP variant)                           │
│ 4. Process start-of-round effects (bleed, burn, regen)    │
└───────────────────────────────────────────────────────────┘
        │
        ▼
┌─ EACH CHARACTER'S TURN (initiative order) ───────────────┐
│ ☐ Declare reactions (AP variant: reserve AP from pool)    │
│ ☐ Take Action (one): Attack / Cast / Dash / Disengage /  │
│   Dodge / Hide / Ready / Use Object / Help                │
│ ☐ Take Bonus Action (one): Off-hand attack / Quick Cast   │
│ ☐ Move (split before/after action; speed in units)        │
│ ☐ End turn → process end-of-turn effects                  │
│   (duration ticks, save ends)                             │
└───────────────────────────────────────────────────────────┘
        │
        ▼
┌─ REACTIONS (triggered, any turn) ────────────────────────┐
│ ☐ Opportunity attack (enemy leaves reach)                 │
│ ☐ Block (incoming melee, declared before hit)             │
│ ☐ Parry (incoming melee, weapon skill, declared before)   │
│ ☐ Evasion / Dodge (incoming attack, declared before)      │
│ ☐ Perk-triggered reactions (intercept, guard ally, etc.)  │
└───────────────────────────────────────────────────────────┘
        │
        ▼
┌─ END OF ROUND ───────────────────────────────────────────┐
│ 1. Process end-of-round effects (duration ticks, saves)   │
│ 2. Refund unused reaction AP (AP variant)                 │
│ 3. Start next round                                       │
└───────────────────────────────────────────────────────────┘
```

### Video Game (Action-Combat) — Combat Loop

```
┌─ ENGAGEMENT ──────────────────────────────────────────────┐
│ 1. Aggro/threat system determines enemy targets           │
│ 2. Player enters combat range → enemies aggro             │
│ 3. Combat music / HUD elements activate                   │
└───────────────────────────────────────────────────────────┘
        │
        ▼
┌─ GAMEPLAY LOOP ──────────────────────────────────────────┐
│ (continuous, player-input-driven)                         │
│                                                          │
│ ☐ Move (free, WASD; sprint costs Stamina)                │
│ ☐ Attack (button press, drains Stamina)                  │
│   → Auto-dodge check (evasion skill %, automatic)         │
│   → Hit → armor DR applied → damage popup                │
│ ☐ Block (hold/timed button; skill determines window)     │
│ ☐ Parry (tap within narrow window, high risk/reward)     │
│ ☐ Use ability / item (cooldown-gated, Stamina cost)      │
│ ☐ Dodge roll (invincibility frames, Stamina cost)        │
│ ☐ Regenerate (Stamina regens from WIL, slow out of combat)│
└───────────────────────────────────────────────────────────┘
        │
        ▼
┌─ DISENGAGEMENT ──────────────────────────────────────────┐
│ 1. All enemies in area defeated → combat ends             │
│ 2. Out-of-combat regen kicks in (HP, Mana, Stamina)       │
│ 3. Cooldowns reset or begin reduced recovery              │
└───────────────────────────────────────────────────────────┘
```

### cRPG (Party-Based Tactical)

For party-based tactical cRPGs, the TTRPG checklist applies with these additions:
- **AI-controlled companions** follow scripted tactics (aggro nearest, protect squishy, focus fire).
- **Auto-pause triggers:** enemy death, OA triggered, new enemy spotted, low HP.
- **Queue system:** actions can be queued during auto-pause for simultaneous execution.
- **Camera/positioning:** grid or free-placement with LoS checking for ranged attacks.

---

## 12. Edge Cases & Common Questions

### Attacking While Prone
- **TTRPG:** −20 penalty to melee attack rolls. Ranged attacks are impossible without the *Point Blank Shot* perk (or similar). Cannot take the Dash or Dodge action.
- **Video game:** Cannot attack during the knockdown animation. Partial attack penalty (−30%) during the recovery (standing) animation.

### Attacking Invisible Targets
- **TTRPG:** −20 penalty on attack rolls (blind swing). A PER-based check (DC 15 + target's stealth skill / 10) may reveal the square. Area-of-effect spells hit regardless. Damage still applies on a hit.
- **Video game:** The invisible target cannot be locked on. Area attacks and cone abilities still hit. A detection effect (e.g. *Detect Life*) reveals the target.
- **Perception perks:** *Blind Fighting* reduces the penalty to −10. *Tremorsense* negates it entirely if the target is on the ground.

### Shooting Into Melee
- **TTRPG:** −10 penalty if an ally is engaged in melee with the target. On a miss (roll > modified skill but within 10 of the target), the attack hits a random adjacent creature in melee (GM determines). The *Sharpshooter* perk negates both penalty and friendly fire.
- **Video game:** Friendly fire is on by default (damaging allies). A perk or difficulty setting may disable it. The penalty is automatic aim deviation if an ally is between the shooter and target.
- **cRPG note:** With grid positioning, check LoS — if an ally occupies a grid cell between attacker and target, apply the penalty.

### Invisible Target in Melee
- An invisible target still provokes OA when leaving reach — you can hear them move. The OA roll takes the −20 blind penalty.
- Invisibility breaks when the target attacks, casts a spell, or takes damage (GM call for TTRPG; automatic for video game).

### Stacks & Durations (Common Rulings)
| Question | Answer |
|----------|--------|
| Do same-condition effects stack? | No — refresh duration (or increase stack counter if `stacking: true`). |
| Do same-effect (id) buffs stack? | No — only the highest magnitude applies (`fortify` default). Effects with `stackable: true` do stack. |
| Does a `cure` effect remove all conditions? | Only if its `parameter` matches the condition key, or is `"all"`. |
| Can you block while stunned? | No. Stunned and incapacitated prevent reactions. |
| Does DR reduce elemental damage? | Only if specified (e.g. fire resistance DR). Physical DR does not reduce magic damage. |

### Mounted Combat (Stub)
- **Mounted character:** Uses the mount's speed and movement. The mount acts on the rider's initiative.
- **Mounted attacks:** The rider attacks with their own weapon skill. Melee attacks against a mounted target can hit either rider or mount (GM decides or random 50/50).
- **Dismounting:** Voluntary (costs half movement) or forced (knockdown check for rider when mount is killed or tripped).
- **Charge attacks:** Moving 4+ units in a straight line before attacking grants +2 damage per unit moved (capped at +10). Requires a Ride check (DC 12) to maintain balance.

### Underwater Combat (Stub)
- **Movement:** Speed halved. Only piercing weapons deal full damage; slashing and bludgeoning deal half.
- **Ranged:** Thrown and projectile weapons are ineffective beyond 1 unit. Crossbows and magic work normally.
- **Breath:** Characters can hold breath for `END × 15` seconds. After that, begin suffocation (1 HP/round, save END DC 10 + rounds without air).
- **Spellcasting:** Fire-phase spells and effects are nullified underwater. Water-phase spells are amplified (×1.5).

---

## 13. Balance Guidelines

These are **reference values** for GMs and developers building encounters, equipment, and character options. Actual tuning depends on campaign difficulty and game mode.

### Expected Damage per Hit by Tier

| Tier | Level Range | Weapon Die (Typical) | Expected Pre-DR Damage | Typical Enemy HP | Hits to Kill |
|------|-------------|---------------------|----------------------|-----------------|--------------|
| Early | 1–5 | 1d6–1d8 (3–5 avg) | 4–7 | 15–30 | 3–6 |
| Mid | 6–14 | 1d8–1d10 (5–7 avg) | 7–12 | 30–60 | 3–7 |
| High | 15–20 | 1d10–2d8 (7–11 avg) | 10–16 | 60–100 | 4–8 |
| Endgame | 21–30 | 2d8–2d10 (9–13 avg) | 14–22 | 100–200 | 5–10 |

### Armor DR by Tier

| Tier | Light Armor DR | Medium Armor DR | Heavy Armor DR | Expected Material |
|------|---------------|-----------------|----------------|-------------------|
| Early | 1–3 | 2–4 | 3–6 | Leather, Iron |
| Mid | 2–4 | 3–6 | 5–10 | Bronze, Steel |
| High | 3–5 | 5–8 | 8–14 | Steel, Elven Wood |
| Endgame | 4–6 | 6–10 | 10–18 | Steel + enchanted materials |

**Total character DR** (all slots) is roughly:
- Light: 3–8 (e.g. studded leather jerkin + boots + gloves)
- Medium: 8–18 (e.g. chainmail + coif + greaves)
- Heavy: 14–28 (e.g. full plate + helm + gauntlets + greaves)

### Hit Points by Role

| Role | END | Level 1 HP | Level 10 HP | Level 20 HP | Level 30 HP |
|------|-----|-----------|-------------|-------------|-------------|
| Glass cannon (mage) | 4 | 47 | 191 | 371 | 551 |
| Skirmisher (rogue) | 5 | 55 | 215 | 415 | 615 |
| Frontline (fighter) | 7 | 71 | 263 | 503 | 743 |
| Tank (paladin) | 9 | 87 | 311 | 591 | 871 |

Formula: `15 + END × 8 + level × 4` (per level after 1st: `END × 8 + 4`).

### How Many Hits Can a Character Take?

Assuming a mid-tier character (level 10, END 7, 263 HP) facing an enemy that deals 9 damage per hit:

| Armor | Net Damage per Hit | Hits to Down |
|-------|-------------------|--------------|
| None | 9 | 29 |
| Light (DR 4) | 5 | 52 |
| Medium (DR 8) | 1 | 263 |
| Heavy (DR 14) | 0 | — |

Heavy armor nearly negates low-to-mid-tier physical damage. This is intentional — heavy armor is the tank's defining feature. Counterplay exists via magic, called shots (head/arm ignore partial DR), effects that expose or corrode armor, and flanking bonuses that improve hit chance.

### Skill Benchmarks

| Level | Expected Weapon Skill | Expected Evasion | Expected Block | d100 Hit Chance vs. Equal Foe |
|-------|----------------------|------------------|----------------|------------------------------|
| 1 | 20–30 | 15–25 | 15–25 | 25–35% |
| 5 | 35–50 | 25–40 | 25–40 | 40–55% |
| 10 | 50–65 | 35–55 | 35–55 | 55–70% |
| 20 | 70–85 | 55–75 | 55–75 | 75–90% |
| 30 | 85–100 | 70–90 | 70–90 | 85–95% |

### Quick Balance Heuristics

- A fair 1v1 fight: both sides should down each other in **4–6 hits** (accounting for misses, blocks, DR).
- A "boss" enemy should take **8–12 party hits** to down (with 3–4 party members).
- A "minion" should go down in **1–2 hits**.
- Armor DR should never exceed the expected enemy damage per hit (or combat becomes stalemate). If DR ≥ expected damage, the enemy needs magic, called shots, armor-piercing traits, or environmental advantages.
- Crit chance of 5–10% (before LCK) means roughly 1 crit per 10–20 attack rolls. With LCK 10, roughly 1 in 10 attacks crits.

---

## 14. Video Game Implementation Notes

### Timed Block / Parry System

| Mechanic | Block | Parry |
|----------|-------|-------|
| Input | Hold button (or tap before hit) | Tap within narrow window |
| Window length | `blockSkill × 0.01` seconds (e.g. skill 70 → 0.7s window) | `weaponSkill × 0.005` seconds (skill 70 → 0.35s) |
| On success | Damage reduced by shield DR % | Attack negated (or reflected with perk) |
| On failure | Full damage (shield provides no passive DR) | Full damage + recovery animation lock |
| Cooldown | None (but Stamina cost per block) | 0.5–1s recovery |
| Visual feedback | Shield flash, clang sound, damage popup with "Blocked!" text | Weapon ring, spark effect, enemy stagger animation |

**Implementation recommendations:**
- Use a **two-phase input model**: the block button enters a "ready" stance (reduced movement, Stamina drain per second), releasing or getting hit triggers the block. This avoids pixel-perfect timing frustration.
- Parry should be strictly **optional** — design the game so that holding block is viable for casual play, while parry rewards mastery.
- Show a **block indicator** (shield icon) when the player is in the block-window frame.

### Auto-Dodge Calculation

Video game evasion is a **passive percentage** computed from:
```
Dodge% = baseEvasion% + (evasionSkill × 0.15) + (AGI × 2)
       − sum(armorEvasionPenalties) − encumbrancePenalty
```

- Rolled automatically on each incoming attack.
- Visually: a "dodge" animation plays (lean back, sidestep, or roll based on distance to attacker).
- Dodging an attack should briefly interrupt the enemy's combo (giving the player a window to counter).
- Overcapped dodge (>100%) grants a chance to **counter-attack** (optional mechanic).

### Aggro / Threat System

| Action | Threat Generated | Notes |
|--------|-----------------|-------|
| Deal damage | Damage × 1.0 | Maintains aggro on current target |
| Taunt (effect id 63) | Fixed high value | Overrides current aggro for duration |
| Heal ally | Heal amount × 0.5 | Causes healer to gain threat from healed target's enemies |
| Block / Parry | Low constant | Does not significantly pull aggro |
| Stay in melee range | Passive decay | Being near the enemy slowly builds "presence" aggro |

- **Threat table:** each enemy tracks a numeric threat score per party member. Attacks the highest.
- **Threat decay:** Out of melee range for 5+ seconds → threat decays by 50%.
- **Tanking:** Tanks should generate 2–3× the threat of DPS characters to hold aggro without taunts.

### Damage Popup UI

Display damage numbers clearly with color coding:
- **White:** Normal physical damage
- **Orange:** Critical hit
- **Blue:** Magic damage
- **Green:** Healing / restoration
- **Red:** Damage over time (tick)
- **Gray:** Damage fully absorbed (0 net damage)

Format: `12 - 4 (DR) = 8` — shows pre-DR, DR, and net. Optionally collapse to just `8` for clean HUD.

### Animation & Timing Guidelines

- **Attack recovery:** After a melee attack, 0.3–0.5s recovery before the next action. Heavy weapons: 0.6–0.8s.
- **Hit stun:** On taking damage, brief animation lock (0.1–0.2s). Prevents infinite stun-locks.
- **Knockdown:** 0.5–0.8s fall animation, 0.5s get-up. Can be interrupted by high-damage attacks during get-up.
- **Dodge roll:** 0.3–0.5s animation with i-frames (invincibility frames) for the middle 60% of the animation.
- **Stamina regen delay:** After spending Stamina, a 0.5s delay before regen begins. Heavy actions add a longer delay (1–2s).

---

## 15. Deferred Combat Sections

The following areas are noted for future expansion. They are **not yet implemented** — the entries below are design notes for later specification.

### Mounted Combat

**Future / Not Yet Implemented.** A mounted character uses the mount's speed and movement instead of their own. The mount acts on the rider's initiative (shared turn). The rider attacks with their own weapon skill. Attacks against a mounted target can hit either rider or mount (GM determination or 50/50 random in TTRPG; automatic target selection in video game). Dismounting costs half movement (voluntary) or requires a knockdown check (forced). **Charge attacks** (moving 4+ units straight before attacking) grant bonus damage: +2 per unit moved (capped at +10) for TTRPG; a flat % multiplier for video game. Requires a Ride check to maintain balance.

- **TTRPG:** Ride skill (AGI-based) for mounted maneuvers; DC 12 base for combat riding. Mount has its own HP pool and acts as additional HP buffer for the rider.
- **Video game:** Mount is a separate entity with its own HP, Stamina, and AI. The rider controls the mount's movement; attacks are made from the saddle. Mount death triggers a dismount animation and stun.

### Underwater Combat

**Future / Not Yet Implemented.** Movement is halved. Only piercing weapons deal full damage; slashing and bludgeoning deal half. Thrown and projectile weapons are ineffective beyond 1 unit; crossbows (with special underwater bolts) and magic work normally. Characters hold their breath for `END × 15` seconds, then begin suffocation (1 HP/round, END save DC 10 + rounds without air). Fire-phase spells and effects are nullified; water-phase spells are amplified (×1.5).

- **TTRPG:** GM tracks breath and movement penalties. Swim skill checks for complex maneuvers. Suffocation begins after breath-hold duration expires.
- **Video game:** Underwater sections use a breath meter UI. Movement speed debuff applied automatically. Weapon damage type check on hit.

### Vehicle Combat

**Future / Not Yet Implemented.** Chases, ship-to-ship, and siege engine operation. Vehicles have their own stat block (speed, HP, crew capacity, weapon mounts). Vehicle-to-vehicle combat uses the crew's relevant skills (gunnery, piloting/navigation). Boarding actions revert to standard combat rules on the vehicle's deck.

### Mass / Group Combat

**Future / Not Yet Implemented.** Proposed design uses a **unit abstraction**: formations of 10–50 combatants are represented as a single "unit" with pooled HP, damage output based on member count and quality, and morale tracking. Commander checks (CHA and relevant tactical skills) modify unit effectiveness. Skirmisher characters can operate independently (as "hero units") using standard combat rules while the mass battle resolves around them.

- **TTRPG:** Resolution uses unit-vs-unit opposed checks (tactical skill of the commander) rather than per-soldier rolls. Morale thresholds trigger rout or surrender. Player characters act as hero units or squad leaders.
- **Video game:** Large battles use simplified crowd AI with hero characters as focal points. Units follow formation waypoints and auto-engage nearby enemies. Performance optimization (LOD, crowd instancing) is engine-specific.

---

## 16. Combat Perks & Traits

Perks (from `data/features/core.json`) are the primary way characters customize their combat capabilities. This section maps existing perks to combat mechanics and provides guidance for designing combat-themed perks.

### Existing Combat Perks

| Perk | Effect on Combat | Type | Prerequisites |
|------|-----------------|------|---------------|
| **Toughness** | +15 max HP per rank (up to 3 ranks). Increases survivability across all combat modes. | perk | END 5 |
| **Alert** | +10 initiative (TTRPG); cannot be surprised. Video game: larger aggro detection radius, faster reaction to flankers. | perk | PER 6 |
| **Power Attack** | +25% melee weapon damage. Stacks additively with other damage multipliers. Most effective with heavy, slow weapons. | perk | STR 6, Blades 40 |
| **Deadeye** | +10% ranged critical chance. Compounds with LCK's crit window expansion. | perk | PER 6, any archery 40 |
| **Shield Wall** | +15% damage blocked. Multiplies the effective DR of a successful block. Works with all shield types. | perk | END 6, Block 40 |
| **Fast Shot** (creation) | Ranged attacks 20% faster. Video game: reduced recovery time. TTRPG: may allow additional shot (GM discretion). Cannot make aimed shots. | creation | — |
| **Field Medic** | +50% healing from Medicine skill. Relevant in combat for stabilizing allies during or after engagement. | perk | Medicine 40 |

### Perk-Equipped Example Builds

**The Tank (Dorn):** Shield Wall (+15% block) + Toughness (+45 HP over 3 ranks) → effective HP pool for holding chokepoints. Synergizes with heavy armor and tower shield.

**The Sharpshooter (Valeria):** Deadeye (+10% crit) + high LCK (expanded crit window) → ~15–20% crit rate on called shots. Video game: auto-crit on headshots with Deadeye active.

**The Berserker:** Power Attack (+25% damage) + high STR (carry heavy weapons without penalty) + low armor (unarmored or light for speed). Glass-cannon melee build.

### Designing Combat Perks

When adding new combat perks, follow these guidelines:

1. **Effect magnitude:** A single perk rank should provide a +10–25% improvement to a specific combat dimension (damage, defense, crit, speed). Rarely exceed +50% in one perk.
2. **Prerequisites:** Gate combat perks behind the relevant attribute (STR for melee damage, PER for ranged, END for defense/block) and a skill threshold of 30–50.
3. **Dual resolution:** Every perk's effect must be expressible in both TTRPG and video game terms. See examples above.
4. **Stacking:** Two perks that modify the same dimension should stack additively, not multiplicatively, to avoid runaway scaling.
5. **Effects integration:** Perks that apply ongoing modifiers should reference the shared effects registry by effect ID where possible (e.g., a "Spell Dodger" perk grants resist effect id 19 with parameter "magic" at magnitude 25).

### Racial Traits Affecting Combat

Racial traits from `data/traits/core.json` also modify combat:

- **Elf — Fey Ancestry:** Resistance to charm and mind-affecting magic (relevant in the magic defense tree, Layer 3–4).
- **Elf — Keen Senses:** +PER-based bonus; helps initiative and perception in combat.
- **Dwarf — Dwarven Resilience:** +25% poison/disease resistance; affects combat vs. poison-using enemies.
- **Orc — Berserker Strength:** Temporary STR boost on taking damage; affects melee damage output.

### Creation Perks (Traits) with Combat Impact

Creation perks chosen at character creation have permanent combat implications:

- **Fast Shot:** Ranged attack speed +20%, but no aimed shots. Build-enabling for rapid-fire archers.
- **Small Frame:** +1 AGI (action economy, evasion) but −25% carry capacity (equipment restrictions).
- **Undead Phobia:** Combat penalties vs undead but +2 WIL (magic resist) and 25% undead/poison resistance.
- **Iron Allergy:** Metal armor/weapons penalized; forces reliance on leather, wood, or exotic materials.
- **Leather Allergy:** Leather armor unusable; forces cloth or metal armor.

---

## 17. Open Design Questions & Resolutions

This section addresses the open questions raised during the combat system design (originally in `plan.md`).

### Slot + Layer Complexity

**Question:** 16+ equipment positions is a lot of equipment management. How should a video game handle this?

**Resolution:** The slot+layer model is the **canonical data model** — it is the single source of truth for what a piece of armor is. Consuming engines may simplify the player-facing UI without changing the data:

- **TTRPG:** All 16 positions are on the character sheet. Paper tracking is fine; groups may use a simplified sheet that folds "on_top" into the back slot and "skin" into the appropriate slot as base clothing.
- **Video game (recommended simplification):** Hide the layer system behind an **auto-undergarment** mechanic. The engine automatically equips the best available skin-layer item when a middle or upper layer item is added to a slot. The player only manages visible armor (middle + upper slots in each body zone). Layer dependency penalties (−1 AGI for missing skin layer) are handled automatically.
  - Alternative: Use a **3-zone UI** (Head, Body, Legs+Feet) with layered slots as sub-tabs within each zone. Hands and Back are separate top-level slots.
- **cRPG:** Show all layers in an inventory grid with slot filtering, similar to Diablo-style equipment screens. Tooltips indicate layer dependencies and total DR contributions.

### Shield Slot and Layer

**Question:** How do shields work in the slot/layer model? Are they an offhand slot? Do they have a layer?

**Resolution:** Shields use `slot: "offhand"` with `layer: null` (no layering). Key rules:

- Shields **do not contribute to passive DR**. Their DR only applies during an active block (see §3).
- Shields occupy the offhand position. Two-handed weapons implicitly occupy both main hand and offhand.
- A character with a shield in the offhand can still carry a light item (wand, torch, dagger) in the shield hand — the shield is strapped to the arm.
- Bucklers allow a light off-hand weapon simultaneously (the buckler is small enough to coexist).
- In video game rendering: the shield appears on the character's arm/side. The block animation brings it to the front.
- Shield `evasionPenalty` applies at all times (the shield is always worn, even when not actively blocking).
- Shield `strengthRequirement` gates effective use (like armor STR requirements).

### Condition → Effect Linking

**Question:** Some conditions (blinded, charmed, etc.) use placeholder effect IDs (47–64). Are these real effects in the registry?

**Resolution:** Yes. Effects 47–64 were added alongside the conditions system (see `data/effects/core.json`):
- id 47 `blind` → condition `blinded` (id 1)
- id 48 `charm` → condition `charmed` (id 2)
- id 49 `deafen` → condition `deafened` (id 3)
- id 50 `grapple` → condition `grappled` (id 5)
- id 51 `incapacitate` → condition `incapacitated` (id 6)
- id 52 `petrify` → condition `petrified` (id 9)
- id 53 `knockdown` → condition `prone` (id 11)
- id 54 `restrain` → condition `restrained` (id 12)
- id 55 `stun` → condition `stunned` (id 13)
- id 56 `unconsciousness` → condition `unconscious` (id 14)
- id 57 `burn` → condition `burning` (id 16)
- id 58 `bleed` → condition `bleeding` (id 17)
- id 59 `slow` → condition `slowed` (id 18)
- id 60 `silence` → condition `silenced` (id 19)
- id 61 `curse` → condition `cursed` (id 21)
- id 62 `expose` → condition `exposed` (id 22)
- id 63 `taunt` → condition `taunted` (id 23)
- id 64 `stagger` → condition `staggered` (id 24)

Every condition's `appliedBy` array references a real effect that exists in the shared registry. No placeholder effects remain.

### Shield Styles & Two-Weapon Fighting

**Clarification on two-weapon fighting and shields:**

- A character with two weapons may use the off-hand weapon to **parry** (using the off-hand weapon skill) but cannot **block** (no shield skill benefit).
- The off-hand weapon in a turn can either **attack** (as a bonus action in TTRPG, or as part of a chain in video game) OR be **reserved for parry** — not both in the same turn.
- Dual-wielding provides no passive DR benefit. The advantage is offensive throughput (extra attack) or defensive flexibility (parry capability), not protection.
- In video game mode, dual-wielding uses a dedicated attack chain animation set. Block is unavailable; the player relies on dodge rolls and parry timing.
