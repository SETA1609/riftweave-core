#### FE_S08_06_bestiary-page A bestiary, and the GM decides what the players know (frontend)
- Status: todo
- Priority: medium
- Needs: BE_S08_05_bestiary-knowledge-cli, APP-26, APP-27
- Was: APP-29
- Do: Frontend half, split by role on 2026-11-18 (PO_S00_01_issue-splitting-by-role): the Angular screens of `BE_S08_05_bestiary-knowledge-cli` and their UX, over the same commands and session steps, built from the shared components (docs/backend/frontend.md#conventions). The behaviour, the keys, and the rules are specified in `BE_S08_05_bestiary-knowledge-cli`; this issue adds no contract or core change.
- Keys: the frontend screen and its route; no new contract keys
- Files: ~3 (frontend/src/app/bestiary/, frontend/src/app/app.spec.ts, frontend/e2e/bestiary.spec.ts)
- See: docs/issues/open/BE_S08_05_bestiary-knowledge-cli.md
- Done when: the frontend lists and filters the bestiary; with "Players see only what they know" off every sheet shows; with it on a player sees only `seen` and `known` entries; the GM's view shows every mark; the specs and the Playwright test pass.
