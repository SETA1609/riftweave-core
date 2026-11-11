#### APP-03 Agent CLI
- Status: done
- Merged: #98, 2026-11-09
- Priority: medium
- Needs: APP-02
- Do: A single static Go binary for agents and scripts: `riftweave list`, `riftweave describe <command>` (prints the contract), and `riftweave <command>` reading JSON from stdin or a file and writing JSON to stdout. Errors are JSON on stderr with a non-zero exit code. No prompts, no colour, fast start, so an agent can call it thousands of times in a loop.
- Keys: `riftweave` binary; exit codes
- Files: ~4 (backend/cmd/riftweave/, docs/backend/cli.md)
- See: docs/backend/README.md
- Done when: every command runs from the CLI; tests run a vector through the CLI and compare the JSON; a bad input returns a JSON error and exit code 2.
