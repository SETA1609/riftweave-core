# APP-07 Angular frontend

Date: 2026-11-14

- **Decision:** the client is typed from the server's own `/openapi.json` with `openapi-typescript`, and calls go through `openapi-fetch`; `src/app/api.ts` is the one place the app reaches the API. The generated types (`src/api/openapi.d.ts`) are checked in, so the app builds without a server, and `npm run generate` refreshes them. [V 2026-11-14]
- **Decision:** the app is zoneless and standalone, three screens and no router: the character sheet (`derived`, then `POST /v1/sessions`), the attack resolver (`combat`, every roll an input field, the result in pipeline order), and the playthrough runner (`playthrough.json` as session steps, then `replay`). It holds no rule: a new character's pools are the maxima `derived` returns, as the playtest driver does. [A]
- **Decision:** the specs call a real `riftweave-server` (`RIFTWEAVE_API`), not a mock, so "the attack result matches the vector" is checked end to end: the first twelve `combat.json` vectors without a `function` or an error go through the app's client and match their `expect`. [V 2026-11-14, 14 specs]
- **Decision:** versions are pinned, each published at least two weeks before this change: Angular 22.2.4 (build and CLI 22.2.5), TypeScript 6.0.3, rxjs 7.8.2, Vitest 4.1.12, jsdom 30.1.3, openapi-fetch 0.17.0, openapi-typescript 7.13.0, Node 24.22.1, npm 11.21.1, and `actions/setup-node` v6.5.0 by commit. [V 2026-11-14, npm publish times]
- **Decision:** `openapi-typescript` is run through `npx`, not installed: it pins TypeScript 5 as a peer, npm's `overrides` crash on it, and Angular 22 needs TypeScript 6. The npm that ships with Node 24 (11.5) crashes resolving Vitest's optional peers; 11.21.1 does not, so CI and the docs use it. [V 2026-11-14]
- **Why:** story APP: the frontend can create a character and resolve an attack through the API.
- **Open:** the server sends no CORS headers, so the app runs behind `ng serve`'s proxy (`proxy.conf.json`) or a proxy in front of both; serving the built app from the server is a later choice. `LICENSING.md` now lists `frontend/` as Apache-2.0; it does not list `backend/` either, which predates this issue. [A]
- **Risk:** the local toolchain here was Node 24.7, below the Angular CLI's minimum; a contributor on an older Node gets a clear error from `ng`, but the docs are the only place the version is stated.
