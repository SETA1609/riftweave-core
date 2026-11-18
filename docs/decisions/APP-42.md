# APP-42 File the observability and dev console issues, and raise the data-integrity priorities

Date: 2026-11-18

- **Decision:** observability uses the OpenTelemetry format (owner, 2026-11-18): its data model for logs, traces, and metrics, OTLP/JSON on disk and OTLP/HTTP on the wire, and W3C Trace Context between the apps and the server. APP-43 defines the records and a `telemetry` setting (`off`, `local`, `export`); APP-44 (server), APP-45 (CLIs), and APP-46 (frontend) emit them; APP-47 is a local stack to look at them.
- **Offline first and private** [A]: telemetry is off by default and, when on, kept on the device unless the setting says `export`; no record names an actor, a session, or a save. The core imports no telemetry package; only the adapters observe the calls they make, as the Charm rule keeps the TUI out of the core.
- **Dev console:** APP-48 exposes the frontend's context, API, saves, and router as `globalThis.rw` in dev builds only, since `globalThis.riftweave` is the Go core's in the page.
- **Priorities for data integrity** (owner, 2026-11-18): M5-70 (1NF), M5-71 (2NF), and M5-74 (3NF, monster HP) are high; M5-75 (3NF, descriptions), M5-58, and M5-59 (the base perk) are medium. M5-40 and M5-52 stay as they are: they touch every file and are best done when few branches are open.
- **Not filed:** an SSH app; APP-45 says it needs its own issue first.
