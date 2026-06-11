# Combat Resolution

**Status:** Design (planned) · data structures exist (skills, conditions, equipment schema) · runtime resolution lives in the engine or GM adjudication.

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

## 10. Worked Combat Examples

### Example 1: TTRPG Turn-Based — Knight vs. Bandit

**Setup:**
- **Sir Aldric** (human knight): Blades 72, Block 55, Evasion 30, heavy_armor 65.
  Wears steel plate (torso/upper, DR 10) + steel helm (head/upper, DR 5) + steel
  gauntlets (hands/upper, DR 3) + steel greaves (feet/upper, DR 4) = **22 total DR**.
  Carries a longsword (1d10 slashing) and a standard shield (DR +2 on block).
  Attributes: STR 8, AGI 6, END 7, LCK 4.
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

## 11. Deferred Combat Sections

The following areas are noted for future expansion but not yet specified:

- **Mounted combat:** Ride checks, mount speed, charge attacks, dismounting.
- **Underwater combat:** Movement penalties, breath tracking, weapon ineffectiveness.
- **Vehicle combat:** Chases, ship-to-ship, siege engine operation.
- **Mass combat:** Unit formations, morale, commander checks.
