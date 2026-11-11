# Issues

One file per issue, named after its id, so two pull requests never edit the same issue list.
An issue's story (its goal and spec) is in [docs/stories/](../stories/README.md).

| Directory | Holds |
| --- | --- |
| [open/](open/) | Issues not merged yet, whatever their `Status` (`todo`, `doing`, `review`, `dropped`) |
| [done/](done/) | Merged issues, each with its `Merged: #<PR>, <date>` line |

## How an issue closes

A pull request sets `Status: review` in its issue's file when it finishes the issue. It does not
move the file, add the `Merged:` line, or touch the generated [order.md](order.md).

After the merge, [close-merged-issue.yml](../../.github/workflows/close-merged-issue.yml) finds the
issue from the branch name, moves it to `done/` when its status is `review`, regenerates the table in
[order.md](order.md), and commits that to `main`. A pull request that
finishes only part of an issue leaves its status at `doing` or `todo`, and the issue stays open.

## Older issues

M0 issues (`SET-NN`) were tracked in Linear before this repo had a tracker and were recovered from
git history; their merge date is the landing commit's. SET-1 to SET-5, SET-23, and SET-24 were not
found in the repo and have no file.
