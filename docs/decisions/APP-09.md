# APP-09 MCP server over the command contract

Date: 2026-11-12

- **Decision:** the server is written on the standard library, not an MCP SDK: JSON-RPC 2.0 on stdio, one message per line, answering `initialize`, `ping`, `tools/list`, and `tools/call`, which is all a tools-only server needs. It speaks protocol versions 2025-06-18, 2025-03-26, and 2024-11-05, and answers a client with its own version when it is one of them. The core and `internal/` stay free of new dependencies, as the human CLI keeps Charm in `cmd/`. [V 2026-11-12, `go test ./mcp`; the version list is from memory of the MCP specification, [U]]
- **Decision:** a tool's `inputSchema` is the contract's `input_schema` with every definition it reaches from `common.json` and the ruleset's `schema.json` copied into `$defs`, refs rewritten to `#/$defs/<name>` (schema.json's with a `ruleset_` prefix), as the OpenAPI document does with components. An MCP client does not resolve a ref to another file, so each schema stands alone. [V 2026-11-12, `go test ./mcp`]
- **Decision:** a rule error and bad input to a command are both tool results with `isError: true` and the `{"error": ...}` JSON, so an agent reads why; a JSON-RPC error is kept for the protocol (not JSON, an unknown method, arguments that are not an object, an unknown tool). [A]
- **Decision:** `session_step` returns the command's `result` and the actor after it, not the whole session, so a long run does not send its log back on every call; `session_save` returns the whole session, the JSON the `session` command builds from the same steps. [V 2026-11-12, `go test ./mcp`]
- **Why:** an agent in an MCP client can play the ruleset without shelling out to the CLI or passing the session back each turn.
- **Open:** the session lives only as long as the process; saving to a file is the client's job. One server holds one session. [A]
- **Risk:** the contract's description key is `doc` until M5-46 renames it `documentation`; the server reads either.
