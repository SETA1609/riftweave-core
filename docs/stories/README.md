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
| [done.md](done.md) | Issues whose PR is merged, with the PR number and merge date (M0 is here) |

## IDs

- **Issue ID:** `RW-NN`. Next id = highest `RW-` number in any file here, plus one. Numbers
  are never reused; a dropped issue stays in its file marked `dropped` with a one-line
  reason.
- **`SET-NN`** issues were tracked in Linear before this tracker. They are archived in
  [done.md](done.md) and keep their ids. The two prefixes never collide, so Linear's own
  numbering is free to continue.
- **Branch:** `RW-NN-<slug>` for human work, `agent/<agent-name>/RW-NN-<slug>` for agent
  work. The slug is the issue title in lowercase kebab-case, shortened if long. No branch
  without an issue; if work has none, add one to the story file first.
- **PR title:** `RW-NN <issue title>`.

## Workflow

The five steps (spec, schema, vector, data, review) and the pull request rules are in
[CONTRIBUTING.md](../../CONTRIBUTING.md#workflow). In short:

1. **File:** add the issue to its story file, `Status: todo`. Optionally create the Linear
   issue with a link to the story anchor and add `Linear:` to the issue.
2. **Work:** set `Status: doing` in the issue's first PR commit.
3. **Review:** set `Status: review` when the PR is opened. Add the issue's line to
   [CHANGELOG.md](../../CHANGELOG.md) in the same PR.
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
#### RW-NN <Title, imperative>
- Status: todo            (todo | doing | review | done | dropped)
- Priority: medium        (high | medium-high | medium | low)
- Needs: <issue ids that must be merged first, or —>
- Do: <what changes and why>
- Keys: <main keys, collections, or schema properties introduced, if any>
- Files: ~<count> (<main paths>)
- See: <related docs and data>
- Done when: <checkable result: validate, a vector, a test, a count>
```

`Keys` and `Files` are proposals; change them if the work shows something better. Issues
moved from the old one-file-per-issue tracker have no `Keys` or `Files` yet; add them when
the issue is next in line.

`validate.py` only walks `ruleset/data/`, so nothing here is validated.
