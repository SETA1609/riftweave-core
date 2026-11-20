# S15 Optional modules

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S15 Optional modules

**Goal:** The post-1.0 modules extend core through modifiers and rows, and core never needs them.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `GD_S15_01_crafting-module` | GD | Crafting module: recipes, quality, stations | POST-01 | M1-07, PO_S01_01_licensing-audit-tag-v1 |
| `GD_S15_02_alchemy-module` | GD | Alchemy module: quality formula and discovery | POST-02 | GD_S15_01_crafting-module |
| `GD_S15_03_enchanting-module` | GD | Magic crafting module: enchanting | POST-03 | GD_S15_01_crafting-module, M3-01 |
| `GD_S15_04_breakthrough-advancement-module` | GD | Breakthrough advancement module | POST-04 | M1-03, PO_S01_01_licensing-audit-tag-v1 |
| `GD_S15_05_companions-module` | GD | Companions and followers module | POST-06 | M1-03, PO_S01_01_licensing-audit-tag-v1 |
| `GD_S15_06_mounts-vehicles-ships-module` | GD | Mounts, vehicles, and ship combat module | POST-07 | M1-02, PO_S01_01_licensing-audit-tag-v1 |
| `GD_S15_07_survival-needs-module` | GD | Survival needs module | POST-08 | M1-04, M1-05, M2-07, PO_S01_01_licensing-audit-tag-v1 |
| `GD_S15_08_weapon-advantage-module` | GD | Weapon advantage module: the guard cycle | POST-09 | M3-03, PO_S01_01_licensing-audit-tag-v1, M5-20 |

**GitHub:** story [#272](https://github.com/SETA1609/riftweave-core/issues/272); tasks [`GD_S15_01_crafting-module` #347](https://github.com/SETA1609/riftweave-core/issues/347), [`GD_S15_02_alchemy-module` #348](https://github.com/SETA1609/riftweave-core/issues/348), [`GD_S15_03_enchanting-module` #349](https://github.com/SETA1609/riftweave-core/issues/349), [`GD_S15_04_breakthrough-advancement-module` #350](https://github.com/SETA1609/riftweave-core/issues/350), [`GD_S15_05_companions-module` #351](https://github.com/SETA1609/riftweave-core/issues/351), [`GD_S15_06_mounts-vehicles-ships-module` #352](https://github.com/SETA1609/riftweave-core/issues/352), [`GD_S15_07_survival-needs-module` #353](https://github.com/SETA1609/riftweave-core/issues/353), [`GD_S15_08_weapon-advantage-module` #354](https://github.com/SETA1609/riftweave-core/issues/354).

**Done when:** every issue above is done, and each module validates on its own and its vectors pass with and without it enabled.
