# APP-15 The frontend is published on GitHub Pages

Date: 2026-11-16
Modified: 2026-11-16

- **Decision:** `.github/workflows/pages.yml` builds the WebAssembly core, the Angular app in a new
  `pages` configuration, and its service worker into one static site on a `v*` tag (or by hand),
  and publishes it with `actions/configure-pages` v6.0.0, `actions/upload-pages-artifact` v5.0.0,
  and `actions/deploy-pages` v5.0.1, each pinned by commit (released 2026-03-24, 2026-04-08, and
  2026-09-01). **[V 2026-11-16, the commits resolved with `git ls-remote` and their dates read
  from the GitHub API; the workflow linted clean with `check-workflows --strict`]**
- **Rules version:** the tag, passed to the build as `RIFTWEAVE_RULES_VERSION` and shown in the
  footer; a build that is not a tag shows `dev`. The tag and the repository name reach the shell
  only through `env:` and are checked against a pattern before use.
- **Offline and installable:** Angular's service worker (`@angular/service-worker` 22.2.4, the
  version of `@angular/core`, published 2026-10-29) prefetches the app, `wasm_exec.js`, and
  `riftweave.wasm` (about 2.9 MB gzipped), registered only in the `pages` build. A web manifest
  and an SVG icon make it installable. `ng add @angular/pwa` was not used: it writes Angular's own
  logo as the icons. **[A]** Whether every phone browser installs from an SVG-only manifest is
  **[U]**.
- **Tested on the Pages build itself:** a Playwright project `chromium-pages` serves `dist/`
  under `/riftweave-core/` (a small Node server, `frontend/scripts/serve-pages.mjs`), plays with no
  server, shows a rules version, and plays again after going offline. The workflow cannot run
  before a tag, so the first real publish is the first release. **[U]** Pages must be set to
  "GitHub Actions" in the repository settings once, by a maintainer.
- **Difference from the issue:** the service worker caches the core, so a new tag reaches a
  returning player on the visit after the update is found, as Angular's service worker updates.
- **Risk:** the frontend CI now also builds the `pages` configuration for the e2e run, a few
  seconds more.
