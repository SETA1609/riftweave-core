#### FE_S09_06_spell-cards-page Spell cards, and custom spells printed as cards (frontend)
- Status: todo
- Priority: low
- Needs: BE_S09_05_spell-cards-print, APP-10, FE_S09_02_spell-editor-page, FE_S09_04_print-character-page
- Was: APP-40
- Do: Frontend half, split by role on 2026-11-18 (PO_S00_01_issue-splitting-by-role): the Angular screens of `BE_S09_05_spell-cards-print` and their UX, over the same commands and session steps, built from the shared components (docs/backend/frontend.md#conventions). The behaviour, the keys, and the rules are specified in `BE_S09_05_spell-cards-print`; this issue adds no contract or core change.
- Keys: the frontend screen and its route; no new contract keys
- Files: ~3 (frontend/src/app/spell-editor/, frontend/src/app/app.spec.ts, frontend/e2e/print.spec.ts)
- See: docs/issues/open/BE_S09_05_spell-cards-print.md
- Done when: the frontend prints a core spell and an authored spell as cards, nine to a page; a spell with problems shows them instead of printing; the specs and the Playwright test pass.
