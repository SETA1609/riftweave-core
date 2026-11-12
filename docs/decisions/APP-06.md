# APP-06 HTTP API server

Date: 2026-11-12

- **Decision:** `backend/cmd/riftweave-server` serves every command as `POST /v1/commands/{command}`, lists them at `GET /v1/commands`, and serves `GET /openapi.json`. Spec: `docs/backend/api.md`. **[V 2026-11-12, `go test ./cmd/riftweave-server`: every vector over HTTP]**
- **Decision:** the session endpoints (`POST /v1/sessions`, `/v1/sessions/step`, `/v1/sessions/replay`) are the `session` command with `operation` filled in from the path. The client holds the session JSON; the server keeps no state. **[V 2026-11-12]**
- **Decision:** status codes: 400 for bad input or a `MissingInputError`, 404 for an unknown command, path, or key (`NotDefinedError`), 405 for the wrong method, 413 over 1 MiB, 422 for any other rule error; every error body is `{ "error": "..." }`, and a rule error's body is the bytes the CLI prints. **[V 2026-11-12]**
- **Decision:** the OpenAPI 3.1 document is built at start-up from the embedded contract and `ruleset/schemas/schema.json`: `common.json` definitions keep their names, `schema.json` ones get `ruleset_`, and every `$ref` is rewritten to point inside the document. A test resolves every `$ref`; the document also validated with `openapi-spec-validator` once, outside CI. **[V 2026-11-12]**
- **Decision:** the HTTP API does not serve `dev`, as APP-08 says: `POST /v1/commands/dev` is a 404, and `dev` is not in the list or the OpenAPI document. The server has no `--dev` flag. **[V 2026-11-12]**
- **Decision:** the server binds to `127.0.0.1:8080` by default and has no authentication, CORS, or TLS. It finds the data as the CLI does. **[A]**
- **Why:** the frontend (APP-07) and other consumers need the core over HTTP with a generated client.
- **Open:** whether CI should run an OpenAPI validator; `rulesetDirectory` is a copy of the CLI's, small enough to keep apart for now. **[A]**
