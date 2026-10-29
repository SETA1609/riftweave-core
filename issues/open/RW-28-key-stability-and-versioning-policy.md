---
id: RW-28
title: Key stability and versioning policy
status: open
milestone: M5
priority: high
labels: [release, policy]
linear:
created: 2026-10-29
---

Cross-references are by key, so renaming a key silently breaks consumers. Define what is breaking before tagging a version.

## Done when

- [ ] `docs/core/versioning.md`: semver, key rename and removal are breaking, deprecation window.
- [ ] A tool diffs keys between tags and lists removals.
- [ ] CHANGELOG marks breaking entries.

See: docs/core/keys.md
