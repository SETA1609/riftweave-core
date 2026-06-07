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

## Submitting

- Work on a branch and open a pull request against `main`.
- Make sure `python ruleset/scripts/validate.py` passes and that any new
  cross-references (effect ids, skill ids) resolve.
- Describe *what rule or content* you changed and why, so both programmers and game
  designers reviewing the PR can follow it.

## Licensing of contributions

Riftweave is dual-licensed (see [LICENSING.md](LICENSING.md)): **Apache-2.0** for
software/tooling and **CC BY 4.0** for content/documentation. Unless you state
otherwise, contributions you submit are provided under these same licenses —
inbound matches outbound.

## Code of Conduct

By participating, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).
