# Changelog fragments

Each pull request adds one file here, `changes/<issue-id>.md`, instead of editing
[CHANGELOG.md](../CHANGELOG.md). Parallel pull requests then never conflict on the
changelog. The owner moves the fragments into the changelog at a release.

## Format

```markdown
### M5-02 Changelog fragments instead of CHANGELOG edits
- Changed: each pull request adds `changes/<issue-id>.md`; the owner releases them into `CHANGELOG.md`.
```

- **File name:** the issue id plus `.md`, spelled as in the story file: `M1-07.md`,
  `POST-05.md`. One fragment per issue. An issue split into a ruleset and a backend pull
  request has one fragment: the ruleset one adds it, the backend one adds a bullet to it.
- **First line:** `### <issue-id> <issue title>`, with the same id as the file name. No
  PR number; the release adds ` (#N)` from the squash-merge commit subject.
- **Then at least one `- ` bullet.** Start each bullet with its
  [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) category: `Added:`, `Changed:`,
  `Deprecated:`, `Removed:`, `Fixed:`, or `Security:`. A renamed or removed key breaks
  consumers, so add **(breaking)** to that bullet.
- **No other headings**, no `Status:` line (status lives in the story file), and no
  `Last updated:` line (a release sets it).
- End the file with a newline.
- Work with no issue gets no fragment: add the issue to its story file first.

This README is not a fragment; the check skips it.

## Commands

From the repo root (`DT` is `uv run --project tools/devops --locked --no-dev devops-tools`,
`S` is `--settings-file tools/devops/pyproject.toml`):

- `$DT changelog-fragments $S check --branch "$(git branch --show-current)"` checks every
  fragment and that this branch's issue has one. CI
  (`.github/workflows/changelog-fragments.yml`) runs the same check on every pull request,
  with the pull request's branch.
- `$DT changelog-fragments $S preview` prints the unreleased entries, newest first.
- `$DT changelog-fragments $S release --section-title "<title>"` moves every fragment into
  `CHANGELOG.md` under `## <title>`, above the first `## ` section, and deletes them.

## Releasing

The owner releases in a pull request of its own, for its own issue. Run `release` on a branch
made from the latest `main`, so the squash-merge subjects that added the fragments give the PR
numbers, then add that issue's own fragment (it goes into the next release), so the branch check
passes.

Settings: `changes-directory` and `changelog-file` in `[tool.devops-tools]` of
`tools/devops/pyproject.toml`. The tool is
[SETA1609/devops_tools](https://github.com/SETA1609/devops_tools), pinned there to a commit.
