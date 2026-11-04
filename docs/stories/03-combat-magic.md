# M3 Combat and magic playable

Roadmap: [M3](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M3 Combat and magic playable

**Goal:** a combat and a casting round can be run at any level using only core rows.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** optional mechanics (cover, called shots, counterspell, and the like) are each either core with vectors or a named module.

Done: M3-04 ([done.md](done.md#m3-combat-and-magic-playable))

#### M3-01 Casting procedure: decide scope, then specify
- Status: todo
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
- Needs: M3-01
- Was: RW-17
- Do: 7 spells exist. 9 colors need a usable list.
- Files: ~2 (ruleset/data/spells/core.json, ruleset/data/effects/core.json)
- See: ruleset/data/spells/core.json
- Done when: at least 5 spells per color (45 total) across levels 1 to 30; each spell composes effects from the shared pool and respects `channels`; every color has a damage, utility, and defensive option or a documented reason it does not.

#### M3-03 Cover, flanking, terrain, and called shots as modifiers
- Status: todo
- Priority: medium
- Needs: M1-01, M1-05
- Was: RW-18
- Do: These are prose in `combat.md`. Decide core or optional, then express them as modifier rows so they stack under the one rule.
- Files: ~5 (docs/modules/combat.md, ruleset/data/combat/, ruleset/tests/vectors/combat.json)
- See: docs/modules/combat.md
- Done when: each is a row set with vectors, or is moved to a named optional module; no special-case stacking rule.

#### M3-05 Spell delivery and save rolls for instant spells
- Status: todo
- Priority: medium
- Needs: M1-02, M3-01
- Do: M1-02 makes the attack's `delivery` (`single`, `area`, `instant`) decide the dodge, but spells do not carry it yet, and an instant spell's save is unspecified. Give each offensive spell a delivery (from its effects' `target` shape, plus an `instant` marker), let each instant spell name its save skill or attribute, and define how the save margin picks the effect's outcome (none, partial, full).
- Keys: `delivery` and `save` on spells (proposal)
- Files: ~6 (docs/core/combat.md, docs/modules/magic.md, ruleset/schemas/spell.schema.json, ruleset/data/spells/core.json, ruleset/tests/vectors/, ruleset/scripts/kernel_reference.py)
- See: docs/core/combat.md (delivery), docs/decisions.md (M1-02)
- Done when: every offensive spell has a delivery; every instant spell names a save; vectors cover a full save, a partial save, and a failed save.
