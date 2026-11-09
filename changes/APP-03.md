### APP-03 Agent CLI
- Added: the `riftweave` agent CLI (`backend/cmd/riftweave`): `list`, `describe <command>`, and `<command> [file]` with JSON in and out; errors as JSON on stderr, with exit codes 0 ok, 1 rule error, 2 bad input, 3 no ruleset; `docs/backend/cli.md`.
- Added: `backend/contract` embeds the contract files in the binary.
