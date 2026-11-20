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
| `M5-61` | GD | Suggested builds: the ranged path, the arcane archer, and the ranger | — | M5-47, M5-60 |

**GitHub:** story [#262](https://github.com/SETA1609/riftweave-core/issues/262); tasks [`DE_S04_01_base-perk-schema` #288](https://github.com/SETA1609/riftweave-core/issues/288), [`DE_S04_02_perks-compose-from-base` #289](https://github.com/SETA1609/riftweave-core/issues/289), [`GD_S04_03_perk-skill-bonuses` #290](https://github.com/SETA1609/riftweave-core/issues/290).

**Done when:** every issue above is done, and perks validate against the base schema, compose from a base perk, and the ranged builds are in `ruleset/data/builds/`.
