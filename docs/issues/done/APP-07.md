#### APP-07 Angular frontend
- Status: done
- Merged: #172, 2026-11-14
- Priority: low
- Needs: APP-06
- Modified: 2026-11-15
- Do: An Angular app in `frontend/` that talks only to the API, with a client generated from the OpenAPI document: a character sheet, an attack resolver (attacker, defender, reaction, delivery, rolls), and a playthrough runner over a session. It shows the rolls and the step-by-step result of each pipeline.
- Files: ~15 (frontend/, docs/backend/frontend.md, .github/workflows/)
- See: docs/backend/api.md
- Done when: the app builds and its tests pass in CI; it creates a character and resolves an attack through the API; the attack result matches the vector for the same inputs.
