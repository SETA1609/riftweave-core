### APP-21 Scripts run the frontend locally or in a container
- Added: `scripts/app.sh` runs the frontend and its API on the host, and `scripts/app-container.sh` builds and runs them as one image (`container/app.Dockerfile`, nginx in front) with docker or podman.
