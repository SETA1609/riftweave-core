# Stories and issues

The [roadmap](../../ROADMAP.md) broken into **stories** (a goal that delivers something
you can check) and **issues** (one role and one pull request each). The repo is the source of truth for
both. Linear mirrors them and links back; it does not own scope.

Open work lives in the story files `S00` to `S20`, each with a full issue table. The milestone
files below (`01-kernel.md` to `08-content.md`) keep the goals of the old milestones; they take
no new issues.

## Files

Open work is grouped in **stories** `S00` to `S20`, one file each: its goal, the table of its
issues (role, old id, `Needs`), and its done-when. The issues themselves are one file each in
[docs/issues/](../issues/README.md), and every issue has one role ([ROLES.md](../issues/ROLES.md)).

| File | Story |
| --- | --- |
| [S00-process-and-conventions.md](S00-process-and-conventions.md) | S00 Process and conventions |
| [S01-release-gates.md](S01-release-gates.md) | S01 Release 1.0 gates |
| [S02-data-integrity.md](S02-data-integrity.md) | S02 Normal forms and data integrity |
| [S03-diagrams.md](S03-diagrams.md) | S03 Diagrams |
| [S04-perk-model-and-builds.md](S04-perk-model-and-builds.md) | S04 Perk model and builds |
| [S05-melee-combat.md](S05-melee-combat.md) | S05 Melee combat |
| [S06-bound-weapons-summoners.md](S06-bound-weapons-summoners.md) | S06 Bound weapons and summoners |
| [S07-files-between-apps.md](S07-files-between-apps.md) | S07 Files and characters between apps |
| [S08-play-sessions.md](S08-play-sessions.md) | S08 Play sessions in both apps |
| [S09-custom-content-printing.md](S09-custom-content-printing.md) | S09 Custom content and printing |
| [S10-online-rooms.md](S10-online-rooms.md) | S10 Online rooms |
| [S11-localization.md](S11-localization.md) | S11 Localization |
| [S12-observability.md](S12-observability.md) | S12 Observability and dev tools |
| [S13-content-perks-spells.md](S13-content-perks-spells.md) | S13 Content parity: perks, effects, and spells |
| [S14-content-bestiary-table.md](S14-content-bestiary-table.md) | S14 Content parity: bestiary, items, and the table |
| [S15-optional-modules.md](S15-optional-modules.md) | S15 Optional modules |
| [S16-table-module.md](S16-table-module.md) | S16 Table module |
| [S17-ranged-combat.md](S17-ranged-combat.md) | S17 Ranged combat |
| [S18-app-environments.md](S18-app-environments.md) | S18 App environments |
| [S19-turn-economy.md](S19-turn-economy.md) | S19 Turn-based action economy |
| [S20-custom-monsters-and-templates.md](S20-custom-monsters-and-templates.md) | S20 Custom monsters and character templates |

The milestone files below hold the goals of the milestones whose done issues kept their old ids
(before 2026-11-18). They take no new issues; the [roadmap](../../ROADMAP.md) still names its
milestones by them.

| File | Milestone |
| --- | --- |
| [01-kernel.md](01-kernel.md) | M1 Kernel closed |
| [02-characters.md](02-characters.md) | M2 Characters |
| [03-combat-magic.md](03-combat-magic.md) | M3 Combat and magic playable |
| [04-world.md](04-world.md) | M4 World |
| [05-release.md](05-release.md) | M5 Release 1.0, plus repo process and tooling |
| [06-modules.md](06-modules.md) | Post-1.0 modules |
| [07-backend.md](07-backend.md) | APP Reference backend: CLIs, API server, Angular frontend |
| [08-content.md](08-content.md) | M6 Content parity: post-1.0 content and tabletop play |

## Order

Focus: S01, S02, S19, S17, S07

The `Focus:` line is the owner's order. S05 and S06 are done on `main`. S01 and S02 are the
highest-priority open work that can start; S19 and S17 finish the turn and ranged rules that S05
left open; S07 unblocks files between the apps. Everything else about order is computed from each
issue's `Status` and `Needs`, so it never goes stale by hand:

| Keyword | Meaning |
| --- | --- |
| **first** | Ready and in a focus story; do these before anything else, in focus order. |
| **ready** | Every `Needs` is merged (`done` or `review`); can start now. |
| **wave N** | How many open issues stand in front of it. Everything in one wave can run in parallel, on separate branches or worktrees (one agent each), once the earlier waves are merged. Wave 0 is ready. |
| **blocked by X** | Waits for the open issues (or stories) X. |
| **critical** | On the longest chain of open issues; a delay there delays everything after it. |

Run these from the repo root (`DT` is `uv run --project tools/devops --locked --no-dev devops-tools`):

- `$DT roadmap-order --settings-file tools/devops/pyproject.toml show` prints the issue view:
  first, ready, the critical path, and the blocked waves with their blockers. It changes
  nothing, so open PRs never conflict on it.
- [docs/issues/order.md](../issues/order.md) is a generated table.
  [close-merged-issue.yml](../../.github/workflows/close-merged-issue.yml) regenerates it after
  each merge; pull requests do not touch it.
- Two ready issues that list the same paths under `Files:` will conflict; run them one after
  the other, or add a `Needs` edge.

## IDs

Issues are named by **role**, **story**, and **number** (since PO_S00_01_issue-splitting-by-role,
2026-11-18):

```
<ROLE>_<STORY>_<ISSUE>_<short-description>
BE_S07_03_character-list-cli
```

| Part | Rule | Example |
| --- | --- | --- |
| `ROLE` | One of `PO`, `GD`, `DE`, `BE`, `FE`, `QA`, `OPS` ([ROLES.md](../issues/ROLES.md)) | `BE` |
| `STORY` | `S` plus two digits: the story file the issue belongs to | `S07` |
| `ISSUE` | Two digits, restarting at `01` in each story, whatever the role. Next id = highest number in that story (in [open/ or done/](../issues/README.md)), plus one. Numbers are never reused; a dropped issue keeps its file in `docs/issues/open`, marked `dropped` with a one-line reason | `03` |
| `short-description` | Lowercase kebab-case, at most five words, fixed when the issue is filed | `character-list-cli` |

The full id is the file name (`docs/issues/open/BE_S07_03_character-list-cli.md`), the heading,
the branch, the changelog fragment (`changes/<id>.md`), and the decision file
(`docs/decisions/<id>.md`).

- **Splitting:** one issue has one role and fits in one pull request
  ([rules](../../CONTRIBUTING.md#pull-request-rules)). Work that needs two roles is two issues in the
  same story, the later one listing the earlier under `Needs:`: a full-stack feature is a `BE` issue
  and an `FE` issue that needs it; a rule that needs Go code is a `GD` issue and a `BE` issue that
  needs it. A story is done when all of its issues are done.
- **New story:** the next free `S` number and a file `docs/stories/S<NN>-<slug>.md` with its goal,
  its issue table, and its done-when. A story groups the issues one checkable goal needs, whatever
  their roles.
- **Older ids:** done issues keep their ids: `M1` to `M6` (milestones), `POST` (modules), `APP`
  (reference backend), `SET-NN` (tracked in Linear), and `RW-NN` (this tracker until M5-01). An open
  issue that was renamed says so in `Was:`, and the table below resolves an old id in a changelog,
  decision, commit message, or doc.
- **Branch:** the full issue id, for people and agents alike (no `agent/<name>/` prefix), for example
  `BE_S07_03_character-list-cli`. An issue that still has an old id keeps the old form,
  `<issue-id>-<slug>`. No branch without an issue; if work has none, add its file to
  `docs/issues/open/` first.
- **PR title:** `<issue-id> <issue title>`.
- **Tooling:** `issue-id-pattern` in `tools/devops/pyproject.toml` accepts both forms. devops-tools
  still reads an issue's story from the text before its last `-`, so the story column of
  [order.md](../issues/order.md) and `Focus:` group the new ids only once devops-tools reads
  `_S<NN>_` ([decision](../decisions/PO_S00_01_issue-splitting-by-role.md)); waves, `ready`, and
  `blocked by` work per issue already.

### Old to new

| Old id | New id | How |
| --- | --- | --- |
| APP-17 | `BE_S10_01_room-log-supabase` | split by role: backend half |
| APP-17 | `FE_S10_02_room-log-client` | split by role: frontend half |
| APP-18 | `BE_S10_03_gm-approval-supabase` | split by role: backend half |
| APP-18 | `FE_S10_04_gm-approval-table` | split by role: frontend half |
| APP-22 | `BE_S09_01_spell-author-wasm` | split by role: backend half |
| APP-22 | `FE_S09_02_spell-editor-page` | split by role: frontend half |
| APP-24 | `FE_S11_01_frontend-i18next` | renamed |
| APP-28 | `BE_S11_02_cli-shared-locales` | split by role: backend half |
| APP-28 | `FE_S11_03_frontend-shared-locales` | split by role: frontend half |
| APP-29 | `BE_S08_05_bestiary-knowledge-cli` | split by role: backend half |
| APP-29 | `FE_S08_06_bestiary-page` | split by role: frontend half |
| APP-30 | `BE_S07_01_file-format-cli` | kept: finished as `APP-30`, merged #241, #242, 2026-11-18; the reserved new id was never filed |
| APP-30 | `FE_S07_02_file-open-frontend` | kept: finished as `APP-30`, merged #241, #242, 2026-11-18; the reserved new id was never filed |
| APP-31 | `BE_S08_01_player-session-cli` | split by role: backend half |
| APP-31 | `FE_S08_02_player-session-page` | split by role: frontend half |
| APP-32 | `BE_S08_03_item-lookup-craft-preview` | split by role: backend half |
| APP-32 | `FE_S08_04_item-lookup-page` | split by role: frontend half |
| APP-33 | `BE_S07_03_character-list-cli` | split by role: backend half |
| APP-33 | `FE_S07_04_character-list-page` | split by role: frontend half |
| APP-34 | `BE_S08_07_encounter-tracker-cli` | split by role: backend half |
| APP-34 | `FE_S08_08_encounter-tracker-page` | split by role: frontend half |
| APP-35 | `BE_S10_05_battle-grid-steps` | split by role: backend half |
| APP-35 | `FE_S10_06_battle-grid-page` | split by role: frontend half |
| APP-38 | `BE_S07_05_character-from-build-cli` | split by role: backend half |
| APP-38 | `FE_S07_06_character-from-build-page` | split by role: frontend half |
| APP-39 | `BE_S09_03_print-character-cli` | split by role: backend half |
| APP-39 | `FE_S09_04_print-character-page` | split by role: frontend half |
| APP-40 | `BE_S09_05_spell-cards-print` | split by role: backend half |
| APP-40 | `FE_S09_06_spell-cards-page` | split by role: frontend half |
| APP-41 | `BE_S09_07_custom-monster-print` | split by role: backend half |
| APP-41 | `FE_S09_08_custom-monster-page` | split by role: frontend half |
| APP-43 | `BE_S12_01_telemetry-format-setting` | kept: finished as `APP-43`, merged #239, 2026-11-18; the reserved new id was never filed |
| APP-44 | `BE_S12_02_server-observability` | renamed |
| APP-45 | `BE_S12_03_cli-observability` | renamed |
| APP-46 | `FE_S12_04_frontend-observability` | renamed |
| APP-47 | `OPS_S12_05_local-telemetry-stack` | renamed |
| APP-48 | `FE_S12_06_browser-dev-console` | kept: finished as `APP-48`, merged #240, 2026-11-18; the reserved new id was never filed |
| M3-09 | `GD_S06_01_bound-weapon-materials` | renamed |
| M3-21 | `GD_S06_02_bonded-summoner-perks` | renamed |
| M3-22 | `GD_S06_03_commander-summoner-perks` | renamed |
| M3-23 | `GD_S06_04_hybrid-summoner-perks` | renamed |
| M3-28 | `GD_S05_01_loadout-attacks-melee-ladders` | renamed |
| M3-29 | `GD_S05_02_fighting-stances` | renamed |
| M3-30 | `GD_S05_03_defense-perk-ladders` | renamed |
| M3-31 | `GD_S05_04_trip-and-disarm` | renamed |
| M5-09 | `PO_S01_01_licensing-audit-tag-v1` | renamed |
| M5-18 | `QA_S01_02_alpha-playtest-human-cli` | renamed |
| M5-19 | `GD_S01_03_perk-throughput-veto` | renamed |
| M5-40 | `DE_S02_01_one-property-casing` | renamed |
| M5-50 | `GD_S01_04_skill-damage-pacing` | renamed |
| M5-52 | `OPS_S00_02_format-data-files` | renamed |
| M5-53 | `OPS_S00_03_doc-last-changed-stamp` | renamed |
| M5-58 | `DE_S04_01_base-perk-schema` | renamed |
| M5-59 | `DE_S04_02_perks-compose-from-base` | renamed |
| M5-61 | `GD_S04_04_ranged-builds` | kept: finished as `M5-61`, merged #237, 2026-11-18; the reserved new id was never filed |
| M5-65 | `DE_S03_01_data-model-diagram-areas` | renamed |
| M5-66 | `OPS_S03_02_diagrams-as-svg` | renamed |
| M5-71 | `DE_S02_02_second-normal-form` | renamed |
| M5-74 | `DE_S02_03_monster-hp-computed` | renamed |
| M5-75 | `DE_S02_04_descriptions-without-numbers` | renamed |
| M5-76 | `M5-76` | kept: finished, merged #217, 2026-11-17 |
| M5-77 | `M5-77` | kept: dropped: superseded by PO_S00_01_issue-splitting-by-role |
| M5-78 | `GD_S04_03_perk-skill-bonuses` | renamed |
| M6-02 | `GD_S13_01_combat-stealth-perks` | renamed |
| M6-03 | `GD_S13_02_magic-social-perks` | renamed |
| M6-04 | `GD_S13_03_effect-pool-150` | renamed |
| M6-05 | `GD_S13_04_spells-red-to-blue` | renamed |
| M6-06 | `GD_S13_05_spells-indigo-to-black` | renamed |
| M6-07 | `GD_S14_01_bestiary-levels-1-15` | renamed |
| M6-08 | `GD_S14_02_bestiary-levels-16-30` | renamed |
| M6-09 | `GD_S14_03_named-magic-items` | renamed |
| M6-10 | `GD_S14_04_gm-encounter-travel-tools` | renamed |
| M6-11 | `GD_S13_06_d100-bonus-multiples-of-5` | renamed |
| M6-12 | `GD_S14_05_prebuilt-characters` | renamed |
| M6-13 | `QA_S14_06_tabletop-playtest` | renamed |
| POST-01 | `GD_S15_01_crafting-module` | renamed |
| POST-02 | `GD_S15_02_alchemy-module` | renamed |
| POST-03 | `GD_S15_03_enchanting-module` | renamed |
| POST-04 | `GD_S15_04_breakthrough-advancement-module` | renamed |
| POST-06 | `GD_S15_05_companions-module` | renamed |
| POST-07 | `GD_S15_06_mounts-vehicles-ships-module` | renamed |
| POST-08 | `GD_S15_07_survival-needs-module` | renamed |
| POST-09 | `GD_S15_08_weapon-advantage-module` | renamed |
| POST-10 | `GD_S16_01_optional-defence-roll` | renamed |
| POST-11 | `GD_S16_02_result-bands-crit-doubles` | renamed |
| POST-12 | `GD_S16_03_milestone-levelling-turns` | renamed |

## Workflow

The five steps (spec, schema, vector, data, review) and the pull request rules are in
[CONTRIBUTING.md](../../CONTRIBUTING.md#workflow). In short:

1. **File:** pick the role ([ROLES.md](../issues/ROLES.md)) and the story, add `docs/issues/open/<issue-id>.md`, `Status: todo`, and add its row to the story's issue table. Optionally create the Linear
   issue with a link to the story anchor and add `Linear:` to the issue.
2. **Work:** set `Status: doing` in the issue's file, in its first PR commit.
3. **Review:** set `Status: review` when the PR is opened. Add the issue's changelog
   fragment, `changes/<issue-id>.md` ([rules](../../changes/README.md)), in the same PR. Do
   not edit [CHANGELOG.md](../../CHANGELOG.md); the owner releases the fragments into it.
4. **Merged issues close themselves:** after the PR is merged, the
   [close-merged-issue](../issues/README.md#how-an-issue-closes) workflow moves an issue in
   `review` to `docs/issues/done/` with `Status: done` and `Merged: #<PR>, <date>`. A PR
   never moves an issue file itself. A PR that finishes only part of an issue leaves its
   status at `doing` or `todo`.

Design questions that need a spec go to `docs/core/` or `docs/modules/` before the issue
starts. Decisions go to `docs/decisions/<issue-id>.md`, one file per issue ([format](../decisions/README.md)).

## Story and issue format

```
## <Story ID> <Title>
**Goal:** <what exists afterwards>
**Issues:** <table: issue id, role, title, was, needs>
**Done when:** every issue above is done, and <checkable result for the whole story>
```

An issue is its own file, `docs/issues/open/<issue-id>.md`:

```
#### <ROLE>_<STORY>_<ISSUE>_<short-description> <Title, imperative>
- Status: todo            (todo | doing | review | done | dropped)
- Priority: medium        (high | medium-high | medium | low)
- Needs: <issue ids that must be merged first, or —>
- Was: <old id, only for renamed or split issues>
- Do: <what changes and why>
- Keys: <main keys, collections, or schema properties introduced, if any>
- Files: ~<count> (<main paths>)
- See: <related docs and data>
- Done when: <checkable result: validate, a vector, a test, a count>
```

`Keys` and `Files` are proposals; change them if the work shows something better. `Needs`
and `Files` are what make parallel work safe: `Needs` decides the waves, and `Files` shows
which ready issues would touch the same paths.

`validate.py` only walks `ruleset/data/`, so nothing here is validated.
