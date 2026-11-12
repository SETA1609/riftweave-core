# HTTP API

`riftweave-server` serves the core over HTTP (APP-06). It is an adapter: it parses the request,
calls `core.Run` with the command name and the body, and writes the result. It holds no rule and
no state, so a call needs no session and two calls never share anything but the ruleset.
[V 2026-11-12, `go test ./cmd/riftweave-server`]

```
riftweave-server [--data <dir>] [--addr <host:port>] [--dev]
```

- `--data` is the ruleset data directory, found as the agent CLI finds it ([cli.md](cli.md)):
  the flag, then `RIFTWEAVE_DATA`, then `ruleset/data` above the working directory.
- `--addr` defaults to `127.0.0.1:8080`, so the server is local unless told otherwise.
- `--dev` turns on dev mode, as the CLI flag does ([dev-mode.md](dev-mode.md)).

## Endpoints

| Method and path | Body | Result |
| --- | --- | --- |
| `POST /v1/commands/{command}` | the command's input ([README.md](README.md#commands)) | the command's result |
| `GET /v1/commands` | none | every command with its contract summary, sorted by name |
| `POST /v1/sessions` | `{ "start": <stored actor> }` | the `session` command's `start` |
| `POST /v1/sessions/step` | `{ "session", "step" }` | the `session` command's `step` |
| `POST /v1/sessions/replay` | `{ "session" }` | the `session` command's `replay` |
| `GET /openapi.json` | none | the OpenAPI 3.1 document |

The session endpoints are the `session` command with its `operation` filled in from the path. The
client holds the session JSON and sends it back with each step; the server keeps nothing between
calls ([sessions.md](sessions.md)).

## Status codes

| Status | When | Body |
| --- | --- | --- |
| 200 | the command ran | the result |
| 400 | the body is not the command's input (`core.ErrInput`), or a rule needs an input it was not given (`MissingInputError`) | `{ "error": "..." }` |
| 404 | no such command or path, or the input names a key that does not exist (`NotDefinedError`) | `{ "error": "..." }` |
| 405 | the method is not the one the path takes | `{ "error": "..." }` |
| 413 | the body is over 1 MiB | `{ "error": "..." }` |
| 422 | the command refused for any other rule reason (`core.ErrRule`) | `{ "error": "..." }`, the same bytes the CLI prints |

## OpenAPI

`GET /openapi.json` is built from `backend/contract/` when the server starts, never written by
hand. Each command is one path, `/v1/commands/<name>`, with its `input_schema` as the request
body and its `result_schema` as the 200 response. The definitions of `common.json` and of
`ruleset/schemas/schema.json` become `components.schemas`, the shared ones prefixed `ruleset_`,
and every `$ref` is rewritten to point inside the document, so it stands alone. A test checks
that every command has its path and that every `$ref` resolves, and the document validated
as OpenAPI 3.1 with `openapi-spec-validator` when it was written. [V 2026-11-12]

## Not covered

Authentication, rate limits, CORS, and TLS: the server binds to localhost by default and is
meant to run beside the frontend or behind a proxy that adds them. [A]
