# Conditions & Status Effects

**Status:** Core (implemented as data + schema) · runtime resolution lives in the engine or GM adjudication.

Conditions are **binary state containers** that are applied and removed by effects from
the shared effect registry ([`data/effects/core.json`](../../ruleset/data/effects/core.json)).
Each condition is an envelope — it does not define its own mechanics inline; instead it
references the effect refs that apply or remove it. The actual penalties and gameplay
consequences are described in the condition entry's `effects` fields (one for TTRPG,
one for video game) and implemented by the consuming engine or GM.

Data lives in [`ruleset/data/conditions/core.json`](../../ruleset/data/conditions/core.json),
governed by `condition.schema.json`.

---

## 1. Design: Effects-Driven Conditions

```
effect (e.g. paralyze core:effect/paralyze)
  → lands on target
    → applies condition (e.g. "paralyzed")
      → condition's mechanical effects take hold (per TTRPG or video_game mode)
        → removed when the applying effect expires, or when a cure effect (core:effect/cure) fires
```

**Rules:**
- A condition entry declares which **effect(s)** apply it via `appliedBy` (array of effect refs).
- The `cure` effect (core:effect/cure) is the primary remover. Its `parameter` specifies which condition to cure.
- Conditions can also be removed by rest, taking damage (charm breaks), or special actions (stand from prone).
- A condition's mechanical effects are described in `effects.ttrpg` and `effects.video_game` — these are narrative/descriptive; the actual implementation is the engine's or GM's responsibility.
- Stacking conditions (like `exhaustion`) have `stacking: true` and `maxStacks`; each additional application increments the stack up to the max.

**Why effects-driven?**
Conditions do not exist in isolation. They are always the *result* of an effect
landing. By tying each condition to specific effect refs, the data stays consistent:
- A spell that applies `paralyze` effect (core:effect/paralyze) automatically puts the target in the `paralyzed` condition.
- A `cure` spell (core:effect/cure) with parameter `"paralyzed"` removes it.
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

Certain conditions can be removed by effects that are mechanically opposite, even if those effects are not classified as `cure`. This is data-driven — each condition's `removedBy` array in `conditions/core.json` lists the effect refs that can remove it.

| Condition | Applying Effect | Dispelling Effect | Rationale |
|-----------|---------------|-------------------|-----------|
| Burning (core:effect/essence_earth) | Burn (core:effect/burn) | Damage Frost (core:effect/damage_frost) | Water (frost) overcomes Fire (burn) |
| Bleeding (core:effect/essence_metal) | Bleed (core:effect/bleed) | Restore Resource (core:effect/restore_resource) | Healing stops bleeding |
| Slowed (core:effect/essence_water) | Slow (core:effect/slow) | Haste Attack (core:effect/haste_attack) | Haste dispels Slow |
| Poisoned (core:effect/feather) | Poison (core:effect/poison) | Restore Resource (core:effect/restore_resource) | Healing flushes poison |
| Any | Any | Cure (core:effect/cure) | Generic cure, the primary remover |

This allows conditions to interact naturally — applying a frost enchantment doesn't just deal damage, it also extinguishes any existing burn on the target.

---

## 2. Resolution Modes

### TTRPG (turn-based)
- Conditions are binary states with explicit duration tracked by the GM (in rounds or real time).
- Applied when the corresponding effect successfully lands on a target (the caster/attacker makes their skill check, the defender fails their resist check).
- Cured via rest, specific spells, or `cure` effects (core:effect/cure).
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
| `combat` | Optional object describing mechanical combat impact. See §9.1 (Combat Object Reference) for all fields, defaults, and examples. Omit for conditions with no direct combat modifiers. |

| # | Key | TTRPG Effect | Video Game Effect | appliedBy (effect keys) |
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
   (via `resist` effect core:effect/resist with appropriate parameter, racial traits, or active buffs).
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
- **`cure` effect (core:effect/cure):** The primary remover. The spell or item declares which
  condition it cures via `parameter` (e.g. `"poisoned"`, `"paralyzed"`, `"all"`).
  In the reference resolver, any effect in a condition's `removedBy` array can dispel
  it — `cure` (core:effect/cure) is in every condition's `removedBy`.
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
  (core:effect/resist) with `parameter` matching the condition key, at 100% magnitude.
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
- The `resist` effect (core:effect/resist) can be parameterized by phase name (e.g. `"fire"`,
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
3. Assign a `removedBy` — typically `cure` (core:effect/cure) with the appropriate `parameter`.
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

Condition checks fire automatically after every effect in four processing loops:

| Loop | Source | Description |
|------|--------|-------------|
| **Base weapon on-hit** | `wspec.get("on_hit_effects", [])` | Effects intrinsic to the weapon type (e.g., a flaming sword base). Currently unused in example data but available for future weapon bases. |
| **Crafted item on-hit** | `item.get("on_hit_effects", [])` | Effects added or overridden during crafting (e.g., an enhancement applied to a base weapon). Overrides/extends the base weapon's on-hit. |
| **Enchantments** | `item.get("enchantments", [])` | Permanent magical effects on a crafted item (burn, frost, resist, etc.). Each has magnitude, duration, and Wuxing interaction. |
| **Coatings** | `item.get("coatings", [])` | Temporary consumable effects applied before combat (poisons, oils). Consumed on use (decrement `uses_left`). |

**Why two on-hit sources?** The base weapon defines its intrinsic behavior (e.g., every "flaming sword" base always burns). The crafted item's `on_hit_effects` represents modifications made during crafting — the item can add new effects or override base ones without modifying the base weapon template. This separation mirrors the crafting system: a base weapon is a template, a crafted item is the specific instance.

For each effect in any of these loops, the resolver first checks if it dispels any existing condition (via `removedBy`), then checks if it applies any new condition (via `appliedBy`). This means applying frost damage on a burning target both deals damage and extinguishes the fire — in a single hit.

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

## 9. Condition Mechanical Impact (Data-Driven)

Condition combat mechanics are **fully data-driven**. Each condition in
`conditions/core.json` that has a `combat` object contributes its modifiers.
The resolver (`combat_reference.py`) reads these at startup via
`_build_combat_modifiers()` and never hardcodes a condition key or value.

### Combat Object Reference

Every condition entry may carry an optional `combat` object describing its
direct effect on combat resolution. All five fields are optional — omit the
entire object for conditions with no combat impact (e.g. `charmed`,
`deafened`, `silenced`).

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `attack_penalty` | integer | 0 | Flat penalty applied to the bearer's attack rolls while active. The resolver adds all active conditions' penalties together. Example: `prone` → `-20` (melee disadvantage). |
| `defense_bonus` | integer | 0 | Flat bonus added to **attackers'** rolls against the bearer. Represents the condition making the bearer an easier target. Example: `blinded` → `+10`. |
| `prevents_action` | boolean | false | If true, the bearer cannot take any actions, bonus actions, or reactions. Used for full-incapacitation conditions. Example: `paralyzed`, `stunned`, `unconscious`. |
| `auto_crit_when_target` | boolean | false | If true, melee attacks against the bearer automatically become critical hits (damage dice doubled). Typically paired with `prevents_action`. Example: `paralyzed` (helpless target), `unconscious`. |
| `attack_penalty_per_stack` | integer | 0 | Per-stack attack penalty for **stacking** conditions. The resolver multiplies this by the current stack count (clamped to `maxStacks`). Each stacking condition defines its own rate, allowing different severities (e.g. exhaustion at `-10`/stack, hypothetical poison at `-5`/stack). |

```json
// Example: a condition with both flat and per-stack penalties
{ "combat": { "attack_penalty": -20, "defense_bonus": 10 } }
// Example: a stacking condition with per-stack penalty only
{ "combat": { "attack_penalty_per_stack": -10 } }
```

Conditions without a `combat` field produce no modifiers. Adding or changing
combat behavior for a condition never requires touching Python code — only
`conditions/core.json` and (if adding a new field) the schema.

The following methods wire these fields into combat resolution:

### Attack Modifier

`_get_condition_attack_modifier()` sums `attack_penalty` and
`attack_penalty_per_stack` from active conditions' `combat` objects:

| Condition | Attack Penalty |
|-----------|---------------|
| `frightened` | −10 (while source in LoS) |
| `prone` | −20 (melee) |
| `restrained` | −20 |
| `staggered` | −10 (next action only) |
| `taunted` | −20 (vs non-taunter) |
| `exhaustion` | −10 per stack level (e.g., exhaustion 3/6 = −30) |

These are subtracted from the attacker's target number (`attacker_skill_value + modifiers + condition_attack_mod`).

### Defense Bonus

`_get_condition_defense_bonus()` returns the sum of `defense_bonus` values from
active conditions' `combat` objects. This bonus is added to attackers' rolls
against the subject. Conditions with `defense_bonus` include:

| Condition | Bonus to Attackers |
|-----------|-------------------|
| `blinded` | +10 |
| `prone` | +10 (ranged) |
| `restrained` | +10 |
| `stunned` | +10 |

This is logged but not yet subtracted from the defender's effective evasion — that
requires full character stat integration (deferred).

### Action Blocking

`_can_take_action()` returns `False` if any active condition has
`combat.prevents_action: true`. Currently, these conditions block all actions:
`incapacitated`, `paralyzed`, `petrified`, `stunned`, `unconscious`.

### Auto-Crit

After all effects are processed in `resolve_attack()`, `_is_auto_crit_target()`
checks whether any active condition has `combat.auto_crit_when_target: true`.
If so, the hit is **upgraded to a critical** automatically — damage dice are
doubled and the `hit_quality` is set to `"critical"`. Currently applies to
`paralyzed` and `unconscious`.

### Stacking Intensity

`_get_condition_intensity(key)` returns:
- The current stack count for stacking conditions (e.g., exhaustion at 3/6)
- `1` for non-stacking active conditions
- `0` if the condition is not active

### Immunity / Resistance (Stub)

`_check_immunity()` is a placeholder that logs but always returns `False`.
A future implementation will check:
- Active `resist` effect (core:effect/resist) with `parameter` matching the condition key
- Racial traits that grant condition immunity
- Phase-based resistance (via Wuxing cycles)

### Parameter Filtering for Dispel

When a `cure` effect (core:effect/cure) has a `parameter` in its `appliedEffect` shape, the
dispel is filtered to only remove conditions whose key matches the parameter
(e.g., `cure { parameter: "poisoned" }` only dispels `poisoned`, not all conditions).
`restore_resource` (core:effect/restore_resource) does not filter by parameter for its secondary dispel
effect (healing always stops bleeding regardless of which resource is restored).

### Duration Type (Data-Driven)

`_get_duration_type()` now uses effect tags and fields instead of hardcoded IDs:

| Condition | Determined By |
|-----------|--------------|
| `permanent` | Effect has tag `"disease"` or `defaultDurationType: "unlimited"` |
| `timed` | Effect has `duration > 0` or `defaultDuration > 0` |
| `until_dispelled` | No duration and no disease tag (fallback) |

---

## Open Items

The following items were addressed in the `feat/conditions-integration` branch:

- ✅ **Cure parameter filtering:** Implemented. `cure` (core:effect/cure) with `parameter` now
  only dispels conditions whose `key` matches the parameter (e.g.,
  `cure { parameter: "poisoned" }` only dispels `poisoned`). Use
  `parameter: "all"` to cure all removable conditions.
- ✅ **Prone special case:** Prone no longer includes `cure` (core:effect/cure) in its
  `removedBy` array. It can only be removed by the "stand" action (engine/GM
  responsibility). The schema was updated to allow empty `removedBy` arrays.
- ✅ **Duration type refactored:** `_get_duration_type()` now checks effect tags
  (`"disease"` → permanent), `defaultDurationType` (`"unlimited"` → permanent),
  and `defaultDuration` (timed), instead of hardcoded effect ID lists.
- ✅ **Condition mechanical impact:** Conditions now affect combat resolution:
  attack modifiers, defense bonuses, action blocking (incapacitating conditions),
  auto-crit on paralyzed/unconscious targets, and exhaustion stacking with
  per-level penalties. See §9 above.
- ✅ **Combat mechanics moved to data:** The `combat` object in each condition
  entry now drives all modifier lookups (`attack_penalty`, `defense_bonus`,
  `prevents_action`, `auto_crit_when_target`, `attack_penalty_per_stack`).
  The resolver builds its maps dynamically at startup and has no hardcoded
  condition-specific values.
- ✅ **Immunity/resistance stub:** `_check_immunity()` placeholder added — logs
  but always passes. Ready for wiring into the `resist` effect system.
- ✅ **Exhaustion stacking demo:** Enhanced to show progressive stack accumulation
  (1→5/6) with per-level attack penalty display.
- ✅ **Condition cross-reference validation:** `validate.py` now validates
  `appliedBy` and `removedBy` IDs against the effects registry, checks stacking
  consistency (stacking + maxStacks), detects duplicate effect refs, and
  enforces the prone-no-cure rule.
- ✅ **Documentation updated:** Both on-hit paths (base weapon vs. crafted item)
  are clearly documented with their purpose and relationship.

### Remaining for future work:

- **ConditionManager class extraction:** A dedicated `ConditionManager` class
  (owning `_build_condition_mappings`, reset/apply/dispel/advance, modifier
  maps, and all `_get_condition_*` / `_can_take_action` / `_is_auto_crit_target`
  methods) is noted via a `# TODO` in the source. Extraction is ~50 lines of
  boilerplate and deferred as it doesn't change behavior.
- **Full character state machine:** The reference resolver tracks conditions
  on a single subject per attack. A full combat loop would need per-character
  condition trackers, so that the attacker's own conditions (e.g., the
  attacker being blinded) affect the roll, and the defender's conditions
  (e.g., paralysis) enable auto-crits.
- **Immunity/resist wiring:** The `_check_immunity()` stub needs to be wired
  into the `resist` effect (core:effect/resist) and racial trait system. This requires
  an active effects tracker that the resolver doesn't yet have.
- **Phase-based condition resistance:** Applying a fire-phased effect to a
  water-phased target should be weakened per the overcoming cycle. This
  requires phase logic in the immunity/resist check.
- **Rest action:** A full rest (removing 1 exhaustion level, curing most
  temporary conditions) needs an explicit method.
