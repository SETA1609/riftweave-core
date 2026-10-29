# devops

Host-side CLI that runs the Python toolchain container (`riftweave:python-3.13.18`) and logs every run. It runs on the host, outside the container. The host needs only [uv](https://docs.astral.sh/uv/). At runtime the CLI uses only the Python standard library.

The container Python version is the repo-root [`.toolchain-python-version`](../../.toolchain-python-version). It must match `ARG PYTHON_VERSION` in [`container/Dockerfile`](../../container/Dockerfile) and `BASE_IMAGE` in [`.devcontainer/devcontainer.json`](../../.devcontainer/devcontainer.json). The file is not named `.python-version`, so pyenv and uv keep using the host interpreter. `tools/devops/.python-version` is only the host CLI (any 3.13).

The root `Dockerfile` and `test.sh` stay the immutable validation image: they copy the ruleset into the image. This toolchain mounts the working tree at `/src`, so it validates the checkout you have now.

## Usage

Use the shims in `scripts/` from the repo root. Each one runs `uv run --project tools/devops --locked --no-dev devops <subcommand>`.

| Shim | Subcommand | What it does |
| --- | --- | --- |
| `./scripts/container-image.sh` | `image` | Checks that `.toolchain-python-version` matches the Dockerfile and the devcontainer, builds the image, prunes dangling images and unused build cache, prints the image id and size. |
| `./scripts/validate.sh [args...]` | `validate` | `python ruleset/scripts/validate.py [args...]` in the container. |
| `./scripts/test.sh [args...]` | `test` | Unit tests in the container. With no arguments: `python -m unittest discover -s ruleset/scripts -p 'test_*.py'`. Arguments are passed to `unittest`. |
| `./scripts/python.sh [args...]` | `python` | `python [args...]` in the container, for example `./scripts/python.sh --version`. |
| `./scripts/shell.sh` | `shell` | Interactive bash in the container, with the repo at `/src`. |
| `./scripts/agent.sh [command...]` | `agent` | Runs a command (default `bash`) in the agent container, which has its own Claude login and git identity. See [Agent container](#agent-container). |
| `./scripts/prune.sh` | `prune` | Removes dangling images and unused build cache. The tagged image and the cache volume stay. |
| `./scripts/clean.sh` | `clean` | End of work only: removes `__pycache__/`, `.pytest_cache/`, `.ruff_cache/`, `.venv/`, and `venv/` under the repo, and the pip cache volume, then prunes. |

`validate`, `test`, `python`, and `shell` build the image first if it is missing.

Run `./scripts/clean.sh` only when the user asks or says the work is finished. No other command calls it. It refuses any path outside the repo and never touches `cache/logs/`.

### Pip cache

Without flags, the pip cache lives in the container's throwaway home and is gone when the container exits. A leading `--keep-cache` on `validate`, `test`, `python`, or `shell` stores it in the named volume `riftweave-pip-cache`, which is kept until `clean`:

```sh
./scripts/validate.sh --keep-cache
```

Only a leading `--keep-cache` counts, so `./scripts/python.sh script.py -- --keep-cache` still passes the flag to the program.

### Environment

- `IMAGE` overrides the image tag.
- `CONTAINER_AS_ROOT=1` runs the container as root instead of the host user (`--user uid:gid`).

The CLI uses docker if its daemon answers, otherwise podman.

## Agent container

`container/agent.Dockerfile` builds `riftweave-agent:python-3.13.18` on top of the toolchain image. It adds Claude Code (pinned), Node 22, `uv`, and an `agent` user with the host's uid and gid, so files the agent writes under `/src` belong to the host user. `agent` builds both images if they are missing. A leading `--rebuild-image` rebuilds the agent image, and a leading `--keep-cache` works as it does for `shell`.

The container's home, `/home/agent`, is the named volume `personal-agent-home`, with `CLAUDE_CONFIG_DIR=/home/agent/.claude`. The host's `~/.claude`, `~/.claude.json`, `~/.gitconfig`, and SSH keys are never mounted, so logging in inside the container leaves the host's Claude Code on its own account. The volume name is generic so other repos can mount the same home and reuse the login. `AGENT_HOME_VOLUME` picks a different volume, and `AGENT_IMAGE` a different tag.

One-time setup, inside `./scripts/agent.sh`:

```sh
claude                       # then /login with the personal account; paste the code from the browser
git config --global user.name  "Your Name"
git config --global user.email "personal@example.com"
```

After that, `./scripts/agent.sh claude` starts straight away. There is no docker inside the container, so run `python ruleset/scripts/validate.py` directly instead of `scripts/*.sh`.

### Commits and pushes

The repo is bind-mounted, so a commit made in the container is a commit in the host checkout, with the container's git identity. The container has git but no GitHub credentials and no SSH keys, so it cannot push or fetch. Push from the host with your own SSH setup. Do not forward the host's `SSH_AUTH_SOCK`: that would give the container every repo the host's key can reach.

`.devcontainer/devcontainer.json` builds the same image for editors that attach to dev containers. It mounts the same home volume. `initializeCommand` builds `riftweave:python-3.13.18` when that image is missing.

## Logs

Every line from the container is prefixed with an ISO-8601 UTC timestamp in milliseconds and a stream tag, on the terminal and in `cache/logs/<utc-timestamp>-<subcommand>.log`:

```text
2026-10-28T21:25:07.166Z out | Python 3.13.18
2026-10-28T21:25:07.293Z run | exit code: 0, duration: 0.412 s
```

- `out`: the child's stdout.
- `err`: the child's stderr, or a devops error.
- `run`: devops bookkeeping (command lines, exit codes, durations).

The log ends with the command line, exit code, and duration. `shell` owns the terminal, so only its start, end, and exit code are logged. The CLI exits with the exit code of the last container command it ran. A child killed by signal N exits with 128 + N.

## Layout

```text
src/devops/
  cli.py           subcommands and main()
  container.py     repo root, docker/podman detection, container command lines
  run_log.py       timestamping, log files, running a child with logged output
  cache_cleanup.py choosing and removing Python cache paths for `clean`
  constants.py     constants
  exceptions.py    exceptions
  models.py        custom types (dataclasses, type aliases) that keep parameter lists short
tests/             pytest
```

## Development

```sh
cd tools/devops
uv run pytest
uv run ruff check
uv run ruff format --check
```

Code style:

- Function names are a descriptive verb plus an object, with words spelled out and units in the name (`measure_directory_size_bytes`).
- Docstrings put `"""` on its own line, start the text on the next line, and have no blank lines inside.
- ruff is configured in `pyproject.toml`, with a line length of 100.
