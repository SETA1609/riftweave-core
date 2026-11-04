# APP Reference backend

Roadmap: [APP](../../ROADMAP.md#app-reference-backend). Format and rules: [README.md](README.md).

## APP Reference backend

**Goal:** one fast backend, written in Go, serves the ruleset: every rule the vectors pin can be called as a command, by an agent CLI, a human CLI, an HTTP API, or the Angular frontend, with the same inputs, results, and rolls. It is a second implementation of the kernel, checked against every vector in CI, so it cannot drift from the rules. It adds no rules; the vectors stay the authority.

**Spec:** `docs/backend/` (written in APP-01). Rules stay in `docs/core/` and `docs/modules/`.

**Done when:** every vector passes under the Python reference and under the Go backend in CI; the agent CLI, the human CLI, and the API return the same result for the same input and rolls; the frontend can create a character and resolve an attack through the API. Not a 1.0 gate.

Done: —

#### APP-01 Command contract: commands and their schemas as data
- Status: doing
- Priority: medium
- Needs: M1-02
- Do: Define the backend's interface without choosing a language, so Go, both CLIs, the OpenAPI document, and the frontend are all built from one source. One file per command in `backend/contract/` (name, summary, input schema, result schema, error shape), one per reference runner (`resolution`, `combat`, `derived`, `progression`, `actions`, `formula`, `modifiers`, and each runner that lands later). Rolls are inputs, or are drawn from a seed and echoed in the result as `rolls`, so any call can be saved as a vector. A Python test checks every vector's input and the reference's result against the command's schemas. Write `docs/backend/README.md`: the layers (contract, Go core, adapters, frontend), the error shape, and that no adapter holds a rule.
- Keys: `backend/contract/<command>.json`; command `name`, `input_schema`, `result_schema`; result field `rolls`
- Files: ~12 (backend/contract/, docs/backend/README.md, ruleset/scripts/test_contract.py)
- See: ruleset/scripts/kernel_reference.py, docs/core/vectors.md
- Done when: every runner has a contract file; every vector input and every reference result validate against its command's schemas; a vector function without a contract file fails the test.

#### APP-02 Go core: the kernel in Go, every vector in CI
- Status: todo
- Priority: medium
- Needs: APP-01, M1-01, M1-04, M1-05
- Do: A Go module in `backend/` that loads `ruleset/data/` (and module data), evaluates formula trees, and implements every command in the contract. It reads coefficients from the data, never from constants. A Go test runs every file in `ruleset/tests/vectors/` and fails on any disagreement, and CI runs it on every PR. It starts after M1 closes, so the kernel stops changing before it is built twice. From then on a rule issue's data step updates the Python reference and Go together; CONTRIBUTING says so.
- Keys: Go module `backend/`; package per kernel area (formula, resolution, combat, ...)
- Files: ~15 (backend/, .github/workflows/, CONTRIBUTING.md, docs/backend/README.md)
- See: docs/core/vectors.md, backend/contract/
- Done when: `go test ./...` passes every vector the Python reference passes; CI runs it; a data change that breaks a rule fails both.

#### APP-03 Agent CLI
- Status: todo
- Priority: medium
- Needs: APP-02
- Do: A single static Go binary for agents and scripts: `riftweave list`, `riftweave describe <command>` (prints the contract), and `riftweave <command>` reading JSON from stdin or a file and writing JSON to stdout. Errors are JSON on stderr with a non-zero exit code. No prompts, no colour, fast start, so an agent can call it thousands of times in a loop.
- Keys: `riftweave` binary; exit codes
- Files: ~4 (backend/cmd/riftweave/, docs/backend/cli.md)
- See: docs/backend/README.md
- Done when: every command runs from the CLI; tests run a vector through the CLI and compare the JSON; a bad input returns a JSON error and exit code 2.

#### APP-04 Session state and the play chain
- Status: todo
- Priority: medium
- Needs: APP-02, M5-04
- Do: Commands are stateless. Add a session: an actor snapshot (HP, MP, AP, XP, level, inventory, conditions) that commands read and update, so a chain like M5-04's playthrough (create, buy, fight, loot, rest, level, cast) runs as calls. Sessions serialize to JSON so a run can be saved, replayed, or turned into vectors.
- Keys: session object and its schema in `backend/contract/`; `session` input on commands
- Files: ~6 (backend/, backend/contract/, docs/backend/sessions.md)
- See: docs/core/loop.md, docs/stories/05-release.md (M5-04)
- Done when: the M5-04 playthrough runs as a sequence of backend calls and matches its vectors; a saved session replays to the same result.

#### APP-05 Human CLI
- Status: todo
- Priority: low
- Needs: APP-03
- Do: A readable, interactive mode of the same binary: prompts for inputs, tables for results, a short log of the rolls. It calls the same core as the agent CLI; it holds no rule.
- Files: ~4 (backend/cmd/riftweave/, docs/backend/cli.md)
- See: docs/backend/cli.md
- Done when: a person can create a character and resolve an attack interactively; the result matches the agent CLI for the same inputs and rolls.

#### APP-06 HTTP API server
- Status: todo
- Priority: medium
- Needs: APP-02, APP-04
- Do: Serve the core over HTTP for consumers: one endpoint per command, session endpoints, and an OpenAPI document generated from `backend/contract/`, not written by hand. Stateless calls need no session.
- Keys: OpenAPI document; endpoint per command
- Files: ~6 (backend/cmd/riftweave-server/, docs/backend/api.md)
- See: docs/backend/README.md
- Done when: every command is callable over HTTP; tests run a vector through the API and compare the JSON; the OpenAPI document validates and lists every command.

#### APP-07 Angular frontend
- Status: todo
- Priority: low
- Needs: APP-06
- Do: An Angular app in `frontend/` that talks only to the API, with a client generated from the OpenAPI document: a character sheet, an attack resolver (attacker, defender, reaction, delivery, rolls), and a playthrough runner over a session. It shows the rolls and the step-by-step result of each pipeline.
- Files: ~15 (frontend/, docs/backend/frontend.md, .github/workflows/)
- See: docs/backend/api.md
- Done when: the app builds and its tests pass in CI; it creates a character and resolves an attack through the API; the attack result matches the vector for the same inputs.
