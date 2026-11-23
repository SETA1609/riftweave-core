### BE_S12_02_server-observability The server is observable: OpenTelemetry traces, metrics, and logs, and health endpoints
- Added: `riftweave-server` traces each request (continuing an incoming `traceparent`) with a child span per command, measures `riftweave.command.duration` and the Go runtime, and logs JSON lines with the trace id, bridged to OpenTelemetry logs; off unless `OTEL_EXPORTER_OTLP_ENDPOINT` or `--telemetry <file>`.
- Added: `GET /healthz`, `/readyz`, and `/version`, in the OpenAPI document.
- Changed: `golang.org/x/text` v0.14.0 to v0.42.0, raised by the OpenTelemetry modules.
