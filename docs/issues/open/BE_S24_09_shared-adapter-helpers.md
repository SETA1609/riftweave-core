#### BE_S24_09_shared-adapter-helpers One copy of the helpers the binaries share
- Status: todo
- Priority: low
- Needs: —
- Modified: 2026-11-24
- Do: Owner request, 2026-11-24: the only copied code the screening found is in the binaries. `rulesetDirectory` is in the agent CLI and the server, `retarget` in the OpenAPI writer and the MCP server, and `moduleRows` in the files command (and the terminal app, which S23 removes). Each moves to one place both read.
- Keys: the shared helpers
- Files: ~7 (backend/cmd/riftweave/main.go, backend/cmd/riftweave-server/main.go, backend/cmd/riftweave-server/openapi.go, backend/mcp/mcp.go, backend/cmd/riftweave/files/files.go, docs/decisions/BE_S24_09_shared-adapter-helpers.md, changes/BE_S24_09_shared-adapter-helpers.md)
- See: backend/cmd/
- Done when: each helper has one definition; the CLI, server, and MCP tests pass.
