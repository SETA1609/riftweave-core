# Roadmap

The core SRD is **finished** when it is tagged `v1.0.0` and every item under
[Definition of finished](#definition-of-finished) is true. This file is the source of truth for
scope; the issues it links to are the source of truth for work. Linear mirrors both and links
back here (see [issues/README.md](issues/README.md)).

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
[issues/done/](issues/done/); what remains is closing the kernel, then filling content.

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

| Issue | Title | Priority | Status |
| --- | --- | --- | --- |
| [RW-3](issues/open/RW-3-reconcile-docs-modules-combat-md-with-the-kernel.md) | Reconcile docs/modules/combat.md with the kernel | high | open |
| [RW-4](issues/open/RW-4-combat-pipeline-as-data-formulas-and-vectors.md) | Combat pipeline as data, formulas, and vectors | high | open |
| [RW-5](issues/open/RW-5-progression-as-data-xp-skill-points-perk-cadence.md) | Progression as data: XP, skill points, perk cadence | high | open |
| [RW-6](issues/open/RW-6-rest-and-recovery-rules-as-data.md) | Rest and recovery rules as data | medium | open |
| [RW-7](issues/open/RW-7-modifier-stacking-and-condition-vectors.md) | Modifier stacking and condition vectors | medium | open |
| [RW-8](issues/open/RW-8-resolve-the-traits-vs-features-fork.md) | Resolve the traits vs features fork | medium | open |
| [RW-9](issues/open/RW-9-one-test-runs-every-vector-file.md) | One test runs every vector file | high | open |
| [RW-10](issues/open/RW-10-ci-runs-unit-tests-and-vectors.md) | CI runs unit tests and vectors | high | open |

## M2 Characters

**Exit:** a character can be built and advanced from level 1 to 30 with real choices. Perk,
skill, race, background, and equipment floors are met, and two multi-level characters are
vectors.

| Issue | Title | Priority | Status |
| --- | --- | --- | --- |
| [RW-11](issues/open/RW-11-perk-catalogue-to-the-content-floor.md) | Perk catalogue to the content floor | high | open |
| [RW-12](issues/open/RW-12-skill-audit-no-orphan-skills.md) | Skill audit: no orphan skills | medium | open |
| [RW-13](issues/open/RW-13-multi-level-example-characters-as-vectors.md) | Multi-level example characters as vectors | high | open |
| [RW-14](issues/open/RW-14-races-and-backgrounds-emit-modifiers-only.md) | Races and backgrounds emit modifiers only | medium | open |
| [RW-15](issues/open/RW-15-equipment-to-the-content-floor.md) | Equipment to the content floor | medium | open |

## M3 Combat and magic playable

**Exit:** a combat and a casting round can be run at any level using only core rows.
Optional mechanics (cover, called shots, counterspell, and the like) are each either core with
vectors or a named module.

| Issue | Title | Priority | Status |
| --- | --- | --- | --- |
| [RW-16](issues/open/RW-16-casting-procedure-decide-scope-then-specify.md) | Casting procedure: decide scope, then specify | high | open |
| [RW-17](issues/open/RW-17-spell-list-to-the-content-floor.md) | Spell list to the content floor | high | open |
| [RW-18](issues/open/RW-18-cover-flanking-terrain-and-called-shots-as-modifiers.md) | Cover, flanking, terrain, and called shots as modifiers | medium | open |
| [RW-19](issues/open/RW-19-wuxing-interaction-vectors.md) | Wuxing interaction vectors | medium | open |

## M4 World

**Exit:** buy, sell, loot, and fight monsters across levels 1 to 30, with hazards. Prices,
treasure, and monster stats are consistent under checkable balance bands.

| Issue | Title | Priority | Status |
| --- | --- | --- | --- |
| [RW-20](issues/open/RW-20-economy-currency-price-formula-merchants.md) | Economy: currency, price formula, merchants | high | open |
| [RW-21](issues/open/RW-21-loot-and-treasure-tables.md) | Loot and treasure tables | high | open |
| [RW-22](issues/open/RW-22-bestiary-to-the-content-floor.md) | Bestiary to the content floor | high | open |
| [RW-23](issues/open/RW-23-traps-and-environmental-hazards.md) | Traps and environmental hazards | medium | open |
| [RW-24](issues/open/RW-24-encounter-and-balance-validation.md) | Encounter and balance validation | medium | open |

## M5 Release 1.0

**Exit:** items 3, 5, 6, 7, and 8 of the definition hold, and `v1.0.0` is tagged.

| Issue | Title | Priority | Status |
| --- | --- | --- | --- |
| [RW-25](issues/open/RW-25-reference-playthrough-vector-chain.md) | Reference playthrough vector chain | high | open |
| [RW-26](issues/open/RW-26-player-and-gm-guide.md) | Player and GM guide | medium | open |
| [RW-27](issues/open/RW-27-entry-docs-and-review-archive-sync.md) | Entry docs and review archive sync | medium | open |
| [RW-28](issues/open/RW-28-key-stability-and-versioning-policy.md) | Key stability and versioning policy | high | open |
| [RW-29](issues/open/RW-29-integrity-gates.md) | Integrity gates | high | open |
| [RW-30](issues/open/RW-30-licensing-and-attribution-audit-tag-v1-0-0.md) | Licensing and attribution audit, tag v1.0.0 | high | open |

## Post-1.0 modules

Each ships as an optional module with a manifest, is written against the kernel contract, and
does not change core. Order is by demand, not commitment.

| Issue | Title | Priority | Status |
| --- | --- | --- | --- |
| [RW-31](issues/open/RW-31-crafting-module-recipes-quality-stations.md) | Crafting module: recipes, quality, stations | medium | open |
| [RW-32](issues/open/RW-32-alchemy-module-quality-formula-and-discovery.md) | Alchemy module: quality formula and discovery | medium | open |
| [RW-33](issues/open/RW-33-magic-crafting-module-enchanting.md) | Magic crafting module: enchanting | low | open |
| [RW-34](issues/open/RW-34-breakthrough-advancement-module.md) | Breakthrough advancement module | medium | open |
| [RW-35](issues/open/RW-35-factions-and-reputation-module.md) | Factions and reputation module | medium | open |
| [RW-36](issues/open/RW-36-companions-and-followers-module.md) | Companions and followers module | medium | open |
| [RW-37](issues/open/RW-37-mounts-vehicles-and-ship-combat-module.md) | Mounts, vehicles, and ship combat module | low | open |
| [RW-38](issues/open/RW-38-survival-needs-module.md) | Survival needs module | low | open |

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

Move work between milestones only with a note in the issue (the `milestone:` field). When an
issue closes, move its file to `done/` and delete its row from the table above. Change the
definition of finished only in a PR that says why.
