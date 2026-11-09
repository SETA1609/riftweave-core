# Findings — Data-oriented design and consistency

*Review date: 2026-11-01, at commit `ac55162` (branch `RW-40-story-files-like-new-world`).*

Scope: whether `ruleset/data/`, `ruleset/schemas/`, `ruleset/scripts/`, and `docs/core/`
follow the data-oriented rules the repo set for itself, and where they contradict each
other. The measure is the repo's own contract, not an outside style guide:

- [docs/core/contract.md](../core/contract.md): one modifier shape, one stacking rule, one
  resolution function, derived stats are functions and never stored, no TTRPG-versus-video
  column.
- [docs/core/keys.md](../core/keys.md): a cross-reference is a namespaced string, never a
  numeric id, and never `oneOf [string, id]`.
- [ROADMAP.md](../../ROADMAP.md#definition-of-finished) items 1, 2, and 5: no rule exists
  only as prose, `docs/core/` is the one authority, and no schema accepts a bare string where
  a key belongs.

Baseline **[V] 2026-11-01**: `python ruleset/scripts/validate.py` passes 30/30 files. All
`core:` references in `ruleset/data/` resolve (checked with a script over every string value).
Every finding below gets past that check.

Each finding has an index (`DOD-NN`) so issues and PRs can cite it. Every claim is **[V]
2026-11-01** (read or run in this review) unless tagged otherwise.

## Index

| ID | Sev | Finding | Area | Related issue |
| --- | --- | --- | --- | --- |
| [DOD-01](#dod-01-the-kernel-rows-and-vectors-have-no-consumer) | High | Kernel rows and vectors have no consumer | scripts, vectors | M1-02, M1-07, M1-08 |
| [DOD-02](#dod-02-the-ap-pool-is-defined-four-times) | High | AP pool defined four times | data | M1-01, M1-02 |
| [DOD-03](#dod-03-crit-and-fumble-are-defined-twice) | High | Crit and fumble defined twice | data | M1-01, M1-02 |
| [DOD-04](#dod-04-the-shared-modifier-definition-is-referenced-by-no-schema) | High | Shared `modifier` definition referenced by no schema | schemas | M1-05, M2-04 |
| [DOD-05](#dod-05-a-second-stacking-rule-lives-in-appliedeffectdurationtype) | High | Second stacking rule in `appliedEffect.durationType` | schemas | M1-05 |
| [DOD-06](#dod-06-ttrpg-and-video-game-columns-survive-in-conditions-and-combat) | High | TTRPG and video-game columns survive | data, schemas, docs | M1-01 |
| [DOD-07](#dod-07-conditions-hold-the-same-rule-three-times) | High | Conditions hold the same rule three times | data | M1-05 |
| [DOD-08](#dod-08-the-two-modifier-specs-disagree) | Medium | `contract.md` and `modifier.md` describe different modifiers | docs/core | M1-05 |
| [DOD-09](#dod-09-the-core-docs-disagree-on-module-overrides) | Medium | Core docs disagree on module overrides | docs/core | M5-06 |
| [DOD-10](#dod-10-monsters-and-crafted-items-store-derived-values) | Medium | Monsters and crafted items store derived values | data | M4-03, POST-01 |
| [DOD-11](#dod-11-feature-effects-are-a-private-effect-language) | Medium | Feature effects are a private effect language | data, schemas | M2-01, M2-04 |
| [DOD-12](#dod-12-materials-carry-two-copies-of-their-numbers-and-they-have-drifted) | Medium | Materials carry two copies of their numbers, and they have drifted | data | — |
| [DOD-13](#dod-13-rules-encoded-as-strings-inside-data) | Medium | Rules stored as strings inside data | data, schemas | M1-02, M1-03 |
| [DOD-14](#dod-14-formula-ast-exists-once-unvalidated-and-unused) | Medium | Formula AST exists once, has no schema, and nothing uses it | data | M1-02, M1-07 |
| [DOD-15](#dod-15-cross-references-that-are-not-namespaced-keys) | Medium | Cross-references that are not namespaced keys | data, schemas | M5-08 |
| [DOD-16](#dod-16-module-rows-are-indexed-under-core-and-module-refs-are-never-checked) | Medium | Module rows indexed under `core:`, `module:` refs never checked | scripts | M5-08 |
| [DOD-17](#dod-17-row-identity-is-not-uniform) | Medium | Row identity (`id`/`key`/`label`) is not uniform | data, schemas | M5-07 |
| [DOD-18](#dod-18-collection-wrapper-shape-is-not-uniform) | Low | Collection wrapper shape is not uniform | data | M5-07 |
| [DOD-19](#dod-19-mixed-units-and-a-d20-difficulty-in-a-d100-system) | Low | Mixed units, and a d20 difficulty in a d100 system | schemas | POST-01 |
| [DOD-20](#dod-20-property-casing-is-mixed-inside-single-schemas) | Low | Property casing is mixed inside single schemas | schemas | open question in decisions.md |
| [DOD-21](#dod-21-schemas-that-validate-almost-nothing) | Low | Schemas that validate almost nothing | schemas | M5-08 |
| [DOD-22](#dod-22-vocabulary-duplicated-instead-of-referenced) | Low | Vocabulary copied into schemas instead of `$ref` | schemas | — |
| [DOD-23](#dod-23-denormalized-labels-in-references) | Low | Labels copied into references | data | M2-04 |
| [DOD-24](#dod-24-entry-docs-still-describe-string-ids) | Low | Entry docs still describe string ids and an unvalidated net | docs | M5-06 |

Severity, by what each one breaks:
- **High:** the "one authority" or "no rule only as prose" gate fails today, or two
  implementations reading the data would legitimately disagree.
- **Medium:** the data model drifts from the contract, or real drift is already in the data.
- **Low:** convention or hygiene.

---

## High

### DOD-01 The kernel rows and vectors have no consumer

The M0 kernel moved rules into rows: `data/resolution/core.json`, `data/actions/core.json`,
`data/derived/core.json`, `data/formulas/ap.json`, and the golden vectors in
`ruleset/tests/vectors/{derived,resolution}.json`. No script reads any of them:

- `git grep 'tests/vectors'` outside docs: no hits. No test loads a vector file.
- No script under `ruleset/scripts/` opens `resolution/core.json`, `actions/core.json`, or
  `derived/core.json`.
- `combat_reference.py:87` loads the **legacy** `data/combat/base_resolution.json`, and
  `combat_reference.py:515` reads `naturalCritRange` from it.

[vectors.md](../core/vectors.md) says "Python runs it now", which is false. The vectors pin
nothing, and the one reference implementation runs on the rows the kernel replaced.

**Fix:** add a single vector runner that evaluates the kernel rows (M1-07), point
`combat_reference.py` at `resolution/core.json` and `actions/core.json`, and run the runner
in CI (M1-08).

### DOD-02 The AP pool is defined four times

| Location | Shape |
| --- | --- |
| `data/actions/core.json` `pool` | `base 2, per_point 1, per_point_divisor 3, attr agi, cap 10` |
| `data/derived/core.json` `ap` | `base 2, per_point 1, per_point_divisor 3` |
| `data/formulas/ap.json` | formula AST `2 + agi / 3` |
| `data/combat/base_resolution.json` `actionEconomy` | `apBasePool 2, apMaxPool 10`, plus attack, dodge, and block costs again |

The values agree today. Nothing forces them to keep agreeing, and `actionEconomy` also
carries `reactionsPerRound: 1`, which [actions.md](../core/actions.md) abolished. There is
also a docs split on the cap: `actions.md` puts the cap in `actions/core.json`, while
`derived.md` says "The turn cap of 10 … stays in `progression.md`".

**Fix:** keep `derived/core.json` (or the formula AST) as the only pool formula, keep costs
and the cap only in `actions/core.json`, and delete `actionEconomy`.

### DOD-03 Crit and fumble are defined twice

`resolution/core.json`: `crit_window 5`, `fumble_at 100`. `combat/base_resolution.json`
`critical`: `naturalCritRange 5`, `fumbleRange 100`, plus `critDamageMultiplier 2.0`.
[resolution.md](../core/resolution.md) says "A crit tag does not itself double damage", but
the legacy row stores a multiplier anyway, and the reference script reads the legacy row
(DOD-01). `hitQuality` margins (solid, glancing) are a second interpretation of `margin`
that `docs/core/` does not mention.

**Fix:** remove `critical` from `base_resolution.json`. Move the damage multiplier and the
hit-quality bands into a consumer row that M1-02 specifies against the crit tag.

### DOD-04 The shared `modifier` definition is referenced by no schema

`schema.json#/definitions/modifier` exists (SET-21, "One modifier schema"), but
`grep 'definitions/modifier' ruleset/schemas/*.json` returns no file. Instead:

- `condition.schema.json` declares its own inline `modifiers` items (`target`, `duration`,
  and `source` are free strings).
- `feature.schema.json` declares its own `effects` items (DOD-11).
- `material.schema.json` has an untyped `modifiers` object marked "Old combined column. Not
  the authority."
- Race `abilityModifiers` is a map, and background `skill_bonuses` is its own shape.

So "the only shared sentence" is not enforced anywhere, and the word *modifiers* means three
different shapes in three schemas.

**Fix:** `$ref` the shared definition from conditions and features, rename the material
column, and decide in M2-04 whether race and background bonuses become modifier rows.

### DOD-05 A second stacking rule lives in `appliedEffect.durationType`

The `appliedEffect.durationType` description in `schema.json` says: `unlimited` effects
stack (magnitudes add), and `temporal` effects with the same effect and parameter
"override each other, taking the highest magnitude" and usually the longest duration.

The kernel says the opposite. Stacking is keyed on **source**: the same source does not
stack and the latest application wins; different sources add
([contract.md](../core/contract.md#modifier)). Highest-magnitude-wins keyed on the effect
is exactly the second stacking rule the contract forbids. `effect.schema.json`
(`defaultDurationType`) and `feature.schema.json` (`durationType`) carry the same field.

**Fix:** delete `durationType`, or redefine it as the modifier `duration`. Express "the best
of two potions" as same-source if that is wanted. Pin the rule with an M1-05 vector.

### DOD-06 TTRPG and video-game columns survive in conditions and combat

The contract forbids "a TTRPG-versus-video column inside the kernel", and
[resolution.md](../core/resolution.md) says "There is no TTRPG column and no video-game
column". Still present:

- `condition.schema.json`: `effects.ttrpg` and `effects.video_game`, filled on all 24
  conditions in `data/conditions/core.json`.
- `combat.schema.json` / `base_resolution.json`: a `videoGame` block (stamina pool, i-frames,
  parry windows).
- [docs/modules/conditions.md](../modules/conditions.md) lines 11 to 73 still describe the
  per-mode model.

M1-01 covers `combat.md` but not the conditions schema or data.

**Fix:** extend M1-01 (or open a new issue) to cover conditions. A real-time client maps AP to
cooldowns ([actions.md](../core/actions.md)). It does not get its own column.

### DOD-07 Conditions hold the same rule three times

Each condition has prose (`effects.ttrpg`, `effects.video_game`), and 12 of 24 also have an
inline `combat` block (`defense_bonus`, `attack_penalty`, `auto_crit_when_target`, …), plus
`modifiers`. The `modifiers` are almost all `grant <flag>`. The only numeric one is
`paralyzed: set ap 0`. Example, `blinded`:

- prose: "-20 on PER-based checks … Attacks against you have +10"
- `combat.defense_bonus: 10` (the attacker's bonus, encoded as a field named *defense*)
- `modifiers`: `grant blinded`

The numbers a second implementation needs (−20 PER, +10 to the attacker) are in prose and
in an ad hoc block, not in modifier rows. The schema's own description says "Conditions do
not define their own mechanics inline", and `combat` does exactly that.

**Fix:** move every number into modifier rows on named targets, and delete `combat` and the
prose columns (keep one `description`). M1-05 then pins them.

---

## Medium

### DOD-08 The two modifier specs disagree

| | [contract.md](../core/contract.md) | [modifier.md](../core/modifier.md) | `schema.json#/definitions/modifier` |
| --- | --- | --- | --- |
| target | `attr \| skill \| resource \| flag \| damage_tag` | `skill, attribute, ap, dr` | free string |
| duration | `turns \| uses \| while_equipped \| permanent` | absent | free string; data uses `while_applied` |
| stack | absent (the rule is global) | `none \| add` | `none \| add` |
| two `set` | "replaces for that source's duration" | "last in load order wins" | last in load order wins |

A per-row `stack` field contradicts the global rule: it lets a row opt out of "different
sources add". `while_applied` appears in neither duration list.

**Fix:** make `modifier.md` the single spec (the contract links to it), give `target` a
pattern or enum per target kind, give `duration` an enum, and drop `stack` or define it.

### DOD-09 The core docs disagree on module overrides

- [contract.md](../core/contract.md#what-a-later-module-is-forbidden-to-do): a module may
  not "Override a core row. Modules add rows. Duplicate ids are conflicts, not patches."
- [overrides.md](../core/overrides.md) (SET-25, newer): a module may override a vanilla row
  by listing its key in `overrides`. Last in load order wins.
- `CLAUDE.md` / `AGENTS.md`: a module overrides simply by reusing the same `core:` key. That
  is the implicit form `overrides.md` says still fails.
- `validate.py` `build_effect_channels`: "later entries override". It applies overrides
  silently, without consulting the manifest `overrides` list.

**Fix:** amend `contract.md` to point to `overrides.md`, fix the entry docs (M5-06), and make
validation fail on an implicit duplicate.

### DOD-10 Monsters and crafted items store derived values

The contract says "Store a derived stat as authority" is forbidden: "HP is a function of END
plus modifiers."

- Monster bases and overrides store `hitPoints` (e.g. `ember_drake: 78`) and grow it with
  `hitPointsPerLevel`. That is a second HP formula next to `derived/core.json`.
- `crafted_items.json` stores pre-baked `attack_bonus`, `defense_bonus`, and `phase`, which
  copy the material (`steel_longsword.attack_bonus 5` = `steel.weapon.attack 5`). The
  contract's item is `spawn_item { base, material, mods[] }`, i.e. computed. Where the copies
  differ (`ruby_engraved_steel_longsword.phase: fire`, material `metal`), the rule that picks
  the phase (the gem wins?) is not written anywhere.

**Fix:** monsters either use the derived function (with a monster modifier) or record why
they are exempt in `docs/decisions.md`. Crafted items become `base + material + gem + mods`
with the stats computed, and the phase-precedence rule goes into a spec.

### DOD-11 Feature effects are a private effect language

`feature.schema.json` effects are `{ type: bonus | penalty | immunity, target, parameter,
value }`. Of 43 feature effect targets, 28 are effect refs (`core:effect/fortify_attribute`)
and 15 are stat names (`hp_max`, `initiative`, `barter_prices`, `brutal_charge_surcharge`,
…). So `target` means "which effect" in one row and "which stat" in the next. Stat names are
unregistered: `hp_max` is not the derived key `hp`, and `heavy_armor_surcharge` is defined
nowhere. `bonus` versus `penalty` repeats the sign of `value`.

**Fix:** feature effects become either `appliedEffect` (an effect-registry ref) or
`modifier` (a registered target). Not a third shape. Derived and action keys form the
registered target list.

### DOD-12 Materials carry two copies of their numbers, and they have drifted

`materials/core.json` keeps the legacy `modifiers {attack, defense, effectMagnitude}` and
`appliesTo` next to the newer `weapon`, `armor`, and `ingredient` blocks from SET-19.
`materials.md` says `appliesTo` "is not the authority", but both columns are filled. They
already disagree: **granite** has `modifiers: {}` but `weapon.attack 1` and
`armor.defense 2`. `material.schema.json` types `modifiers` as an untyped object, so nothing
catches this.

**Fix:** delete `modifiers` and `appliesTo` from data and schema in one PR.

### DOD-13 Rules stored as strings inside data

Logic stored as opaque strings that a second implementation has to parse or hard-code:

- `resolution/core.json`: `"success": "roll <= target and roll < fumble_at"`,
  `"margin": "target - roll"` (the schema `const`-pins them, so they are documentation, not
  data).
- `actions/core.json` `pool`: `"carry": "min(unspent, cap)"`, `"rest": "refill"`.
- `wuxing/core.json` `cycles.condition`: an English sentence ("Fires only when the actor's
  magnitude exceeds the recipient's …") is the only encoding of the insulting-cycle trigger.
- Spell → skill is a naming convention (`color` → `<color>_magic`), not a field, while the
  contract's `cast` reads `spell.skill`.

Roadmap item 1 says "No rule exists only as prose". Strings inside JSON are still prose.

**Fix:** express each of these as formula AST nodes or enum opcodes (M1-02, M3-04), or move
them to the spec and out of the row.

### DOD-14 Formula AST exists once, unvalidated and unused

`data/formulas/ap.json` is the only formula-AST instance. It has no `$schema`, so
`validate.py` prints "no $schema field, skipping". There is no `formula.schema.json`, and
nothing evaluates it (DOD-01). It also duplicates `derived.ap` (DOD-02). `derived.md` still
specifies its own coefficient function (`base + attr*per_point/divisor + level*…`), so the
repo has two formula languages: coefficients and the AST.

**Fix:** add `formula.schema.json` and pick one language. Either derived rows hold an `expr`
AST, or the coefficient form is the AST's documented sugar. Make `validate.py` fail, not
warn, on a missing `$schema`.

### DOD-15 Cross-references that are not namespaced keys

[keys.md](../core/keys.md) says every cross-reference is `core:<category>/<key>`. Remaining
exceptions:

| Field | Value example | Problem |
| --- | --- | --- |
| `materials[].tier`, `gems[].tier` | `1`, `3` | numeric id reference (tiers have no `key`) |
| `examples/loop.json` `loop.material` | `"iron"` | bare key, in the vertical-loop record |
| `equipment/bases/weapons/*.json` `valid_materials` | `["copper", "iron", …]` | bare material keys |
| `equipment[].base`, `monsters[].base` | `"dagger"`, `"beast"` | resolved by **file name**, not a key namespace |
| `materials[].foundAs`, `forms.recipes[].from/to` | `"ore"` | bare form keys |
| feature `effects.parameter` | `"animal_handling"` | bare skill key; `appliedEffect.parameter` says use `core:skill/<key>`, and one row does |
| race `traits[].id`, feature prerequisite `skills[].id` | `core:trait/…` | the field is named `id` but holds a key ref, against the `id` = numeric-own-identity rule |
| `entryId` description (`schema.json`) | "string keys (bare or namespaced)" | contradicts keys.md |

`feature.schema.json` prerequisite `skills.id/label` is still a `oneOf`, which keys.md bans.

**Fix:** give tiers and forms keys, namespace every column above, rename `id`-named refs to
`ref` (or the category name), and add a base category (`core:weapon_base/dagger`) if bases
stay separate files.

### DOD-16 Module rows are indexed under `core:` and `module:` refs are never checked

- `build_namespaced_ref_index` and `build_effect_channels` (`validate.py:65`, `:216`) index
  module data with a `core:` prefix. A module effect `frost_sword` becomes
  `core:effect/frost_sword`, so core data could cite a module row, which breaks "core never
  requires a module".
- `resolve_namespaced_ref` and `equipment_ref_valid` return `True` for any string starting
  with `module:`, with no existence check. A typo in a module ref passes.
- `materials/forms.json` declares a `recipes` collection. `COLLECTION_CATEGORIES` maps
  `recipes` → `recipe`, so the forms `smelt` and the crafting `forge_steel_longsword` share
  `core:recipe/` while following different schemas (`material_form` vs `recipe`). The keys
  do not collide today; one shared namespace for two shapes is a trap.

**Fix:** index module rows under `module:<id>/<category>/<key>` and resolve them. Rename the
forms collection (`form_steps` → `core:form_step/`).

### DOD-17 Row identity is not uniform

CLAUDE.md says every entry has a numeric `id`, a `key`, and a `label`. Actual:

| Collection | `id` | `key` | `label` |
| --- | --- | --- | --- |
| effects (77) | **none** | yes | yes |
| derived, actions, forms | none | yes | derived only |
| tiers | int | **none** (`name`) | none |
| abilities, wuxing phases and cycles | **string** | none | none |
| equipment weapons, armor, accessories; monsters | int | yes | in the base or overrides, not the row |
| everything else | int | yes | yes |

Effects are the most-referenced collection and have no `id`. If `id` is only internal
tracking (keys.md: "A numeric id is not a reference"), it is either needed everywhere or
nowhere.

**Fix:** decide in M5-07. The data-oriented answer is that `key` is the identity and `id` is
dropped, or else `id` is required by a shared row definition. Either way, `tiers` needs a
`key`.

---

## Low

### DOD-18 Collection wrapper shape is not uniform

The convention is "one wrapper object with one named array". Exceptions:
`resolution/core.json` and `combat/base_resolution.json` (a single object), `actions/core.json`
(an array plus a `pool` object), `materials/forms.json` (two arrays: `forms`, `recipes`),
`wuxing/core.json` (two arrays), `examples/loop.json` (an object), and `formulas/ap.json` (a
bare row with no wrapper). Singletons are fine, but the convention should name them, so a
generic loader can tell a table from a config.

### DOD-19 Mixed units, and a d20 difficulty in a d100 system

- Race `speed`: feet per round. Equipment `weight` and derived `carry`: pounds.
  `appliedEffect.radius`: meters. `appliedEffect.duration`: seconds. Recipe `time`: hours.
  Conditions and actions: turns and AP.
- `recipe.schema.json` `difficulty`: integer 1 to 30, "Target number", which is a d20 DC
  range in a d100 roll-under system (data: `difficulty: 14`).

**Fix:** record one unit per dimension in `docs/core/`, and make recipe difficulty a skill
modifier on the d100 target.

### DOD-20 Property casing is mixed inside single schemas

Already an open question in `docs/decisions.md`. It is worse than "older vs newer": the same
schema mixes both. `condition.schema.json` has `appliedBy`, `maxStacks` next to
`attack_penalty`, `video_game`. `equipment.schema.json` has `drBase`, `attackReach` next to
`enchantment_slots`, `on_hit_effects`. `effect.schema.json` has `magnitudeUnit` next to
`sub_effects`. The kernel schemas (`action`, `derived`, `resolution`) are all snake_case.

### DOD-21 Schemas that validate almost nothing

- `loop.schema.json`: `loop` is `{ "type": "object" }`, so any content validates.
- `equipment.schema.json#/definitions/equipmentData`, `monster.schema.json#/definitions/monsterData`
  and `coreDropShape`, and `combat.schema.json` `resolution`: no `additionalProperties: false`
  (review L3, still open), so typos in `overrides` pass.
- Equipment and monster base templates (`*/bases/*.json`) have no `$schema` and are skipped
  by the walker (review L4, still open).
- `material.schema.json` `phase` and `foundAs` are free strings, while every other `phase`
  `$ref`s `schema.json#/definitions/phase`.

### DOD-22 Vocabulary duplicated instead of referenced

- `derived.schema.json` `attr` restates the 8-ability enum instead of
  `$ref: schema.json#/definitions/ability`. `action.schema.json` `pool.attr` is a free
  string.
- `schema.json#/definitions/damageType` restates the keys of `combat/damage_types.json`.
  `validate.py` has its own loader for that file (`load_canonical_damage_types`), so the set
  lives in two places that must be edited together.
- `percentile` and `creatureCategory` are still defined and unused (review L2).

### DOD-23 Denormalized labels in references

Race `traits[]` items are `{ id: "core:trait/…", label: "Darkvision" }`. The label is a copy
of the trait's own label. They match today (checked for all 12 races), but nothing keeps them
in sync. A reference should be the key alone, with the display label read from the target.

### DOD-24 Entry docs still describe string ids

`CONTRIBUTING.md:62` says "IDs are `snake_case` matching `^[a-z_][a-z0-9_]*$` … cross-
references … are **not** schema-validated". `AGENTS.md` and `CLAUDE.md` say refs are "by
string ID only and are not validated", then list what `validate.py` does validate. Both
predate the keys model (review docs H1, still open; tracked as M5-06).

---

## Fixed since the July review

Checked against [findings-data-schemas.md](findings-data-schemas.md) **[V] 2026-11-01**:

- **H1** duplicate equipment ids: fixed. Equipment refs are namespaced keys, and
  `check_equipment_uniqueness` rejects duplicates.
- **H2** dangling crafting refs: fixed. All `core:` refs resolve.
- **H3** effect channels: enforced (`CHANNEL_REQUIREMENTS` in `validate.py`), and the
  current data passes.
- **M2** damage-type split: fixed. `damageType` is shared and aliases are rejected (but see
  DOD-22).

Still open from July: M1 traits vs features (M1-06), M3 `oneOf` (partly, DOD-15), M4 skill
base formula, M5 equipment type ↔ sub-object, M6 unused effects (M5-08), L2 to L4 (DOD-21,
DOD-22).

## Suggested order

1. DOD-01, then DOD-02 and DOD-03: make one consumer read one set of kernel rows. Without
   that, no vector proves anything.
2. DOD-04, DOD-05, DOD-07, DOD-08: one modifier shape, one stacking rule, conditions as rows.
   This is M1-05 widened.
3. DOD-06 with M1-01.
4. DOD-15, DOD-16, DOD-17 before M5-07 freezes keys. After the stability policy, each one
   costs a major version.
5. The rest as cleanup in the issues they are listed against.
