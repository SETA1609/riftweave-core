# PO_S23_02_cli-issues-to-command-and-mcp Rewrite the open terminal-screen issues as a command and an MCP tool

Date: 2026-11-23

- **Decision:** the terminal app is frozen. No open issue outside S23 builds a terminal screen: each keeps its core command and gains its MCP tool, and its web page (the `FE_` half) is how people use it. The issue ids and titles stay, so no `Needs` edge breaks; a title is a frozen name.
  - **Rewritten:** `BE_S07_05`, `BE_S08_01`, `BE_S08_03`, `BE_S08_07`, `BE_S08_15`, `BE_S09_05`, `BE_S09_07`, `BE_S10_01`, `BE_S10_03`, `BE_S10_05`, `BE_S20_03`, `BE_S20_08` (each screen step, each `play/` file, and each TUI test becomes the MCP tool and an MCP test); `FE_S07_06` and `FE_S08_08` compare with the commands, not the human CLI; `M5-77`'s role row.
  - **Observability:** `BE_S12_03` keeps the agent CLI and server paths, and the human CLI's file records move to the MCP server, whose stdout carries the protocol; the SSH app line is gone.
  - **Dropped:** `BE_S11_02_cli-shared-locales` and `FE_S11_03_frontend-shared-locales`, which moved the locale files so the Go terminal app could read them. The commands print JSON and keys, so the frontend's own files (FE_S11_01) have no second reader. Nothing else needs either.
  - **The playtest:** `QA_S01_02` runs in the web app; a run already under way in the terminal app finishes there.
  - **The rule** is in AGENTS.md and `docs/backend/README.md`, and the story goals of S07, S08, S09, and S11 no longer name the human CLI.
- **The check:** `test_no_terminal_screens.py` fails an open issue outside S23 that is not dropped and asks for a terminal screen (TUI, human CLI, a CLI screen, both apps, a file in `backend/cmd/riftweave/play/`, `play_test.go`), titles left out. [V 2026-11-23, red with 18 issues, then green]
- **Not done here:** removing the terminal app (`BE_S23_05`) and moving its rule steps and tests (`BE_S23_03`, `QA_S23_04`); `docs/backend/README.md` still describes the human CLI, marked frozen, until then.
