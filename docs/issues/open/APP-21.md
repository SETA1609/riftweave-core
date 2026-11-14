#### APP-21 Scripts run the frontend locally or in a container
- Status: doing
- Priority: medium
- Needs: APP-07
- Do: Owner request, 2026-11-14: one command should start the frontend and its API, on the host or in a container. `scripts/app.sh` builds `riftweave-server`, starts it, installs the frontend's packages when they are missing or stale, and runs `ng serve` with its proxy; Ctrl-C stops both. It needs Go and Node; with a Node below Angular 22's minimum (24.15) it runs Node 24.22.1 through `npx`, so a contributor does not have to change their Node. `scripts/app-container.sh` builds one image (`container/app.Dockerfile`: the server, the built app, and the ruleset data, behind nginx, so the page and the API share one origin and need no CORS) and runs it with docker or podman, bound to localhost. Neither script holds a rule or changes the repo outside `cache/`.
- Keys: `scripts/app.sh`, `scripts/app-container.sh`, `container/app.Dockerfile`
- Files: ~8 (scripts/app.sh, scripts/app-container.sh, container/app.Dockerfile, container/app-nginx.conf, container/app-entrypoint.sh, .dockerignore, docs/backend/frontend.md, AGENTS.md)
- See: docs/backend/frontend.md, docs/backend/api.md
- Done when: `./scripts/app.sh` serves the app at localhost:4200 with a working API; `./scripts/app-container.sh` serves it at localhost:8080 with docker, and with podman; in both, the attack screen's default input returns the first combat vector's result.
