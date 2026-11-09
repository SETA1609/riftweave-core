# Findings — Data & Schemas

Scope: `ruleset/schemas/*.schema.json`, `ruleset/data/**/*.json`, `ruleset/modules/`.
All cross-references below were verified with scripts against the actual data, not
assumed. Baseline: schema validation passes 23/23; the built-in referential check in
`validate.py` covers numeric-id refs for races/features/spells/backgrounds/monsters.
These findings are what that net does not catch.

## High severity

### H1. Duplicate equipment ids 22–25 across files, with live ambiguous references

`equipment/weapons.json` (22 shortsword, 23 war_pick, 24 light_crossbow, 25 sling)
and `equipment/armor.json` (22 padded_tunic, 23 cloth_robe, 24 leather_jerkin,
25 studded_leather_jerkin) reuse the same numeric ids. `entryId` is documented as a
"stable numeric primary identifier … used for all cross-references", and the id index
in `validate.py` merges all three equipment files into one `equipment` collection —
so these are real collisions, not namespacing.

Five backgrounds in `backgrounds/core.json` reference the ambiguous ids: `scout` (22),
`former_soldier` (24), `hedge_witch` (22), `street_urchin` (22), `wilderness_hunter`
(22). It is unresolvable whether `scout` gets a shortsword or a padded tunic; context
suggests armor was intended, meaning weapons 22–25 should be renumbered.

**Recommendation:** renumber weapons 22–25 (e.g. to 45–48), fix the background refs,
and add a cross-file duplicate-id check per collection to `validate.py`.

### H2. Dangling key references in crafting data

- `crafting/crafted_items.json`: `enchanted_gold_amulet` has
  `base_key: "gold_necklace"` — no such equipment exists. In fact **zero**
  `accessory`-type equipment entries exist anywhere, so this item is unresolvable.
- `crafting/basic_recipes.json`: recipe `engrave_ruby_longsword` requires input
  `base_item: "steel_longsword"` — not an equipment key (the equipment key is
  `longsword`; `steel_longsword` is a *crafted_item* key, so the input type/key is
  wrong).

`validate.py`'s referential check does not cover key-based refs at all
(`base_key`/`material_key`/`gem_key`, recipe inputs/outputs), so these ship silently.

**Recommendation:** fix both entries; extend the integrity check to key-based
crafting refs.

### H3. Effect channel violations (channels are documented as enforced, but aren't)

- `ingredients/core.json`: `red_mushroom`, `ironvine`, `emberbloom` list effect 32
  (`fortify_attribute`), whose `channels` in `effects/core.json` do not include
  `ingredient`.
- `crafting/crafted_items.json`: enchantments use effect 57 `burn`
  (`ruby_engraved_steel_longsword`, `poisoned_obsidian_dagger`) and effect 2
  `damage_frost` (`sapphire_engraved_elven_bow`), neither of which has `enchantment`
  in `channels`.

**Recommendation:** add the missing channels to effects 2/32/57 (or swap effects),
and implement channel enforcement in the integrity check (see tooling H2).

## Medium severity

### M1. Trait/feature system fork

CLAUDE.md asserts "`features` is the single perk table … there is no separate 'trait'
type." Reality: `trait.schema.json` + `traits/core.json` (18 entries, all racial:
darkvision, catman_claws, …) exist as a separate registry; `race.schema.json`
references *traits* by id; and `features/core.json` contains **zero**
`type: "racial_trait"` entries (13 perk, 10 creation) even though the schema enum
still allows the type. Two parallel systems for the same concept. The trait schema
also has no `effects` block, so traits are pure prose with no mechanics.

**Recommendation:** either fold traits into features as `racial_trait` (per the
documented design) or update the docs/feature enum to acknowledge the separate
registry — and give traits a mechanical `effects` shape either way.

### M2. Damage-type vocabulary split: `cold`/`lightning` vs `frost`/`shock`

`equipment.schema.json`'s weapon `damage.type` enum uses `cold` and `lightning`
(D&D vocabulary), while the canonical registry `combat/damage_types.json` defines
`frost` and `shock` (armor `resistances` docs also use `"frost"`). The weapon enum
isn't `$ref`'d to anything shared, and `damage_types` entries are never referenced
by id from any other file.

**Recommendation:** hoist a `damageType` definition into `schema.json`, align the
enum with the damage_types keys, and reference it from equipment.

### M3. `oneOf: [{string}, {entryId}]` escape hatches gut validation

`background.schema.json` (`skill_bonuses.skill`, `starting_equipment.item`,
`starting_spells`, `suggested_tag_skills`) and `feature.schema.json`
(`effects.target`, `parameter`) accept either a numeric id **or an arbitrary
string**. All current background data uses numeric ids, so the string branch only
serves to let typos pass both the schema and `validate.py` (which only checks
`isinstance(x, int)` refs).

**Recommendation:** drop the free-string branches where data is already numeric, or
constrain strings to a pattern/enum.

### M4. Contradictory skill-base formula between schemas

`skill.schema.json` (`associatedAbility` description): base = `2 + attr*2 + Luck/2`.
`background.schema.json` (`skill_bonuses` description): base = `5 + associatedAbility × 2`.
One of these is wrong; both claim to be the rule.

**Recommendation:** pick one (check against the consuming engine's gameplay spec)
and fix the other.

### M5. Equipment `type` enum over-specified relative to the `oneOf`

The enum allows `accessory`, `wand`, `tool`, `adventuring_gear`, `ammunition`, but
the entry-level `oneOf` requires one of the `weapon`/`armor`/`consumable`/`accessory`
sub-objects — so `wand`/`tool`/`adventuring_gear`/`ammunition` items are
unrepresentable (none exist, which is why backgrounds can't grant rope or torches
and `street_urchin` gets 2 daggers). Nothing enforces type↔sub-object consistency
either: a `type: "weapon"` entry carrying only an `armor` block validates.

**Recommendation:** add conditional (`if type then required sub-object`) constraints
and either implement or drop the unused type values.

### M6. 25 of 77 effects (32%) are referenced nowhere

Including the whole weapon-combat cluster (`armor_pierce`, `sunder_armor`, `rend`,
`cripple`, `cleave`, `interrupt`, `life_steal`, `mana_leech`, …) and both dedicated
poison composites (`poison_weakness`, `poison_health_drain` — the consumable
`poison_weakness` in `consumables.json` uses different effect ids instead of its
namesake composite). Not broken, but a sign the registry and its consumers are
drifting apart.

**Recommendation:** wire the poison composites into the poison consumables; add data
(weapons/perks) that consume the combat effects, or tag them as reserved.

### M7. Content gaps / stub collections

- `spells/core.json`: 7 spells covering only 5 of 9 color schools (no
  orange/yellow/indigo/violet spells despite 9 `<color>_magic` skills).
- `crafting`: 3 recipes, 6 crafted items.
- No legendary-tier (tier 5) gem exists.
- `ingredients`: only 7 brewing reagents (plus 5 spirit cores).
- Accessories are schema-only (see M5).

**Recommendation:** prioritize at least one spell per color school and one recipe
per crafting category.

## Low severity

- **L1.** CLAUDE.md's id convention (`^[a-z_][a-z0-9_]*$` string ids) and its claim
  that weapons/armor are "still on the D&D AC shape" are both stale — schemas moved
  to numeric `entryId` + snake_case `key`, and armor uses `drBase`/`evasionPenalty`/
  layers with no AC fields anywhere. Doc-only, but it will steer future edits wrongly.
- **L2.** Dead definitions in `schema.json`: `percentile` and `creatureCategory` are
  never `$ref`'d (and `creatureCategory`'s "enum" lives only in its description).
- **L3.** `monster.schema.json#/definitions/monsterData` and `coreDropShape` omit
  `additionalProperties: false` (every other object sets it), so override typos pass.
  The documented `{"$extend": [...]}` merge convention for `abilities`/`tags` is
  schema-incompatible with the declared `type: array` — unused in data today, a
  latent trap.
- **L4.** Monster base templates (`data/monsters/bases/*.json`) carry no `$schema`
  and `validate.py` skips the dir; they're only exercised indirectly through
  base+override resolution. Give bases their own schema (a `growthDef` definition
  already exists that only they use).
- **L5.** Nits: `tiers/core.json` entries have no `key` (only collection without);
  `ability.schema.json` still uses string enum ids (undocumented exception);
  armor ids out of order (43/44 spliced between 28 and 29); `force_dart` deals force
  damage under red (fire) magic — reads odd next to `fire_bolt`; feature prerequisite
  `skills` items require only `min`, so an entry with neither id nor label validates.

## Strengths

- **Shared vocabulary discipline:** `schema.json` centralizes `ability`, `phase`,
  `color`, `qualityGrade`, `diceExpression`, `entryId`, `appliedEffect`, `sourceRef`,
  and every schema consistently `$ref`s it — changing the attribute set really is a
  one-enum edit.
- **The shared effect registry works:** spells, potions, poisons, coatings, oils,
  food, ingredients, monster abilities, weapon on-hit effects, and enchantments all
  flow through `appliedEffect`; effect ids 1–77 are contiguous with no dangling refs.
- **Conditions layer is clean:** all 24 conditions' `appliedBy`/`removedBy` effect
  ids resolve; the envelope model (conditions carry no inline mechanics) is applied
  consistently.
- **Wuxing matrix is correct:** all four cycles carry exactly the canonical 5 edges
  (verified against the classical relations), with sane multipliers and a documented
  backlash condition.
- **Race lineage model is coherent in data:** all 12 races validate the
  parent/subrace vs template/kin split; conditional requiredness works; all
  `parentRace` and race→trait refs resolve.
- **Monster Base + Override:** all 11 monsters resolve to existing bases; no
  `coreDrop` on undead/humanoids (the hard rule holds); growth curves live only in
  bases as designed.
- **`validate.py` goes beyond pure schema** with a real id-index referential layer —
  more than most JSON-Schema pipelines bother with.
