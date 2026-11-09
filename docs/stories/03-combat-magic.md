# M3 Combat and magic playable

Roadmap: [M3](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M3 Combat and magic playable

**Goal:** a combat and a casting round can be run at any level using only core rows.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** optional mechanics (cover, called shots, counterspell, and the like) are each either core with vectors or a named module.

Done: M3-01, M3-03, M3-04 ([done.md](done.md#m3-combat-and-magic-playable))

#### M3-08 Summon spells name defined creatures
- Status: todo
- Priority: high
- Needs: M3-01, M4-03
- Do: A summon spell may exist only once what it summons is defined as data. Today `summon`'s `parameter` is a free word (`undead`), so nothing says what appears; `bound_weapon` already names a row (`core:weapon/longsword`). Add a collection of summon groups: each row has a key, a label, and `variants`, a list of `core:monster/<key>` refs (the caster picks one when casting [A]). A summon's `parameter` becomes `core:summon/<key>`, and `magnitude` stays the count. Three groups to start: familiars, elementals, and undead. Familiars have no creature yet, so add a familiar base and at least three rows (proposal: small beasts with no attacks worth a fight [A]). Elementals use the existing Fire Elemental and Ice Elemental plus variants built from the elemental base and the element templates (`storm`, `acid`, `venom`). Undead use Skeleton and Zombie. Decide whether a stronger spell needs its own group with stronger variants, or scales one group by the spell's level. Convert `raise_thrall`, the one summon on `main`. M3-02 adds four more (Summon Familiar, Summon Elemental, Conjured Legion, Raise Horde); it rebases on this issue and converts them before it merges.
- Keys: `core:summon/<key>` as the `summon` parameter; the `summons` collection; `ruleset/schemas/summon.schema.json`; a familiar base in `ruleset/data/monsters/bases/`
- Files: ~9 (docs/modules/magic.md, docs/modules/monsters.md, ruleset/schemas/summon.schema.json, ruleset/data/summons/core.json, ruleset/data/monsters/bases/familiar.json, ruleset/data/monsters/core.json, ruleset/data/spells/core.json, ruleset/scripts/validate.py, ruleset/scripts/test_spells.py)
- See: docs/modules/monsters.md#templates, ruleset/data/monsters/templates.json, `bound_weapon` on Bound Blade
- Done when: every group's variants resolve to monster rows; a test fails on a `summon` whose `parameter` is not an existing `core:summon/<key>`; Raise Thrall names its group; `validate.py` and the spell tests pass.

#### M3-02 Spell list to the content floor
- Status: review
- Priority: high
- Needs: M3-01, M3-08, M4-04
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

#### M3-07 Cast cost by skill and perks
- Status: todo
- Priority: medium
- Needs: M3-02
- Do: A spell's cost is its effects' (M3-02); what the caster pays differs only by skill and perks, never by level. Apply the skill discount in the cast, `<color>_magic / 200` capped at 0.45 (Daggerfall lowers spell points with skill too, `gold × (110 − skill) / 400`), and the caster's modifiers on `spell_cost` (a perk, a robe), stacked by the kernel rules; the cast spends the result, rounded up, at least 1. The cast and craft thresholds read the cost before the discount.
- Keys: `spell_cost` modifier target; the discount in `casting/core.json`
- Files: ~7 (docs/core/casting.md, docs/modules/magic.md, ruleset/data/casting/core.json, ruleset/schemas/casting.schema.json, ruleset/scripts/casting_reference.py, backend/internal/casting/, ruleset/tests/vectors/casting.json)
- See: docs/modules/magic.md#the-casters-side
- Done when: vectors show the same spell costing less at a higher skill, a perk on `spell_cost` lowering it further, the cap at 0.45, the minimum of 1, and two casters of different levels and the same skill paying the same; Python and Go pass them.
