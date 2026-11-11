# APP Reference backend

Roadmap: [APP](../../ROADMAP.md#app-reference-backend). Format and rules: [README.md](README.md).

## APP Reference backend

**Goal:** one fast backend, written in Go, serves the ruleset, built alongside the rules rather than after them: every rule the vectors pin can be called as a command, by an agent CLI, a human CLI, an HTTP API, or the Angular frontend, with the same inputs, results, and rolls. It is a second implementation of the kernel, checked against every vector in CI, so it cannot drift from the rules. It adds no rules; the vectors stay the authority.

**Spec:** `docs/backend/` (written in APP-01). Rules stay in `docs/core/` and `docs/modules/`.

**Done when:** every vector passes under the Python reference and under the Go backend in CI; the agent CLI, the human CLI, and the API return the same result for the same input and rolls; the frontend can create a character and resolve an attack through the API. Not a 1.0 gate, except the issues the alpha playtest (M5-18) runs on: APP-04, APP-05, and APP-08.
