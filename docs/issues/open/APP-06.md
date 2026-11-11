#### APP-06 HTTP API server
- Status: todo
- Priority: medium
- Needs: APP-02, APP-04
- Do: Serve the core over HTTP for consumers: one endpoint per command, session endpoints, and an OpenAPI document generated from `backend/contract/`, not written by hand. Stateless calls need no session.
- Keys: OpenAPI document; endpoint per command
- Files: ~6 (backend/cmd/riftweave-server/, docs/backend/api.md)
- See: docs/backend/README.md
- Done when: every command is callable over HTTP; tests run a vector through the API and compare the JSON; the OpenAPI document validates and lists every command.
