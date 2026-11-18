# APP-48 The frontend's objects and functions in the browser console, in a dev build

Date: 2026-11-18

- **Decision:** a dev build sets `globalThis.rw` to `{ context, api, saves, router, screen(), help() }` (`frontend/src/app/dev-console.ts`), installed from `main.ts` after bootstrap when `RIFTWEAVE_DEV && isDevMode()`. [V 2026-11-18, the unit spec; the end-to-end specs run in CI]
- **Which builds are dev:** a new `development` build configuration (the core in the page) and the existing `api` one set `RIFTWEAVE_DEV` to true, and `ng serve` now builds `development` by default (it built the default options before, which are the same but for that define). `ng build` and `pages` leave it false. [A]
- **Left out of the bundle:** the console module is loaded by a dynamic `import()` inside the `RIFTWEAVE_DEV` branch, so a build where the define is the constant false drops the branch and the module. The Pages spec checks `rw` is absent from the page. [A: a check of the bundle's files would be stronger]
- **`rw.screen()`** reads the routed element through Angular's dev-mode `ng.getComponent`, so it works in `ng serve` and returns null elsewhere. [A]
- **No dev operations from the console yet:** the core in the page runs without dev mode (`EnableDev` is the CLI's), so `rw` changes the app through its context and API only. Editing a session's actor from the console needs a dev switch in `backend/cmd/riftweave-wasm`, which this issue does not add. [U]
- **Not run here:** the Vitest and Playwright specs; the sandbox cannot install the frontend's packages. CI runs them. [U]
