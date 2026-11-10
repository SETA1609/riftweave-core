# M5 Release 1.0

Roadmap: [M5](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M5 Release 1.0

**Goal:** the SRD is documented, gated, legally clean, and tagged.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** items 3, 5, 6, 7, and 8 of the [definition of finished](../../ROADMAP.md#definition-of-finished) hold, and `v1.0.0` is tagged.

Done: RW-1, RW-2, RW-39, RW-40, RW-41, M5-01, M5-02, M5-03, M5-07, M5-08, M5-10, M5-11, M5-12, M5-13, M5-14, M5-15, M5-16, M5-17, M5-21 ([done.md](done.md#m5-release-10))

#### M5-04 Reference playthrough vector chain
- Status: todo
- Priority: high
- Needs: M1-02, M1-03, M1-04, M1-05, M3-01, M4-01, M4-02
- Was: RW-25
- Do: Extend the vertical loop into a full chain: create, buy, fight, loot, rest, level, cast, take a condition, recover. It is the executable proof that the core is a game.
- Files: ~4 (ruleset/data/examples/, ruleset/tests/vectors/playthrough.json, docs/core/loop.md)
- See: docs/core/loop.md, ruleset/data/examples/loop.json
- Done when: one chain uses only core rows; every step is a vector; the chain passes under the Python reference.

#### M5-05 Player and GM guide
- Status: todo
- Priority: medium
- Needs: M5-04
- Was: RW-26
- Do: There is no 'how to play' or 'how to run' text. An SRD needs a quick start for a table.
- Files: ~3 (docs/guide/)
- See: docs/archive/issues.md section 6
- Done when: quick-start of 10 pages or fewer: make a character, take a turn, run an encounter; GM section: pacing, difficulty, adjudicating unlisted actions with the one resolution function.

#### M5-06 Entry docs and review archive sync
- Status: review
- Priority: medium
- Needs: M5-02, M5-03
- Was: RW-27
- Do: The July review found README, AGENTS, CLAUDE, and CONTRIBUTING lagging the repo. States which of `docs/core/` and `docs/modules/` wins on conflict, and retires `plan.md`, `issues.md`, and `docs/review/` as history. Adds `docs/onboarding/`, with the AI rules: AI is welcome as a tool, and a human always reviews before committing. `AGENTS.md` is the only agent rules file.
- Files: ~12 (README.md, AGENTS.md, CONTRIBUTING.md, docs/archive/, docs/onboarding/, plan.md, issues.md, docs/review/)
- See: docs/archive/review/findings-docs-hygiene.md
- Done when: README tree, AGENTS, and CONTRIBUTING match the repo; authority order is written down once; historical files are moved under `docs/archive/` with a note; `docs/onboarding/` takes a new contributor from a clone to a pull request and states the AI rules.

#### M5-09 Licensing and attribution audit, tag v1.0.0
- Status: todo
- Priority: high
- Needs: M2-01, M2-02, M2-03, M2-04, M2-05, M3-02, M3-03, M3-04, M4-03, M4-04, M4-05, M5-01, M5-05, M5-06, M5-07, M5-08, M5-19, M5-20
- Was: RW-30
- Do: Final pass: licence headers, `source` fields, NOTICE and CONTRIBUTORS, and the changelog. Then tag.
- Files: ~6 (LICENSING.md, NOTICE, CONTRIBUTORS.md, CHANGELOG.md, ROADMAP.md)
- See: LICENSING.md, ROADMAP.md
- Done when: all checklist items in `ROADMAP.md` 'Definition of finished' are ticked; `v1.0.0` tagged and `[Unreleased]` in the changelog becomes `[1.0.0]`.

#### M5-18 Alpha playtest through the human CLI
- Status: todo
- Priority: high
- Needs: M2-01, M2-06, M3-02, M3-06, M4-03, M5-04, APP-04, APP-05, APP-08
- Do: The alpha, before the 1.0 tag: the content floors are met and the playthrough chain passes, so play the rules as a person would, with dev mode (APP-08) to reach the states a run is slow to reach. Run at least five characters through the human CLI from creation to level 30, across builds (melee, ranged, each magic color, a hybrid), each run a saved session. Log what a person had to decide, what never came up, and where a rule was missing or read two ways. File each finding as an issue or a decision; turn a run that exposes a wrong result into a vector.
- Files: ~4 (docs/decisions/M5-18.md, docs/playtests/, ruleset/tests/vectors/playthrough.json)
- See: docs/stories/07-backend.md (APP-04, APP-05), docs/core/loop.md
- Done when: five saved sessions reach level 30 and replay to the same result; every finding is an issue, a decision, or a vector; the decision record lists, for every perk, how often the runs took it and what it changed.

#### M5-19 Perk throughput veto and distillation
- Status: todo
- Priority: high
- Needs: M5-18
- Do: Use the alpha's perk tallies to judge every perk. Veto perks that inflate throughput (damage, actions, or recovery a level band cannot answer) or that no run took; distil perks that overlap into one; rename perks whose label or key does not say what they do; remove the vetoed ones. Each perk gets one verdict: keep, merge, rename, or cut. A rename or removal is a breaking key change and is marked so in the changelog; references in examples, vectors, and docs move with it.
- Files: ~8 (ruleset/data/features/core.json, docs/modules/progression.md, docs/decisions/M5-19.md, ruleset/data/examples/, ruleset/tests/vectors/, ruleset/scripts/test_features.py)
- See: docs/decisions/M5-18.md, docs/stories/done.md (M2-01)
- Done when: every perk has a verdict and a reason in the decision record; no cut or renamed key is still referenced; the perk floor (60) still holds or the decision record moves it; vectors pass.

#### M5-20 Wuxing becomes a module, the reference for cross-cutting modules
- Status: todo
- Priority: high
- Needs: M3-02, M4-03
- Do: The contract already names Wuxing a module and says a phase is a tag a module may read, but core still requires `phase` on monsters, monster templates, playable races, gems, crafted items, and engraving jewels, and ships the cycles in `data/wuxing/core.json`. Split it. Core keeps `phase` as an optional tag (the vocabulary in `schema.json`), so rows keep their phases and nothing changes with the module off. `ruleset/modules/wuxing/` takes the cycles and multipliers, the interaction rules (`magic.md` §6, `docs/modules/wuxing.md`), `wuxing_reference.py` and its vectors, and the bestiary's phase-coverage test; its effect on an attack or a cast is a modifier through the contract, never a call into combat or casting. The Go backend learns to load modules (it loads none today), and `internal/wuxing` reads the module's data. Moving `data/wuxing/` out of core is a breaking change, so it lands before the 1.0 tag. Write `docs/modules/module-system.md` up as the guide a module author follows, with Wuxing as the worked example of a module that adds a collection, reads tags on core rows, and emits modifiers.
- Keys: `ruleset/modules/wuxing/` with its manifest; `phase` optional in the six core schemas
- Files: ~16 (ruleset/schemas/, ruleset/data/wuxing/, ruleset/modules/wuxing/, ruleset/scripts/wuxing_reference.py, ruleset/scripts/test_wuxing.py, ruleset/scripts/test_bestiary.py, ruleset/scripts/crafting_reference.py, ruleset/tests/vectors/wuxing.json, backend/internal/data/data.go, backend/internal/wuxing/, docs/modules/wuxing.md, docs/modules/magic.md, docs/modules/module-system.md, docs/core/contract.md)
- See: docs/core/contract.md, docs/core/materials.md, docs/modules/module-system.md, ruleset/modules/factions/
- Done when: validate, the unit tests, the vectors, and the Go tests pass with the module and without it; without it no core rule reads a phase; with it the wuxing vectors pass under Python and Go from the module's data; the change is marked breaking in the changelog; `module-system.md` walks through Wuxing step by step.

#### M5-22 Move fetch-docs to devops-tools
- Status: review
- Priority: low
- Needs: M5-03
- Do: devops-tools ships `devops-tools fetch-docs` (url, github, list), the same downloader as `tools/fetch-docs/`, into the gitignored docs cache, pinned to a commit. Set `docs-cache-directory = "cache/docs"` in `[tool.devops-tools]`, remove `tools/fetch-docs/` and its tests, and point AGENTS, CLAUDE, and CONTRIBUTING at the new command. Split from M5-03.
- Keys: `docs-cache-directory` in `[tool.devops-tools]`
- Files: ~6 (tools/fetch-docs/, tools/devops/pyproject.toml, AGENTS.md, CONTRIBUTING.md)
- See: https://github.com/SETA1609/devops_tools (fetch-docs)
- Done when: a cited source downloads into `cache/docs/` through `devops-tools fetch-docs`; no local fetch-docs code remains; the docs name the new command.

#### M5-23 Drop the old material columns
- Status: review
- Priority: medium
- Needs: —
- Do: Materials still carry two columns from before SET-19 that the schema calls "not the authority": `appliesTo` and `modifiers`. Nothing reads either. `modifiers` repeats the `weapon`, `armor`, and `ingredient` blocks exactly in all 11 rows that have it, and `default_materials` on each base now says what an item is found and sold in, while a recipe does not refuse a material. Granite, the newest row, has neither. Remove both from `materials/core.json` and from `material.schema.json`. Module docs still treat them as live and contradict `docs/core/materials.md`: `armor.md` says a material without `"armor"` in `appliesTo` cannot be used, and `armor.md`, `weapons.md`, `materials.md`, `magic-crafting.md`, and the `crafted_item` schema descriptions read `material.modifiers.*`. Point them at `material.armor.defense`, `material.weapon.attack`, and `material.ingredient.effectMagnitude`. Removing a field breaks consumers, so mark it **(breaking)** in the changelog.
- Keys: removes `appliesTo` and `modifiers` from materials
- Files: ~9 (ruleset/data/materials/core.json, ruleset/schemas/material.schema.json, ruleset/schemas/crafted_item.schema.json, docs/core/materials.md, docs/modules/materials.md, docs/modules/armor.md, docs/modules/weapons.md, docs/modules/crafting.md, docs/modules/magic-crafting.md)
- See: docs/core/materials.md, docs/archive/review/findings-data-oriented-design.md (it recommended this)
- Done when: no row or schema has `appliesTo` or `modifiers` on a material; the `weapon`, `armor`, and `ingredient` values are unchanged; no doc outside `docs/archive/` and the frozen logs names either column as a material field; `validate.py` and the tests pass; the changelog fragment marks the removal **(breaking)**.

#### M5-28 Monsters and players share one actor shape
- Status: review
- Priority: high
- Needs: —
- Do: A monster stores hand-written HP and no mana or skills, while a player's are computed, so the same idea has two shapes. A monster row becomes an actor's starting sheet: derived stats use the player function, `hitPoints` is an optional authored value that replaces computed HP, and a reduced `skills` sheet lists what it fights or casts with, with other skills reading a player's starting value. Existing rows keep their HP for now.
- Keys: the `monster_sheet` command; `skills` on monsters; `skill_start` in progression
- Files: ~12 (docs/core/actor-sheet.md, docs/core/vectors.md, ruleset/schemas/monster.schema.json, ruleset/schemas/progression.schema.json, ruleset/data/progression/core.json, ruleset/scripts/kernel_reference.py, ruleset/tests/vectors/monster_sheet.json, backend/contract/monster_sheet.json, backend/core/core.go, backend/internal/data/data.go, backend/internal/derived/derived.go)
- See: docs/core/contract.md, docs/core/derived.md, docs/modules/monsters.md
- Done when: vectors show a monster and a player with the same attributes and level getting the same stats, authored HP replacing only HP, and written and unwritten skills; Python and Go pass them.

#### M5-27 One way to compose a row
- Status: review
- Priority: high
- Needs: —
- Do: Monsters, equipment, and soon bound weapons and summons each build a row from a base, layers, and overrides, each in its own code. Define one composition (`compose(base, layers, overrides)`) with one set of merge rules, and use it for monsters and equipment without changing any resolved row, so M3-09 and M3-10 add layers and no new merge code.
- Keys: `compose_row` in `validate.py`
- Files: ~5 (docs/core/composition.md, docs/modules/monsters.md, ruleset/scripts/validate.py, ruleset/scripts/test_composition.py)
- See: docs/modules/monsters.md#merge-rules, docs/core/overrides.md
- Done when: monsters and equipment both use `compose_row`; every resolved row is unchanged; tests cover order, attribute merge, `$extend`, and identity.

#### M5-26 References are typed by category
- Status: review
- Priority: high
- Needs: —
- Do: Every reference field accepts any `core:` row, so a field that needs a monster would take a weapon, and an effect's free `parameter` let summons name a plain word. Give `schema.json` one typed reference per category and use it on every field. Add `parameter_category` on an effect, for the effects whose `parameter` is a reference, and fail validation when a use names anything else.
- Keys: `monsterRef`, `effectRef`, and the rest in `schema.json`; `itemRef`; `parameter_category` on effects
- Files: ~18 (ruleset/schemas/, ruleset/data/effects/core.json, ruleset/data/equipment/accessories.json, ruleset/scripts/validate.py, ruleset/scripts/test_integrity.py, docs/core/integrity.md)
- See: docs/core/integrity.md, docs/core/keys.md
- Done when: a schema rejects a reference of the wrong category; only fields that really hold several kinds keep the general reference; an effect with `parameter_category` refuses another kind; `validate.py` and the tests pass.

#### M5-29 The key is the only row identity
- Status: review
- Priority: medium
- Needs: M5-27
- Do: Rows carry both a numeric `id` and a `key`, yet references reject numeric ids and 14 of the 35 keyed tables have none. Make the key the identity: the module loader detects conflicts by key, nothing requires `id`, and the first half of the data drops it (monsters, spells, features, skills, equipment). M5-30 finishes the data and removes the field. Built on M5-27, which changes the same merge.
- Keys: removes numeric `id` from rows **(breaking)**
- Files: ~16 (docs/core/keys.md, docs/core/versioning.md, AGENTS.md, docs/modules/module-system.md, ruleset/schemas/, ruleset/scripts/module_loader.py, ruleset/scripts/validate.py, ruleset/data/)
- See: docs/core/keys.md, docs/archive/review/findings-data-oriented-design.md (DOD-17)
- Done when: module conflicts work with no ids; no schema requires `id`; the listed data has none and is otherwise unchanged; `validate.py` and the tests pass.

#### M5-30 Remove numeric row ids
- Status: review
- Priority: medium
- Needs: M5-29
- Do: Finish what M5-29 started. Drop the numeric `id` from the remaining data files and from modules, then remove the field from the schemas, so a new row cannot add one.
- Keys: removes numeric `id` from rows **(breaking)**
- Files: ~16 (ruleset/data/, ruleset/modules/, ruleset/schemas/, AGENTS.md, docs/core/keys.md)
- See: docs/core/keys.md
- Done when: no row in `ruleset/data/` or `ruleset/modules/` has a numeric `id`; a schema rejects one; `validate.py` and the tests pass.

#### M5-25 Game state is a snapshot and a log
- Status: review
- Priority: high
- Needs: —
- Do: The ruleset is static, but a running game changes on every command, and nothing says what that state's model is. Before APP-04 builds sessions, define it: a session is the ruleset version, a starting actor, an ordered log of commands with their inputs, and the current actor as a cached result. Rolls are already inputs, so a log replays exactly, which gives save files, replays, reproducible bug reports, and vectors from one mechanism.
- Keys: none; the session schema is APP-04's
- Files: ~3 (docs/core/state.md, docs/stories/07-backend.md, docs/stories/05-release.md)
- See: docs/core/contract.md, docs/core/resolution.md, docs/stories/07-backend.md (APP-04)
- Done when: `docs/core/state.md` defines the session and its rules, and APP-04 builds on it.

#### M5-24 Define the data model
- Status: review
- Priority: medium
- Needs: —
- Do: The repo calls itself data-oriented, but no document says what that means; the pieces are spread across `contract.md`, `keys.md`, and the backend's `data.go`. Write one page that states what a data model is, the common types, and which type ours is: relational in its logical design (tables, keys, references, one fact in one place), stored as JSON documents with some nested lists (a spell's effects), and loaded by the backend into dense in-memory tables. Say what it is not: not ECS today, and not an inheritance tree (base + template + overrides is composition). List its parts (rows, schemas, references, shared pools, composition, stored versus computed, the one modifier shape, formulas as data, vectors), why we chose it, what it costs us, and the known exceptions (authored monster HP). Link to the documents that own each part instead of repeating their rules. A second page, `docs/theory/data-models.md`, holds what a data model is, the types, and why ours is shaped as it is.
- Keys: none
- Files: ~5 (docs/core/data-model.md, docs/theory/data-models.md, docs/onboarding/README.md, AGENTS.md, README.md)
- See: docs/core/contract.md, docs/core/keys.md, docs/core/integrity.md, backend/internal/data/data.go, docs/archive/review/findings-data-oriented-design.md
- Done when: `docs/core/data-model.md` and `docs/theory/data-models.md` cover all of the above, every claim about the repo is tagged **[V]** with a date or **[A]**, and the onboarding reading list links it.

#### M5-31 File the data model review
- Status: review
- Priority: medium
- Needs: —
- Do: A review of the data against `docs/core/data-model.md` on 2026-11-10 found places where the data does not follow its own model, all of which pass `validate.py`. File each one as an issue (M5-32 to M5-41) so it can be ordered and worked, and record the review in a decision file.
- Keys: none
- Files: ~4 (docs/stories/05-release.md, docs/stories/07-backend.md, docs/decisions/M5-31.md, ROADMAP.md)
- See: docs/core/data-model.md, docs/archive/review/findings-data-oriented-design.md
- Done when: every finding has an issue with `Needs`, `Files`, and a checkable `Done when`, and the roadmap table is regenerated.

#### M5-32 Modifier targets are rows
- Status: review
- Priority: high
- Needs: —
- Do: The modifier is the only shared sentence, but its `target` is free text: the data names 145 targets, and 92 of them are a key in no table. A typo is a valid target, and nothing says whether a target is a number or a flag, or whether any rule reads it. A target names a row: an attribute, a skill, a derived stat, an action (its AP cost), a condition, a damage type, or a row of a new `modifier_targets` table for everything else (rule parameters such as `crit_window`, flags such as `darkvision`). Each `modifier_targets` row says its kind (`number` or `flag`) and what it changes, and is marked `pending` while no golden vector applies a modifier to it. Add a gate: every target resolves to one row, `grant` names a flag, `add`, `mul`, and `set` name a number, `resist` names a damage type, every `modifier_targets` row is used, and `pending` matches the vectors in both directions, so the flag cannot go stale. Register the targets used today without renaming any; renames are M5-33 and M5-37.
- Keys: `modifier_targets` collection, `modifier_target.schema.json`, `pending`
- Files: ~9 (docs/core/modifier.md, docs/core/integrity.md, docs/core/data-model.md, ruleset/schemas/modifier_target.schema.json, ruleset/data/modifier_targets/core.json, ruleset/modules/factions/, ruleset/scripts/validate.py, ruleset/scripts/test_integrity.py)
- See: docs/core/contract.md, docs/core/modifier.md, docs/core/integrity.md
- Done when: an unknown target, a `grant` on a number, an unused `modifier_targets` row, or a stale `pending` fails `validate.py`; every target in `ruleset/data/` and `ruleset/modules/` resolves; the tests pass.

#### M5-33 Percentages are `mul`, and distances are units
- Status: todo
- Priority: medium
- Needs: M5-32
- Do: 29 `add` modifiers carry a percentage in the target's name (`melee_damage_percent`, `mana_cost_percent`) instead of using `mul` on the base target, so there are two ways to scale a value. Distances in metres break the rule that every distance is in units (`docs/core/contract.md#distance`): `safe_fall_meters` says so in its name, and `darkvision_range` holds metres under a units description (`night_eyes` adds 9 for "6 units", `elf_darkvision` 18 for "12 units"). Rewrite each as `mul` on the target it scales (+10% is `mul 1.1`), or, where nothing to scale exists yet, as a target with a unit in its description, convert every distance to units, and remove the old `modifier_targets` rows. Renaming a target breaks consumers.
- Keys: removes the `*_percent` targets and `safe_fall_meters` **(breaking)**
- Files: ~6 (ruleset/data/features/core.json, ruleset/data/targets/core.json, docs/core/modifier.md, docs/modules/progression.md)
- See: docs/core/modifier.md, docs/core/contract.md#distance
- Done when: no target ends in `_percent` or names a unit other than units; `darkvision_range` values match their descriptions in units; each rewritten perk gives the same number at its listed rank, shown by a test or vector; `validate.py` and the tests pass.

#### M5-34 Retire the legacy combat table
- Status: todo
- Priority: high
- Needs: —
- Do: `ruleset/data/combat/base_resolution.json` still holds a second copy of the crit and fumble windows (`critical`), the AP pool and action costs (`actionEconomy`), and a `videoGame` column the contract forbids. `ruleset/scripts/combat_reference.py` reads it, and on a crit it rolls a second d100 to confirm, which `docs/core/resolution.md` and the Go backend do not do. No test runs that script. Move what still has no owner (`defense`, `hitQuality`, `calledShots`) to its owner or drop it, delete the copies and the `videoGame` block, and remove `combat_reference.py` or make it read the kernel tables and drop the confirmation roll.
- Keys: removes `combat/base_resolution.json` or most of it **(breaking)**
- Files: ~8 (ruleset/data/combat/, ruleset/schemas/combat.schema.json, ruleset/scripts/combat_reference.py, docs/core/resolution.md, docs/modules/combat.md, docs/combat/)
- See: docs/archive/review/findings-data-oriented-design.md (DOD-02, DOD-03, DOD-06), docs/decisions.md (the `base_resolution.json` risk)
- Done when: the crit window, fumble, crit damage, and AP pool each live in one table; no file has a `videoGame` key; no code rolls to confirm a crit; `validate.py` and the tests pass.

#### M5-35 A feature's bonus is written once
- Status: todo
- Priority: medium
- Needs: M5-32
- Do: A feature can state a bonus twice: in `effects` (a `target` and `value` for people) and in `modifiers` (the rows the rules apply). 14 features have both, and their targets differ (`hp_max` against `hp`, `melee_damage` against `melee_damage_percent`). 10 creation traits (`gifted`, `genius`, the phobias and allergies) have only `effects`, which name a shared effect and a value but often not what it applies to: `gifted` gives `fortify_attribute` 1 without naming an attribute, so the rule exists only in the description. Make `modifiers` the one statement of a feature's mechanics: give those 10 their modifiers (or applied effects with a `parameter`), and reduce `effects` to text, or drop it where `description` already says the same.
- Keys: `effects` on features **(breaking)**
- Files: ~4 (ruleset/data/features/core.json, ruleset/schemas/feature.schema.json, ruleset/scripts/validate.py, docs/modules/progression.md)
- See: docs/archive/review/findings-data-oriented-design.md (DOD-11), ROADMAP.md (definition of finished, item 1)
- Done when: no feature states a numeric bonus outside `modifiers`; every feature with a mechanical bonus has modifiers; `validate.py` and the tests pass.

#### M5-36 Composition bases are typed references
- Status: todo
- Priority: medium
- Needs: —
- Do: A monster's `base` is a bare creature-type name (`undead`) or a reference (`core:monster/skeleton`) in the same field, and an equipment row's `base` is a bare file name. A base's `default_materials` lists bare material keys. None of them is checked by the reference gate. Give creature types and equipment bases a category each, write every `base` and `default_materials` entry as a typed reference, and let `compose_row` take references.
- Keys: `core:creature_type/<key>`, `core:weapon_base/<key>`, `core:armor_base/<key>`, `core:accessory_base/<key>` [A] **(breaking)**
- Files: ~12 (ruleset/data/monsters/, ruleset/data/equipment/, ruleset/schemas/schema.json, ruleset/schemas/monster.schema.json, ruleset/schemas/equipment.schema.json, ruleset/scripts/validate.py, docs/core/keys.md, docs/core/integrity.md, docs/core/composition.md)
- See: docs/core/keys.md, docs/core/composition.md, docs/archive/review/findings-data-oriented-design.md (DOD-15)
- Done when: every `base` and every `default_materials` entry is a typed reference that the gate resolves; every composed row is unchanged; `validate.py` and the tests pass.

#### M5-37 One word per concept
- Status: todo
- Priority: low
- Needs: M5-34, M5-35
- Do: Some words mean two things. `abilities` is the eight attributes (`schema.json`, a feature's `prerequisite.abilities`) and also a monster's list of applied effects. `sp` is stamina (`derived`, `growth`) and silver pieces (`economy`). Called shots are `calledShots` in one file and `called_shots` in another (M5-34 removes one). The modifier target `attack` is the attack action's AP cost (`docs/core/actions.md`) and also a bonus on the attack roll (cover and called shots in `combat/situations.json`), and `block` is both a skill and an action, so a modifier on either is ambiguous (M5-32 rejects `block` until it is renamed). Rename the monster field, one of the two `sp`, the attack-roll target, and one `block`, in data, schemas, docs, and the backend.
- Keys: renames monster `abilities`, one `sp`, the attack-roll `attack` target, and one `block` **(breaking)**
- Files: ~12 (ruleset/data/monsters/, ruleset/schemas/monster.schema.json, ruleset/data/economy/core.json or ruleset/data/derived/core.json, ruleset/data/combat/situations.json, ruleset/tests/vectors/, backend/, docs/)
- See: CONTRIBUTING.md#naming ("one concept, one word")
- Done when: each word names one concept in data, schemas, docs, and code; `validate.py`, the tests, and `go test ./...` pass.

#### M5-38 The stored actor has a schema
- Status: doing
- Priority: high
- Needs: M5-25
- Do: `docs/core/state.md` says a session holds an actor snapshot, and `contract.md` lists what an actor has: resources, flags, inventory. The actor the backend reads (`backend/contract/common.json`) has only attributes, skills, level, and growth. Before APP-04 stores sessions, define the stored actor: what is stored (attributes, level, skills and tags, growth, current HP, MP, SP, and XP, active conditions with their stacks, active modifiers with source, origin, and turns elapsed, flags, inventory as item instances of base, material, and mods, what is equipped) and what is computed (maximum pools and every derived stat). A row in `ruleset/data/` is a definition; the actor holds instances that point at rows.
- Keys: the stored actor in `backend/contract/common.json`
- Files: ~5 (docs/core/state.md, docs/core/contract.md, backend/contract/common.json, ruleset/tests/vectors/)
- See: docs/core/state.md, docs/core/contract.md, docs/stories/07-backend.md (APP-04)
- Done when: one schema says what an actor stores; a vector round-trips a stored actor with a condition, a timed modifier, and an equipped item; `contract.md`'s actor matches it.

#### M5-39 Core docs match the data
- Status: todo
- Priority: medium
- Needs: —
- Do: `docs/core/contract.md` still writes a modifier `source` as `perk:12 | item:44`, numeric ids that `keys.md` rejects; still forbids a module to override a core row, which `overrides.md` allows; and lists durations without `while_applied`. `docs/core/data-model.md` lists numeric row ids as a known exception, but M5-30 removed them; it does not list authored HP growth on creature types (`growth.hitPointsPerLevel`). Fix both docs against the data.
- Keys: none
- Files: ~3 (docs/core/contract.md, docs/core/data-model.md)
- See: docs/archive/review/findings-data-oriented-design.md (DOD-08, DOD-09), docs/core/keys.md, docs/core/overrides.md
- Done when: every claim in the two docs about the data is tagged **[V]** with a date and true; no core doc contradicts another on sources, overrides, or durations.

#### M5-40 One property casing
- Status: todo
- Priority: low
- Needs: M5-37
- Do: The data mixes 72 `camelCase` property names (`hitPoints`, `appliedBy`) with 85 `snake_case` ones (`enchantment_slots`, `crit_window`), sometimes inside one schema. `CONTRIBUTING.md` defers unifying them to an issue of its own because it renames keys. Pick one casing, rename every property in data, schemas, reference code, the backend's JSON tags, and docs, and fail validation on the other.
- Keys: every property in the other casing **(breaking)**
- Files: ~20, splits by collection (ruleset/schemas/, ruleset/data/, ruleset/scripts/, backend/, docs/)
- See: CONTRIBUTING.md#naming, docs/archive/review/findings-data-oriented-design.md (DOD-20)
- Done when: every property in data and schemas uses one casing; a test fails on the other; `validate.py`, the tests, and `go test ./...` pass.

#### M5-41 A base with no changes is its own row
- Status: todo
- Priority: low
- Needs: M5-36
- Do: 31 of 33 weapon rows and 23 of 24 armor rows are `{ key, base, overrides: {} }` with the same key as their base, so every such key is written twice. Let a base be an item row itself, so a row exists only where it changes something, and keep every resolved row as it is.
- Keys: none
- Files: ~6 (ruleset/data/equipment/, ruleset/scripts/validate.py, docs/core/composition.md, docs/modules/equipment.md)
- See: docs/core/composition.md
- Done when: no equipment row has empty overrides; every resolved item is unchanged; `validate.py` and the tests pass.
