# M3 Combat and magic playable

Roadmap: [M3](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M3 Combat and magic playable

**Goal:** a combat and a casting round can be run at any level using only core rows.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** optional mechanics (cover, called shots, counterspell, and the like) are each either core with vectors or a named module.

Done: M3-01, M3-03, M3-04 ([done.md](done.md#m3-combat-and-magic-playable))

#### M3-08 Summon spells name defined creatures
- Status: review
- Priority: high
- Needs: M3-01, M4-03
- Do: A summon spell may exist only once what it summons is defined as data. Today `summon`'s `parameter` is a free word (`familiar`, `elemental`, `undead`), so nothing says what appears; `bound_weapon` already names a row (`core:weapon/longsword`). Add a collection of summon groups: each row has a key, a label, and `variants`, a list of `core:monster/<key>` refs (the caster picks one when casting [A]). A summon's `parameter` becomes `core:summon/<key>`, and `magnitude` stays the count. Three groups to start: familiars, elementals, and undead. Familiars have no creature yet, so add a familiar base and at least three rows (proposal: small beasts with no attacks worth a fight [A]). Elementals use the existing Fire Elemental and Ice Elemental plus variants built from the elemental base and the element templates (`storm`, `acid`, `venom`). Undead use Skeleton and Zombie. Decide whether a stronger spell needs its own group with stronger variants, or scales one group by the spell's level. Convert all five summons on `main`: Summon Familiar, Summon Elemental, Conjured Legion, Raise Thrall, and Raise Horde. They shipped with M3-02 ahead of this issue.
- Keys: `core:summon/<key>` as the `summon` parameter; the `summons` collection; `ruleset/schemas/summon.schema.json`; a familiar base in `ruleset/data/monsters/bases/`
- Files: ~9 (docs/modules/magic.md, docs/modules/monsters.md, ruleset/schemas/summon.schema.json, ruleset/data/summons/core.json, ruleset/data/monsters/bases/familiar.json, ruleset/data/monsters/core.json, ruleset/data/spells/core.json, ruleset/scripts/validate.py, ruleset/scripts/test_spells.py)
- See: docs/modules/monsters.md#templates, ruleset/data/monsters/templates.json, `bound_weapon` on Bound Blade
- Done when: every group's variants resolve to monster rows; a test fails on a `summon` whose `parameter` is not an existing `core:summon/<key>`; all five summon spells name their group; `validate.py` and the spell tests pass.

#### M3-02 Spell list to the content floor
- Status: review
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
- Do: Spellcrafting composes only effects the character knows. Learning a spell (scroll, trainer, perk) teaches every craftable effect in it, not its target shape: shapes stay free and are priced by the target multiplier, so Fire Bolt lets you craft a fire burst. A downtime action, Research, teaches one craftable effect without a spell: days and gold at a library or guild, and a `<color>_magic` check against the effect's craft threshold (numbers [A]; the only downtime activity until M4). Effects gain `craftable` (default true): false means only an authored spell carries it. Uncraftable: `teleport` (new; Divine Intervention, one spell per deity, sends the caster to the nearest temple of that deity, its `parameter` a faction ref), `mark` and `recall` (new), `storage` (new; a pocket inventory), `bound_weapon`, and `summon`. Where an effect is craftable for some parameters only, `uncraftable_parameters` lists the others: `restore_resource` lists `mana`, so no crafted spell restores mana, instant or over time. Absorb and drain effects (`spell_absorption`, `life_steal`, `mana_leech`, `drain_*`, `damage_resource`) stay craftable: they take from a target, not from nothing. Define the `world` target in `docs/core/casting.md`: it is a valid target with a cost multiplier of 3.0, but no rule says what it reaches and no spell uses it; decide which of Mark, Recall, and Divine Intervention use it. Spells gain `template` (default false): a small set of plain, learnable spells, one per target shape (bolt, touch, burst, ward, mend), whose mana cost and cast threshold equal the spellcrafting cost formula exactly, so a crafted spell reads as a template with other effects. The regen family (health and stamina over time) uses `duration` with magnitude per second, as M2-05 set.
- Keys: `craftable` and `uncraftable_parameters` on effects, `template` on spells (proposal)
- Files: ~10 (docs/modules/magic.md, docs/core/casting.md, ruleset/schemas/effect.schema.json, ruleset/schemas/spell.schema.json, ruleset/data/effects/core.json, ruleset/data/spells/core.json, ruleset/data/actions/, ruleset/tests/vectors/spellcraft.json, ruleset/scripts/casting_reference.py, ruleset/scripts/test_spells.py)
- See: docs/modules/magic.md (spell crafting), docs/core/rest.md (downtime), docs/modules/factions.md
- Done when: a crafted spell with an unknown, uncraftable, or excluded effect or parameter is rejected; learning a spell teaches exactly its craftable effects; Research teaches one effect on a passed check; a crafted mana regen is rejected and a crafted mana leech passes; every template's cost and threshold match the formula (a test); Divine Intervention, Mark, Recall, Pocket Storage, and a bound weapon spell exist as authored spells; `docs/core/casting.md` says what the `world` target reaches; vectors cover each case. Resolving a temple, a mark, or a pocket inventory is engine state (M4 and the backend), not this issue.

#### M3-07 Cast cost by skill and perks
- Status: todo
- Priority: medium
- Needs: M3-02
- Do: A spell's cost is its effects' (M3-02); what the caster pays differs only by skill and perks, never by level. Apply the skill discount in the cast, `<color>_magic / 200` capped at 0.45 (Daggerfall lowers spell points with skill too, `gold × (110 − skill) / 400`), and the caster's modifiers on `spell_cost` (a perk, a robe), stacked by the kernel rules; the cast spends the result, rounded up, at least 1. The cast and craft thresholds read the cost before the discount.
- Keys: `spell_cost` modifier target; the discount in `casting/core.json`
- Files: ~7 (docs/core/casting.md, docs/modules/magic.md, ruleset/data/casting/core.json, ruleset/schemas/casting.schema.json, ruleset/scripts/casting_reference.py, backend/internal/casting/, ruleset/tests/vectors/casting.json)
- See: docs/modules/magic.md#the-casters-side
- Done when: vectors show the same spell costing less at a higher skill, a perk on `spell_cost` lowering it further, the cap at 0.45, the minimum of 1, and two casters of different levels and the same skill paying the same; Python and Go pass them.

#### M3-09 Bound weapons have a material, perks, and enchants
- Status: todo
- Priority: medium
- Needs: M3-02, M3-06
- Do: `bound_weapon` names only the weapon it copies, so it gets no stats from a material. A bound weapon starts with the stats of `core:material/iron` (tier 2, attack +3), written once as data and not on each spell. Perks swap in another base material and add enchants; decide how a perk names a material, since a modifier holds a number. The material gives stats only: a bound weapon overrides the material's phase and has none, so it sits out the five-phase cycles, and it deals `magic` damage in place of its weapon's physical type. `magic` is a new, plain damage type for non-elemental magic. `force` does not fit: its `bypasses_armor` and `unerring` tags would make a starting bound weapon too strong. `resist` with parameter `magic` reduces the new type, so a mage's magic resistance works against a bound weapon, and armor DR applies in full at the start (tag `armor_dr_applies`). A coating or an elemental enchantment overrides the damage type and gives the weapon that element's phase. Stronger enchants come one step at a time through perks: armor penetration (`core:effect/armor_pierce`), more damage to undead, banishing summoned creatures on a hit (a new effect), and damage over time of the weapon's current damage type, so `magic` by default and the coating's or enchant's type when one applies. The same rules apply to every bound weapon, including Bound Bow.
- Keys: the `magic` damage type; the default material for `bound_weapon`; bound weapon perks in `features/core.json`; `core:effect/banish_summoned`
- Files: ~11 (docs/core/combat.md, docs/modules/magic.md, docs/modules/weapons.md, ruleset/schemas/schema.json, ruleset/schemas/effect.schema.json, ruleset/data/combat/damage_types.json, ruleset/data/effects/core.json, ruleset/data/features/core.json, ruleset/tests/vectors/casting.json, ruleset/scripts/casting_reference.py, backend/internal/casting/)
- See: ruleset/schemas/crafted_item.schema.json (base, material, enchantments), ruleset/data/combat/damage_types.json, docs/modules/weapons.md, docs/modules/combat.md (armor DR and resist)
- Done when: vectors cover a bound weapon with iron's stats and no phase, `magic` damage reduced by `resist` and by armor DR and able to miss, a perk changing its material, a coating and an elemental enchantment each overriding the damage type and phase, and each perk step (armor penetration lowering the armor DR that a starting bound weapon takes in full, bonus against undead, banishing a summoned creature and failing on a creature that was not summoned, damage over time following the current type); Python and Go pass them.

#### M3-10 Power and nature templates for summons and monsters
- Status: todo
- Priority: medium
- Needs: M3-08
- Do: Summons stay static and uncraftable: a spell names rows, a row appears at the caster's level, and only a GM or dev adds a new type. A row sets the stats, so a summoned dog stays weaker than a summoned wolf at the same level, whatever is added on top. Templates gain a `slot`: `element` (the five that exist), `power` (`dire`, `ancient`, `elder`, and so on), and `nature` (`demonic`, `angelic`, `otherworldly`, and so on), applied in that order. `phase` and `damage_type` become optional, and a template may carry attribute adds, resistances, tags, and a spell list. Attributes stay capped at 10. A monster has one template per slot. Power is a ladder of perks: buying `elder` needs the `dire` perk (ordinary perk prerequisites), and the latest bought replaces the one before. Nature is kept simple here: the latest bought replaces the one before, and only it is tracked; whether a game locks it as a one-time choice is left for later. The caster's current power and nature apply when the creature is summoned, through the same merge as a row's templates, so no row is written per combination. A perk names what it unlocks in a new `unlocks` list of refs, because a modifier holds only a number; M3-09 can use the same field for a bound weapon's material. A template never changes the creature type: an undead horse has the `undead` base, not `beast`, so there is no `undead` template, and `dire` is not applied twice. Monster stats read as a player's do: a template's attribute adds move HP and mana by the derived rows' `per_point` (END +2 is +16 HP) [A], and a creature with a spell list gets mana from the `mp` row. Monster rows gain a reduced skill sheet, `skills`: only the skills the creature fights or casts with are written (its weapon skill, `evasion`, `block`, a `<color>_magic`), and any other skill reads as a player's starting value, `5 + associatedAbility × 2`. Existing rows keep their authored HP. The same merge generates monsters: any row plus a power and a nature is a monster, with no row written for the combination, so no one has to author an `ancient_otherworldly_dragon`. A generated monster is named by its row ref, `power`, and `nature` [A]. Authored rows can use the new templates too, and only rows kept as examples are written out: `dire_wolf` becomes a `wolf` row with `templates: ["dire"]`. Keep some written-out rows as examples: at least a dog and a wolf at the same level, a dire and an elder wolf, a demonic hound, and an undead horse. An example's HP and mana come from the player formulas with average rolls: the derived function plus `(level − 1) × ⌊LCK / 2⌋ / 2` growth, rounded down; check them against the balance bands. Fix the stale seed formula in `skill.schema.json`. Summon groups are not split by tier: the caster's level does that job (M3-08).
- Keys: `slot`, `attributes`, and `spells` on monster templates; `skills` on monsters; `unlocks` on features; `dire`, `ancient`, `elder`, `demonic`, `angelic`, `otherworldly`; the power and nature the caster holds
- Files: ~14 (docs/modules/monsters.md, docs/modules/magic.md, ruleset/schemas/monster_template.schema.json, ruleset/schemas/monster.schema.json, ruleset/schemas/skill.schema.json, ruleset/schemas/feature.schema.json, ruleset/data/monsters/templates.json, ruleset/data/monsters/core.json, ruleset/data/loot/core.json, ruleset/data/summons/core.json, ruleset/data/features/core.json, ruleset/scripts/validate.py, ruleset/scripts/test_bestiary.py, ruleset/scripts/test_spells.py)
- See: docs/modules/monsters.md#templates, ruleset/data/summons/core.json, ruleset/data/features/core.json (`conjurer`), M3-09
- Done when: a test shows a dog weaker than a wolf with the same template; `elder` cannot be bought without `dire`; buying a later power or nature replaces the earlier one; two templates in one slot on a row fail validation; a monster generated from a row plus `ancient` and `otherworldly` composes with no row written for it; no template changes a creature type; `dire_wolf` comes from `wolf` plus `dire` with its stats unchanged or the change listed in the PR; the example rows exist and their HP and mana match the formulas at average rolls; a skill missing from a monster's sheet reads `5 + associatedAbility × 2`; `validate.py` and the tests pass.
