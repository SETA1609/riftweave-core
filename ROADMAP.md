# Roadmap

The core SRD is **finished** when it is tagged `v1.0.0` and every item under
[Definition of finished](#definition-of-finished) is true. This file is the source of truth for
scope; the story files it links to are the source of truth for work. Linear mirrors both and links
back here (see [docs/stories/README.md](docs/stories/README.md)).

## What "core SRD" means

Riftweave is data plus schemas plus docs. There is no engine here. The core is finished when a
programmer who has never seen this repo could build a game, or a referee could run a table,
from `ruleset/data/`, `ruleset/schemas/`, `docs/core/`, and the golden vectors alone, and a
second implementation that disagrees with the vectors is wrong.

**In core (1.0):** the kernel contract, d100 resolution, derived stats, the AP action budget,
modifiers and conditions, creation and level 1 to 30 progression, combat, casting, equipment,
economy and loot, a bestiary, hazards, and the guides to use them.

**Not in core (post-1.0 modules):** crafting, alchemy, magic crafting, breakthrough
advancement, factions, companions, mounts and vehicles, survival needs, psionics, and alignment.
A module may extend core through modifiers and rows but core never requires one
([contract](docs/core/contract.md), [overrides](docs/core/overrides.md)).

**Out of scope for this repo:** engines, UIs, and virtual tabletop code. A Zig or other second
implementation is the long-term conformance proof, but it lives elsewhere. The 1.0 gate is that
the vectors fully specify the kernel, not that a second implementation exists.

## Definition of finished

All of these are checked by a person or by CI. None is a feeling.

1. **Closed kernel.** Every core rule is rows plus a formula or modifier, and has at least one
   golden vector. No rule exists only as prose. (M1)
2. **One authority.** `docs/core/` wins over `docs/modules/`. No doc contradicts it. (M1, M5)
3. **Playable end to end.** One actor can be created, equipped, led from level 1 to 30, fight,
   cast, take and clear conditions, rest, loot, buy, and sell using only core rows. The
   reference playthrough chain passes. (M2 to M5)
4. **Content floors met.** Proposed targets, to be adjusted by decision, not by drift:
   60 perks, 45 spells (5 per color), 40 monsters across levels 1 to 30 and all five phases,
   at least 3 weapons per weapon skill, 12 accessories, 20 consumables, 10 hazards.
5. **Gates green.** CI runs validate, unit tests, and every vector on each PR. All
   cross-reference categories are checked. No schema accepts a bare string where a key belongs.
6. **Stable keys.** A versioning policy exists. Renaming or removing a key is a breaking change,
   marked in the changelog and detectable by a key diff between tags. (M5)
7. **Documented for use.** A quick start for players and a GM section exist. README, AGENTS, and
   CONTRIBUTING match the repo. (M5)
8. **Legally clean.** Licence, attribution, and `source` audit is done. (M5)

## Where we are

Foundations are done. Milestone M0 below is complete and archived in
[docs/stories/done.md](docs/stories/done.md); what remains is closing the kernel, then filling content.

| Milestone | State |
| --- | --- |
| M0 Foundations | Done: kernel contract, resolution, derived stats, action budget, modifiers, conditions, formula AST, golden vectors (resolution, derived HP), module overrides, one namespaced key format (SET-6 to SET-25) |
| M1 Kernel closed | Not started |
| M2 Characters | Not started |
| M3 Combat and magic playable | Not started |
| M4 World | Not started |
| M5 Release 1.0 | Not started |

Content today, for scale: 23 features, 7 spells, 11 monsters, 3 recipes, 3 accessories,
7 consumables, 44 skills, 12 races, 8 backgrounds.

## M1 Kernel closed

**Exit:** items 1 and 2 of the definition hold. Every kernel function has a vector, combat and
progression are data, and CI enforces it. This is the highest-leverage milestone: content
written before it must be rechecked after.

Issues, in order: [docs/stories/01-kernel.md](docs/stories/01-kernel.md).

## M2 Characters

**Exit:** a character can be built and advanced from level 1 to 30 with real choices. Perk,
skill, race, background, and equipment floors are met, and two multi-level characters are
vectors.

Issues, in order: [docs/stories/02-characters.md](docs/stories/02-characters.md).

## M3 Combat and magic playable

**Exit:** a combat and a casting round can be run at any level using only core rows.
Optional mechanics (cover, called shots, counterspell, and the like) are each either core with
vectors or a named module.

Issues, in order: [docs/stories/03-combat-magic.md](docs/stories/03-combat-magic.md).

## M4 World

**Exit:** buy, sell, loot, and fight monsters across levels 1 to 30, with hazards. Prices,
treasure, and monster stats are consistent under checkable balance bands.

Issues, in order: [docs/stories/04-world.md](docs/stories/04-world.md).

## M5 Release 1.0

**Exit:** items 3, 5, 6, 7, and 8 of the definition hold, and `v1.0.0` is tagged.

Issues, in order: [docs/stories/05-release.md](docs/stories/05-release.md).

## Post-1.0 modules

Each ships as an optional module with a manifest, is written against the kernel contract, and
does not change core. Order is by demand, not commitment.

Issues, in order: [docs/stories/06-modules.md](docs/stories/06-modules.md).

Also parked with no issue yet: psionics, alignment and morality, birthsigns. Open one when it
is picked up.

## Assumptions (change any of these)

- Content floors in item 4 are proposals. They are sized for a first playable SRD, not a
  complete game.
- Casting extras (components, concentration, counterspell, ritual, spell failure) are left to
  RW-16 to decide.
- Crafting, alchemy, and breakthroughs are modules, not core, because core level-up must not
  name them (SET-14, SET-17).
- A second implementation is not a 1.0 gate; complete vectors are.
- Repo-originated issues use the `RW-` prefix so they cannot collide with Linear's `SET-`
  numbers. Archived `SET-` ids are unchanged.

## Updating this file

Move work between milestones only by moving the issue between story files, with a note in
the issue. Issues close in their story file ([docs/stories/README.md](docs/stories/README.md#workflow)),
so this file does not change when an issue closes. Change the definition of finished only in
a PR that says why.
