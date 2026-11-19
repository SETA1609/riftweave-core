# Suggested builds

Modified: 2026-11-18

Riftweave has no classes. A **suggested build** is a worked plan for reaching one style of play, from character creation to level 30, for a player who does not yet know the perk table. It is a suggestion and nothing more.

**Nothing is gated by a build.** Any character can use any weapon, cast any spell it knows, and raise any skill. Leaving a build is only less efficient, never forbidden: a fencer who picks up a maul still hits with it, just with a lower skill and no maul perks behind it. A build names no new rule, perk, or item; every perk it suggests is a row in [`features/core.json`](../../../ruleset/data/features/core.json), and every number it relies on is in the module docs it links to.

## What a build holds

The builds are rows in [`ruleset/data/builds/core.json`](../../../ruleset/data/builds/core.json) (`build.schema.json`). Each one holds:

- **`path` and `sub_path`.** `melee`, `magic`, `ranged`, or `hybrid`, and the style inside it.
- **`attributes`.** The point buy of [character creation](../character-creation.md#2-assign-attributes) step 2: 4 in every attribute plus 10 points, none over 10. Race modifiers come on top; a build names no race, so any race can follow it.
- **`creation_perks`.** Level 1: up to two creation perks (the trade-off traits of step 4).
- **`suggested_tag_skills`.** The three tag skills of step 5 (four with Genius).
- **`skill_priorities`.** Where the skill points go, in order: each entry raises one skill to a value before the next entry gets anything. An entry can come back later with a higher value. Points left over carry to the next level.
- **`perks_by_level`.** A perk for every level that grants one, levels 2 to 30. A perk with ranks appears once per rank.
- **`spells`** (a caster's). The spells it leans on, in about the order it learns them (M5-60).

## How a build is checked

`build_reference.play` plays a build from creation to level 30 with the kernel's own rules: the skill-point pool of [progression.md](../progression.md) at the **lowest luck roll** (0) every level, the tag and normal costs, the skill cap, and each perk through the same `take_perk` check a character's level-up uses (level, attributes, skills, perks, rank). A build that passes is legal for any luck roll; race, background, and gear only add to it. `test_builds.py` fails on any build that breaks a rule at any level. [V 2026-11-16, `test_builds.py`]

Where a path has no perk worth taking at some level, the build still names a legal perk, and the gap is listed in the issue's decision record instead of inventing a perk ([M5-47](../../decisions/M5-47.md)).

## How to read a build

1. Assign the attributes, take the creation perks, and tag the skills.
2. Each level, spend the pool down the priority list and take the level's perk.
3. Read the build's page for why each perk is there and what to do at the table.

## How to leave a build

Every build page ends with what the build can still do outside its style, and at what cost. To change course, keep the perks you have (a perk never stops working), move the next skill points to the new style, and pick perks from the new style's page once you meet their prerequisites. Nothing is lost but the points already spent.

## The builds

| Path | Build | Page |
| --- | --- | --- |
| Melee | Shield-bearer | [melee.md](melee.md#shield-bearer) |
| Melee | Two-handed fighter | [melee.md](melee.md#two-handed-fighter) |
| Melee | Fencer | [melee.md](melee.md#fencer) |
| Magic | Blaster | [magic.md](magic.md#blaster) |
| Magic | Summoner | [magic.md](magic.md#summoner) |
| Magic | Support | [magic.md](magic.md#support) |
| Ranged | Archer | [ranged.md](ranged.md#archer) |
| Ranged | Arbalist | [ranged.md](ranged.md#arbalist) |
| Ranged | Gunslinger | [ranged.md](ranged.md#gunslinger) |
| Hybrid | Magic swordsman | [hybrids.md](hybrids.md#magic-swordsman) |
| Hybrid | Arcane trickster | [hybrids.md](hybrids.md#arcane-trickster) |
| Hybrid | Ranger | [hybrids.md](hybrids.md#ranger) |
