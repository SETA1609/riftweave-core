Authority is the `modifiers` bundle in `ruleset/data/conditions/core.json`. A condition is a named bundle of contract modifiers plus a flag. `paralyzed` is `set ap 0` plus a flag. A point on the d100 is the unit. Write `-20`, not `-20%`. The percent sign is a second scale, and this ruleset does not have one.

# Conditions & Status Effects

**Status:** Core (implemented as data + schema) · runtime resolution lives in the engine or GM adjudication.

Conditions are **binary state containers** that are applied and removed by effects from
the shared effect registry ([`data/effects/core.json`](../../ruleset/data/effects/core.json)).
Each condition is an envelope — it does not define its own mechanics inline; instead it
references the effect refs that apply or remove it. Its penalties are its `modifiers` and its
`combat` block, one rule for a table and a real-time client alike: a table counts the duration
in turns, `ceil(seconds / 5)`, and a client in seconds ([actions.md](../core/actions.md#one-rule-set-two-clocks)).
There is no per-mode text (M3-11).

Data lives in [`ruleset/data/conditions/core.json`](../../ruleset/data/conditions/core.json),
governed by `condition.schema.json`.

---

## 1. Design: Effects-Driven Conditions

```
effect (e.g. paralyze core:effect/paralyze)
  → lands on target
    → applies condition (e.g. "paralyzed")
      → its modifiers and combat block take hold
        → removed when the applying effect expires, or when a cure effect (core:effect/cure) fires
```

**Rules:**
- A condition entry declares which **effect(s)** apply it via `appliedBy` (array of effect refs).
- The `cure` effect (core:effect/cure) is the primary remover. Its `parameter` specifies which condition to cure.
- Conditions can also be removed by rest, taking damage (charm breaks), or special actions (stand from prone).
- A condition's rules are its `description` in words and its `modifiers` and `combat` block in numbers; a rule with no modifier target yet is described only, and the GM or engine applies it.
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

## 2. One rule set

A table and a real-time client run the same condition: the same modifiers, the same `combat`
block, the same duration in seconds. A table rounds the duration up to whole turns; a client
counts seconds and shows the state as an icon. A table's GM still rules on edge cases the data
does not cover, such as whether a blinded character hears an invisible foe. [V 2026-11-12]

---

## 3. Complete Condition Catalog

Data in `conditions/core.json` (25 entries). Each entry:

| Field | Meaning |
|-------|---------|
| `key` | Stable snake_case identifier |
| `label` | Human-readable name |
| `description` | The rule in words, the same for every client |
| `modifiers` | The rule in numbers: a flag named after the condition, and any modifiers |
| `appliedBy` | Effect refs that can apply this condition |
| `removedBy` | Effect refs that can remove this condition |
| `stacking` | Whether multiple instances stack |
| `maxStacks` | Maximum stack count |
| `combat` | Optional object describing mechanical combat impact. See §9.1 (Combat Object Reference) for all fields, defaults, and examples. Omit for conditions with no direct combat modifiers. |

| # | Key | Rule | Modifiers beside its flag |
|---|-----|------|-----------|
| 1 | blinded | Cannot see: checks that need sight fail, −20 on PER checks, and attacks against it have +10. | `add perception_checks -20` |
| 2 | charmed | Cannot attack the charmer, and the charmer has +20 on social checks against it. Damage from the charmer ends it. | `add social_defense -20` |
| 3 | deafened | Cannot hear: checks that need hearing fail, and −10 on initiative. | `add initiative -10` |
| 4 | frightened | −10 on attacks and checks while the source is in sight, and cannot move closer to it. Losing sight of it ends it. | `add checks -10` |
| 5 | grappled | Held: speed 0, −10 on attacks and AGI checks. An athletics check against the grappler breaks free. | `set speed 0`, `add agility_checks -10` |
| 6 | incapacitated | Cannot spend AP, react, or concentrate. | `set ap 0` |
| 7 | invisible | Cannot be seen: attacks against it have −20 and its own attacks +10. Attacking or taking damage ends it. | — |
| 8 | paralyzed | Cannot move or act: STR and AGI checks fail, and a melee hit from within 1 unit is a crit. | — |
| 9 | petrified | Turned to stone: cannot act, resists 20 of every damage type, is immune to poison and disease, and beneficial effects do not land. | `set ap 0`, `resist bludgeoning 20`, `resist piercing 20`, `resist slashing 20`, `resist fire 20`, `resist frost 20`, `resist shock 20`, `resist acid 20`, `resist poison 20`, `resist necrotic 20`, `resist radiant 20`, `resist psychic 20`, `resist force 20`, `resist thunder 20`, `add poison_resistance 100`, `add disease_resistance 100` |
| 10 | poisoned | The poison's sub_effects are its symptoms; a poison resistance lessens them. | — |
| 11 | prone | On the ground: −20 on melee attacks, and ranged attacks against it have +10. Standing costs the stand_up action. | — |
| 12 | restrained | Bound: speed 0, −20 on attacks and AGI checks, and attacks against it have +10. A STR check or a teleport frees it. | `set speed 0`, `add agility_checks -20` |
| 13 | stunned | Cannot act, −20 on every check, and attacks against it have +10. | `set ap 0`, `add checks -20` |
| 14 | unconscious | Out cold: cannot act, lies prone, STR and AGI checks fail, and a melee hit from within 1 unit is a crit. Healing above 0 HP or time wakes it. | `set ap 0` |
| 15 | exhaustion | Stacks to 6: −10 on checks for each level; at 3 speed is halved, at 5 speed is 0, at 6 the actor dies. A full rest removes one level. | `add checks -10` |
| 16 | burning | On fire: the burn effect deals its fire damage each second until it ends. Water, frost, or a cure puts it out. | — |
| 17 | bleeding | The bleed effect deals its damage each second until a bandage, healing, or a Medicine check stops it. | — |
| 18 | slowed | Speed halved, −10 on AGI checks, and every action costs 1 AP more (5 stamina in real time). | `multiply speed 0.5`, `add agility_checks -10`, `add action_cost 1` |
| 19 | silenced | Cannot cast spells, and −20 on social checks that need speech. | `add persuasion -20`, `add deception -20`, `add intimidation -20` |
| 20 | diseased | The disease effect's symptoms apply; END resists them, and a cure with parameter disease ends it. | — |
| 21 | cursed | −2 on every attribute; a rest does not end it. | `add str -2`, `add per -2`, `add end -2`, `add int -2`, `add wil -2`, `add agi -2`, `add cha -2`, `add lck -2` |
| 22 | exposed | Defenses down: −10 evasion, armor DR halved, and damage taken ×1.25. | `add evasion -10`, `multiply armor_dr 0.5`, `multiply incoming_damage 1.25` |
| 23 | taunted | −20 on attacks against anyone but the taunter, and cannot willingly move away from it. | — |
| 24 | staggered | Cannot react until its next turn, and −10 on the next action. | — |

`overdosed` is the 25th, applied by a rule rather than an effect ([modifier.md](../core/modifier.md#toxicity)).
The new targets (`checks`, `perception_checks`, `agility_checks`, `social_defense`,
`incoming_damage`, `armor_dr`, `action_cost`) are `modifier_targets` rows marked pending until a
rule reads them. [V 2026-11-12]

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

A runtime tracks conditions in a table keyed by condition key. Each entry stores:

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
| Stack effects | Each level has cumulative penalties described in the condition's `description` |
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
material, or active buff phases. The cycles in `modules/wuxing/data/wuxing.json` determine
whether the condition is amplified, weakened, or unaffected.

---

## 7. Adding New Conditions

To add a new condition:

1. Create a new entry in `conditions/core.json` with the next available `id`.
2. Assign an `appliedBy` effect — either an existing one from `effects/core.json`
   or a new one (if new, add the effect entry first).
3. Assign a `removedBy` — typically `cure` (core:effect/cure) with the appropriate `parameter`.
   Optionally add opposite-effect dispels (e.g., healing for bleeding, frost for burn).
4. Write the rule once in `description`, and its numbers as `modifiers` (add a `modifier_targets` row if no table holds the target).
5. Set `stacking` and `maxStacks` as appropriate.
6. Run `validate.py`.

---

## 8. Reference Implementation

The reference is the `conditions` function in `ruleset/scripts/kernel_reference.py`, checked by the
vectors in `ruleset/tests/vectors/conditions.json`. The lifecycle below is the shape a runtime follows.

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
A resolver reads these and never hardcodes a condition key or value.

### Combat Object Reference

Every condition entry may carry an optional `combat` object describing its
direct effect on combat resolution. All five fields are optional — omit the
entire object for conditions with no combat impact (e.g. `charmed`,
`deafened`, `silenced`).

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `attack_penalty` | integer | 0 | Flat penalty applied to the bearer's attack rolls while active. The resolver adds all active conditions' penalties together. Example: `prone` → `-20` (melee disadvantage). |
| `defense_bonus` | integer | 0 | Flat bonus added to **attackers'** rolls against the bearer. Represents the condition making the bearer an easier target. Example: `blinded` → `+10`. |
| `prevents_action` | boolean | false | If true, the bearer cannot spend AP or take a reaction. Used for full-incapacitation conditions. Example: `paralyzed`, `stunned`, `unconscious`. |
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
| `timed` | Effect has `duration > 1` or `defaultDuration > 1` (1 is instant) |
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
