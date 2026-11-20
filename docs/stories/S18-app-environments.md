# S18 App environments

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S18 App environments

**Goal:** The frontend reads one typed environment per build, so a local run and the GitHub Pages site each use their own API mode, dev tools, rooms project, telemetry defaults, and log level, and the Pages build can never ship a local value.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `PO_S18_01_file-app-environment-issues` | PO | File the app environment issues | — | — |
| `FE_S18_02_frontend-build-environments` | FE | One typed environment per frontend build: local, local with the server, container, and Pages | — | — |
| `OPS_S18_03_environment-values-in-ci-scripts` | OPS | The Pages workflow and the local scripts supply each environment's values | — | FE_S18_02_frontend-build-environments |

**GitHub:** story [#275](https://github.com/SETA1609/riftweave-core/issues/275); tasks [`PO_S18_01_file-app-environment-issues` #371](https://github.com/SETA1609/riftweave-core/issues/371), [`FE_S18_02_frontend-build-environments` #372](https://github.com/SETA1609/riftweave-core/issues/372), [`OPS_S18_03_environment-values-in-ci-scripts` #373](https://github.com/SETA1609/riftweave-core/issues/373).

**Done when:** every issue above is done, `ng serve`, `./scripts/app.sh`, `./scripts/app-container.sh`, and the Pages build each run with their own environment, and CI fails a Pages build that carries a localhost URL or the dev console.
