# The frontend, its API, and the ruleset data in one image (APP-21, docs/backend/frontend.md#running-it).
# nginx serves the built app and proxies /v1 and /openapi.json to riftweave-server on the same
# origin, so the server needs no CORS. Build from the repository root: scripts/app-container.sh.

FROM docker.io/library/golang:1.27.2-alpine AS api
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -trimpath -o /out/riftweave-server ./cmd/riftweave-server

FROM docker.io/library/node:24.22.1-alpine AS web
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
# npm 11.21.1: the npm that ships with Node 24 fails to resolve Vitest's optional peers.
RUN npx --yes npm@11.21.1 ci --no-audit --no-fund
COPY frontend/ ./
# The playthrough screen reads its steps from the vectors.
COPY ruleset/tests/vectors/playthrough.json /src/ruleset/tests/vectors/playthrough.json
RUN npm run build

FROM docker.io/library/nginx:1.29.3-alpine
COPY --from=api /out/riftweave-server /usr/local/bin/riftweave-server
# The server reads the data, and the schemas for its OpenAPI document.
COPY ruleset/data /srv/ruleset/data
COPY ruleset/schemas /srv/ruleset/schemas
COPY --from=web /src/frontend/dist/riftweave/browser /usr/share/nginx/html
COPY container/app-nginx.conf /etc/nginx/conf.d/default.conf
COPY container/app-entrypoint.sh /usr/local/bin/app-entrypoint.sh
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/app-entrypoint.sh"]
