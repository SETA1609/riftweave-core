# riftweave-core

Riftweave — A data-driven, open-source cRPG ruleset for games and TTRPGs.

All game data is defined as JSON files, validated against JSON Schemas. The data
and schemas live under `ruleset/`.

It is a **classless**, **d100 roll-under** system: characters are defined by eight
attributes, point-buy skills with tagging, and perks. See
[docs/modules/progression.md](docs/modules/progression.md) for the system overview.

## Project structure

```
docs/
  modules/            Documentation for optional, composable extensions
    README.md
    progression.md    Core system: attributes, skills, perks, dice, magic
    crafting.md
    alchemy.md
    magic-crafting.md
ruleset/
  schemas/            JSON Schema definitions (draft-07)
    schema.json       Shared types & definitions (incl. color + appliedEffect)
    ability.schema.json
    race.schema.json
    skill.schema.json
    effect.schema.json    Shared effect registry (the pool)
    ingredient.schema.json
    equipment.schema.json
    spell.schema.json
    feature.schema.json   Perks, traits, and racial features
  data/               Game data files (JSON)
    abilities/core.json   Eight attributes (STR PER END INT WIL AGI CHA LCK)
    races/core.json
    skills/core.json
    effects/core.json     Shared effect pool (magic + alchemy)
    ingredients/core.json Alchemy reagents (effects from the pool)
    equipment/weapons.json
    equipment/armor.json
    equipment/consumables.json  Potions/poisons/coatings (effects from the pool)
    spells/core.json      Color-school spells (effect compositions)
    features/core.json    Perks and traits
  scripts/
    validate.py       Schema validation script
  requirements.txt    Python dependencies
```

See [docs/modules/README.md](docs/modules/README.md) for the current thinking on
extensibility and planned optional modules (crafting, alchemy, magic crafting, etc.).

## Validation

```bash
pip install -r ruleset/requirements.txt
python ruleset/scripts/validate.py
```

Or with Docker:

```bash
docker build -t riftweave-validate .
docker run --rm riftweave-validate
```

CI runs automatically via `.github/workflows/validate-ruleset.yml` on every push
or PR that touches the `ruleset/` directory.

## License

Riftweave is dual-licensed so it can be used freely in both software and
games/tabletop:

- **Software & tooling** (`ruleset/scripts/`, `ruleset/schemas/`, `Dockerfile`, CI)
  — [Apache License 2.0](LICENSE).
- **Game content & documentation** (`ruleset/data/`, `docs/`, prose) —
  [Creative Commons Attribution 4.0](LICENSE-CONTENT).

See [LICENSING.md](LICENSING.md) for the full breakdown, attribution guidance, and a
note on trademarks. Contributions are accepted under these same licenses.
