# Frontend

`frontend/` is an Angular app that talks only to the HTTP API ([api.md](api.md)) (APP-07). It is an
adapter like the CLIs: it builds a command's input from a form, sends it, and shows the result. It
holds no rule. A number it shows comes from a response, never from a formula of its own. [A]

## The client

The client is typed from the server's own OpenAPI document, never by hand:

```
npm run generate   GET /openapi.json from a running riftweave-server
                   → openapi-typescript → src/api/openapi.d.ts (checked in)
src/app/api.ts     openapi-fetch over those types: one function per call the app makes
```

The generated file is checked in, so the app builds without a server; `npm run generate` refreshes
it after a contract change, and a stale file fails the type check where a field moved. [A]
`api.ts` takes the API's base URL, so the tests point it at a server and the app at the dev
proxy. The server sends no CORS headers (api.md, Not covered), so `ng serve` proxies `/v1` and
`/openapi.json` to `127.0.0.1:8080` (`proxy.conf.json`).

## Screens

| Screen | Calls | Shows |
| --- | --- | --- |
| Character sheet | `derived` on the attributes and level; `POST /v1/sessions` to create the character | every derived stat, and the stored actor of the new session |
| Attack resolver | `combat` with the attacker (skill, roll, LCK), the defender (evasion or block, armor DR, HP, reaction), the delivery, and the damage roll | each step of the result: the target, the margin, crit, the reaction, the damage after DR, and HP after |
| Playthrough | `POST /v1/sessions/step` for each step of `ruleset/tests/vectors/playthrough.json`, then `replay` | each step's input and result, and whether the replay matches |

Every roll is an input field, as in the vectors: the app rolls nothing unless asked, and a rolled
value is written into the field before the call, so a result can be read back against a vector.
[A]

## Tests

`npm test` runs the specs (Vitest, through `@angular/build:unit-test`). The API specs call a real
`riftweave-server` at `RIFTWEAVE_API` (default `http://127.0.0.1:8080`):

- **Create a character:** `derived` returns every stat, and `POST /v1/sessions` returns a session
  whose actor is the one sent.
- **Resolve an attack:** the first vectors of `ruleset/tests/vectors/combat.json` go through the
  app's attack call, and each result matches the vector's `expect`.

CI (`.github/workflows/frontend.yml`) builds the server, starts it, and runs `npm ci`, the build,
and the tests on every change under `frontend/`, `backend/`, or `ruleset/`.

## Not covered

Styling past a readable layout, a saved list of characters, and the human CLI's level-up flow.
The app reads the session JSON the server returns and keeps it in memory only. [A]
