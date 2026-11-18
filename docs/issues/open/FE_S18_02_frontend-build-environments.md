#### FE_S18_02_frontend-build-environments One typed environment per frontend build: local, local with the server, container, and Pages
- Status: todo
- Priority: medium
- Needs: —
- Modified: 2026-11-18
- Do: Owner request, 2026-11-18: a local run and the GitHub Pages site use different configuration. Replace the four `define` values that each configuration of `frontend/angular.json` repeats with one typed `Environment`, one file per build, swapped in by `fileReplacements` (Angular's environment pattern [A]; the spec step checks it against the Angular 22 docs and records the pinned source with `fetch-docs`, as docs/backend/frontend.md does for its other claims).
  - **The shape** (`frontend/src/environments/environment.model.ts`), read through one injectable token, never a global:
    - `api`: `null` (the core runs in the page as WebAssembly) or a base URL (`''` is the same origin, behind the proxy or nginx).
    - `dev`: the dev console `rw` and dev-only routes (docs/backend/dev-mode.md).
    - `serviceWorker`: register `ngsw-worker.js`.
    - `rulesVersion`: shown in the footer; `'dev'` in every file, replaced at build time with the release tag on Pages (the one value that stays a `define`, since a static file cannot know the tag [A]).
    - `rooms`: `null` (no online rooms) or `{ url, anonKey }` of the Supabase project (S10). The anon key is public by design (row-level security guards the data), so it may be in a built file; a service-role key never is.
    - `telemetry`: the default of the telemetry setting (`off`, `local`, or `export`) and the default OTLP/HTTP endpoint (docs/backend/observability.md). A player's own setting always wins over the default.
    - `logLevel`: `debug`, `info`, `warn`, or `error`; a small logger reads it, so no `console.log` stays in app code.
  - **The environments** [A]:

    | Field | `local` (`ng serve`, WebAssembly) | `local-api` (`app.sh`, with the server) | `container` (`app-container.sh`, nginx) | `pages` (GitHub Pages) |
    | --- | --- | --- | --- | --- |
    | `api` | `null` | `''` | `''` | `null` |
    | `dev` | `true` | `true` | `false` | `false` |
    | `serviceWorker` | `false` | `false` | `false` | `true` |
    | `rooms` | local Supabase, `http://127.0.0.1:54321` and its local anon key | the same | `null` [A] | the hosted project, from repository variables (OPS_S18_03) |
    | `telemetry` | `local`, endpoint `http://localhost:4318` | `local`, `http://localhost:4318` | `off` | `off`, no endpoint |
    | `logLevel` | `debug` | `debug` | `info` | `warn` |

    The container today builds the `api` configuration, which turns the dev console on; it gets its own `container` configuration with `dev: false` [A].
  - **Offline first:** every environment plays with no network; `rooms` and `telemetry` only add online features and never block play.
  - **Pages never carries a local value:** a Vitest spec reads `environment.pages.ts` and fails on `localhost`, `127.0.0.1`, `dev: true`, or `logLevel: debug`; OPS_S18_03 adds the same check on the built site.
  - **Configurations and scripts:** `angular.json` gets `local`, `local-api`, `container`, and `pages` (renaming `development` and `api`, so `defaultConfiguration` is `local` and a plain `ng build` no longer runs with no defines [A]); `package.json` scripts and `container/app.Dockerfile` name them.
  - **Docs:** docs/backend/frontend.md gains an "Environments" section with the table; dev-mode.md and observability.md point at it.
- Keys: `Environment`; `ENVIRONMENT` injection token; `environment.ts`, `environment.local-api.ts`, `environment.container.ts`, `environment.pages.ts`; the `local`, `local-api`, `container`, `pages` configurations
- Files: ~17 (frontend/src/environments/environment.model.ts, environment.ts, environment.local-api.ts, environment.container.ts, environment.pages.ts, environment.spec.ts, frontend/src/app/logger.ts, frontend/src/main.ts, frontend/src/app/api.ts, frontend/src/app/app.ts, frontend/angular.json, frontend/package.json, container/app.Dockerfile, docs/backend/frontend.md, docs/backend/dev-mode.md, docs/decisions/FE_S18_02_frontend-build-environments.md, changes/FE_S18_02_frontend-build-environments.md)
- See: frontend/angular.json (today's defines), frontend/src/main.ts, frontend/src/app/api.ts, docs/backend/frontend.md#github-pages, docs/backend/observability.md, docs/issues/open/BE_S10_01_room-log-supabase.md
- Done when:
  - No `RIFTWEAVE_API`, `RIFTWEAVE_DEV`, or `RIFTWEAVE_SERVICE_WORKER` define remains; `RIFTWEAVE_RULES_VERSION` is the only one.
  - Each configuration builds, and the Vitest spec checks each environment's values against the table and fails on a local value in `pages`.
  - The Playwright projects still pass: the WebAssembly run on `local`, the server run on `local-api`, `chromium-pages` on `pages`.
  - The container image has no dev console.
