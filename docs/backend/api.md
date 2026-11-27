# HTTP API

Modified: 2026-11-23

`riftweave-server` serves the core over HTTP (APP-06). It is an adapter: it parses the request,
calls `core.Run` with the command name and the body, and writes the result. It holds no rule and
no state, so a call needs no session and two calls never share anything but the ruleset.
[V 2026-11-12, `go test ./cmd/riftweave-server`]

```
riftweave-server [--data <dir>] [--addr <host:port>]
```

- `--data` is the ruleset data directory, found as the agent CLI finds it ([cli.md](cli.md)):
  the flag, then `RIFTWEAVE_DATA`, then `ruleset/data` above the working directory.
- `--addr` defaults to `127.0.0.1:8080`, so the server is local unless told otherwise.
- Dev mode is CLI-only: the server does not serve `dev` and has no `--dev` flag ([dev-mode.md](dev-mode.md)).

## Endpoints

What the API should become, resource by resource, is the proposal in
[api-guide.yml](api-guide.yml) (PO_S26_01); this file is what the server serves today.


| Method and path | Body | Result |
| --- | --- | --- |
| `POST /v1/commands/{command}` | the command's input ([README.md](README.md#commands)) | the command's result; every command but `dev` |
| `GET /v1/commands` | none | every command with its contract summary, sorted by name |
| `POST /v1/sessions` | `{ "start": <stored actor> }` | the `session` command's `start` |
| `POST /v1/sessions/step` | `{ "session", "step" }` | the `session` command's `step` |
| `POST /v1/sessions/replay` | `{ "session" }` | the `session` command's `replay` |
| `GET /openapi.json` | none | the OpenAPI 3.1 document |

The session endpoints are the `session` command with its `operation` filled in from the path. The
client holds the session JSON and sends it back with each step; the server keeps nothing between
calls ([sessions.md](sessions.md)).

## Health

Three GET endpoints say whether the server can answer, for a load balancer, a container's health
check, or a person (BE_S12_02). They are in the OpenAPI document.

| Path | Answers |
| --- | --- |
| `/healthz` | `{"status": "ok"}`: the process answers |
| `/readyz` | `{"status": "ready"}`: the ruleset loaded; the server listens only after it has |
| `/version` | `{"build", "rules_version"}`: the build (`-ldflags "-X main.build=..."`, `dev` otherwise) and `core.RulesVersion` |

[V 2026-11-23, `TestTheHealthEndpointsAnswerAndAreDocumented`]

## Telemetry

The server reports traces, metrics, and logs in OpenTelemetry's format
([observability.md](observability.md#the-server)): off unless `OTEL_EXPORTER_OTLP_ENDPOINT` is set
(OTLP/HTTP) or `--telemetry <file>` names a file; `OTEL_SDK_DISABLED=true` turns it off whatever
else is set. With neither, it writes only its log lines, as JSON on stderr.

## Status codes

| Status | When | Body |
| --- | --- | --- |
| 200 | the command ran | the result |
| 400 | the body is not the command's input (`core.ErrInput`), or a rule needs an input it was not given (`MissingInputError`) | `{ "error": "..." }` |
| 404 | no such command or path, `dev`, or the input names a key that does not exist (`NotDefinedError`) | `{ "error": "..." }` |
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

## Authoring

`author` (APP-10) checks draft spells and exports a data file ([authoring.md](authoring.md)). It needs `ruleset/schemas` beside the data directory, which `core.Open` finds; a draft that fails is a result with `problems`, not an error.

## Catalog

`catalog` (BE_S08_10) reads the tables a player or a GM studies to plan a build: weapons, armor,
accessories, consumables, gear, materials, gems, monsters, spells, features (perks and traits),
races, backgrounds, skills, conditions, and effects. It holds no rule; it reads rows.

- **`collections`** lists them in that order, each with its label, its row count, and its columns.
- **`list`** returns one collection's rows. A row is resolved as validate.py composes it
  ([composition.md](../core/composition.md)): its base (a base file under `bases/`, or another row
  of the collection, such as `core:monster/wolf` or `core:spell/unlock`), then its overrides
  merged field by field, then its own fields. A base that is an item (`is_item: true`) is a row
  of its own key. Each row has its `ref` and its `source`: `core`, or the module it comes from.
- **Modules:** every loaded module's rows follow core's, as `module:<id>/<category>/<key>`. A row
  whose key the module's manifest lists in `overrides` replaces the core row and keeps its ref
  ([overrides.md](../core/overrides.md)). A module's equipment row goes to the collection of its
  type, or of the base it names.
- **Filters** are generic, so no collection needs code: `field` is a dotted path
  (`weapon.damage.type`); `equals` matches the value or an item of a list; `min` and `max` bound
  a number; `text` is found in the label or the description, in any case. A row matches every
  filter. **`sort`** orders by a field, `descending` if set; a row without the field sorts last.
- **`get`** returns one row by its ref.
- **Columns** are display, not rules: `collections` in `backend/internal/catalog/catalog.go`
  names each collection's, for example weapons: label, skill, damage, type, properties, range,
  cost, weight.
- **Not composed yet:** a monster's level growth (attributes and damage scaled from its creature
  type's or base monster's level) and its templates. A monster row is its base merged with its
  overrides; `monster_sheet` and BE_S08_05 read the rest. [A]

An unknown collection, operation, or ref is a 404; `list` or `get` without a collection is a 400.
[V 2026-11-21, `catalog_test.go`, `TestTheCatalogOverHTTP`, `TestTheCatalogThroughCall`]
