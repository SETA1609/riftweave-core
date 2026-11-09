# devops

The toolchain container, the agent container, and the repository workflow tools all come from [devops-tools](https://github.com/SETA1609/devops_tools), pinned to a full commit in [`pyproject.toml`](pyproject.toml) and `uv.lock`. This directory is only that pin and its settings: no local runner code. The host needs only [uv](https://docs.astral.sh/uv/). [V 2026-11-08]

The settings are `[tool.devops-tools]` (roadmap order, changelog fragments, the rebase workflow) and `[tool.devops-tools.container]` (the images and the named commands) in `pyproject.toml`; every command passes `--settings-file tools/devops/pyproject.toml`.

The root `Dockerfile` stays the immutable validation image: it copies the ruleset into the image. The toolchain container mounts the working tree, so it validates the checkout you have now.

## Images

- **`devops-tools-base`** is shipped by devops-tools: Ubuntu with git, uv, and a uv-managed Python. `devops-tools container` builds it and passes it to our build as `BASE_IMAGE`.
- **`riftweave:python-<version>`** is [`container/Dockerfile`](../../container/Dockerfile), `FROM ${BASE_IMAGE}`: it installs the pinned Python and the ruleset requirements into `/opt/riftweave-venv`, first on `PATH`. The version is the repo-root [`.toolchain-python-version`](../../.toolchain-python-version); it must equal `ARG PYTHON_VERSION` in the Dockerfile, and the runner checks it before every build. The file is not named `.python-version`, so pyenv and uv keep using the host interpreter.
- **`riftweave-agent:<tag>`** is devops-tools' agent image on top of ours: Node, Claude Code, and an `agent` user with the host's uid and gid. No agent Dockerfile lives here.

## Usage

Each shim in `scripts/` is one line: `uv run --project tools/devops --locked --no-dev devops-tools container --settings-file tools/devops/pyproject.toml <action> "$@"`.

| Shim | Action | What it does |
| --- | --- | --- |
| `./scripts/container-image.sh` | `image` | Checks the Python version, builds the base and our image, prunes dangling layers and build cache, prints the image id and size. |
| `./scripts/validate.sh [args...]` | `validate` | `python ruleset/scripts/validate.py [args...]` in the container. |
| `./scripts/test.sh [args...]` | `test` | `python -m unittest discover -s ruleset/scripts -p 'test_*.py' [args...]` in the container. |
| `./scripts/python.sh [args...]` | `python` | `python [args...]` in the container, for example `./scripts/python.sh --version`. |
| `./scripts/shell.sh` | `shell` | Interactive bash in the container. |
| `./scripts/agent.sh [command...]` | `agent` | A command (default `bash`) in the agent container, with its own Claude login and git identity. |
| `./scripts/prune.sh` | `prune` | Removes dangling images and unused build cache. Tagged images and volumes stay. |
| `./scripts/clean.sh` | `clean` | End of work only: removes the `clean-directories` in `pyproject.toml` (Python caches, the local virtualenv) and the cache volume, then prunes. |
| `./scripts/devcontainer-image.sh` | `agent true` | Builds the agent image and tags it `riftweave-agent:devcontainer` for `.devcontainer/devcontainer.json`. |

`validate`, `test`, `python`, and `shell` build the image first if it is missing. Every run is timestamped and logged to `cache/logs/`. Run `./scripts/clean.sh` only when the user asks or says the work is finished; nothing else calls it.

- `--keep-cache` (leading) keeps uv's cache in the volume `riftweave-uv-cache` until `clean`.
- `IMAGE` overrides the image tag; `CONTAINER_AS_ROOT=1` runs as root instead of the host user.
- The container mounts the checkout at its host path, not `/src`; the devcontainer still uses `/src`.

## Agent container

`./scripts/agent.sh` runs devops-tools' agent image. Its home is the named volume `personal-agent-home`, so the Claude login and git identity survive the container and never touch the host's `~/.claude`, `~/.gitconfig`, or SSH keys: the agent can commit but not push. A leading `--rebuild-image` rebuilds the agent image. On every start a script installs the marketplaces and plugins `.claude/settings.json` declares, if any are missing.

One-time setup, inside `./scripts/agent.sh`:

```sh
claude                       # then /login with the personal account; paste the code from the browser
git config --global user.name  "Your Name"
git config --global user.email "personal@example.com"
```

There is no docker inside the container, so run `python ruleset/scripts/validate.py` directly instead of `scripts/*.sh`.

## Devcontainer

`.devcontainer/devcontainer.json` runs the agent image as a devcontainer. Its `initializeCommand` is `./scripts/devcontainer-image.sh`, which builds the agent image through devops-tools (whose tag carries a content hash) and tags it `riftweave-agent:devcontainer`, the image the devcontainer names. [V 2026-11-08]

## Shared devops-tools

The same pin runs the repository workflow tools: `devops-tools roadmap-order`, `changelog-fragments`, and the reusable rebase workflow (`.github/workflows/`). The workflows pin the same commit as `pyproject.toml`; bump them together in one PR. The run log and the rebase script are tested in devops-tools, not here.
