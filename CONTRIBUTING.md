# Contributing to Riftweave

Thanks for your interest in Riftweave! This repository is a **data-oriented,
engine-agnostic cRPG ruleset**. The product *is* the data: a set of JSON files
describing attributes, skills, perks, effects, spells, items, and ancestries,
governed by JSON Schemas.

Because the rules live as plain data rather than code, the same ruleset can drive:

- **Computer games** — game engines that load these definitions at runtime.
- **Tabletop / TTRPG play** — virtual tabletops, character-sheet tools, and
  generators that consume the same JSON, or referees reading it directly.

Keep this dual audience in mind: prefer rules expressed as **declarative data** over
behavior that only makes sense inside one engine, and write descriptions that read
clearly to a human at the table as well as to a program.

## Getting set up

```bash
pip install -r ruleset/requirements.txt
python ruleset/scripts/validate.py        # validate every data file against its schema
```

The validator checks **all** data files and exits non-zero on any failure. CI
(`.github/workflows/validate-ruleset.yml`) runs it on every push/PR that touches
`ruleset/`. A change is not ready until validation is green.

You can also run it via Docker: `docker build -t riftweave-validate . && docker run --rm riftweave-validate`.

For day-to-day work, the dev toolchain mounts your checkout (uv, plus docker or podman):

```bash
./scripts/container-image.sh
./scripts/validate.sh
./scripts/test.sh
```

See [tools/devops/README.md](tools/devops/README.md) for the shell, the agent container, and logs.

## How the data is organized

- `ruleset/schemas/*.schema.json` — JSON Schema (draft-07) definitions.
- `ruleset/data/<collection>/*.json` — the game data, each wrapped in a single named
  array (`{ "skills": [...] }`, `{ "effects": [...] }`, …).
- `ruleset/schemas/schema.json` — **shared vocabulary**: the `ability` enum, the
  `color` magic-school enum, the `appliedEffect` shape, `diceExpression`,
  `percentile`, and `sourceRef`. Reference these via `schema.json#/definitions/...`
  rather than redefining them.

See [`docs/modules/progression.md`](docs/modules/progression.md) for the full system
overview (classless, d100 roll-under, 8 attributes, point-buy + tagged skills, perks,
9 color magic schools, and the shared effect pool).

## Rules for changes

1. **Schema first.** Every schema sets `additionalProperties: false`. To add a new
   field to any data entry, add it to the schema *before* the data, or validation
   fails.
2. **Declare `$schema`.** Each data file must have a `$schema` field with the
   relative path to its schema (files without it are skipped, not validated).
3. **IDs are `snake_case`** matching `^[a-z_][a-z0-9_]*$` and should be stable —
   cross-references between files are by string id and are **not** schema-validated,
   so renaming an id can silently break links.
4. **Use the shared effect pool.** Spells, consumables, and ingredients reference
   effects by id from `data/effects/core.json`; respect each effect's `channels`.
   Don't define effects locally in another system.
5. **Keep the established conventions.** Magic uses color-named schools
   (`red_magic`…`black_magic`), never Elder Scrolls school names; skills are
   point-buy (not use-leveled); tag skills are a +2/point discount. See
   [`docs/modules/progression.md`](docs/modules/progression.md) before reworking
   magic/skills/perks.
6. **Match surrounding style.** Mirror the existing JSON formatting (e.g. compact
   one-line entries where a collection already uses them) and comment/description
   density.

## Workflow

Every issue that adds or changes a rule follows five steps, in this order. Each step is
its own commit in the issue's PR, so a reviewer can see that the spec and the failing
vector existed before the rows.

1. **Spec.** Write the rule in `docs/core/` (kernel) or `docs/modules/` (module) in your
   own words, with [verification tags](#verification-tags). Name the inspiration it
   mirrors, if any, and how it differs.
2. **Schema.** Add or extend the schema and the shared vocabulary in `schema.json`.
   Existing data still validates.
3. **Vector.** Write the golden vector(s) in `ruleset/tests/vectors/`
   ([vectors.md](docs/core/vectors.md)): input actor, modifiers, seeded roll, expected
   result. They fail against the Python reference; the commit message records the
   failing run.
4. **Data.** Add rows and the reference code (`ruleset/scripts/*_reference.py`) until
   `validate.py` and the vectors pass, without changing the vectors. If a vector turns
   out wrong, fix it in its own commit before this one and say why.
5. **Review.** Read the result against the spec and list every difference in the PR
   description, as a bug (fix it) or intended (say why).

Content-only issues (new rows under an existing schema and kernel, such as more perks or
monsters) skip steps 1 to 3. Tooling issues use the same steps with tests in place of
vectors.

### Pull request rules

1. **One issue, one PR.** The PR does what the issue says and nothing else. No branch
   without an issue; if work has none, file it first ([issues/](issues/README.md)).
2. **At most 20 changed files** per PR, as `git diff --stat` counts them. A file with 5 or
   fewer changed lines does not count. If the counted files grow past 20, split the issue.
   Content-floor issues (perks, spells, monsters) split by color, level band, or skill.
3. **Each PR leaves `main` green.** `validate.py` and `./scripts/test.sh` pass at the PR's
   last commit.
4. **Docs in the same PR.** A change to a rule, a decision, or a tag updates the matching
   doc. Record decisions in [docs/decisions.md](docs/decisions.md).
5. **Changelog and roadmap.** Closing an issue adds its [CHANGELOG.md](CHANGELOG.md) line
   and deletes its [ROADMAP.md](ROADMAP.md) row in the same PR.
6. **Names.** Branch `RW-NN-<slug>` (the issue title in lowercase kebab-case, shortened if
   long), PR title `RW-NN <title>`. Agents prefix the branch with `agent/<name>/`.
7. **Describe it for both audiences.** Say *what rule or content* changed and why, so
   programmers and game designers reviewing the PR can follow it.

Open pull requests are rebased onto `main` after every push to `main`
(`.github/workflows/rebase-open-branches.yml`). Conflicts where both sides only added
`CHANGELOG.md` lines keep both; any other conflict leaves the branch for a person. Because
of it, an open branch can move under you: run `git pull --rebase` before committing on it.

### Verification tags

Design docs mix checked facts with intentions. Tag claims that a reader might rely on:

- **[V]** verified: checked against the data, a vector, or a cited source in this change,
  with the date (`[V 2026-10-31]`).
- **[A]** assumption: a design intent or a default not yet backed by a vector or decision.
- **[U]** unverified: taken from memory or an outside source not yet checked.

Check **[A]** and **[U]** before building on them. A PR that confirms one changes it to
**[V]**. `tools/fetch-docs/` fetches a cited source, pinned to a commit, for checking
([README](tools/fetch-docs/README.md)).

### Naming

A name says what the thing is or does without opening it. One concept, one word, the same
in every layer: the schema property, the JSON key, the doc term, and the reference code use
one name.

- **Entry keys** are `snake_case`, spelled out, with the established prefixes
  (`damage_fire`, `fortify_strength`, `potion_minor_healing`). No invented abbreviations or
  dropped vowels (`dmg`, `pot`, `mgr`).
- **Allowed abbreviations** are the game's own terms: the ability enum (`str`, `per`, `end`,
  `int`, `wil`, `agi`, `cha`, `lck`), `hp`, `mp`, `ap`, `dr`, `xp`, `crit`, `id`, `min`,
  `max`. Anything else is spelled out.
- **Schema property casing** is mixed today: older schemas use `camelCase` (`hitPoints`,
  `durationType`), newer ones `snake_case` (`base_key`, `enchantment_slots`). A new property
  follows the schema it extends. Unifying them renames keys, so it is a breaking change for
  its own issue.
- **Units in the name** where a number has one: `duration_rounds`, `range_meters`,
  `weight_kilograms`.
- **Booleans read as questions:** `is_ranged`, `has_reach`, `requires_concentration`.
- **No vague names:** not `data`, `info`, `misc`, `common`, `other`, `extra`, `stuff`.
- **Python** (scripts, tools) follows PEP 8 casing and the same rules: verb + object
  function names (`resolve_equipment_bases`, `compute_derived_hit_points`), tests named as
  sentences (`test_crit_on_roll_under_five_is_tagged`).
- **Renaming a key is breaking.** It goes in its own issue and is marked **(breaking)** in
  the changelog.

## Submitting

- Work on an issue branch and open a pull request against `main`.
- Make sure `python ruleset/scripts/validate.py` passes and that any new
  cross-references (effect ids, skill ids) resolve.

## Licensing of contributions

Riftweave is dual-licensed (see [LICENSING.md](LICENSING.md)): **Apache-2.0** for
software/tooling and **CC BY 4.0** for content/documentation. Unless you state
otherwise, contributions you submit are provided under these same licenses —
inbound matches outbound.

## Code of Conduct

By participating, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).
