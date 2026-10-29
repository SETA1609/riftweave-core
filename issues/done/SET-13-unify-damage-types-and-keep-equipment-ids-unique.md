---
id: SET-13
title: Unify damage types and keep equipment ids unique
status: done
closed: 2026-10-04
pr: "#19"
---

`cold`/`lightning` are aliases of `frost`/`shock`, not a second vocabulary. Equipment ids 22-25 stay with armor; weapons stay at 45-48.

Landed in #19. See `ruleset/schemas/damage_type.schema.json`.

_Imported from git history; `closed` is the merge date of the landing commit._
