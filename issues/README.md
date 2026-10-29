# Issues

The tracker lives in the repo so it versions with the data it describes. The repo is the
source of truth for issues, the [roadmap](../ROADMAP.md), and the [changelog](../CHANGELOG.md).
Linear mirrors them and links back; it does not own scope.

```text
issues/
  open/   RW-NN-slug.md   live issues (status: open | in-progress)
  done/   *.md            archive (status: done | wontfix)
  TEMPLATE.md
```

One markdown file per issue with front matter (see [TEMPLATE.md](TEMPLATE.md)). The directory
is the status: close an issue by moving it, never by deleting it.

## Ids

- `RW-NN`: issues created in this repo. Next id = highest `RW-` number in `open/` and `done/`,
  plus one. Never reused.
- `SET-NN`: issues that were tracked in Linear before this tracker. They are archived in `done/`
  and keep their ids.

The two prefixes never collide, so Linear's own numbering is free to continue.

## Workflow

1. **File:** copy `TEMPLATE.md` to `open/RW-<next>-<slug>.md`. Set a `milestone` from the
   roadmap.
2. **Mirror (optional):** create the Linear issue, put the repo file link in its description,
   and set `linear:` in the front matter.
3. **Work:** `status: in-progress`. Put the id in the branch, commits, and PR title
   (`RW-12 combat pipeline vectors (#41)`).
4. **Close:** in the same PR, `git mv` the file to `done/`, set `status: done`,
   `closed: YYYY-MM-DD`, `pr: "#NN"`, add one line to [CHANGELOG.md](../CHANGELOG.md), and delete its
   row from the roadmap table. Use `status: wontfix` with a one-line reason to drop an issue.

Design questions that need a spec go to `docs/modules/` or `docs/core/` before the issue starts.

## Notes

- `done/` holds SET-6 to SET-25, recovered from git history, plus RW-1 and RW-2. SET-1 to
  SET-5, SET-23, and SET-24 were not found in the repo and have no files. `closed` on imported
  issues is the merge date of the landing commit.
- `../issues.md` is the older narrative backlog, kept for context. Its items are now in `open/`.
- `validate.py` only walks `ruleset/data/`, so nothing under `issues/` is validated.
