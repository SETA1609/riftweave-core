# Integrity gates

What `python ruleset/scripts/validate.py` refuses to pass. JSON Schema alone cannot check that a reference points at something, so the validator adds four gates on top of it. They close the open items of the data review (`docs/archive/review/findings-data-schemas.md` M3 and M6) and are tested in `ruleset/scripts/test_integrity.py`.

## 1. Every reference resolves

Every string in `ruleset/data/` and `ruleset/modules/` that is a namespaced reference, `core:<category>/<key>`, must name an entry that exists. The check walks every string value in every data file, so a reference in a field nobody wrote a per-collection check for is still caught. The per-collection checks stay for what the walk cannot know (which field must hold which category, effect `channels`). `module:` references are accepted without a lookup, as before. [A]

Categories (singular, as in `docs/core/keys.md`): `weapon`, `armor`, `accessory`, `consumable` (from `equipment/`), and `effect`, `spell`, `background`, `condition`, `ingredient`, `material`, `gem`, `skill`, `feature`, `race`, `monster`, `recipe`, `crafted_item`. A reference to any other category is an error. [V 2026-11-02, 161 references in the data resolve at the start of this change]

## 2. No bare string where a key is required

A schema property that holds a reference uses a typed reference from `schema.json`, such as `monsterRef` or `effectRef`, so a field that needs a monster accepts only `core:monster/<key>` or `module:<module>/monster/<key>`. `itemRef` accepts the five equipment categories. A field that really may hold several kinds keeps the general `namespacedRef`: only a recipe input's `key`. None of them is `type: string`. That includes the crafting fields that read like keys (`base_key`, `material_key`, `gem_key`, a recipe's `skill`, `crafted_item_key`, and an input's `key`): the data already writes them as `core:<category>/<key>`. A typo that is not even shaped like a reference fails the schema, and a well-shaped one fails gate 1 or the per-collection check.

Free text stays free where it is not a key: a modifier `target` (`hp_max`, `melee_damage`), an effect `parameter` where the effect names a vocabulary word (`health`, `fire`, an ability), labels, and descriptions. [A] An effect whose `parameter` is a reference says which kind in `parameter_category`, and `validate.py` then fails on any other value: `fortify_skill` takes a `skill`, `bound_weapon` a `weapon`, `summon` a `summon`.

`ruleset/scripts/test_integrity.py` lists the property names that hold references in the data today and fails if a schema defines one of them as a plain string.

## 3. Effects are used or reserved

Every entry in `effects/core.json` is referenced from some other data file or module, or carries `"reserved": true`. A reserved effect is one the registry defines for content that does not exist yet (the weapon-combat cluster, for example). The check fails in both directions: an unreferenced effect that is not reserved, and a reserved effect that something now references. So the flag cannot go stale. A reference from another effect (`sub_effects`) does not count as use, because the composite itself may be unused. [A]

## 4. Every modifier target names one row

A modifier's `target` is a key, not a `core:` reference, so gate 1 does not see it. Every modifier in `ruleset/data/` and in a module's data (an object with an `op` and a `target`) must name one row of the kind its op needs, as [modifier.md](modifier.md#target) lists: an attribute, a skill, a derived stat, an action, a condition, a damage type, or a row of `modifier_targets`. The gate also fails a `modifier_targets` row that no modifier names, a row whose key repeats a key of another target table, and a `pending` flag that disagrees with the golden vectors. Vectors are inputs, so their own modifiers are not checked against the table. [V 2026-11-10]

## Not covered

Cross-collection numeric consistency (a perk's skill minimum against the cap), `oneOf` branches that select a structure rather than a key (equipment and monster entries), and the free-text fields above.
