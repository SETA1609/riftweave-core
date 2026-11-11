#### SET-13 Unify damage types and keep equipment ids unique
- Status: done
- Merged: #19, 2026-10-04
- Needs: —
- Do: `cold`/`lightning` are aliases of `frost`/`shock`, not a second vocabulary. Equipment ids 22-25 stay with armor; weapons stay at 45-48. Landed in #19. See `ruleset/schemas/damage_type.schema.json`.
