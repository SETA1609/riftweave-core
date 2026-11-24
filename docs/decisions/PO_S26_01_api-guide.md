# PO_S26_01_api-guide Write the resource API guide

Date: 2026-11-24

- **Decision** (owner, 2026-11-24): a written guide to the API before more endpoints are built, as `docs/backend/api-guide.yml`: endpoints by resource, no code. It has 55 endpoints in 12 resources: 6 served today, 3 served but proposed to change (the session endpoints), and 46 new. [V 2026-11-24, parsed with PyYAML]
- **Resources over commands:** today every rule is `POST /v1/commands/{command}`, and the client holds the session. The guide puts the things a person works with (characters, creation drafts, the profile, saved games, the encounter, the bestiary) at their own paths and keeps the generic endpoint for calculators and agents. [A]
- **One shape, two stores:** offline the frontend keeps a resource in IndexedDB and runs the core in the browser; with a server, the server keeps it. The frontend's API service has one method per endpoint and picks the store. [A]
- **Shapes stay in the contracts:** the guide names a body or a result by its contract definition and copies no schema, so it cannot drift from `backend/contract/`.
- **Characters and creation follow story S25:** the roster, a draft built one step at a time in the core's order, and the profile (the current character). The S25 issues (filed separately) give those endpoints their commands.
- **Open:** where saved games live, whether the profile is per device or per person, whether the generic command endpoint stays public, and roles once there is sign-in [U]. The endpoint issues are filed once the owner has agreed the guide.
