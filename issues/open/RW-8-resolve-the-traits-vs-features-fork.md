---
id: RW-8
title: Resolve the traits vs features fork
status: open
milestone: M1
priority: medium
labels: [data, schema]
linear:
created: 2026-10-29
---

`traits/core.json` (18) and `features/core.json` (23) both exist, and the review (data-schemas M1) flagged the fork. CLAUDE.md says `features` is the single perk table. Pick one model before more racial content lands.

## Done when

- [ ] One table owns racial traits; the other is removed or documented as a view.
- [ ] All references migrated; `validate.py` is green.
- [ ] `docs/modules/race.md` and the schemas describe the decision.

See: docs/review/findings-data-schemas.md M1; docs/review/findings-docs-hygiene.md M3
