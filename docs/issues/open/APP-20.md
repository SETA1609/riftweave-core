#### APP-20 Playwright end-to-end tests for the frontend
- Status: todo
- Priority: medium
- Needs: APP-19
- Do: Owner rule, 2026-11-14: unit tests run on Vitest and end-to-end tests on Playwright (docs/backend/frontend.md#conventions). Add `@playwright/test` (a release at least two weeks old, 1.64.1 when filed), `playwright.config.ts`, and `e2e/*.spec.ts`, run by `npm run e2e`. The specs drive the built app in Chromium through `ng serve` and its proxy against a real `riftweave-server`, as a player would: fill the character sheet and create a character, fill the attack form with a vector's inputs and read the result in the page against that vector's `expect`, and run the playthrough to a matching replay. The Frontend workflow installs the browser (`npx playwright install --with-deps chromium`) and runs them after the unit specs; a failed run uploads the Playwright report as an artifact (actions pinned by commit). Select elements by their label or role, not by CSS class, so a restyle does not break a test [A].
- Keys: `npm run e2e`; `frontend/playwright.config.ts`; `frontend/e2e/`
- Files: ~6 (frontend/package.json, frontend/package-lock.json, frontend/playwright.config.ts, frontend/e2e/app.spec.ts, .github/workflows/frontend.yml, docs/backend/frontend.md)
- See: docs/backend/frontend.md, docs/issues/open/APP-19.md (or done/ once merged)
- Done when: `npm run e2e` creates a character, resolves an attack whose shown result matches its vector, and runs the playthrough to a matching replay, in a browser, locally and in CI.
