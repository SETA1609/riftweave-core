# Stories and issues

The [roadmap](../../ROADMAP.md) broken into **stories** (a milestone that delivers something
you can check) and **issues** (one pull request each). The repo is the source of truth for
both. Linear mirrors them and links back; it does not own scope.

Stories near the top (M1, M2) have detailed issues. Later stories are outlines. Refine a
story's issues when it is next in line, not before; the design will have moved by then.

## Files

One file per milestone. Issues inside a file are in the order they should be done.

| File | Story |
| --- | --- |
| [01-kernel.md](01-kernel.md) | M1 Kernel closed |
| [02-characters.md](02-characters.md) | M2 Characters |
| [03-combat-magic.md](03-combat-magic.md) | M3 Combat and magic playable |
| [04-world.md](04-world.md) | M4 World |
| [05-release.md](05-release.md) | M5 Release 1.0, plus repo process and tooling |
| [06-modules.md](06-modules.md) | Post-1.0 modules |
| [07-backend.md](07-backend.md) | APP Reference backend: CLIs, API server, Angular frontend |
| [done.md](done.md) | Issues whose PR is merged, with the PR number and merge date (M0 is here) |

## Order

Focus: M1

The `Focus:` line is the owner's order. Everything else about order is computed from each
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
- `$DT roadmap-order --settings-file tools/devops/pyproject.toml write` regenerates the story
  table in [ROADMAP.md](../../ROADMAP.md#story-order). A PR that changes a `Status` or
  `Needs` runs it; CI fails while the table is stale.
- Two ready issues that list the same paths under `Files:` will conflict; run them one after
  the other, or add a `Needs` edge.

## IDs

- **Story ID:** the milestone in the story heading: `M1` to `M5`, `POST` for post-1.0
  modules, and `APP` for the reference backend.
- **Issue ID:** story ID plus a two-digit number: `M1-01`, `POST-03`. Next id = highest
  number in that story (open or in [done.md](done.md)), plus one. Numbers are never reused;
  a dropped issue stays in its file marked `dropped` with a one-line reason.
- **Older ids:** `SET-NN` (tracked in Linear) and `RW-NN` (this tracker until M5-01) keep
  their ids in [done.md](done.md). An open issue that was renumbered says so in `Was:`.
- **Branch:** `<issue-id>-<slug>` for human work, `agent/<agent-name>/<issue-id>-<slug>` for
  agent work, for example `M1-07-one-test-runs-every-vector-file`. The slug is the issue title
  in lowercase kebab-case, shortened if long. No branch without an issue; if work has none,
  add one to the story file first.
- **PR title:** `<issue-id> <issue title>`.

## Workflow

The five steps (spec, schema, vector, data, review) and the pull request rules are in
[CONTRIBUTING.md](../../CONTRIBUTING.md#workflow). In short:

1. **File:** add the issue to its story file, `Status: todo`. Optionally create the Linear
   issue with a link to the story anchor and add `Linear:` to the issue.
2. **Work:** set `Status: doing` in the issue's first PR commit.
3. **Review:** set `Status: review` when the PR is opened. Add the issue's changelog
   fragment, `changes/<issue-id>.md` ([rules](../../changes/README.md)), in the same PR. Do
   not edit [CHANGELOG.md](../../CHANGELOG.md); the owner releases the fragments into it.
4. **Move done issues:** after the PR is merged, the next PR that touches that story file
   moves the issue to [done.md](done.md) with its full text, `Status: done`, and
   `Merged: #<PR>, <date>`, and adds its id to the `Done:` line under the story heading.
   Story files list only open work.

Design questions that need a spec go to `docs/core/` or `docs/modules/` before the issue
starts. Decisions go to [docs/decisions.md](../decisions.md).

## Story and issue format

```
## <Milestone> <Title>
**Goal:** <what exists afterwards>
**Spec:** <where the rules are specified>
**Done when:** <checkable result for the whole story>

Done: <ids moved to done.md>
```

```
#### <Issue ID> <Title, imperative>
- Status: todo            (todo | doing | review | done | dropped)
- Priority: medium        (high | medium-high | medium | low)
- Needs: <issue ids that must be merged first, or —>
- Was: <old id, only for renumbered issues>
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
