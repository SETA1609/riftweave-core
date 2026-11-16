### APP-14 The core runs in the browser as WebAssembly
- Added: `backend/cmd/riftweave-wasm` and `scripts/build-wasm.sh`: the Go core with the ruleset embedded, built to WebAssembly (9.3 MB, 2.3 MB gzipped), with one `riftweave.run(command, input)` entry point over `core.Run`.
- Changed: the frontend's default build runs the core in the page and needs no server; the `api` configuration (`npm start`, `scripts/app.sh`, the container) keeps using the HTTP API. The specs and the Playwright tests run through both.
- Added: `docs/theory/webassembly.md`, how WebAssembly and Go's js/wasm target work, with the upstream sources quoted.
