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

**AP pool variant:** AGI-based Action Points (`2 + floor(AGI / 3)` per turn, max 10,
carry over between turns) can replace the discrete action/bonus action model for
groups that prefer a resource-budget approach.

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
- **Block:** Declared reaction before the attack roll. The blocker makes a block skill
  check against the attacker's margin of success. On success, the shield's DR is
  subtracted from incoming damage. On failure, no block benefit.
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

### TTRPG

- **Confirmed crit:** Roll damage dice twice and sum both (alternatively: max dice + roll).
- **Optional:** Special effects table (disarm, bleed, stagger, sever limb) — GM discretion.
  Effects may include:
  | d20 | Effect |
  |-----|--------|
  | 1–5 | Extra damage only |
  | 6–10 | + target bleeds (1 HP/round until healed) |
  | 11–15 | + target disarmed (weapon dropped) |
  | 16–19 | + target staggered (loses next reaction) |
  | 20 | + sever limb or permanent injury |
- **Fumble:** Only on a natural **100** (not 96–00). Flavor-only for TTRPG:
  drop weapon, stumble, hit an ally by accident. No mechanical penalty enforced.
  High Luck does not affect fumble range.

### Video Game

- **Confirmed crit:** Fixed multiplier (×1.5 or ×2) applied after DR.
  `Final damage = (base damage × crit multiplier) - armor DR`
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

## 6. Initiative

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

## 7. Mode Comparison Table

| Aspect | TTRPG (turn-based) | Video game (action-combat) |
|--------|-------------------|---------------------------|
| **Actions** | Discrete Action/Bonus/Reaction | Player input per button; Stamina cost |
| **Movement** | Speed in units per action | Speed rating, free + sprint costs Stamina |
| **Reactions** | 1/round, declared | Timed block/parry input (optional); cooldowns |
| **OA** | Auto-trigger on disengage | Bonus on next attack vs disengaging enemy |
| **Initiative** | d20 + PER | Aggro/threat or proximity |
| **Physical defense** | Evasion check, Block reaction, flat DR | Auto-dodge %, timed block/parry, flat DR |
| **Magic defense** | Sequential rolls per layer (evasion only with perk) | Auto-mitigation from stats + effects (evasion only with perk) |
| **Critical hits** | Confirmation roll, double dice (or effects table) | Confirmation roll, ×1.5–2 multiplier |
| **Fumble** | On 100 only; flavor table | None — 100 is a miss |
| **Cover** | −10/−20 to-hit | LoS blockers only (no to-hit modifier) |
| **Flanking** | +10 to-hit | +25% damage |
| **Terrain** | Double movement cost | % speed slow |
| **Conditions** | GM tracks duration | Automated timers, visual indicators |

---

## 8. Deferred Combat Sections

The following areas are noted for future expansion but not yet specified:

- **Mounted combat:** Ride checks, mount speed, charge attacks, dismounting.
- **Underwater combat:** Movement penalties, breath tracking, weapon ineffectiveness.
- **Vehicle combat:** Chases, ship-to-ship, siege engine operation.
- **Mass combat:** Unit formations, morale, commander checks.
