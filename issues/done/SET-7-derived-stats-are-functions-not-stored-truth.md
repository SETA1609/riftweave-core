---
id: SET-7
title: Derived stats are functions, not stored truth
status: done
closed: 2026-10-04
pr: "#20"
---

Coefficients live in `derived/core.json`. HP, MP, AP, carry, and initiative recompute from the snapshot. A module emits a modifier; it does not write a second max.

Landed in #20. See `ruleset/data/derived/core.json`.

_Imported from git history; `closed` is the merge date of the landing commit._
