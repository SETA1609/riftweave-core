# Frontend

Modified: 2026-11-23

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
it after a contract change. The frontend workflow regenerates it against the server it starts and
fails when the checked-in file differs, keeping the regenerated file as the `openapi-types`
artifact, so a contract change cannot leave it behind (FE_S00_05). [V 2026-11-20, this workflow]
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

The app opens on the home page, the main menu of [app-layout.md](app-layout.md), and the Angular
router shows one screen at a time (APP-27). The routes use the hash (`#/play/attack`), so the
static Pages build reloads any page and works offline with no server-side fallback. [A]

| Route | Shows |
| --- | --- |
| `/` | the home page: Start, Characters (APP-33), Settings, Modules, Network |
| `/play` | the tools: GM tools, Character creation, Spell creation (APP-22), Skill resolution, Bestiary (APP-29), Saved games |
| `/play/gm_tools` | Resolve an attack, Encounter tracker (APP-34), Playthrough |
| `/play/<tool>` | the tool's screen below |
| `/settings`, `/modules`, `/network` | the settings file of APP-25, one screen each |

The three menus are one component, `rw-menu-grid` (`ui/menu-grid`), which reads
`backend/contract/menu.json` and shows a menu's entries as cards; a pending entry is a disabled
card with its issue id.

| Screen | Calls | Shows |
| --- | --- | --- |
| Skill resolution | `resolution` with the skill and its value, LCK, one target modifier, and the roll | the target, success, margin, and tags |
| Character sheet | `derived` on the attributes and level; `POST /v1/sessions` to create the character | every derived stat, and the stored actor of the new session |
| Attack resolver | `combat` with the attacker (skill, roll, LCK), the defender (evasion, armor DR, a carried shield's DR and the shields skill, HP, and a dodge or block stance), the delivery, and the damage roll | each step of the result: the target, the margin, crit, the evasion and block bands, evaded or blocked, the damage after DR, and HP after (M3-27) |
| Playthrough | `POST /v1/sessions/step` for each step of `ruleset/tests/vectors/playthrough.json`, then `replay` | each step's input and result, and whether the replay matches |
| Saved games | `replay` on an opened file | the saves in this browser (name, rules, last played, last backup), a backup file for each, and what opening a file found |

Every roll is an input field, as in the vectors: the app rolls nothing unless asked, and a rolled
value comes from `crypto.getRandomValues` and is written into the field before the call, so a result can be read back against a vector.
The playthrough runner sends each vector's input as a session step, as given, and all twelve
steps match their vectors and the replay matches. [V 2026-11-14, by hand against the server]

## Saves

A player's games live in their browser (APP-16), in two layers. [V 2026-11-16, `saves.spec.ts` and
`e2e/app.spec.ts` over both transports]

- **Autosave.** `Saves` (`src/app/saves.ts`, generated with `ng generate service saves`) writes a
  save to IndexedDB (database `riftweave`, store `saves`) when a session starts and after every
  step, and the next step runs only once the write is done. A save is `{ id, name, rules_version,
  last_played, last_backup, session }`. The app resumes the latest save when it starts, and asks
  the browser to keep its storage (`navigator.storage.persist()`), since a browser may clear site
  data [U: which browsers clear it, and when, is not checked]. `localStorage` is not used for
  saves: it is small, and it blocks the page.
- **Characters.** The character list keeps its own database, `riftweave-characters` (store
  `characters`), so adding it needed no upgrade of the saves' database
  ([files.md](files.md#in-the-frontend)).
- **Backup files.** "Download backup" writes `<name>.riftweave.json`: the session JSON itself,
  with its `rules_version` ([sessions.md](sessions.md#the-rules-version)), so the agent CLI
  replays it as it is: `jq '{operation: "replay", session: .}' mara.riftweave.json | riftweave
  session`. "Open file", or dropping a file on the saved games, replays it first: a file that
  replays to its actor is kept as the replayed game; one that does not is flagged as edited or
  broken, and the player may keep the saved character anyway; a file from other rules says which
  rules it was made with and that it was replayed under the current ones. A save with no backup,
  or a backup older than 7 days, is marked, and the screen says that saves live in this browser
  only.

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

## GitHub Pages

The published app (APP-15) is one static site: the WebAssembly core with the ruleset of the same
commit built in, the Angular app, and a service worker. `.github/workflows/pages.yml` builds it on
a release tag (`v*`) and publishes it with the official Pages actions, pinned by commit; a push to
`main` does not publish, so players see a stable rules version. The data is in the core, not
fetched from GitHub at run time. [V 2026-11-16, `e2e/pages.spec.ts` on the Pages build; the
workflow itself runs only on a tag, so its first real run is the first release]

- **The `pages` build** (`npm run build:pages`): `angular.json` configuration `pages` turns on the
  service worker (`ngsw-config.json`, which prefetches the app, `wasm_exec.js`, and
  `riftweave.wasm`) and the `--base-href /<repo>/` of a project site. The workflow passes the tag as
  `RIFTWEAVE_RULES_VERSION`, which the footer shows; any other build shows `dev`.
- **Offline and installable.** After one visit the service worker serves everything, so the app
  plays with no network; `public/manifest.webmanifest` and its icon make it installable on phones
  and desktops. The service worker is registered only in the `pages` build, so `ng serve` and the
  tests are unchanged.
- **Preview.** `npm run preview:pages` serves `dist/` under `/riftweave-core/` on port 4400 as
  Pages does; the Playwright project `chromium-pages` runs on it, plays with no server, and plays
  again after going offline.

## Running it

| Command | Needs | Serves |
| --- | --- | --- |
| `./scripts/app.sh [--port 4200] [--api-port 8080]` | Go and Node (with Node below 24.15, it runs Node 24.22.1 through `npx`) | `ng serve` at `localhost:4200`, proxying `/v1` and `/openapi.json` to a `riftweave-server` it builds and starts; reloads on a frontend change |
| `./scripts/build-wasm.sh`, then `npm run start:wasm` in `frontend/` | Go and Node | `ng serve` at `localhost:4200` with no server: the core runs in the page as WebAssembly |
| `./scripts/build-wasm.sh`, then `npm run build:pages && npm run preview:pages` in `frontend/` | Go and Node | the Pages build at `localhost:4400/riftweave-core/`, as GitHub Pages serves it, with its service worker |
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
- **The design guide** (owner, 2026-11-23): colors, fonts, components, and layouts follow
  [docs/design/](../design/README.md), from the owner's mockups; their Tailwind is rewritten as plain CSS.
- **Small reusable components, built up.** Fields, cards, alerts, and lists live in `src/app/ui/`
  and take their values as signal inputs and models; a screen is a composition of them. Add a new
  `ui/` component when a second screen needs the same piece, rather than repeating markup.
- **Tests.** Unit and API specs run on Vitest (`npm test`, through `@angular/build:unit-test`),
  beside what they test (`name.spec.ts`; `app.spec.ts` covers the ui components and the screens).
  End-to-end tests run on Playwright (`npm run e2e`,
  `e2e/*.spec.ts`): a real browser on the built app, the app's proxy, and a real server.
- **Generate, then fill in** (owner, 2026-11-15). Components, services, pipes, directives, and
  guards are made with the Angular CLI, not written from scratch: from `frontend/`, run
  `npx ng generate component ui/<name> --flat` for a shared component (it writes `<name>.ts`,
  `.html`, and `.css`, standalone, beside the others in `ui/`), `npx ng generate component <name>`
  for a screen, `npx ng generate service <name> --skip-tests` for a service, and likewise
  `pipe`, `directive`, and `guard`. Then rename the selector to the `rw-` prefix and fill the files
  in. Keep a generated `*.spec.ts` only when it holds a real spec; the shared specs live in
  `app.spec.ts`. The CLI needs Node 24.15 or later.
- **Layout.** `src/app/api.ts` is the API and its `API` token, `src/app/context.ts` the app context,
  `src/app/ui/` the shared components, and each screen its three files in `src/app/` (`sheet`,
  `attack`, `playthrough`). [A]
- **Icons** (owner, 2026-11-22): Font Awesome's free solid set through `fa-icon`
  (`@fortawesome/angular-fontawesome`), drawn as inline SVG; flags from `flag-icons`, copied into the
  build only for the languages offered (`flags/<code>.svg`). Both work offline. A control that shows
  only an icon has an `aria-label`.
- **Explanations beside the code, not in it** (owner, 2026-11-22). A `.ts` file starts with one
  pointer line (`// See i18n.md.`) and carries no other comments; what it does, why, and the
  links to upstream docs go in a sibling Markdown file of the same name (`i18n.ts` and `i18n.md`),
  with a `Modified:` line. Names and types should say the rest. A spec (`*.spec.ts`) is the
  exception: its comments stay in it, beside the checks they explain, and it has no sibling. First done for the i18n files (FE_S11_01); the rest of `frontend/` moves in
  FE_S00_08_sibling-docs-for-code.

## The app context

`AppContext` (`src/app/context.ts`, generated with `ng generate service context`) is the one place a
screen reads the viewer's settings, as signals: `theme` (`light`, `dark`, or `system`) and `locale`
(an i18next language code, [Languages](#languages)). The mode starts from the saved choice, else the
system's mode (`prefers-color-scheme`), and the language from the saved one, else the browser's;
each is saved each change in `localStorage`, a per-viewer convenience; every
read and write is guarded, so a blocked storage still renders, in the system's mode. No component
reads `localStorage` or `matchMedia` itself. [V 2026-11-16, `app.spec.ts` and `e2e/app.spec.ts`
over both transports]

The mode is `data-theme` on `<html>`: `light` or `dark` when chosen, and absent for `system`.
`styles.css` keeps one set of tokens; the dark ones apply under `:root[data-theme="dark"]`, and
under `prefers-color-scheme: dark` only when no light mode is chosen
(`:root:not([data-theme="light"])`). `rw-theme-toggle` (`src/app/ui/theme-toggle`) in the header
cycles light, dark, and the system's mode, and its `aria-label` names the mode it switches to.

Settings: `exportSettings()` and `importSettings()` trade the settings file with the human CLI
([settings.md](settings.md), APP-25).

## Languages

The app's own text is in English, German, and Spanish (FE_S11_01). [V 2026-11-21, `i18n.spec.ts`
and `e2e/app.spec.ts` over both transports]

- **One source.** `AppContext.locale` holds the language. `I18n` (`src/app/i18n.ts`) is one
  [i18next](https://www.i18next.com) instance (26.7.2, no Angular wrapper) that follows it, and
  sets `<html lang>`. A code the app does not offer shows English.
- **Start language:** the saved one, else the browser's (`navigator.language`, `de-DE` names `de`)
  if it is offered, else English. The choice is saved like the mode and survives a reload.
- **Files.** One JSON file per language under `frontend/src/locales/`: `en.json` is the source,
  then `de.json` and `es.json`. They are bundled with the app, so a static build needs no fetch. A
  key missing from a language falls back to English. Keys are grouped by screen (`menu.*`,
  `settings.*`, `network.*`); a menu entry's key is `menu.<key of backend/contract/menu.json>`.
- **In a template:** `{{ 'menu.back' | t }}`, or `'app.rules' | t: { version: v }` for a key with
  `{{version}}` in it. The `t` pipe (`src/app/translate-pipe.ts`) is not pure, and the language
  signal it reads re-renders the view, so a switch changes every string without a reload. In code,
  `I18n.translate(key, values)` inside a `computed` does the same.
- **Page titles.** A route's `title` is a key, shown by `TranslatedTitle` in the viewer's language.
- **The select.** `rw-language-select` (`src/app/ui/language-select`) in the header lists the
  languages by their own names (English, Deutsch, Español), from `LANGUAGES` in `context.ts`.
- **Not translated:** ruleset content (spell, perk, and monster labels and descriptions) is data
  and shows as the data writes it; localizing content is its own issue. Numbers show as the core
  returns them. The settings file's rule messages are the Go CLI's words, shared with it
  ([settings.md](settings.md)). [A]
- **Checked by tests.** `i18n.spec.ts` fails when a key of `en.json` is missing from another
  language (or a language has a key `en.json` lacks), when a menu entry has no key, and when a
  template shows a literal word or a written-out `label`, `title`, `message`, `placeholder`,
  `aria-label`, or `alt`. The game's name is the one word allowed. The screens
  `FE_S11_04_frontend-screen-strings` still moves to keys are listed in the spec's `PENDING`.

**Who reviews a language** (owner decision, 2026-11-15): the owner reviews German and Spanish. Any
contributor may add a language, or change one other than those two, in a pull request that adds
only its locale file and its entry in `LANGUAGES`; the missing-key test checks it.

## Not covered

Styling past a readable layout, sharing saves between devices (APP-17 rooms, APP-30 files), and the human CLI's level-up flow.
The app reads the session JSON the server returns and keeps it in memory only. [A]
