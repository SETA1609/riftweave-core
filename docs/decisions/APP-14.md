# APP-14 The core runs in the browser as WebAssembly

Date: 2026-11-15
Modified: 2026-11-15

- **Decision:** `backend/cmd/riftweave-wasm` builds with the standard Go toolchain
  (`GOOS=js GOARCH=wasm`) and sets one JavaScript function, `riftweave.run(command, input)`,
  over `core.Run`. The ruleset data and modules are copied beside it by `scripts/build-wasm.sh`
  and embedded with `go:embed`, read through the loader's `fs.FS`, so a `.wasm` file holds the
  core and the data of one commit. `dev` is refused, as over HTTP. **[V 2026-11-15, `go test
  ./cmd/riftweave-wasm` runs every golden vector through `call`]**
- **Decision:** `api.ts` sends every call through a transport, HTTP or WebAssembly, chosen at
  build time by angular.json's `define` of `RIFTWEAVE_API`: the default build has none, so it
  runs the core in the page (a static site, APP-15); the `api` configuration uses the dev proxy.
  `npm start`, `scripts/app.sh`, and the container image use `api`, so they behave as before.
  **[V 2026-11-15, `npm test` 41 specs over both transports; `npm run e2e` 6 specs, 3 per
  project]**
- **Why the standard toolchain, not TinyGo:** it builds the core unchanged and passes every
  vector; TinyGo supports a subset of the standard library, and the core uses `encoding/json`
  reflection and `html/template`. **[A]**
- **Size:** 9,326,790 bytes, 2,321,453 gzipped, about 1 MB of it data; the core is ready about
  0.1 s after the module compiles. **[V 2026-11-15]**
- **Theory:** `docs/theory/webassembly.md` quotes the WebAssembly specification (commit
  `943c5df`), MDN (commit `742f106`), the Go wiki, and `syscall/js` at go1.26.1, from copies
  fetched with `devops-tools fetch-docs`.
- **Difference from the issue:** no schema step; the contract and the OpenAPI types are
  unchanged, since both transports follow them.
- **Risk:** `syscall/js` is experimental and outside Go's compatibility promise, and
  `wasm_exec.js` must come from the Go that built the module; the build script copies both from
  one toolchain, and CI builds them on every run. A 2.3 MB download on a slow connection is a few
  seconds of "Starting the rules engine…"; Brotli or TinyGo could cut it later. **[A]**
