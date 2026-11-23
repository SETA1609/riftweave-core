# BE_S12_02_server-observability The server is observable: OpenTelemetry traces, metrics, and logs, and health endpoints

Date: 2026-11-23

- **Decision:** `riftweave-server` emits traces, metrics, and logs with the OpenTelemetry Go SDK, in `backend/cmd/riftweave-server` only ([observability.md#the-server](../backend/observability.md#the-server)), and answers `/healthz`, `/readyz`, and `/version` ([api.md#health](../backend/api.md#health)).
- **Versions pinned:** otel v1.47.0 (the log exporters v0.23.0), contrib v0.72.0, otelslog v0.21.0. `go get` also raised `golang.org/x/text` to v0.42.0 and added gRPC's protobuf modules the OTLP exporters need. [V 2026-11-23, `go test ./...`]
- **A refusal's kind, never its message:** a core refusal can repeat the player's input (`actor has no skill "..."`), so spans, metrics, and logs carry a fixed `error.type` only. [V 2026-11-23, the test types a skill name and finds it in no record]
- **Off by default:** an endpoint exports OTLP/HTTP; `--telemetry <file>` writes the stdout exporters' JSON, which is the SDK's own shape, not OTLP/JSON [A]; `OTEL_SDK_DISABLED=true` wins. A server with none writes only its slog JSON lines.
- **Tests do not observe:** `newHandler` stays unobserved, so the existing server tests run as before; `newObservedHandler` wraps it in `otelhttp` and a 4xx/5xx log line.
- **The core stays pure:** no OpenTelemetry package under `backend/core` or `backend/internal`. [V 2026-11-23, `TestTelemetryStaysOutOfTheCore`, and `go list -deps`]
- **Open:** writing the `--telemetry` file as OTLP/JSON, the diagnostics shape the settings page describes; `/readyz` turning `not ready` while a reload runs, once the server reloads rules.
