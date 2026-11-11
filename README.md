# riftweave-core

Riftweave — A data-driven, open-source cRPG ruleset for games and TTRPGs.

All game data is defined as JSON files, validated against JSON Schemas. The data
and schemas live under `ruleset/`.

It is a **classless**, **d100 roll-under** system: characters are defined by eight
attributes, point-buy skills with tagging, and perks. See
[docs/modules/progression.md](docs/modules/progression.md) for the system overview.

The shared sentence between systems is the kernel contract:
[docs/core/contract.md](docs/core/contract.md). A system may read an actor snapshot
and emit modifiers or events. It may not call another system's procedure.

## Project structure

```
docs/
  onboarding/         How to get started, and the AI rules
  core/               Kernel rules. Wins over docs/modules/ on a conflict.
    contract.md       Actor snapshot, modifier, event. The only shared sentence.
  modules/            Core systems in detail, and optional modules
  stories/            Milestones and the issues to work on
  decisions/          One decision file per issue
  backend/            The reference backend and its CLI
  combat/             Worked combat examples
  theory/             Background: data models and why ours is shaped as it is
  archive/            Retired plans and the July 2026 review
ruleset/
  schemas/            JSON Schema (draft-07)
    schema.json       Shared vocabulary (ability, color, phase, appliedEffect, ...)
  data/               Core game data (JSON), one directory per collection
  modules/            Optional modules, each with a manifest (factions, example-module)
  tests/vectors/      Golden vectors every implementation must pass
  scripts/
    validate.py       Schema, reference, and integrity checks
    *_reference.py    Python reference for each system
    test_*.py         Unit tests and the vector runner
  requirements.txt    Python dependencies
backend/              Reference backend in Go; it adds no rules
changes/              Changelog fragments, one per issue
scripts/              Dev toolchain wrappers (validate, test, shell, agent)
tools/                The devops-tools pin
```

See [docs/modules/README.md](docs/modules/README.md) for how modules work and which are
planned (crafting, alchemy, magic crafting, etc.). Scope and milestones are in
[ROADMAP.md](ROADMAP.md). New contributors start at
[docs/onboarding/](docs/onboarding/README.md).
Modules must speak the kernel contract. They must not call another system's procedure.

## Validation

```bash
pip install -r ruleset/requirements.txt
python ruleset/scripts/validate.py
```

Or with Docker. This image copies the ruleset in, so rebuild it after edits:

```bash
docker build -t riftweave-validate .
docker run --rm riftweave-validate
```

The dev toolchain mounts this checkout instead. It needs [uv](https://docs.astral.sh/uv/) and docker or podman on the host:

```bash
./scripts/container-image.sh   # build riftweave:python-<version>
./scripts/validate.sh
./scripts/test.sh
```

`./scripts/shell.sh` opens a shell in that image. `./scripts/agent.sh` runs the agent image (Claude Code and uv, with its own login). See [tools/devops/README.md](tools/devops/README.md).

CI runs automatically via `.github/workflows/validate-ruleset.yml` on every push
or PR that touches `ruleset/`, `docs/core/`, `docs/stories/`, `docs/issues/`, `docs/backend/`, `backend/`,
`.github/`, or the `Dockerfile`.

## License

Riftweave is dual-licensed so it can be used freely in both software and
games/tabletop:

- **Software & tooling** (`ruleset/scripts/`, `ruleset/schemas/`, Dockerfiles, `tools/devops/`, CI)
  — [Apache License 2.0](LICENSE).
- **Game content & documentation** (`ruleset/data/`, `docs/`, prose) —
  [Creative Commons Attribution 4.0](LICENSE-CONTENT).

See [LICENSING.md](LICENSING.md) for the full breakdown, attribution guidance, and a
note on trademarks. Contributions are accepted under these same licenses.
