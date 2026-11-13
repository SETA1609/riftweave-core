# MCP server

`riftweave mcp` is a third front end in the same binary (APP-09): a [Model Context
Protocol](https://modelcontextprotocol.io) server on stdio, for an agent in an MCP client. Like
the agent CLI and the HTTP API, it parses, calls `core.Run`, and formats; it holds no rule
([README.md](README.md)).

```
riftweave [--data <dir>] [--dev] mcp
```

The client starts the binary and talks JSON-RPC 2.0 on its stdin and stdout, one message per
line. Nothing else is written to stdout; a log line goes to stderr. [V 2026-11-12, `go test ./mcp`]

## Messages

| Method | Answer |
| --- | --- |
| `initialize` | the protocol version (the client's, when the server speaks it, else the newest it speaks), `capabilities.tools`, and `serverInfo` |
| `notifications/initialized` | none (a notification) |
| `ping` | `{}` |
| `tools/list` | every tool, below |
| `tools/call` | the tool's result as one `text` content block holding the JSON, and the same JSON as `structuredContent`; a rule error is a result with `isError: true` and the `{"error": ...}` JSON |

An unknown method is JSON-RPC error `-32601`, a line that is not JSON `-32700`, and arguments that
are not a JSON object `-32602`. Bad input to a command (an unknown field, a wrong type) is a tool
result with `isError: true`, as a rule error is, so the agent reads why. [A]

## Tools

**One tool per contract command**, generated from `backend/contract/` at start-up as the OpenAPI
document is (docs/backend/api.md#openapi), never written by hand. The tool's `name` is the
command's, its `description` the contract's `summary` and the docs file that owns the rule, and its `inputSchema`
the contract's `input_schema` with every `$ref` it reaches (in `common.json` and the ruleset's
`schema.json`) copied into `$defs`, so the schema stands alone. Calling it runs the command on the
arguments: the same JSON the agent CLI gives for the same input. [V 2026-11-12, every vector through
`tools/call`]

**Session tools** hold one open session between calls, so an agent plays a run without passing the
session back each time ([sessions.md](sessions.md)):

| Tool | Arguments | Result |
| --- | --- | --- |
| `session_start` | `start`, a stored actor | the new session, now held |
| `session_step` | `step` (`command`, `input`, optional `as`) | the command's `result` and the held session's `actor` |
| `session_save` | none | the held session, the same JSON the `session` command returns (APP-04) |
| `session_load` | `session`, a saved session | the session, now held |
| `session_replay` | none | the held session rebuilt from its start and log, and `matches` |

A session tool other than `session_start` and `session_load` with no session held is an error
result. The session lives only as long as the process.

**Dev tools** exist only when the server starts with `--dev` ([dev-mode.md](dev-mode.md)): the
`dev` command's tool, and `session_dev`, which runs one dev operation (`operation` and its fields)
on the held session. Without `--dev` they are not in `tools/list`, and calling one is an unknown
tool. [V 2026-11-12, `go test ./mcp`]

## Checks

`go test ./mcp` runs every vector in `ruleset/tests/vectors/` through `tools/call` and compares the
JSON with `core.Run`; it checks that every contract command is a tool with a standalone input
schema, that a session persists across calls and saves to the JSON the `session` command builds
from the same steps, and that the dev tools are absent without `--dev` and present with it.
