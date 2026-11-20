# Integrity gates

Modified: 2026-11-20

What `python ruleset/scripts/validate.py` refuses to pass. JSON Schema alone cannot check that a reference points at something, so the validator and its tests add eight gates on top of it. They close the open items of the data review (`docs/archive/review/findings-data-schemas.md` M3 and M6) and are tested in `ruleset/scripts/test_integrity.py`.

## 1. Every reference resolves

Every string in `ruleset/data/` and `ruleset/modules/` that is a namespaced reference, `core:<category>/<key>`, must name an entry that exists. The check walks every string value in every data file, so a reference in a field nobody wrote a per-collection check for is still caught. The per-collection checks stay for what the walk cannot know (which field must hold which category, effect `channels`). `module:` references are accepted without a lookup, as before. [A]

Base files count: the check also walks `monsters/bases/` and `equipment/bases/`, and `creature_type`, `weapon_base`, `armor_base`, and `accessory_base` name them. Categories (singular, as in `docs/core/keys.md`): `weapon`, `armor`, `accessory`, `consumable` (from `equipment/`), and `effect`, `spell`, `background`, `condition`, `ingredient`, `material`, `gem`, `skill`, `feature`, `race`, `monster`, `recipe`, `crafted_item`. A reference to any other category is an error. [V 2026-11-02, 161 references in the data resolve at the start of this change]

## 2. No bare string where a key is required

A schema property that holds a reference uses a typed reference from `schema.json`, such as `monsterRef` or `effectRef`, so a field that needs a monster accepts only `core:monster/<key>` or `module:<module>/monster/<key>`. `itemRef` accepts the five equipment categories. A field that really may hold several kinds keeps the general `namespacedRef`: only a recipe input's `key`. None of them is `type: string`. That includes the crafting fields that read like keys (`base_key`, `material_key`, `gem_key`, a recipe's `skill`, `crafted_item_key`, and an input's `key`): the data already writes them as `core:<category>/<key>`. A typo that is not even shaped like a reference fails the schema, and a well-shaped one fails gate 1 or the per-collection check.

Free text stays free where it is not a key: a modifier `target` (`hp_max`, `melee_damage`), an effect `parameter` where the effect names a vocabulary word (`health`, `fire`, an ability), labels, and descriptions. [A] An effect whose `parameter` is a reference says which kind in `parameter_category`, and `validate.py` then fails on any other value: `fortify_skill` takes a `skill`, `bound_weapon` a `weapon`, `summon` a `summon`.

`ruleset/scripts/test_integrity.py` lists the property names that hold references in the data today and fails if a schema defines one of them as a plain string.

## 3. Effects are used or reserved

Every entry in `effects/core.json` is referenced from some other data file or module, or carries `"reserved": true`. A reserved effect is one the registry defines for content that does not exist yet (the weapon-combat cluster, for example). The check fails in both directions: an unreferenced effect that is not reserved, and a reserved effect that something now references. So the flag cannot go stale. A reference from another effect (`sub_effects`) does not count as use, because the composite itself may be unused. [A]

## 4. Every modifier target names one row

A modifier's `target` is a key, not a `core:` reference, so gate 1 does not see it. Every modifier in `ruleset/data/` and in a module's data (an object with an `operation` and a `target`) must name one row of the kind its operation needs, as [modifier.md](modifier.md#target) lists: an attribute, a skill, a derived stat, an action, a condition, a damage type, or a row of `modifier_targets`. The gate also fails a `modifier_targets` row that no modifier names, a row whose key repeats a key of another target table, and a `pending` flag that disagrees with the golden vectors. Vectors are inputs, so their own modifiers are not checked against the table. [V 2026-11-10]

## 5. An override lists only what differs

Third normal form for composed rows ([normal-forms.md](../theory/normal-forms.md#third-normal-form-3nf-no-fact-about-another-non-key-field)). An equipment or monster override whose field, or sub-field of an object, equals what the row inherits after its base and layers fails: removing it would change nothing, so it is a copy. A derived monster's `level` is exempt, because it must state one. A subrace field equal to its parent race's fails the same way. `$extend` is never a copy, and `$replace` is one only when it replaces with the same value. Tested in `test_composition.py`. [V 2026-11-17]

## 6. A domain is written once

A field's domain is a fact too ([normal-forms.md](../theory/normal-forms.md#a-domain-written-once)). Each enum of
shared vocabulary lives once, in `ruleset/schemas/schema.json`, and every other schema names it
with a `$ref`: `derived.schema.json`'s `attribute` and `ability.schema.json`'s `id` are
`schema.json#/definitions/ability` (M5-72). Where a table holds the rows, the enum and the table
must name the same keys, so a new value cannot land in one place only: `test_integrity.py`
fails when they differ.

| Enum in `schema.json` | Checked against |
| --- | --- |
| `ability` (8) | the `id`s of `abilities/core.json` |
| `damageType` (13) | `combat/damage_types.json` |
| `phase` (5) | the `id`s of the wuxing module's `phases` (`modules/wuxing`) |
| `color` (9) | the magic skills: each color has its `<color>_magic` skill, and each `*_magic` skill a color |
| `qualityGrade` (8) | nothing: it has no table, and `tiers` is another scale (Common to Legendary) [A] |

`color`, `phase`, and `qualityGrade` stay enum-only, with no core table of their own: a color is
already a skill, a phase is the wuxing module's row, and a grade has no field beyond its name [A].
The `spellcraft` command contract repeats `color` twice; it moves to a `$ref` with the contract
diagram regenerated (APP-36) [U].

## 7. One kind of value per field

First normal form ([normal-forms.md](../theory/normal-forms.md#first-normal-form-1nf-one-value-of-one-kind-per-field)): a field holds one domain in every row. `test_integrity.py` walks every file under `ruleset/data/` (bases included) and every module's data files, reads each field by its path from the collection (`features[].prerequisite.skills[].any_of`), and fails when one path holds two JSON types across rows. A null is no type, and an integer and a float count as one, a number. The failure lists each path and the types it found. Golden vectors are inputs and are not walked. [V 2026-11-18]

A field that needs "one or many" names a list, and one is a list of one: a perk's skill requirement is `{ "any_of": [ref, ...], "min": n }`, and a prerequisite perk is `{ "perk": ref, "rank": n }` with `rank` left out for 1 (M5-70). [V 2026-11-18]

## 8. A reference is the key alone

Second normal form ([normal-forms.md](../theory/normal-forms.md#second-normal-form-2nf-no-fact-about-part-of-the-key)). An object that names a row by a `core:` reference is a child row keyed by its parent and that row, so a field beside the reference that is a fact about the named row alone is a copy, and it goes stale when the row changes. `validate.py` fails an object that holds a `core:` reference and also a field with the same name and value as a field of the referenced row (`label`, `description`, `phase`, ...); text compares without case and spaces, so `Short Sword` beside a row labelled `Shortsword` is a copy. The referenced row is read after composition: an equipment item with its base, a creature type as its base file. Fields that are about the pair, such as a skill requirement's `min`, a perk's `rank`, or an attack's own `damage`, stay. A consumer that shows the named row's label reads it from that row, or, for a module override, from the overriding row. Data files, bases, and module data are walked; golden vectors are inputs and are not. Tested in `test_integrity.py`. [V 2026-11-20]

What this removed (DE_S02_02): the `label` beside each race trait (`races[].traits[]` is `{ "id" }`), the `label` beside a natural weapon attack that only repeated the weapon's (an attack's `label` is optional and defaults to its weapon's), and a crafted item's `phase` and its engraving jewel's `phase`, which repeated the material's or the gem's (`crafting_reference.py` already reads them from there). [V 2026-11-20]

## Not covered

Cross-collection numeric consistency (a perk's skill minimum against the cap), `oneOf` branches that select a structure rather than a key (equipment and monster entries), and the free-text fields above.
