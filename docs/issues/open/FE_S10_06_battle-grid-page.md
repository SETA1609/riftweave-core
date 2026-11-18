#### FE_S10_06_battle-grid-page Online table: a shared battle grid, and the initiative list in sync (frontend)
- Status: todo
- Priority: low
- Needs: BE_S10_05_battle-grid-steps, FE_S10_04_gm-approval-table, FE_S08_08_encounter-tracker-page
- Was: APP-35
- Do: Frontend half, split by role on 2026-11-18 (PO_S00_01_issue-splitting-by-role): the Angular screens of `BE_S10_05_battle-grid-steps` and their UX, over the same commands and session steps, built from the shared components (docs/backend/frontend.md#conventions). The behaviour, the keys, and the rules are specified in `BE_S10_05_battle-grid-steps`; this issue adds no contract or core change.
- Keys: the frontend screen and its route; no new contract keys
- Files: ~4 (frontend/src/app/table/, frontend/src/app/table.spec.ts, frontend/e2e/table.spec.ts, docs/backend/frontend.md) [A]
- See: docs/issues/open/BE_S10_05_battle-grid-steps.md
- Done when: in a room of a GM and four players, each browser shows the same grid, tokens, initiative order, and current turn, and a move appears on every screen once appended; the specs and the Playwright test pass.
