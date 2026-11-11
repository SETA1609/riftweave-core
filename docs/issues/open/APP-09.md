#### APP-09 MCP server over the command contract
- Status: todo
- Priority: medium
- Needs: APP-03, APP-04
- Do: A third front end in the same binary, `riftweave mcp`, serving the Model Context Protocol over stdio for agents in an MCP client. Each command is one tool, generated from `backend/contract/` (name, description, input schema) as the OpenAPI document is in APP-06, never written by hand. The server holds the open session between calls, so an agent plays a run without passing the session back each time; tools to open, save, and load a session wrap APP-04. The dev operations (APP-08) are tools only when the server starts with `--dev`. It calls the same core as the agent CLI and holds no rule. Not a 1.0 gate.
- Keys: `riftweave mcp` subcommand; one tool per contract command; session tools
- Files: ~5 (backend/cmd/riftweave/, backend/mcp/, docs/backend/mcp.md, docs/backend/cli.md)
- See: docs/stories/07-backend.md (APP-03, APP-04, APP-06, APP-08), backend/contract/
- Done when: every contract command is a tool with the contract's input schema; a test runs a vector through MCP and gets the same JSON as the agent CLI; a session persists across calls and saves to the same JSON as APP-04; dev tools are absent without `--dev` and present with it.
