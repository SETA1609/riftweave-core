# Onboarding

How to get from a fresh clone to your first pull request. Riftweave is a data-driven cRPG
ruleset: the product is the JSON under `ruleset/data/`, the schemas that govern it, and the
docs. There is no game engine here.

## 1. Set up

You need Python 3.13. The Go backend also needs Go 1.27, and the dev containers need
[uv](https://docs.astral.sh/uv/) and docker or podman.

```bash
pip install -r ruleset/requirements.txt
python ruleset/scripts/validate.py
```

The last line should end with every file passed and 0 failed. If you would rather not install
Python packages on your machine, use the dev toolchain, which mounts your checkout:

```bash
./scripts/container-image.sh
./scripts/validate.sh
./scripts/test.sh
```

## 2. Read, in this order

1. [README.md](../../README.md): what the project is and how it is laid out.
2. [ROADMAP.md](../../ROADMAP.md): what is in scope and what "finished" means.
3. [docs/core/contract.md](../core/contract.md): the one rule every system follows.
4. [docs/core/data-model.md](../core/data-model.md): how the information is shaped, with the
   [theory](../theory/data-models.md) behind it.
5. [docs/modules/progression.md](../modules/progression.md): the game itself: classless,
   d100 roll-under, eight attributes, point-buy skills, and perks.
6. [CONTRIBUTING.md](../../CONTRIBUTING.md): the rules for changes, the workflow, and
   [which document wins](../../CONTRIBUTING.md#which-document-wins).
7. [AGENTS.md](../../AGENTS.md): the short version of all of this. Its section on how
   validation works is worth reading before you edit a schema.

## 3. Pick an issue

Work comes from [docs/issues/open/](../issues/README.md), grouped by the [stories](../stories/README.md). This command lists what can start now:

```bash
uv run --project tools/devops --locked --no-dev devops-tools roadmap-order --settings-file tools/devops/pyproject.toml show
```

Take one from `first`, or from `ready` if you have a reason to. If your idea has no issue, add
its file to `docs/issues/open/` first.

## 4. Do the work

- Branch `<issue-id>-<slug>`, for example `M2-07-effects-over-time-and-food`.
- Set the issue's `Status` to `doing` in your first commit.
- A rule change follows five steps, one commit each: spec, schema, failing vector, data,
  review ([details](../../CONTRIBUTING.md#workflow)). Content-only issues skip the first three.
- Record decisions in `docs/decisions/<issue-id>.md`.
- Add a changelog fragment, `changes/<issue-id>.md` ([format](../../changes/README.md)).

## 5. Before you open the pull request

- `python ruleset/scripts/validate.py` and `./scripts/test.sh` pass.
- If you touched `backend/`, `cd backend && go test ./...` passes.
- The pull request does one issue and has at most 20 counted changed files.
- The issue's `Status` is `review`.
- If you used AI, a person has reviewed every change ([AI rules](ai.md)).

Open it against `main`, titled `<issue-id> <issue title>`. Open branches are rebased
automatically after every push to `main`, so run `git pull --rebase` before committing on one.

Contributions are accepted under the project's licenses; see
[LICENSING.md](../../LICENSING.md).
