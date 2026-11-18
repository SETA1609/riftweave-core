# APP-43 Telemetry in the OpenTelemetry format, and a telemetry setting both apps share

Date: 2026-11-18

- **Decision:** every component reports in OpenTelemetry's data model as OTLP/JSON, exported by OTLP/HTTP; `docs/backend/observability.md` names the resource, the span per command, one duration histogram, and the log records each component emits. The settings file gains `telemetry` (`mode` off, local, or export; `endpoint`), checked alike by `backend/internal/settings` and `frontend/src/app/settings-file.ts`, with `off` in both apps' defaults. [V 2026-11-18, Go settings tests and the shared fixture]
- **Sources:** the OTLP JSON encoding, the exporter's default endpoint and environment variables, and the stable `service.*` and `error.type` attributes were read from shallow clones of the three OpenTelemetry repositories at the commits named in observability.md. `devops-tools fetch-docs github` was refused (HTTP 403 from the GitHub API in the sandbox), so the sources are not in `cache/docs/`. [V 2026-11-18, read; the cache is the gap]
- **`endpoint` must be http or https:** the schema and both checks refuse anything else, so a typo (`collector:4318`) fails on import rather than at export. [A]
- **`error.type` is the refusal's code:** low-cardinality, as the convention asks; a refusal also sets the span status to ERROR. [A]
- **What this issue does not emit:** no component sends anything yet. The server's emitters are APP-44; the dev console and the diagnostics file are APP-45 [A]. This issue is the format and the setting.
- **The purity test:** `TestTelemetryStaysOutOfTheCore` sits beside the Charm test and fails when `go list -deps` of the core and `internal/` names `go.opentelemetry.io`. [V 2026-11-18, added; it passes today, since nothing imports OpenTelemetry]
- **Not run here:** the Vitest spec and the command tests that need module downloads; the sandbox has no npm or Go module network. CI runs them. [U]
