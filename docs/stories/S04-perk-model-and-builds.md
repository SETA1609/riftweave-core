# S04 Perk model and builds

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S04 Perk model and builds

**Goal:** Perks share one base schema and compose from it, perks can grant skill bonuses, and every play style has a suggested build.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `DE_S04_01_base-perk-schema` | DE | A base perk schema, and kinds that extend it | M5-58 | — |
| `DE_S04_02_perks-compose-from-base` | DE | Perks compose from a base perk | M5-59 | DE_S04_01_base-perk-schema |
| `GD_S04_03_perk-skill-bonuses` | GD | Perk skill bonuses: two skills or +10, an extra tag skill, and a skill-point ladder | M5-78 | M5-67 |
| `GD_S04_04_ranged-builds` (reserved) | GD | see `M5-61` | M5-61 | open pull request #237: keeps `M5-61` until it merges |

**Done when:** every issue above is done, and perks validate against the base schema, compose from a base perk, and the ranged builds are in `ruleset/data/builds/`.
