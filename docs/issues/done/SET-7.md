#### SET-7 Derived stats are functions, not stored truth
- Status: done
- Merged: #20, 2026-10-04
- Needs: —
- Do: Coefficients live in `derived/core.json`. HP, MP, AP, carry, and initiative recompute from the snapshot. A module emits a modifier; it does not write a second max. Landed in #20. See `ruleset/data/derived/core.json`.
