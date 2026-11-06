# M3 Combat and magic playable

Roadmap: [M3](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M3 Combat and magic playable

**Goal:** a combat and a casting round can be run at any level using only core rows.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** optional mechanics (cover, called shots, counterspell, and the like) are each either core with vectors or a named module.

Done: M3-03, M3-04 ([done.md](done.md#m3-combat-and-magic-playable))

#### M3-01 Casting procedure: decide scope, then specify
- Status: review
- Priority: high
- Needs: M1-02, M1-05
- Was: RW-16
- Do: Magic emits a roll request and an effect list (SET-16). Still undecided: components, concentration, counterspell and dispel, ritual casting, and spell failure from armor. Decide which are core and which are optional modules, and write the core ones as data.
- Files: ~6 (docs/modules/magic.md, docs/core/, ruleset/data/spells/core.json, ruleset/schemas/spell.schema.json, ruleset/tests/vectors/casting.json)
- See: issues.md section 4, docs/modules/magic.md
- Done when: A decision record lists each mechanic as core, module, or cut; core mechanics have rows, formulas, and vectors; `magic.md` matches.

#### M3-02 Spell list to the content floor
- Status: todo
- Priority: high
- Needs: M3-01, M4-04
- Was: RW-17
- Do: 7 spells exist. 9 colors need a usable list. Include spells that work on hazards and locks: disarm trap (level N), dispel (area, arcane hazards of level N or below), and unlock (level N). For them, hazards gain a `kind` (`mechanical`, `natural`, `arcane`), a `color` for arcane ones, and a `level`; the kind picks the detect and disarm check, so an arcane trap uses `<color>_magic` (or a dispel check) instead of lockpick. A skill buff on lockpick or investigation already works through the hazard check's modifiers.
- Files: ~8 (ruleset/data/spells/core.json, ruleset/data/effects/core.json, ruleset/data/hazards/core.json, ruleset/data/checks/core.json, ruleset/schemas/hazard.schema.json, docs/core/hazards.md, ruleset/tests/vectors/hazard.json, backend/internal/hazard/hazard.go)
- See: ruleset/data/spells/core.json
- Done when: at least 5 spells per color (45 total) across levels 1 to 30; each spell composes effects from the shared pool and respects `channels`; every color has a damage, utility, and defensive option or a documented reason it does not; disarm trap, dispel, and unlock spells exist, an arcane trap is detected and disarmed with its color's skill, and hazard vectors cover each spell and each kind in Python and Go.

#### M3-05 Spell delivery and save rolls for instant spells
- Status: todo
- Priority: medium
- Needs: M1-02, M3-01
- Do: M1-02 makes the attack's `delivery` (`single`, `area`, `instant`) decide the dodge, but spells do not carry it yet, and an instant spell's save is unspecified. Give each offensive spell a delivery (from its effects' `target` shape, plus an `instant` marker), let each instant spell name its save skill or attribute, and define how the save margin picks the effect's outcome (none, partial, full).
- Keys: `delivery` and `save` on spells (proposal)
- Files: ~6 (docs/core/combat.md, docs/modules/magic.md, ruleset/schemas/spell.schema.json, ruleset/data/spells/core.json, ruleset/tests/vectors/, ruleset/scripts/kernel_reference.py)
- See: docs/core/combat.md (delivery), docs/decisions.md (M1-02)
- Done when: every offensive spell has a delivery; every instant spell names a save; vectors cover a full save, a partial save, and a failed save.

#### M3-06 Spell knowledge, template spells, and uncraftable effects
- Status: todo
- Priority: medium
- Needs: M3-01
- Do: Spellcrafting composes only effects the character knows. Learning a spell (scroll, trainer, perk) teaches every craftable effect in it, not its target shape: shapes stay free and are priced by the target multiplier, so Fire Bolt lets you craft a fire burst. A downtime action, Research, teaches one craftable effect without a spell: days and gold at a library or guild, and a `<color>_magic` check against the effect's craft threshold (numbers [A]; the only downtime activity until M4). Effects gain `craftable` (default true): false means only an authored spell carries it. Uncraftable: `teleport` (new; Divine Intervention, one spell per deity, sends the caster to the nearest temple of that deity, its `parameter` a faction ref), `mark` and `recall` (new), `storage` (new; a pocket inventory), `bound_weapon`, and `summon`. Where an effect is craftable for some parameters only, `uncraftable_parameters` lists the others: `restore_resource` lists `mana`, so no crafted spell restores mana, instant or over time. Absorb and drain effects (`spell_absorption`, `life_steal`, `mana_leech`, `drain_*`, `damage_resource`) stay craftable: they take from a target, not from nothing. Spells gain `template` (default false): a small set of plain, learnable spells, one per target shape (bolt, touch, burst, ward, mend), whose mana cost and cast threshold equal the spellcrafting cost formula exactly, so a crafted spell reads as a template with other effects. The regen family (health and stamina over time) uses `duration` with magnitude per second, as M2-05 set.
- Keys: `craftable` and `uncraftable_parameters` on effects, `template` on spells (proposal)
- Files: ~10 (docs/modules/magic.md, docs/core/casting.md, ruleset/schemas/effect.schema.json, ruleset/schemas/spell.schema.json, ruleset/data/effects/core.json, ruleset/data/spells/core.json, ruleset/data/actions/, ruleset/tests/vectors/spellcraft.json, ruleset/scripts/casting_reference.py, ruleset/scripts/test_spells.py)
- See: docs/modules/magic.md (spell crafting), docs/core/rest.md (downtime), docs/modules/factions.md
- Done when: a crafted spell with an unknown, uncraftable, or excluded effect or parameter is rejected; learning a spell teaches exactly its craftable effects; Research teaches one effect on a passed check; a crafted mana regen is rejected and a crafted mana leech passes; every template's cost and threshold match the formula (a test); Divine Intervention, Mark, Recall, Pocket Storage, and a bound weapon spell exist as authored spells; vectors cover each case. Resolving a temple, a mark, or a pocket inventory is engine state (M4 and the backend), not this issue.
