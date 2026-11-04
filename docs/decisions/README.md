# Decisions

One file per issue that decides something: `docs/decisions/<issue-id>.md`. Each pull request adds its own file, so two open pull requests never edit the same lines and GitHub can merge them in any order. The older log, up to 2026-11-04, stays in [`../decisions.md`](../decisions.md) and is not edited any more; its open questions and risks stay open until an issue settles them in its own file.

## Format

```markdown
# M5-15 One file per decision

Date: 2026-11-04

- **Decision:** what was decided, and where the rule now lives.
- **Why:** the problem it solves.
- **Open:** what is still an assumption [A] or unverified [U], if anything.
- **Risk:** what could go wrong, if anything.
```

- **File name:** the issue id plus `.md`, spelled as in the story file (`M1-05.md`).
- **First line:** `# <issue-id> <issue title>`.
- **Date:** the day of the decision, `YYYY-MM-DD`.
- Bullets as in the old log; tag claims **[V date]**, **[A]**, or **[U]** ([CONTRIBUTING](../../CONTRIBUTING.md#verification-tags)).
- A later issue that changes a decision adds its own file and names the one it replaces; it does not edit the old file.
- An issue split into a ruleset and a backend pull request has one file: the ruleset one adds it, the backend one adds bullets to it.

Newest first: `grep -H '^Date:' docs/decisions/*.md | sort -t: -k3 -r`.
