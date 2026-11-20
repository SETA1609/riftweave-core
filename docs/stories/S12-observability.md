# S12 Observability and dev tools

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S12 Observability and dev tools

**Goal:** The server, the CLIs, and the frontend send OpenTelemetry data you can look at locally, and the dev build opens the app in the browser console.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `APP-43` | BE | Telemetry in the OpenTelemetry format, and a telemetry setting both apps share | — | — |
| `BE_S12_02_server-observability` | BE | The server is observable: OpenTelemetry traces, metrics, and logs, and health endpoints | APP-44 | APP-43 |
| `BE_S12_03_cli-observability` | BE | The agent CLI and the human CLI are observable | APP-45 | APP-43 |
| `FE_S12_04_frontend-observability` | FE | The frontend is observable, and exports its diagnostics | APP-46 | APP-43, FE_S18_02_frontend-build-environments |
| `OPS_S12_05_local-telemetry-stack` | OPS | A local stack to look at the telemetry | APP-47 | BE_S12_02_server-observability |
| `APP-48` | FE | The frontend's objects and functions in the browser console, in a dev build | — | — |

**GitHub:** story [#269](https://github.com/SETA1609/riftweave-core/issues/269); tasks [`BE_S12_02_server-observability` #331](https://github.com/SETA1609/riftweave-core/issues/331), [`BE_S12_03_cli-observability` #332](https://github.com/SETA1609/riftweave-core/issues/332), [`FE_S12_04_frontend-observability` #333](https://github.com/SETA1609/riftweave-core/issues/333), [`OPS_S12_05_local-telemetry-stack` #334](https://github.com/SETA1609/riftweave-core/issues/334).

**Done when:** every issue above is done, and traces, metrics, and logs from every app show in the local stack, and the dev console works.
