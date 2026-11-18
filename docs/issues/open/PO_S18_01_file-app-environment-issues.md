#### PO_S18_01_file-app-environment-issues File the app environment issues
- Status: todo
- Priority: medium
- Needs: —
- Modified: 2026-11-18
- Do: Owner request, 2026-11-18: the app needs environments, so the build deployed to GitHub Pages uses some configuration and a local run another. Today the differences are four `define` values repeated in each configuration of `frontend/angular.json` (`RIFTWEAVE_API`, `RIFTWEAVE_DEV`, `RIFTWEAVE_SERVICE_WORKER`, `RIFTWEAVE_RULES_VERSION`). The environments carry those four, the rooms project (Supabase URL and public anon key), the telemetry defaults, and the log level. File story S18 with `FE_S18_02` (the environment files) and `OPS_S18_03` (the workflow and scripts that supply the values), and make the rooms and frontend telemetry issues need `FE_S18_02`.
- Keys: none
- Files: ~10 (docs/stories/S18-app-environments.md, docs/stories/README.md, ROADMAP.md, docs/issues/open/PO_S18_01_file-app-environment-issues.md, FE_S18_02_frontend-build-environments.md, OPS_S18_03_environment-values-in-ci-scripts.md, FE_S10_02_room-log-client.md, FE_S12_04_frontend-observability.md, docs/decisions/PO_S18_01_file-app-environment-issues.md, changes/PO_S18_01_file-app-environment-issues.md)
- See: docs/backend/frontend.md#github-pages, docs/backend/dev-mode.md, docs/backend/observability.md, frontend/angular.json
- Done when: story S18 and its two issues exist with their `Needs`; `FE_S10_02` and `FE_S12_04` need `FE_S18_02`; the stories README lists S18; `roadmap-order show` lists the new issues.
