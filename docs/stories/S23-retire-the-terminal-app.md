# S23 Retire the terminal app

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S23 Retire the terminal app

**Goal:** One app for people and one interface for agents. The interactive terminal app (the Charm TUI, APP-05) is removed: people play in the web app, which runs the same core in the browser, and agents use the JSON commands and the MCP server. Nothing the terminal app did is lost: its rule steps live in core commands, and its flow tests live on as vectors and web end-to-end tests. Each feature is then built once in the core and once in the web app, not twice.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `PO_S23_01_file-tui-retirement-issues` | PO | File the issues that retire the terminal app | — | — |
| `PO_S23_02_cli-issues-to-command-and-mcp` | PO | Rewrite the open terminal-screen issues as a command and an MCP tool | — | — |
| `BE_S23_03_tui-flow-audit` | BE | Move every rule step that lives only in the terminal app into the core | — | — |
| `QA_S23_04_port-tui-flow-tests` | QA | Keep the terminal app's flow tests as core vectors and web end-to-end tests | — | BE_S23_03_tui-flow-audit |
| `BE_S23_05_remove-tui` | BE | Remove the terminal app | — | PO_S23_02_cli-issues-to-command-and-mcp, BE_S23_03_tui-flow-audit, QA_S23_04_port-tui-flow-tests |

**Done when:** every issue above is done; `go.mod` has no Charm module; no open issue asks for a terminal screen; the agent CLI and the MCP server pass their tests as before.
