# APP Reference backend

Roadmap: [APP](../../ROADMAP.md#app-reference-backend). Format and rules: [README.md](README.md).

## APP Reference backend

**Goal:** one fast backend, written in Go, serves the ruleset, built alongside the rules rather than after them: every rule the vectors pin can be called as a command, by an agent CLI, a human CLI, an HTTP API, or the Angular frontend, with the same inputs, results, and rolls. It is a second implementation of the kernel, checked against every vector in CI, so it cannot drift from the rules. It adds no rules; the vectors stay the authority.

**Spec:** `docs/backend/` (written in APP-01). Rules stay in `docs/core/` and `docs/modules/`.

**Done when:** every vector passes under the Python reference and under the Go backend in CI; the agent CLI, the human CLI, and the API return the same result for the same input and rolls; the frontend can create a character and resolve an attack through the API. Not a 1.0 gate, except the issues the alpha playtest (M5-18) runs on: APP-04, APP-05, and APP-08.

Done: APP-01 ([done.md](done.md#app-reference-backend))

#### APP-02 Go core: the kernel in Go, every vector in CI
- Status: review
- Priority: medium
- Needs: APP-01
- Do: A Go module in `backend/` that loads `ruleset/data/` (and module data), evaluates formula trees, and implements every command in the contract. It reads coefficients from the data, never from constants. A Go test runs every file in `ruleset/tests/vectors/` and fails on any disagreement, and CI runs it on every PR. It starts with the commands that exist and grows with the rules: from then on a rule issue that adds or changes a function updates the Python reference, the Go core, and the contract in the same pull request; CONTRIBUTING says so.
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
- Needs: APP-02
- Do: Commands are stateless. Add a session: an actor snapshot (HP, MP, AP, XP, level, inventory, conditions) that commands read and update, so a chain of commands (create, buy, fight, loot, rest, level, cast) runs as calls. The chain grows with the rules; M5-04's playthrough runs through it once it lands. Sessions serialize to JSON so a run can be saved, replayed, or turned into vectors.
- Keys: session object and its schema in `backend/contract/`; `session` input on commands
- Files: ~6 (backend/, backend/contract/, docs/backend/sessions.md)
- See: docs/core/loop.md, docs/stories/05-release.md (M5-04)
- Done when: a chain of the commands that exist (at least derived, combat, rest, progression) runs as backend calls and matches their vectors; a saved session replays to the same result.

#### APP-05 Human CLI
- Status: todo
- Priority: low
- Needs: APP-03
- Do: A readable, interactive mode of the same binary, built as a terminal UI with the Charm libraries (Bubble Tea for the program loop, Bubbles for inputs and tables, Lip Gloss for styling): prompts for inputs, tables for results, a short log of the rolls. It calls the same core as the agent CLI; it holds no rule. The agent CLI path imports none of it, so its start stays fast, and `backend/core` and `backend/internal/` stay free of Charm.
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

#### APP-08 Dev mode: inflict state on a session for testing
- Status: todo
- Priority: medium
- Needs: APP-03, APP-04
- Do: A dev mode in both CLIs, off unless the binary runs with `--dev`, that changes a session's actor directly so a tester can reach a state without playing to it. Operations: set or adjust a resource (HP, MP, stamina, AP); grant XP or set the level; set an attribute or skill; grant or remove a perk; learn a spell or an effect; add or remove an item or gold; inflict or clear a condition or an effect (any effect from the registry, with magnitude, duration, and an origin from `origins/core.json`, so it stacks by the real rule); advance time in turns or rests; fix the next rolls (a d100 sequence) or the seed. Each operation runs through the same core and checks its refs like any command; it holds no rule. Agent CLI: `riftweave dev <operation>` with JSON in and out. Human CLI: a dev panel and short commands (`:give`, `:inflict`, `:level`, `:roll`). The HTTP API does not serve dev operations. Each operation is logged in the session, which is marked `dev: true`, so a replay repeats it and a dev run is never taken for a played one.
- Keys: `--dev` flag; `dev` command and its operations in `backend/contract/`; `dev` mark and log on the session
- Files: ~6 (backend/cmd/riftweave/, backend/core/dev/, backend/contract/, docs/backend/dev-mode.md, docs/backend/cli.md)
- See: docs/stories/07-backend.md (APP-03, APP-04, APP-05), ruleset/data/origins/core.json
- Done when: without `--dev`, every dev operation is refused with a JSON error; with it, each operation has a test; an inflicted effect stacks with a real one by its origin's policy; a bad ref is refused; a dev session replays to the same result and stays marked `dev: true`; the human CLI and the agent CLI give the same result for the same operations.
