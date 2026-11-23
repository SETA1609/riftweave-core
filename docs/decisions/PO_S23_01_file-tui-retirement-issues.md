# PO_S23_01_file-tui-retirement-issues File the issues that retire the terminal app

Date: 2026-11-23

- **Decision** (owner, 2026-11-23): the interactive terminal app (the Charm TUI, APP-05) is an extra and is retired.
  - **People** use the web app. It runs the same core through `riftweave-wasm`, so offline play stays.
  - **Agents** use the JSON commands (`list`, `describe`, `command`) and `riftweave mcp`.
- **Why:**
  - **Built twice:** each feature was built twice, a terminal screen and a web page. Eleven open issues carry a `_cli` screen beside their `FE_` page, and the terminal app has its own locale (`BE_S11_02`) and observability (`BE_S12_03`) work.
  - **Size:** `play/` is about 2,300 lines of Go and about 650 lines of tests, over half of `backend/cmd/riftweave/` [V 2026-11-23, wc -l].
- **Order:** freeze the screens first (`PO_S23_02`), move any rule step out of `play/` (`BE_S23_03`), keep its flow tests (`QA_S23_04`), then remove it (`BE_S23_05`). Nothing is removed before the steps and tests it holds have a home.
- **Kept:** the agent CLI, the MCP server, and `backend/cmd/riftweave/characters/` and `files/`, which the agent commands use.
- **Assumed** [A]:
  - `riftweave play` prints that play is in the web app and agents use `command` or `mcp`, with exit code `exitInput`.
  - An alpha playtest run already under way in the terminal app finishes there; new runs use the web app with dev mode.
