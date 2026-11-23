### PO_S23_02_cli-issues-to-command-and-mcp Rewrite the open terminal-screen issues as a command and an MCP tool
- Changed: every open issue that built a terminal screen now builds its core command and its MCP tool, with the web page as the way people use it; AGENTS.md and the backend spec say a feature is a command, an MCP tool, and a web page.
- Removed: the shared-locale issues `BE_S11_02` and `FE_S11_03` are dropped, since the terminal app was their only second reader.
- Added: `test_no_terminal_screens.py`, which fails an open issue outside S23 that asks for a terminal screen.
