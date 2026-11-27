# S01 Release 1.0 gates

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S01 Release 1.0 gates

**Goal:** The SRD is playtested, balanced, legally clean, and tagged `v1.0.0`.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `PO_S01_01_licensing-audit-tag-v1` | PO | Licensing and attribution audit, tag v1.0.0 | M5-09 | M2-01, M2-02, M2-03, M2-04, M2-05, M3-02, M3-03, M3-04, M4-03, M4-04, M4-05, M5-01, M5-05, M5-06, M5-07, M5-08, GD_S01_03_perk-throughput-veto, M5-20 |
| `QA_S01_02_alpha-playtest-human-cli` | QA | Alpha playtest through the human CLI | M5-18 | M2-01, M2-06, M3-02, M3-06, M4-03, M5-04, APP-04, APP-05, APP-08 |
| `GD_S01_03_perk-throughput-veto` | GD | Perk throughput veto and distillation | M5-19 | QA_S01_02_alpha-playtest-human-cli, BE_S08_10_compendium-catalog-cli, FE_S08_11_compendium-tables-page, BE_S08_05_bestiary-knowledge-cli, FE_S08_06_bestiary-page |
| `GD_S01_04_skill-damage-pacing` | GD | Skill and damage pacing over thirty levels | M5-50 | QA_S01_02_alpha-playtest-human-cli |

**GitHub:** story [#259](https://github.com/SETA1609/riftweave-core/issues/259); tasks [`PO_S01_01_licensing-audit-tag-v1` #278](https://github.com/SETA1609/riftweave-core/issues/278), [`QA_S01_02_alpha-playtest-human-cli` #279](https://github.com/SETA1609/riftweave-core/issues/279), [`GD_S01_03_perk-throughput-veto` #280](https://github.com/SETA1609/riftweave-core/issues/280), [`GD_S01_04_skill-damage-pacing` #281](https://github.com/SETA1609/riftweave-core/issues/281).

**Done when:** every issue above is done, and the alpha playtest is logged, the perk throughput and pacing checks pass, the licence audit is done, and `v1.0.0` is tagged.
