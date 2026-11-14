# APP-20 Playwright end-to-end tests for the frontend

Date: 2026-11-14

- **Decision:** `@playwright/test` 1.64.1, Chromium only, specs in `frontend/e2e/`, run by `npm run e2e`. `playwright.config.ts` starts `riftweave-server` (`go run`) and `ng serve` unless each already answers, so a local run needs nothing started by hand, and CI and `./scripts/app.sh` reuse their servers. [V 2026-11-14, local run]
- **Decision:** three specs act as a player: the sheet computes stats and creates a character; the attack form, filled with the first `combat.json` vector's inputs, shows that vector's `outcome`, `target`, `damage`, and `hp`; the playthrough runs to a matching replay. Elements are found by label and role. [V 2026-11-14]
- **Decision:** a card is now named by its title (`aria-label` on its `<section>`), so it is a `region` a spec, and a screen reader, finds by name. No other component changed. [V 2026-11-14, the 23 Vitest specs still pass]
- **Decision:** CI installs Chromium with its system libraries (`npx playwright install --with-deps chromium`), runs the specs after the unit specs, and keeps the HTML report for 7 days when a run fails (`actions/upload-artifact` v4.6.2, pinned by commit). [A, the workflow has not run in GitHub yet]
- **Why:** owner rule, 2026-11-14: unit tests on Vitest, end-to-end tests on Playwright.
- **Open:** only Chromium is tested; Firefox and WebKit are one line each in `projects` when wanted. [A]
