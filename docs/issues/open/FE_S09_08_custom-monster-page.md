#### FE_S09_08_custom-monster-page Custom NPCs and monsters, printed (frontend)
- Status: todo
- Priority: low
- Needs: BE_S09_07_custom-monster-print, FE_S08_06_bestiary-page, FE_S09_04_print-character-page
- Was: APP-41
- Do: Frontend half, split by role on 2026-11-18 (PO_S00_01_issue-splitting-by-role): the Angular screens of `BE_S09_07_custom-monster-print` and their UX, over the same commands and session steps, built from the shared components (docs/backend/frontend.md#conventions). The behaviour, the keys, and the rules are specified in `BE_S09_07_custom-monster-print`; this issue adds no contract or core change.
- Keys: the frontend screen and its route; no new contract keys
- Files: ~3 (frontend/src/app/bestiary/, frontend/src/app/app.spec.ts, frontend/e2e/print.spec.ts)
- See: docs/issues/open/BE_S09_07_custom-monster-print.md
- Done when: the frontend prints a custom monster's stat block and an NPC as a sheet and as a block, with the numbers `monster_sheet` gives; an invalid file shows its problems; the specs and the Playwright test pass.
