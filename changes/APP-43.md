### APP-43 Telemetry in the OpenTelemetry format, and a telemetry setting both apps share
- Added: `docs/backend/observability.md`, the OpenTelemetry resource, spans, metrics, and log records each component emits, and that none is about the player.
- Added: the `telemetry` key of the settings file (`mode` off, local, or export; `endpoint`), checked alike in Go and TypeScript, `off` by default.
- Added: a test that keeps OpenTelemetry out of the core.
