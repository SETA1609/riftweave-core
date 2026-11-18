# S12 Observability and dev tools

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S12 Observability and dev tools

**Goal:** The server, the CLIs, and the frontend send OpenTelemetry data you can look at locally, and the dev build opens the app in the browser console.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `BE_S12_01_telemetry-format-setting` (reserved) | BE | see `APP-43` | APP-43 | open pull request #239: keeps `APP-43` until it merges |
| `BE_S12_02_server-observability` | BE | The server is observable: OpenTelemetry traces, metrics, and logs, and health endpoints | APP-44 | APP-43 |
| `BE_S12_03_cli-observability` | BE | The agent CLI and the human CLI are observable | APP-45 | APP-43 |
| `FE_S12_04_frontend-observability` | FE | The frontend is observable, and exports its diagnostics | APP-46 | APP-43, FE_S18_02_frontend-build-environments |
| `OPS_S12_05_local-telemetry-stack` | OPS | A local stack to look at the telemetry | APP-47 | BE_S12_02_server-observability |
| `FE_S12_06_browser-dev-console` (reserved) | FE | see `APP-48` | APP-48 | open pull request #240: keeps `APP-48` until it merges |

**Done when:** every issue above is done, and traces, metrics, and logs from every app show in the local stack, and the dev console works.
