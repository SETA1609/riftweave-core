# Frontend

Modified: 2026-11-15

`frontend/` is an Angular app over the core (APP-07): through the HTTP API ([api.md](api.md)), or
through the Go core compiled to WebAssembly in the page, so it can be served as a static site with
no server (APP-14). It is an adapter like the CLIs: it builds a command's input from a form, sends
it, and shows the result. It holds no rule. A number it shows comes from a response, never from a
formula of its own. [A]

## The client

The client is typed from the server's own OpenAPI document, never by hand:

```
npm run generate   GET /openapi.json from a running riftweave-server
                   → openapi-typescript → src/api/openapi.d.ts (checked in)
src/app/api.ts     openapi-fetch over those types: one function per call the app makes
```

The generated file is checked in, so the app builds without a server; `npm run generate` refreshes
it after a contract change, and a stale file fails the type check where a field moved. [A]
`npm run generate` runs `openapi-typescript` 7.13.0 through `npx`, not as a dependency: it pins
TypeScript 5 as a peer, and Angular 22 needs TypeScript 6. [V 2026-11-14]
`api.ts` sends every call through a transport: `httpTransport(baseUrl)` over openapi-fetch, or
`wasmTransport(core)`, which maps the same paths onto the commands the server would run
(`/v1/commands/<name>` to `<name>`, the session paths to `session` with their `operation`). The
build chooses: angular.json `define`s `RIFTWEAVE_API`, `null` by default (WebAssembly) and `''` in
the `api` configuration (the dev proxy). The server sends no CORS headers (api.md, Not covered), so
`ng serve --configuration api` proxies `/v1` and `/openapi.json` to `127.0.0.1:8080`
(`proxy.conf.json`).

## The core in the browser

How WebAssembly works, and why this is built the way it is, with the upstream sources quoted:
[docs/theory/webassembly.md](../theory/webassembly.md).

```
./scripts/build-wasm.sh   copy ruleset/data and ruleset/modules into backend/cmd/riftweave-wasm/ruleset
                          GOOS=js GOARCH=wasm go build → frontend/public/riftweave.wasm
                          copy the same Go's lib/wasm/wasm_exec.js → frontend/public/
src/app/wasm.ts           load wasm_exec.js, instantiateStreaming(fetch('riftweave.wasm')), go.run,
                          wait for riftweaveReady; run(command, input) → result or Error
```

- **One entry point.** `backend/cmd/riftweave-wasm` sets `globalThis.riftweave.run(command,
  input)` over `core.Run`, the function the CLI, the HTTP API, and MCP call. `dev` is refused, as
  over HTTP. A refusal comes back as the rule's message, the same text the API sends in `error`.
- **The data is embedded,** read through the loader's `fs.FS`, so the core and the data in one
  `.wasm` file come from one commit. The copy and the build outputs are gitignored.
- **Size:** `riftweave.wasm` is 9,326,790 bytes, 2,321,453 gzipped (Go 1.26.1, `-trimpath
  -ldflags="-s -w"`); about 1 MB is the data. The app shows "Starting the rules engine…" until
  the core is ready, about 0.1 s after the file compiles. [V 2026-11-15]

## Screens

| Screen | Calls | Shows |
| --- | --- | --- |
| Character sheet | `derived` on the attributes and level; `POST /v1/sessions` to create the character | every derived stat, and the stored actor of the new session |
| Attack resolver | `combat` with the attacker (skill, roll, LCK), the defender (evasion or block, armor DR, HP, reaction), the delivery, and the damage roll | each step of the result: the target, the margin, crit, the reaction, the damage after DR, and HP after |
| Playthrough | `POST /v1/sessions/step` for each step of `ruleset/tests/vectors/playthrough.json`, then `replay` | each step's input and result, and whether the replay matches |

Every roll is an input field, as in the vectors: the app rolls nothing unless asked, and a rolled
value is written into the field before the call, so a result can be read back against a vector.
The playthrough runner sends each vector's input as a session step, as given, and all twelve
steps match their vectors and the replay matches. [V 2026-11-14, by hand against the server]

## Tests

`npm test` runs the specs (Vitest, through `@angular/build:unit-test`). The API specs run twice,
through each transport: a real `riftweave-server` at `RIFTWEAVE_API` (default
`http://127.0.0.1:8080`), and the WebAssembly core read from `public/` (run
`./scripts/build-wasm.sh` first). Both must give the same results:

- **Create a character:** `derived` returns every stat, and `POST /v1/sessions` returns a session
  whose actor is the one sent.
- **Resolve an attack:** the first vectors of `ruleset/tests/vectors/combat.json` go through the
  app's attack call, and each result matches the vector's `expect`.

`npm run e2e` runs the end-to-end specs (`e2e/*.spec.ts`, Playwright, Chromium) twice: the
`chromium` project against the app served by `ng serve --configuration api` and its proxy, with a
real `riftweave-server` behind it, and the `chromium-wasm` project against the default build on
port 4300, with no proxy, where every call runs in the page. All three are started by
`playwright.config.ts` unless one is already running, so `./scripts/app.sh` and CI reuse theirs.
[V 2026-11-15, `npm run e2e` locally: 6 specs pass, 3 per project]
The specs act as a player: fill the character sheet and create a character, fill the attack form
with the first `combat.json` vector's inputs and read the shown result against its `expect`, and run
the playthrough to a matching replay. They find elements by label and role, not by CSS class, so a
restyle does not break them. A first run installs the browser: `npx playwright install chromium`.
[V 2026-11-14, `npm run e2e` locally: 3 specs pass, both servers started and stopped by Playwright]

CI (`.github/workflows/frontend.yml`) builds the server and the WebAssembly core, starts the
server, and runs `npm ci`, the build, and the tests on every change under `frontend/`, `backend/`,
or `ruleset/`. [V 2026-11-14, the
workflow's steps run by hand: 14 specs pass]

The toolchain is Node 24.15 or later (the Angular 22 CLI's minimum; CI pins 24.22.1) and npm
11.21.1 for installs: the npm that ships with Node 24.7 fails to resolve Vitest's optional peers.
npm 11.21.1 keeps install scripts blocked, its default. [V 2026-11-14]

## Running it

| Command | Needs | Serves |
| --- | --- | --- |
| `./scripts/app.sh [--port 4200] [--api-port 8080]` | Go and Node (with Node below 24.15, it runs Node 24.22.1 through `npx`) | `ng serve` at `localhost:4200`, proxying `/v1` and `/openapi.json` to a `riftweave-server` it builds and starts; reloads on a frontend change |
| `./scripts/build-wasm.sh`, then `npm run start:wasm` in `frontend/` | Go and Node | `ng serve` at `localhost:4200` with no server: the core runs in the page as WebAssembly |
| `./scripts/app-container.sh [--port 8080] [--build-only]` | docker or podman (`RIFTWEAVE_ENGINE` picks one) | one image with the server, the built app, and the ruleset data behind nginx, at `localhost:8080` |

`app.sh` builds into `cache/` and installs the frontend's packages (npm 11.21.1, `npm ci`) only
when `node_modules` is missing or older than the lockfile. Ctrl-C stops the server and the app.
The image (`container/app.Dockerfile`) is built in three stages, the server (Go), the app (Node),
and nginx, which serves the app and proxies the API on one origin, so the server needs no CORS.
Both bind to localhost only. Image names are fully qualified (`docker.io/library/...`,
`localhost/riftweave-app:local`), since podman does not resolve short names. Rootless podman also
needs subordinate ids for the user in `/etc/subuid` and `/etc/subgid`; the script checks and
prints the fix. [V 2026-11-14: `app.sh`, and `app-container.sh` with docker, each served the app
and the first combat vector's result; with podman, only the check ran, see docs/decisions/APP-21.md]

## Conventions

Every change to `frontend/` follows these (owner, 2026-11-14):

- **Angular 22, standalone components.** No NgModules.
- **Signals first.** State, derived values, inputs, outputs, and two-way bindings are signals
  (`signal`, `computed`, `input`, `output`, `model`). RxJS only where signals do not fit: streams,
  cancellation, debouncing, a Realtime or WebSocket feed.
- **Three files per component.** `name.ts`, `name.html`, `name.css`, joined with `templateUrl` and
  `styleUrl`. No inline `template` or `styles`.
- **Bootstrap's look, in our own CSS.** No Bootstrap package, CDN, or script. `src/styles.css` holds
  the design tokens (custom properties: colors, spacing, radius, font) and the few shared classes
  (`.btn`, `.btn-primary`, `.btn-outline`, `.form-control`, `.card`, `.alert`, `.table`, `.nav`); a
  component's own layout is in its own `.css`.
- **Small reusable components, built up.** Fields, cards, alerts, and lists live in `src/app/ui/`
  and take their values as signal inputs and models; a screen is a composition of them. Add a new
  `ui/` component when a second screen needs the same piece, rather than repeating markup.
- **Tests.** Unit and API specs run on Vitest (`npm test`, through `@angular/build:unit-test`),
  beside what they test (`name.spec.ts`; `app.spec.ts` covers the ui components and the screens).
  End-to-end tests run on Playwright (`npm run e2e`,
  `e2e/*.spec.ts`): a real browser on the built app, the app's proxy, and a real server.
- **Layout.** `src/app/api.ts` is the API and its `API` token, `src/app/ui/` the shared components,
  and each screen its three files in `src/app/` (`sheet`, `attack`, `playthrough`). [A]

## Not covered

Styling past a readable layout, a saved list of characters, and the human CLI's level-up flow.
The app reads the session JSON the server returns and keeps it in memory only. [A]
