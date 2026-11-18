# PO_S18_01_file-app-environment-issues File the app environment issues

Date: 2026-11-18

- **Decision:** a new story, S18 App environments: `FE_S18_02` replaces the four `define` values repeated in `frontend/angular.json` with one typed environment per build (`local`, `local-api`, `container`, `pages`), and `OPS_S18_03` makes the Pages workflow and the local scripts pick and fill them. The environments carry the API mode, the dev console, the service worker, the rules version, the rooms project, the telemetry defaults, and the log level (owner, 2026-11-18). [V 2026-11-18, `frontend/angular.json` on `development`]
- **Rooms and telemetry need it:** `FE_S10_02` (the rooms client) and `FE_S12_04` (frontend telemetry) now need `FE_S18_02`, so each reads its URL or endpoint from the environment instead of adding its own define.
- **A fourth environment, `container`** [A]: `container/app.Dockerfile` builds the `api` configuration, which sets `RIFTWEAVE_DEV` true, so the image ships the dev console [V 2026-11-18]. It gets its own configuration with the dev tools off.
- **Public values only in the build:** the Supabase anon key is public by design and may be in the built site; it comes from repository variables, not secrets. A service-role key never reaches the frontend, and the built-site guard checks for one [A].
- **The rules version stays a define** [A]: a static environment file cannot know the release tag.
- **Open:** how `pages.yml` writes the variables into the build (a generated file or a define), and whether the container gets rooms; both are the spec steps' (`OPS_S18_03`, `FE_S18_02`).
