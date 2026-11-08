# APP-03 Agent CLI

Date: 2026-11-08

- **Decision:** `riftweave`, built from `backend/cmd/riftweave`, is the agent CLI: `list`,
  `describe <command>`, and `<command> [file]` with JSON from the file or stdin, the result
  as one line of JSON on stdout. Spec: `docs/backend/cli.md`. **[V 2026-11-08,
  `go test ./cmd/riftweave`: every vector runs through the CLI and matches `core.Run`]**
- **Decision:** exit codes 0 ok, 1 rule error, 2 bad input or usage, 3 no ruleset. An error is
  one JSON line on stderr; a rule error is the core's `{"error": ...}`, the sentence a vector
  pins. **[V 2026-11-08]**
- **Decision:** `backend/contract/` gains a small Go package that embeds its JSON files, so
  `list` and `describe` work from the binary alone. The ruleset is read from `--data`, then
  `RIFTWEAVE_DATA`, then the first `ruleset/data` above the current directory. **[A]**
- **Why:** an agent needs one fast command to call in a loop; the core was only callable from
  Go.
- **Open:** each call loads the ruleset again (about 6 ms a call on the owner's machine, 50
  calls in 0.29 s) **[V 2026-11-08]**; a long-running mode or the MCP server (APP-09) would
  load it once. The CLI draws no rolls; seeded rolls come with sessions (APP-04) **[A]**.
  Machine-readable error codes are still not specified; the exit code is the only
  classification **[U]**.
