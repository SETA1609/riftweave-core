#### APP-15 The frontend is published on GitHub Pages
- Status: todo
- Priority: medium
- Needs: APP-14
- Do: Owner request, 2026-11-14: a player with no technical background should open one link and play, with nothing installed or hosted. A workflow builds the WebAssembly core (APP-14), the Angular app, and the `ruleset/data` of the same commit into one static site and publishes it with the official Pages actions (pinned by commit), on release tags, not on every push to `main`, so players see a stable rules version [A]. The data is bundled, not fetched from GitHub at run time: same origin, no rate limit, cached, and always the rules the core was built with. The footer shows the rules version (the tag). A service worker (Angular's PWA support) makes the app work offline after the first visit and installable on phones and desktops [A]. The app runs under the Pages base path (`/<repo>/`). Write how a player opens and installs it in `docs/guide/`.
- Keys: `.github/workflows/pages.yml`; the service worker config
- Files: ~7 (.github/workflows/pages.yml, frontend/angular.json, frontend/ngsw-config.json, frontend/src/app/app.ts, frontend/package.json, docs/backend/frontend.md, docs/guide/quick-start.md)
- See: docs/issues/open/APP-14.md, docs/backend/frontend.md
- Done when: a release tag publishes the site; the published app shows the tag as its rules version, creates a character and resolves an attack with no server, and works offline after one visit; the workflow lints clean and pins every action by commit.
