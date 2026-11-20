### FE_S00_05_regenerate-openapi-types Regenerate the frontend's API types from the current contract
- Changed: `frontend/src/api/openapi.d.ts` is regenerated from the current contract with `npm run generate`; it gains the `attack_sequence`, `command_summons`, and `stance` commands and the enums and fields added since 2026-11-16.
- Added: the frontend workflow regenerates the API types against the server it starts and fails when `openapi.d.ts` differs, keeping the regenerated file as an artifact (`openapi-types`).
