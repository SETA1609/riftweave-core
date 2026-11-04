# APP Reference backend

Roadmap: [APP](../../ROADMAP.md#app-reference-backend). Format and rules: [README.md](README.md).

## APP Reference backend

**Goal:** one backend serves the ruleset: every rule the vectors pin can be called as a command, by an agent CLI, a human CLI, an HTTP API, or the Angular frontend, with the same inputs, results, and rolls. It adds no rules; the vectors stay the authority.

**Spec:** `docs/backend/` (written in APP-01). Rules stay in `docs/core/` and `docs/modules/`.

**Done when:** every vector runs through the backend's command registry and passes; the agent CLI, the human CLI, and the API return the same result for the same input and rolls; the frontend can create a character and resolve an attack through the API. Not a 1.0 gate.

Done: —

#### APP-01 Backend core: one command registry over the reference
- Status: todo
- Priority: medium
- Needs: M1-02
- Do: Add a Python package in `backend/` with a command registry: one command per reference runner (`resolution`, `combat`, `derived`, `progression`, `actions`, `formula`, `modifiers`, and each runner that lands later), each with a JSON Schema for its input and result. Rolls are inputs, or are drawn from a seeded RNG and echoed in the result, so any call can be saved as a vector. The registry wraps `kernel_reference.RUNNERS` and does not move them, so open runner branches do not conflict. Write `docs/backend/README.md`: the layers (registry, adapters, frontend), the command and error shapes, and that no adapter holds a rule.
- Keys: `backend/` package; command `name`, `input_schema`, `result_schema`; result field `rolls`
- Files: ~8 (backend/, docs/backend/README.md, ruleset/scripts/test_vectors.py, .github/workflows/validate-ruleset.yml)
- See: ruleset/scripts/kernel_reference.py, docs/core/vectors.md
- Done when: every vector file also runs through the registry and passes; every command's input and result validate against its schemas; a seeded call returns the rolls it used; backend tests run in CI.

#### APP-02 Agent CLI
- Status: todo
- Priority: medium
- Needs: APP-01
- Do: A non-interactive CLI for agents and scripts: `riftweave list`, `riftweave describe <command>` (prints the schemas), and `riftweave <command>` reading JSON from stdin or a file and writing JSON to stdout. Errors are JSON on stderr with a non-zero exit code. No prompts and no colour.
- Keys: `riftweave` entry point; exit codes
- Files: ~4 (backend/, docs/backend/cli.md)
- See: docs/backend/README.md
- Done when: every command runs from the CLI; tests run a vector through the CLI and compare the JSON; a bad input returns a JSON error and exit code 2.

#### APP-03 Session state and the play chain
- Status: todo
- Priority: medium
- Needs: APP-01, M5-04
- Do: Commands are stateless today. Add a session: an actor snapshot (HP, MP, AP, XP, level, inventory, conditions) that commands read and update, so a chain like M5-04's playthrough (create, buy, fight, loot, rest, level, cast) runs as calls. Sessions serialize to JSON so a run can be saved, replayed, or turned into vectors.
- Keys: session object and its schema; `session` input on commands
- Files: ~6 (backend/, docs/backend/sessions.md, ruleset/tests/vectors/playthrough.json)
- See: docs/core/loop.md, docs/stories/05-release.md (M5-04)
- Done when: the M5-04 playthrough runs as a sequence of backend calls and matches its vectors; a saved session replays to the same result.

#### APP-04 Human CLI
- Status: todo
- Priority: low
- Needs: APP-02
- Do: A readable, interactive front door over the same commands: prompts for inputs, tables for results, a short log of the rolls. It calls the registry like the agent CLI; it holds no rule.
- Files: ~4 (backend/, docs/backend/cli.md)
- See: docs/backend/cli.md
- Done when: a person can create a character and resolve an attack interactively; the result matches the agent CLI for the same inputs and rolls.

#### APP-05 HTTP API server
- Status: todo
- Priority: medium
- Needs: APP-01, APP-03
- Do: Serve the registry over HTTP for consumers: one endpoint per command, session endpoints, and an OpenAPI document generated from the command schemas, not written by hand. Stateless calls need no session.
- Keys: OpenAPI document; endpoint per command
- Files: ~6 (backend/, docs/backend/api.md)
- See: docs/backend/README.md
- Done when: every command is callable over HTTP; tests run a vector through the API and compare the JSON; the OpenAPI document validates and lists every command.

#### APP-06 Angular frontend
- Status: todo
- Priority: low
- Needs: APP-05
- Do: An Angular app in `frontend/` that talks only to the API: a character sheet, an attack resolver (attacker, defender, reaction, delivery, rolls), and a playthrough runner over a session. It shows the rolls and the step-by-step result of each pipeline.
- Files: ~15 (frontend/, docs/backend/frontend.md, .github/workflows/)
- See: docs/backend/api.md
- Done when: the app builds and its tests pass in CI; it creates a character and resolves an attack through the API; the attack result matches the vector for the same inputs.
