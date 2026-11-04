# Decisions, risks, and open questions

A dated log of process and design decisions that are not obvious from the data. Newest
first. A decision that changes a rule also updates the doc for that rule; this file only
records that it was decided, when, and why. Open questions move to **Decided** once an
issue settles them.

## Decided

### 2026-11-04: A reference backend is in scope (M5-12)

- **Decision (owner):** the repo gets one backend that serves the ruleset: a Go implementation
  of the kernel behind one command contract, with an agent CLI, a human CLI, an HTTP API,
  and an Angular frontend as thin adapters (story APP, `docs/stories/07-backend.md`). Game
  engines and VTT code stay out of scope. It is not a 1.0 gate.
- **Why:** the ruleset could only be called through the vector test. A backend lets agents and
  people play it, find broken builds, and turn any run into a vector, and gives consumers an
  API.
- **Go (owner, 2026-11-04):** the backend is written in Go for speed where it matters: a
  static binary that starts in milliseconds for agent CLI loops, and bulk simulation of many
  fights. Per call, the rule work is microseconds in any language; JSON and HTTP dominate
  **[U]**: estimates, not measured. The cost is a second implementation of every rule, so:
  - APP-01 defines a language-neutral command contract (`backend/contract/`, JSON Schemas)
    that Go, both CLIs, the OpenAPI document, and the frontend are built from;
  - the Go core (APP-02) starts after M1 closes, so the kernel stops changing before it is
    written twice, and CI runs every vector against both Python and Go;
  - the Python reference stays the reference until 1.0; whether Go replaces it is decided
    then.
- **Defaults [A]:** Go module in `backend/`, Angular in `frontend/` calling only the API
  through a client generated from the OpenAPI document; rolls are inputs or seeded and echoed.
  Backend and frontend code are Apache-2.0, as tooling ([LICENSING.md](../LICENSING.md)).
- **Risk:** a backend invites rules in adapters. Rule: an adapter only parses, calls the
  core, and formats; anything else is a bug.

### 2026-11-02: Economy rows and one price formula (M4-01)

- **Decision:** `ruleset/data/economy/core.json` holds the denominations (copper is the base
  unit), the barter check, and two merchant rows with stock. `ruleset/data/formulas/price.json`
  is the one price tree. A purchase and a sale differ only in which merchant percent they read
  and which way the barter shift goes. Spec: `docs/core/economy.md`.
- **Why:** equipment had `cost` and materials `valueFactor`, but no rule turned them into a
  price or said what a merchant pays.
- **Open:** every number is a first choice **[U]**: the 1/10/100/1000 ratios, 125 and 40
  percent for the general store, the barter divisor and cap. `value_percent` is
  `valueFactor * 100` supplied by the caller, not looked up **[A]**. Prices round down, so a
  1 cp item sells for 0 **[A]**.
- **Risk:** `validate.py` does not check merchant stock refs; `test_economy.py` does, until
  [M5-08](stories/05-release.md#m5-08-integrity-gates) covers every category. The issue's
  `See: issues.md section 1` points at a file that is not in the repo.

### 2026-11-02: Rest is rows per occasion (M1-04)

- **Decision:** `ruleset/data/rest/core.json` holds one row per occasion (`short_rest`,
  `long_rest`, `end_of_combat`) with a recovery rule for each of `hp`, `mp`, `stamina`, `ap`,
  and `conditions`. A resource that does not recover says `none`. The `rest` vector runner
  reads the rows and the derived maxima. Spec: `docs/core/rest.md`.
- **Why:** the only recorded rule was "a short rest refills AP". Prose recovery cannot be
  pinned by a vector, and a second implementation would invent its own numbers.
- **Open:** the fractions (a quarter of hp and mp on a short rest) are a first choice **[U]**.
  Stamina has no derived row, so the vector supplies its maximum **[A]**. `pool.rest` in
  `actions/core.json` still says `refill`; it agrees with the rest row for `ap` but is a second
  place to change it.

### 2026-11-02: One physical attack pipeline (M1-02)

- **Decision:** `docs/core/combat.md` fixes the order attack roll, one defender reaction
  (evade or block), damage, crit, block reduction, armor DR, resist, HP.
  `ruleset/data/combat/pipeline.json` holds the crit damage and pierce trees and the
  damage-type tags the steps read. `combat.json` vectors pin hit, miss, crit, fumble, evade,
  block, DR, resist, and damage-type interaction. The reference is `run_combat` in
  `kernel_reference.py`.
- **Why:** the order lived only in `docs/modules/combat.md` prose, with worked examples that
  disagree on small points (for example whether a crit multiplies before DR).
- **Open:** these rules are choices the prose left open **[A]**: crit multiplies before any
  reduction; a piercing hit halves DR; evade and block apply to physical types only. The
  magic defense tree, called shots, cover, hit quality, and what happens at 0 HP are not
  specified **[U]**.
- **Evasion after the attack (owner, 2026-11-04):** evasion is a defender reaction rolled
  after a hit, contested on margin like block, not a passive penalty on the attacker's
  target. A defender takes one reaction per attack, declared before the attack roll. The
  evade target is the raw `evasion` skill, as block reads the block skill; the old tree
  `10 + evasion_skill * 0.15 + AGI * 2` is gone, because Evasion skill points barely moved
  it and it topped out at 45.
- **Delivery decides the dodge (owner, 2026-11-04):** not the damage type. A `single`
  attack or spell (sword, arrow, fire bolt, touch) is evaded whole; an evade halves an `area`
  attack (fireball, breath, bomb), rounded down, after a crit; an `instant` spell cannot be
  evaded or blocked, and its target makes a save on a skill or attribute the spell names.
  Block stays physical-only. Spells do not carry `delivery` yet, and saves are unspecified:
  M3-05. **[V 2026-11-04]** vectors pin the three deliveries.
- **Reaction rolls (owner, 2026-11-04):** evade and block roll through the one d100
  function with the defender's skill and LCK: a 100 fumbles and fails, and a roll in the
  defender's crit window is tagged `crit`. A crit reaction has no effect yet; later perks
  may read the tag. **[V 2026-11-04]** vectors pin both.
- **Balance (owner, 2026-11-04):** at equal skill an evade avoids all damage and a block
  only subtracts shield DR. That is balanced outside the pipeline: perks gate stronger dodges
  and blocks, armor penalties land on evasion, area attacks are only halved, and instant
  spells skip the dodge for a save **[A]**. The perk ladders are M2-06: block against magic
  at half shield DR, then full; block-side resistance scales with the block, not a general
  resist; instant spells stay unblockable. Armor evasion penalties are not in the vectors
  **[U]**.
- **Risk:** `base_resolution.json` still holds `defense.*` and `critDamageMultiplier`.
  The evasion keys (`baseEvasion`, `evasionSkillFactor`, `agiEvasionFactor`) are now read by
  nothing; `critDamageMultiplier` is restated as a literal in the crit tree, so changing one
  drifts from the other. Removing keys is breaking, so it is a follow-up issue.
  `combat_reference.py` is the older crafted-item simulator and is not the vector reference.

### 2026-11-02: Integrity gates (M5-08)

- **Decision:** `validate.py` gains two checks: every `core:<category>/<key>` string in any
  data or module file must resolve (it walks all strings, so no field escapes), and every
  core registry effect is referenced from another file or marked `"reserved": true`. A
  reserved effect that is referenced also fails. Every schema property that holds a
  reference uses `schema.json#/definitions/namespacedRef`, a pattern. Spec:
  `docs/core/integrity.md`; tests: `ruleset/scripts/test_integrity.py`.
- **Why:** the data review found 25 unreferenced effects (M6) and free-string branches
  that let typos through (M3). 161 references resolved when the walk was first run, so the
  gate went in green.
- **Open:** the 26 unreferenced effects are marked reserved, not wired into content. The
  review recommended wiring the two poison composites into the poison consumables **[U]**;
  that changes content and is left to a content issue. The `oneOf`s that choose a structure
  (equipment, monster) stay. A reference from another effect (`sub_effects`) does not count
  as use **[A]**. `module:` references are not looked up **[A]**.
- **Risk:** the crafting schemas described `base_key`, `material_key`, `gem_key`, and a
  recipe's `skill` as "keys", but the data and the validator already used namespaced refs;
  the descriptions were wrong, the data was right.

### 2026-11-02: Progression is data and formula trees (M1-03)

- **Decision:** `ruleset/data/progression/core.json` holds the level and skill caps, the XP
  and skill-point formulas as trees, the tag-skill cost ratio, and the perk cadence. The
  formula AST gains a `var` leaf for named inputs (`level`, `luck_roll`), and `schema.json`
  gains `formulaNode`. The luck roll is an input, so vectors pin it. Spec:
  `docs/core/progression.md`.
- **Why:** the numbers were prose in `docs/modules/progression.md`, which a second
  implementation cannot read.
- **Open:** a character gets a skill-point pool at level 1 as well as at each level-up
  (30 pools by level 30) **[A]**. Points that cannot buy a whole rating stay in the pool
  **[A]**. `CLAUDE.md` says tag skills get "+2 per point instead of +1"; the module doc says
  2:1 cost. Same ratio, so the data uses cost. Breakthrough bonuses and stamina growth are
  not in the data.

### 2026-11-02: The rebase keeps both sides of decision-log additions (M5-10)

- `docs/decisions.md` is in `keep-both-added-lines`: when two open PRs each add a section at
  the top, the rebase keeps both, the branch's first, like the changelog and story files.
  Before this, the rebase workflow failed on every push to `main` since 2026-11-01 and left
  each such branch for a person. Reproduced locally with `rebase-open-branches --no-push`;
  `docs/decisions.md` was the only conflicting file. **[V 2026-11-02]**
- new-world never hit this: it has no shared append-only log; decisions go into the topic
  doc they change. Kept the single log here because it is short and newest-first. **[A]**
- devops-tools, the rebase workflow, and the fragments workflow move to `v0.2.0`
  (`15c1bf1`), the tag new-world uses; the code is the same as `fde087b` apart from the
  version bump.
- Risk: two PRs that edit the same existing lines of the log still need a person.

### 2026-11-02: Changelog fragments (M5-02)

- Each PR adds `changes/<issue-id>.md` instead of a `CHANGELOG.md` line, as in new-world. Every
  PR used to add a line at the same places in the changelog, so parallel PRs conflicted there;
  the rebase workflow only resolves pure additions. Rules: [changes/README.md](../changes/README.md).
- CI runs `devops-tools changelog-fragments check --branch` on every PR through the reusable
  workflow at the devops-tools pin (`fde087b`, unchanged; it already has the command and the
  workflow). A branch whose issue has no fragment fails.
- The owner releases the fragments into `CHANGELOG.md` with `changelog-fragments release` in
  a PR of its own. The release writes `### <issue-id> <title> (#N)` entries under a new
  `## <title>` section above `[Unreleased]`; the categorized `[Unreleased]` list holds the
  entries written before M5-02 and is folded into the first release by hand. **[A]**
- Risk: a release PR deletes every fragment, its own included, so it adds its own fragment
  after `release` or its branch check fails.

### 2026-11-02: Racial traits are features (M1-06)

- `features` is the only table for racial traits. The 18 rows of `data/traits/core.json`
  move to `data/features/core.json` as `type: racial_trait`, keeping their keys;
  `traits/core.json` and `trait.schema.json` are removed. Races keep their `traits` field,
  which now holds `core:feature/<key>` refs.
- Why: `AGENTS.md` names `features` the single perk table, `feature.schema.json` already had
  the `racial_trait` type, and the traits table repeated the row shape without `effects`.
  One table means one ref category and one effect shape for everything a character has.
- Rejected: keeping `traits` as its own table and dropping `racial_trait`. It would make
  the docs match the data, but keep two row shapes and two ref categories for one concept.
- Breaking: `core:trait/<key>` refs no longer resolve; use `core:feature/<key>`. The race
  field name `traits` is unchanged, so no data key is renamed.
- `validate.py` checks that a race's `traits` point at `racial_trait` features and that
  every core `racial_trait` is used by a race.

### 2026-11-02: Every skill has a consumer; checks are rows (M2-02)

- A skill needs a governing attribute and at least one consumer: a weapon, an armor
  category, a spell or effect color, a check, or a perk ([checks.md](core/checks.md)).
  Background bonuses do not count. `ruleset/scripts/skill_audit.py` enforces it.
- Checks get their own collection, `ruleset/data/checks/core.json`. There was no row a
  non-combat skill could be rolled through, so 18 of the 44 skills had no consumer.
- No skill is removed. Every orphan was a skill the design docs already use (TTRPG
  social, knowledge, and survival skills, the crafting trades, the nine color schools), so
  each gets a check row, and `orange_magic` gets its first effect. Removing them would
  have been a breaking key change for no gain.
- The audit is a separate script, not part of `validate.py`, so it can land without
  touching the validator while M1-06 changes it. Folding it into `validate.py` is M5-08.
### 2026-11-02: One vector test and a kernel checklist (M1-07)

- `ruleset/scripts/test_vectors.py` globs `ruleset/tests/vectors/*.json`; a vector's function
  is its `function` field or the file name. The reference runners live in
  `ruleset/scripts/kernel_reference.py`, one per kernel function.
- The checklist of kernel functions is the table in `docs/core/vectors.md`, and the test reads
  it, so the doc and the test cannot disagree. A row is `covered` or `pending <issue>`.
- Progression stays `pending M1-03`: it has no data yet, and writing its vectors here would do
  M1-03's work against prose that M1-03 replaces. Risk: the M1-07 "no function without a
  vector" check holds for five of six functions until M1-03 merges.
- The modifier vectors cover only `add`, `mul`, and same-source duplicates. The order of
  `set`, `grant`, and `resist`, and whether two `mul` from different sources multiply or add,
  is open for M1-05; the reference multiplies [A].

### 2026-11-02: Factions and reputation module (POST-05)

- Reputation is one signed integer per faction on the actor (`reputation[faction ref]`),
  bounded by the faction row. It is primary state, like XP. Standing, promotion
  eligibility, and the emitted modifiers are recomputed from it, never stored.
- Rank is held state written by a consumer event (join, promote, expel), not computed from
  reputation, so quests can gate promotion. A held rank below its `min_reputation` is
  suspended: it emits nothing until reputation recovers.
- Every standing of a faction shares the source `faction_standing:<key>`, every rank
  `faction_rank:<key>`. The kernel's same-source rule then replaces the old bundle when the
  standing changes; no second stacking rule. The schema enforces the prefix.
- Spillover to related factions is one hop, `trunc(delta * share_percent / 100)` toward
  zero, from the unclamped delta. One hop keeps it loop-free.
- No price or barter modifiers until the economy issue (M4) defines a target for them.

### 2026-11-01: Story-scoped issue IDs and a computed order (M5-01)

- Open issues are renumbered per story (`M1-01`, `POST-03`), as in new-world. Owner decision;
  it reverses the RW-40 line "Issue ids stay `RW-NN`". Done `RW-` and `SET-` ids stay in
  `done.md`, and each renumbered issue keeps its old id in `Was:`.
- Order is computed by `devops-tools roadmap-order` from `Status`, `Needs`, and the `Focus:`
  line, so parallel waves and the critical path are never maintained by hand. CI fails when
  the story table in `ROADMAP.md` is stale.
- The rebase workflow keeps both sides of a story-file conflict where both only added lines.
  Status edits are changes, not additions, and still go to a person.
- Changelog fragments (M5-02) and the container runner move (M5-03) are their own issues.

### 2026-11-01: Shared devops-tools (RW-41)

- Generic devops code (run logs, log pruning, open-PR rebasing, changelog fragments) lives in
  [SETA1609/devops_tools](https://github.com/SETA1609/devops_tools) and is shared by several
  repos. This repo depends on it through `tools/devops/pyproject.toml`, pinned to a full commit
  SHA, and keeps only Riftweave-specific code (the containers, cache cleanup).
- The pin moves only in its own PR, after its tests pass here. The rebase workflow calls the
  shared reusable workflow at the same SHA as `tools/devops/pyproject.toml`; the two move together.
- Not adopted yet: changelog fragments (`changes/<id>.md`) and `keep-both-added-lines` for
  `docs/stories/*.md`. Either changes the RW-39 and RW-40 process, so each needs its own issue
  (adopted in M5-01 and M5-02).

### 2026-10-31: Story files (RW-40)

- Issues live in one story file per milestone in `docs/stories/`, with a `done.md`
  archive, as in project-new-world. It replaces one file per issue (RW-2). Owner decision:
  the story's goal and the issue order belong next to the issues. Accepted cost: two open
  PRs in the same milestone can conflict on the story file; the rebase workflow leaves those
  for a person.
- Issue ids stay `RW-NN` and `SET-NN`; they are not renumbered per story.

### 2026-10-31: Process (RW-39)

- One issue, one PR, at most 20 counted changed files, five steps (spec, schema, vector,
  data, review), one commit per step ([CONTRIBUTING.md](../CONTRIBUTING.md#workflow)).
- Design docs tag claims **[V]**, **[A]**, or **[U]**
  ([CONTRIBUTING.md](../CONTRIBUTING.md#verification-tags)).
- Open pull requests are rebased onto `main` after every push to `main`, keeping both
  sides of an add-only `CHANGELOG.md` conflict.
- GitHub Actions are pinned to a commit SHA with the tag in a comment.

### 2026-10-29: Tracking (RW-2)

- The repo owns issues, the roadmap, and the changelog. Linear mirrors them.
- The changelog has one entry per closed issue, not per commit.

## Open questions

- **Opposed checks.** A check row may name an `opposed_by` skill, but how two results
  compare (margin against margin, ties) is not specified
  ([checks.md](core/checks.md)). Decide with the casting and combat contest rules in M3.

- **Faction cross-row rules in the validator.** Band order, the source key matching the
  faction, and relation targets are checked by `faction_reference.py`, not `validate.py`
  (POST-05). Move them into the integrity gates
  ([M5-08](stories/05-release.md#m5-08-integrity-gates)) or a module-aware check.
- **Module ref spelling.** `docs/core/keys.md` writes a module key as `mod:copper_fiber`;
  `AGENTS.md` and `validate.py` use `module:<module>/<category>/<key>`. POST-05 uses the
  latter (`module:factions/faction/lantern_wardens`). One of the two docs should change.
- **Modifier `duration` vocabulary.** The contract lists `turns | uses | while_equipped |
  permanent`; conditions and factions use `while_applied`. The schema leaves `duration` a
  free string.
- **Traits versus features.** Whether creation traits stay a `features` type or become
  their own table ([M1-06](stories/01-kernel.md#m1-06-resolve-the-traits-vs-features-fork)).
- **Casting scope.** What the core casting procedure includes and what is a module
  ([M3-01](stories/03-combat-magic.md#m3-01-casting-procedure-decide-scope-then-specify)).
- **Content floors.** The targets in [ROADMAP.md](../ROADMAP.md#definition-of-finished) are
  proposals **[A]**, to be changed only by a decision recorded here.
- **Schema property casing.** Older schemas use `camelCase`, newer ones `snake_case`.
  Unifying them is breaking; decide before 1.0, because after the key-stability policy
  ([M5-07](stories/05-release.md#m5-07-key-stability-and-versioning-policy)) it costs a major
  version.

## Risks

- **Two stacking implementations.** `faction_reference.py` carries a small
  `apply_modifiers` for its stacking vectors. When M1-05 lands the shared stacking
  reference, the factions reference should call it instead (POST-05).
- **Module vectors live in the module.** The factions vectors are in
  `ruleset/modules/factions/tests/vectors.json`, run by `test_faction_reference.py`, not in
  `ruleset/tests/vectors/`: the core runner (M1-07) checks only kernel functions against the
  checklist in `docs/core/vectors.md`. [V 2026-11-02] (POST-05)
- **Silent cross-reference breaks.** Refs are strings. `validate.py` checks most namespaced
  refs, but not every category yet
  ([M5-08](stories/05-release.md#m5-08-integrity-gates)).
- **Rebased branches are not revalidated by CI.** Pushes made with the workflow's
  `GITHUB_TOKEN` do not trigger other workflows, so a rebased PR's new head has no
  validation run. Run validation (or push a commit) before merging a rebased PR.
- **Docs outrunning data.** `docs/modules/` holds long design prose with few vectors. Until
  it is tagged, treat untagged claims there as **[A]**.
