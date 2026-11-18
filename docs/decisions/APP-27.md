# APP-27 The Angular app opens on a home page, and Start opens the tools

Date: 2026-11-18

- **Decision:** the app uses the Angular router (`@angular/router` 22.2.4, added) with hash URLs (`#/play/attack`), so the static Pages build reloads any route and works offline with no server fallback [A]. The routes are in `app.routes.ts`; a route's data sets the component's inputs (`withComponentInputBinding`).
- **One menu component:** `rw-menu-grid` reads `backend/contract/menu.json` (imported as JSON) and shows the menu a route names; the home page, the tools page, and GM tools are the same component. [V 2026-11-18, `app.spec.ts`, `e2e/menus.spec.ts`]
- **`apps` on a menu entry** (added here): Quit is CLI-only, and Saved games and Playthrough are web-only, because the browser keeps saves in IndexedDB and the CLI has no such screen. The CLI filters on it too. This keeps one file for both menus while each app lists only what it can open [A].
- **One component for three settings routes:** `/settings`, `/modules`, and `/network` are `rw-settings-screens` with a `section` input, since all three edit the one settings file through `AppContext` [A]. `AppContext` now holds the whole file (`file`), with `update()` checking every edit by the APP-25 rules before saving.
- **Modules** are the three manifests imported by `settings-screens.ts` (`MODULE_IDS`); a new module adds its manifest import there [A]. Loading only the checked modules is not part of this issue [U], as in APP-26.
- **The network switch:** `switchableTransport` reads the setting on each call: `local` is the build's transport (WASM, or the dev API), `server` is `httpTransport(server_url)`. A server on another origin needs CORS, which the API does not send (docs/backend/api.md#not-covered) [U].
- **Stacked on APP-26:** this branch starts from APP-26, which adds `menu.json` and `app-layout.md`; it merges after APP-26.
