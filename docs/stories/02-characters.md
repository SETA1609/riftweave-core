# M2 Characters

Roadmap: [M2](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M2 Characters

**Goal:** a character can be built and advanced from level 1 to 30 with real choices.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** perk, skill, race, background, and equipment floors are met, and two multi-level characters are vectors.

Done: —

#### M2-01 Perk catalogue to the content floor
- Status: todo
- Priority: high
- Needs: M1-05, M1-06
- Was: RW-11
- Do: 23 features exist (13 perks, 10 creation). A character gains a perk about every 3 levels, so 10 or more by level 30; real choice needs a wide pool.
- Files: ~3 (ruleset/data/features/core.json, docs/modules/progression.md)
- See: ruleset/data/features/core.json, docs/modules/progression.md
- Done when: at least 60 perks, spread across all 8 attributes and the combat, magic, and social skill groups; every perk emits modifiers; prerequisites are keys; no perk is dominated by another at the same level.

#### M2-02 Skill audit: no orphan skills
- Status: doing
- Priority: medium
- Needs: —
- Was: RW-12
- Do: 44 skills. Each needs a governing attribute and at least one consumer: a weapon, spell color, check, or perk.
- Files: ~4 (ruleset/data/skills/core.json, docs/modules/attributes.md, ruleset/scripts/)
- See: ruleset/data/skills/core.json, docs/modules/attributes.md
- Done when: script lists skills with zero references; each is wired up or removed; `attributes.md` seeds match the data.

#### M2-03 Multi-level example characters as vectors
- Status: todo
- Priority: high
- Needs: M1-03, M1-05
- Was: RW-13
- Do: Single-level creation examples exist. Walk two characters (one martial, one caster) from creation to level 10, with perks, tag skills, gear, and spells, and encode them as vectors.
- Files: ~4 (ruleset/tests/vectors/characters.json, docs/modules/character-creation.md)
- See: docs/modules/character-creation.md, docs/core/loop.md
- Done when: two characters, each with a level-by-level record; each level's derived stats match a vector; prose in `character-creation.md` is generated from or checked against the record.

#### M2-04 Races and backgrounds emit modifiers only
- Status: todo
- Priority: medium
- Needs: M1-05, M1-06
- Was: RW-14
- Do: 12 races and 8 backgrounds. Confirm every racial effect and background bonus is a modifier row, not prose, and that lineage inheritance is covered by tests.
- Files: ~6 (ruleset/data/races/core.json, ruleset/data/backgrounds/core.json, ruleset/schemas/race.schema.json, ruleset/schemas/background.schema.json, ruleset/scripts/test_validate.py)
- See: docs/modules/race.md, docs/modules/backgrounds.md
- Done when: no mechanical statement exists only in a description string; lineage (parent, subrace, template, kin) has validator tests for each role.

#### M2-05 Equipment to the content floor
- Status: todo
- Priority: medium
- Needs: M1-02
- Was: RW-15
- Do: 20 weapons, 24 armor, 3 accessories, 7 consumables. Accessories and consumables are far below playable.
- Files: ~6 (ruleset/data/equipment/, ruleset/schemas/equipment.schema.json)
- See: ruleset/data/equipment/
- Done when: each weapon skill has at least 3 weapons across tiers; at least 12 accessories and 20 consumables; adventuring gear and tools exist as rows.
