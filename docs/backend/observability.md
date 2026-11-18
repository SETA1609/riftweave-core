# Observability

Modified: 2026-11-18

Every component reports in one format, OpenTelemetry's, so a trace from the frontend through
the server reads as one trace, and a diagnostics file reads the same whichever app wrote it
(APP-43). This page says what each component emits. It is a spec: the server's traces, metrics,
and logs are APP-44, and the dev console's diagnostics export is APP-45 [A]. What a process does
with the setting is below; the setting itself is in [settings.md](settings.md).

Sources, read on 2026-11-18 from shallow clones of the OpenTelemetry repositories (the API path
of `devops-tools fetch-docs` was refused with HTTP 403 in the sandbox):

- OTLP specification, `docs/specification.md` of `open-telemetry/opentelemetry-proto` at
  `ef48090`: "JSON Protobuf Encoding".
- OpenTelemetry specification, `open-telemetry/opentelemetry-specification` at `a7b22a8`:
  `specification/protocol/exporter.md` and `specification/configuration/sdk-environment-variables.md`.
- Semantic conventions, `open-telemetry/semantic-conventions` at `3535c02`:
  `docs/registry/attributes/service.md` and `error.md`.

## The format

Logs, traces, and metrics use OpenTelemetry's data model, written as **OTLP/JSON**: the proto3
JSON mapping, with `traceId` and `spanId` as hex strings, enum values as integers, and object
keys in lowerCamelCase (OTLP specification, "JSON Protobuf Encoding"). [V 2026-11-18] Export is
**OTLP/HTTP**; its default endpoint is `http://localhost:4318`, with the signal paths
`/v1/traces`, `/v1/metrics`, and `/v1/logs` appended to a base URL (`exporter.md`). [V 2026-11-18]

## Resource: who is reporting

Every record carries the resource of the process that wrote it.

| Attribute | Value | Source |
| --- | --- | --- |
| `service.name` | `riftweave-server`, `riftweave-cli`, `riftweave-play`, or `riftweave-web` | semantic conventions, stable [V 2026-11-18] |
| `service.version` | the build (`git describe`, or `dev`) [A] | semantic conventions, stable [V 2026-11-18] |
| `riftweave.rules_version` | the rules the core runs (`core.RulesVersion`) | this project |

| Component | `service.name` | What it is |
| --- | --- | --- |
| HTTP API | `riftweave-server` | `backend/cmd/riftweave-server` |
| Agent CLI | `riftweave-cli` | `riftweave <command>` with JSON in and out (docs/backend/cli.md) |
| Human CLI | `riftweave-play` | `riftweave play`, the TUI (APP-26) |
| Frontend | `riftweave-web` | the Angular app, with the core in the page or a server (docs/backend/frontend.md) |

The MCP server runs inside `riftweave-server` and reports as it [A].

## Spans: what a command did

Each call to the core is one span, made by the adapter around `core.Run` (or the page's
`riftweave.run`):

| Field | Value |
| --- | --- |
| name | `riftweave.command <name>`, such as `riftweave.command attack` |
| kind | `INTERNAL` (1) in the CLIs and the page; a child of the request's `SERVER` span on the server (APP-44) |
| `riftweave.command.name` | the command, such as `attack` |
| `error.type` | on a refusal only: the refusal's code (low cardinality, as the convention asks) [V 2026-11-18] |
| status | `ERROR` (2) on a refusal, unset otherwise |

A frontend call to a server sends W3C `traceparent`, so the page's span and the server's are one
trace (APP-44). [A]

## Metrics

| Instrument | Kind, unit | Attributes |
| --- | --- | --- |
| `riftweave.command.duration` | histogram, `s` | `riftweave.command.name`, `error.type` on a refusal |

Its count is the number of calls; the server adds the HTTP server metrics of the semantic
conventions and the Go runtime's (APP-44). [A]

## Log records

A log record names an event, not a person: the startup line (`service.version`,
`riftweave.rules_version`), a refusal (`riftweave.command.name`, `error.type`), and an error of
the adapter itself (a file it could not read, a server it could not reach). A record inside a
span carries its `traceId` and `spanId`.

## Nothing about the player

An event names the command, its duration, its outcome, and the kind of error. It never carries
an actor, a session, a save, a roll, or any free-text input: no attribute, body, or span name
holds them, whatever the mode. [A] A test of each emitting component checks one call's records
for the actor's name and a typed text (APP-44, APP-45).

## The setting, and the environment

The `telemetry` key of the settings file decides what a component does ([settings.md](settings.md)):

| `mode` | What happens |
| --- | --- |
| `off` (the default) | nothing is recorded |
| `local` | records stay on the device, in memory or a file, and can be exported as a diagnostics file: one OTLP/JSON object with `resourceSpans`, `resourceMetrics`, and `resourceLogs` [A] |
| `export` | records are sent by OTLP/HTTP to `endpoint`, or to `http://localhost:4318` with none |

A process that has an environment follows OpenTelemetry's variables over the file [A]:

| Variable | Effect | Source |
| --- | --- | --- |
| `OTEL_SDK_DISABLED` | `true` records nothing, whatever the mode | sdk-environment-variables.md [V 2026-11-18] |
| `OTEL_SERVICE_NAME` | replaces `service.name` | sdk-environment-variables.md [V 2026-11-18] |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | replaces `endpoint` for `export` | exporter.md [V 2026-11-18] |

The page has no environment, so the frontend reads the file only.

## The core stays pure

No telemetry package is imported by `backend/core` or `backend/internal/`: the adapters observe
the calls they make. `TestTelemetryStaysOutOfTheCore` (`backend/cmd/riftweave/play_test.go`)
fails if one is, as `TestCharmStaysOutOfTheCore` does for Charm. [V 2026-11-18, the test added]
