---
id: RW-5
title: Progression as data: XP, skill points, perk cadence
status: open
milestone: M1
priority: high
labels: [progression, data]
linear:
created: 2026-10-29
---

XP thresholds, skill points per level, tag-skill ratio, perk cadence, and the level 30 skill cap live in `progression.md`; `derived.md` says they stay there. A second implementation cannot read prose. Move them to data and formulas.

## Done when

- [ ] `ruleset/data/progression/` (or equivalent) holds the XP table and per-level gains, validated by a schema.
- [ ] Level-up 1 to 30 has vectors, including the tag-skill and cap rules.
- [ ] `progression.md` becomes commentary that cites the data.

See: docs/modules/progression.md; docs/core/derived.md; docs/core/loop.md
