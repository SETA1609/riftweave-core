# Conditions & Status Effects

**Status:** Core (implemented as data + schema) · runtime resolution lives in the engine or GM adjudication.

Conditions are **binary state containers** that are applied and removed by effects from
the shared effect registry ([`data/effects/core.json`](../../ruleset/data/effects/core.json)).
Each condition is an envelope — it does not define its own mechanics inline; instead it
references the effect IDs that apply or remove it. The actual penalties and gameplay
consequences are described in the condition entry's `effects` fields (one for TTRPG,
one for video game) and implemented by the consuming engine or GM.

Data lives in [`ruleset/data/conditions/core.json`](../../ruleset/data/conditions/core.json),
governed by `condition.schema.json`.

---

## 1. Design: Effects-Driven Conditions

```
effect (e.g. paralyze id 11)
  → lands on target
    → applies condition (e.g. "paralyzed")
      → condition's mechanical effects take hold (per TTRPG or video_game mode)
        → removed when the applying effect expires, or when a cure effect (id 20) fires
```

**Rules:**
- A condition entry declares which **effect(s)** apply it via `appliedBy` (array of effect IDs).
- The `cure` effect (id 20) is the primary remover. Its `parameter` specifies which condition to cure.
- Conditions can also be removed by rest, taking damage (charm breaks), or special actions (stand from prone).
- A condition's mechanical effects are described in `effects.ttrpg` and `effects.video_game` — these are narrative/descriptive; the actual implementation is the engine's or GM's responsibility.
- Stacking conditions (like `exhaustion`) have `stacking: true` and `maxStacks`; each additional application increments the stack up to the max.

**Why effects-driven?**
Conditions do not exist in isolation. They are always the *result* of an effect
landing. By tying each condition to specific effect IDs, the data stays consistent:
- A spell that applies `paralyze` effect (id 11) automatically puts the target in the `paralyzed` condition.
- A `cure` spell (id 20) with parameter `"paralyzed"` removes it.
- New conditions can be added without schema changes — just add a new entry and point existing or new effects at it.

---

## 1b. Duration Types

Conditions follow one of three duration behaviors:

| Type | How It Works | Example |
|------|-------------|---------|
| `timed` | Has a finite duration in seconds. Expires automatically when the timer elapses. | Poison, burn, bleed (most combat conditions) |
| `permanent` | Stays indefinitely until explicitly removed by a `removedBy` effect. | Diseases (blight, fever, plague), curses |
| `until_dispelled` | Permanent in practice, but removable by a specific opposite effect rather than a generic cure. | Haste dispels Slow, frost dispels Burning |

### Opposite-Effect Dispel Pairs

Certain conditions can be removed by effects that are mechanically opposite, even if those effects are not classified as `cure`. This is data-driven — each condition's `removedBy` array in `conditions/core.json` lists the effect IDs that can remove it.

| Condition | Applying Effect | Dispelling Effect | Rationale |
|-----------|---------------|-------------------|-----------|
| Burning (id 16) | Burn (id 57) | Damage Frost (id 2) | Water (frost) overcomes Fire (burn) |
| Bleeding (id 17) | Bleed (id 58) | Restore Resource (id 7) | Healing stops bleeding |
| Slowed (id 18) | Slow (id 59) | Haste Attack (id 70) | Haste dispels Slow |
| Poisoned (id 10) | Poison (id 46) | Restore Resource (id 7) | Healing flushes poison |
| Any | Any | Cure (id 20) | Generic cure, the primary remover |

This allows conditions to interact naturally — applying a frost enchantment doesn't just deal damage, it also extinguishes any existing burn on the target.

---

## 2. Resolution Modes

### TTRPG (turn-based)
- Conditions are binary states with explicit duration tracked by the GM (in rounds or real time).
- Applied when the corresponding effect successfully lands on a target (the caster/attacker makes their skill check, the defender fails their resist check).
- Cured via rest, specific spells, or `cure` effects (id 20).
- The GM adjudicates edge cases (e.g. "can a blinded character still sense the invisible enemy?").
- Stacking conditions are tracked as counters (e.g. exhaustion at level 3/6).

### Video game (action-combat)
- Conditions are timed debuff/buff states. Applied automatically when an effect procs.
- Duration tracked by the engine in seconds or game ticks.
- Visual and audio indicators replace GM adjudication (screen flash, icon in HUD, sound cue).
- Stacking conditions show an intensity counter (e.g. "Exhaustion III").
- Immunity and resistance are computed from stats and active effects before the condition is applied.

---

## 3. Complete Condition Catalog

Data in `conditions/core.json` (24 entries). Each entry:

| Field | Meaning |
|-------|---------|
| `id` | Numeric primary key |
| `key` | Stable snake_case identifier |
| `label` | Human-readable name |
| `description` | Brief summary |
| `effects.ttrpg` | Mechanical effect in TTRPG/turn-based mode |
| `effects.video_game` | Mechanical effect in video game/action-combat mode |
| `appliedBy` | Effect IDs that can apply this condition |
| `removedBy` | Effect IDs that can remove this condition |
| `stacking` | Whether multiple instances stack |
| `maxStacks` | Maximum stack count |

| # | Key | TTRPG Effect | Video Game Effect | appliedBy (effect ids) |
|---|-----|-------------|-------------------|----------------------|
| 1 | blinded | −20 PER checks, auto-fail sight- dependent rolls, +10 to attackers | −20% accuracy, +10% enemy crit chance | 47* |
| 2 | charmed | Cannot attack charmer; charmer +20 social; damage breaks | Cannot target charmer; damage from charmer breaks | 48* |
| 3 | deafened | −10 initiative, auto-fail hearing PER | −10% detect range, no audio cues | 49* |
| 4 | frightened | −10 attack/checks while source in LoS, cannot approach | −10% damage, forced retreat if source closes | 22 (phobia) |
| 5 | grappled | Speed 0, −10 attacks/AGI, contested athletics to escape | Immobilized, −10% attack, escape via skill | 50* |
| 6 | incapacitated | No actions, bonus, or reactions | Cannot act or use abilities | 51* |
| 7 | invisible | Cannot be seen; −20 enemy attacks, +10 your attacks | Stealthed; attacks break stealth | 8 (invisibility) |
| 8 | paralyzed | Incapacitated + immobile; auto-fail STR/AGI; melee auto-crit | Immobilized + defenseless; melee hits guaranteed | 11, 43 |
| 9 | petrified | Incapacitated; DR 20 vs all; immune poison/disease | DR +90%, cannot act, no HP regen | 52* |
| 10 | poisoned | Sub_effects of the source poison apply as symptoms; resist reduces | DoT/debuff from poison's sub_effects; automated | 46 (poison) |
| 11 | prone | −20 melee attack, +10 vs ranged, half speed to stand | Knockdown; 1s to stand, ranged +15% vs you | 53* |
| 12 | restrained | Speed 0, −20 attacks/AGI, +10 attacker | Immobilized, −20% attack, incoming +15% damage | 54* |
| 13 | stunned | Incapacitated + immobile, −20 all, +10 attacker | Staggered (1–3s), cannot act, incoming +25% | 55* |
| 14 | unconscious | Incapacitated + prone; auto-fail STR/AGI; melee auto-crit | KO state; 0 HP, revive with healing | 56* |
| 15 | exhaustion | Stacking (6 levels). L1: −10 checks. L3: −20, speed halved. L5: speed 0. L6: death | Stacking debuff with diminishing stats. L6: death | 40, 41, 42 |
| 16 | burning | 1d4 fire/round; action to extinguish (AGI check, difficulty 10) | Fire DoT every 2s; extinguished by water/cure | 57 |
| 17 | bleeding | 1 HP/round; Medicine check (difficulty 10) or bandage stops | Health DoT every 2s; severity 1–3 | 58 |
| 18 | slowed | Speed halved; −10 AGI; no Dash; +1 AP cost | Speed −30%; attack speed −20%; dodge −10% | 59 |
| 19 | silenced | Verbal spells blocked; −20 social speech | Spellcasting disabled for verbal skills | 60 |
| 20 | diseased | Symptoms from disease effect (blight/fever/plague) | Periodic stat drain; worsens if untreated | 40, 41, 42 |
| 21 | cursed | −2 all stats; needs remove curse or ritual | −X stats; requires specialized cure | 61 |
| 22 | exposed | −10 evasion/block; DR halved; ×1.25 damage | Evasion −15%; resist −20%; incoming +X% | 62 |
| 23 | taunted | −20 attacks vs non-taunter; cannot flee | Forced aggro; −20% damage to others | 63 |
| 24 | staggered | Loses reaction; −10 next action; ends next turn | Cannot block/dodge 1–2s | 64 |

---

## 4. Application & Removal Flow

### Application
1. An effect lands on a target (attack roll, spell check, poison injection, etc.).
2. The engine or GM checks the target's **resistance** or **immunity** to the condition
   (via `resist` effect id 19 with appropriate parameter, racial traits, or active buffs).
3. If not resisted/immune, the condition is added to the target's `active_conditions` tracker
   (or its stack counter incremented for stacking conditions).
4. Duration is derived from the applying effect's `defaultDuration` field. If the effect has
   no duration, the condition uses `defaultDuration` from the applying effect entry.

### Active Tracking (Reference Implementation)

In `combat_reference.py`, conditions are tracked in a `self.active_conditions` dict keyed by
condition key. Each entry stores:

```python
{
    "key": "bleeding",
    "label": "Bleeding",
    "applied_by_effect": 58,
    "duration_type": "timed",        # timed | permanent | until_dispelled
    "remaining_duration": 15,        # seconds remaining (0 = indefinite)
    "applied_at": 0,                 # simulation_time when applied
    "source_label": "bleed weapon",  # human-readable source context
}
```

The `apply_condition()` method adds conditions; `try_dispel_condition()` removes them.
`reset_conditions()` clears the tracker between combat simulations.

### Removal
- **`cure` effect (id 20):** The primary remover. The spell or item declares which
  condition it cures via `parameter` (e.g. `"poisoned"`, `"paralyzed"`, `"all"`).
  In the reference resolver, any effect in a condition's `removedBy` array can dispel
  it — `cure` (id 20) is in every condition's `removedBy`.
- **Opposite-effect dispel:** Certain conditions can be removed by their mechanical
  opposite (frost dispels burn, healing dispels bleeding, haste dispels slow).
  This is data-driven: the dispelling effect ID is listed in the condition's `removedBy`.
- **Rest:** A full rest removes exhaustion (1 level) and most temporary conditions.
- **Special actions:** Standing removes prone. Breaking line of sight can remove
  frightened. Dealing damage to the charmer breaks charm.
- **Duration expiry (timed conditions):** When the applying effect's duration runs out,
  the condition is removed automatically. In the reference resolver, `_advance_time(seconds)`
  decrements `remaining_duration` on all timed conditions and removes expired entries.

### Immunity & Resistance
- Racial traits can grant immunity to specific conditions via the `resist` effect
  (id 19) with `parameter` matching the condition key, at 100% magnitude.
- Temporary immunity comes from spells or potions that apply `resist` with the
  condition as parameter.
- Resistance reduces the duration or effectiveness but does not prevent application.

---

## 5. Stacking Conditions

Some conditions (notably **exhaustion**) are designed to stack:

| Property | Behavior |
|----------|----------|
| `stacking: true` | Multiple applications increment the stack counter |
| `maxStacks` | Hard cap (exhaustion caps at 6 — death) |
| Stack effects | Each level has cumulative penalties described in the condition's `effects` field |
| Removal | Rest removes 1 level per full rest. `cure` with `parameter` can remove all levels or 1 level depending on implementation |

Non-stacking conditions (blinded, paralyzed, etc.) simply refresh their duration on
reapplication — they do not stack.

---

## 6. Interaction with Wuxing Phases

Conditions themselves are **not phased** — they are binary states. However, the
**effects** that apply them may carry a `phase`:

- A `fire`-phased spell that applies `blinded` (via a flash effect) interacts with
  the five-phase cycles. A target with a `water`-phased racial trait would resist
  the fire-phased effect via the overcoming cycle (water overcomes fire).
- The `resist` effect (id 19) can be parameterized by phase name (e.g. `"fire"`,
  `"wood"`) for conditional resistance to elemental conditions.

In practice: check the applying effect's `phase` against the target's race, armor
material, or active buff phases. The cycles in `data/wuxing/core.json` determine
whether the condition is amplified, weakened, or unaffected.

---

## 7. Adding New Conditions

To add a new condition:

1. Create a new entry in `conditions/core.json` with the next available `id`.
2. Assign an `appliedBy` effect — either an existing one from `effects/core.json`
   or a new one (if new, add the effect entry first).
3. Assign a `removedBy` — typically `cure` (id 20) with the appropriate `parameter`.
   Optionally add opposite-effect dispels (e.g., healing for bleeding, frost for burn).
4. Write the `effects.ttrpg` and `effects.video_game` strings.
5. Set `stacking` and `maxStacks` as appropriate.
6. Run `validate.py`.

---

## 8. Reference Implementation

The reference condition system lives in `combat_reference.py` and is exercised
via `demo_condition_interactions()` when run with `--verbose`:

```bash
python ruleset/scripts/combat_reference.py --verbose
```

### Lifecycle

```
effect lands on target
  → _process_effect_conditions(eff, source_label)
    → try_dispel_condition(key, eid, label)    # check removedBy
    → apply_condition(key, eid, dur, label)     # check appliedBy
      → adds to active_conditions dict
        → later: _advance_time(n) removes expired timed conditions
```

### Key Methods in `CombatResolver`

| Method | Function |
|--------|----------|
| `apply_condition(key, effect_id, duration, source)` | Adds to `active_conditions`. Refreshes if already active. Derives duration type and remaining time from the effect. |
| `try_dispel_condition(key, effect_id, source)` | Checks `removedBy` on the condition. If the effect ID matches and the condition is active, removes it. |
| `_advance_time(seconds)` | Decrements `remaining_duration` on all timed conditions. Removes expired entries. |
| `_process_effect_conditions(eff, label)` | Dispatch: calls dispel first, then apply for any matching condition mappings. |
| `_list_active_conditions()` | Pretty-prints all active conditions with durations and sources. |
| `reset_conditions()` | Clears the tracker and simulation clock. Called at the start of each `resolve_attack`. |

### Integration Points

Condition checks fire automatically after every effect in these processing loops:

- **Base weapon on-hit effects** (`source: "base weapon on-hit"`)
- **Crafted item on-hit effects** (`source: "crafted item on-hit"`)
- **Enchantments** (`source: "enchantment"`)
- **Coatings** (`source: "coating"`)

For each effect, the resolver first checks if it dispels any existing condition
(via `removedBy`), then checks if it applies any new condition (via `appliedBy`).
This means applying frost damage on a burning target both deals damage and
extinguishes the fire — in a single hit.

### Example Output

```
Enchantment: Burn (base mag 3)
  -> Condition [Burning] applied (duration: 10s)

Coating: Paralytic Poison (mag 4, uses 3)
  -> Condition [Paralyzed] applied (duration: timed)

  Conditions active after hit:
    [Burning] (10s)
    [Paralyzed] (0s)
```

```
  --- Phase 2: Opposite-effect dispels ---
  Healing (restore_resource) dispels bleeding:
  -> Condition [Bleeding] dispelled by restore_resource
  Haste (haste_attack) dispels slowed:
  -> Condition [Slowed] dispelled by haste_attack
```

```
  Advancing time by 6 seconds...
  Advancing time by 3 seconds...
  -> Condition [Paralyzed] expired (duration elapsed)
```

---

## Open Items

- **Cure parameter convention:** The `cure` effect (id 20) cures by `parameter`.
  The convention is for `parameter` to match the condition's `key` string
  (e.g. `cure { parameter: "poisoned" }`). This should be standardized into a
  formal rule.
- **Prone as special case:** Prone is the only condition not removed by a `cure`
  effect — it requires the "stand" action. This is intentional but worth
  documenting as an exception in any engine implementation.
- **ConditionManager class:** The active tracking logic currently lives inside
  `CombatResolver` in `combat_reference.py`. A future refactor should extract
  it into a dedicated `ConditionManager` class.
- **Full character state machine:** Conditions currently have no mechanical
  impact on combat resolution (e.g., paralyzed targets still act normally).
  Integrating condition effects into damage/action resolution is the next step.
- **Effect `parameter` filtering:** When an effect like `restore_resource` (id 7)
  or `cure` (id 20) has a `parameter`, the resolver should filter dispel by that
  parameter (e.g., cure with `parameter: "bleeding"` should only dispel bleeding).
