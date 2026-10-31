# Changelog

One entry per **closed issue**, not per commit. Format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/). Each line is
`**ID** Title (#PR)`, written in the PR that moves the issue to `issues/done/`. Work with no
issue does not get an entry: open the issue first.

Categories: **Added**, **Changed**, **Deprecated**, **Removed**, **Fixed**, **Security**.
A renamed or removed key breaks consumers, so mark it **(breaking)**.

There are no tagged releases yet. `[Unreleased]` becomes `[1.0.0]` when the roadmap's
[definition of finished](ROADMAP.md#definition-of-finished) holds; see
[ROADMAP.md](ROADMAP.md) for what remains. Ids: `SET-NN` were tracked in Linear,
`RW-NN` are tracked in this repo ([issues/README.md](issues/README.md)).

## [Unreleased]

### Added
- **RW-1** Dev toolchain container and host devops CLI.
- **RW-2** In-repo changelog, issue tracker, archive, and roadmap.
- **RW-39** Five-step workflow, PR size rule, verification tags, naming rules, decision log, open-PR rebasing, and `fetch-docs`; Actions pinned to SHAs.
- **SET-6** Kernel contract: actor, modifier, event. (#18)
- **SET-20** Formula AST; the AP formula is a JSON tree. (#37)
- **SET-22** Golden vectors for resolution and derived HP. (#36)
- **SET-25** Module overrides: a module may override a vanilla row by key. (#39)
- **SET-15** Vertical loop example: create, equip, attack, DR, XP, perk. (#27)

### Changed
- **SET-7** Derived stats are functions of the actor snapshot, not stored truth. (#20)
- **SET-8** One d100 resolution function; crit and fumble are tags on the result. (#21)
- **SET-9** Action budget replaces Action, Bonus, and Reaction. (#34)
- **SET-10** Conditions are modifier bundles. (#23)
- **SET-11** Modules may add effect and condition rows. (#24)
- **SET-12** Armor core is slot and DR; layering is a module. (#25)
- **SET-13** Damage-type vocabulary unified; `cold`/`lightning` are aliases of `frost`/`shock`. (#19)
- **SET-14** Level-up from 1 to 2 requires no alchemy, core, or phase. (#26)
- **SET-16** Magic emits a roll request and an effect list. (#30)
- **SET-17** Alchemy is a crafting profile. (#31)
- **SET-18** Wuxing is a tag reader, not a core dependency. (#32)
- **SET-19** Material identity is separate from form. (#28)
- **SET-21** One modifier schema: `add`, `mul`, `set`, `grant`, `resist`. (#35)

## Before issue tracking (2026-06-07 to 2026-10-18)

Work done before issues were tracked, summarised by capability rather than listed.

- Classless d100 ruleset, shared effect pool, five-phase (Wuxing) system, and race lineage.
- Materials, gems, tiers, weapon classification, and the base + override model for
  weapons, armor, accessories, and monsters.
- Combat documentation and conditions, plus the first crafting composition model.
- Module system foundation: manifests, loader, example module.
- Validator referential integrity: key references, effect channel enforcement, duplicate-id
  checks. (Fixed dangling crafting refs and equipment id collisions found in the July review.)
- Cross-references migrated to namespaced `core:<category>/<key>` strings; bare keys rejected.
  (#38, #11)
