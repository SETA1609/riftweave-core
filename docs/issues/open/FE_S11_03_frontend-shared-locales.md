#### FE_S11_03_frontend-shared-locales One set of i18next locale files for the frontend and the human CLI (frontend)
- Status: todo
- Priority: medium
- Needs: BE_S11_02_cli-shared-locales, FE_S11_01_frontend-i18next, APP-26
- Was: APP-28
- Do: Frontend half, split by role on 2026-11-18 (PO_S00_01_issue-splitting-by-role): the Angular screens of `BE_S11_02_cli-shared-locales` and their UX, over the same commands and session steps, built from the shared components (docs/backend/frontend.md#conventions). The behaviour, the keys, and the rules are specified in `BE_S11_02_cli-shared-locales`; this issue adds no contract or core change.
- Keys: the frontend screen and its route; no new contract keys
- Files: ~3 (frontend/angular.json, frontend/src/app/i18n.ts, frontend/src/app/app.spec.ts)
- See: docs/issues/open/BE_S11_02_cli-shared-locales.md
- Done when: the frontend bundles `backend/locales/<lng>.json` instead of its own files; one key with interpolation and one plural render the same text in a Vitest test as in the Go test, in en, de, and es; a file with an unsupported i18next feature fails the test; the tests pass.
