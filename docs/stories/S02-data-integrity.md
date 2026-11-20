# S02 Normal forms and data integrity

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S02 Normal forms and data integrity

**Goal:** The schemas and data meet the third normal form, with one casing for properties, and an integrity gate keeps them there.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `DE_S02_01_one-property-casing` | DE | One property casing | M5-40 | M5-37 |
| `DE_S02_02_second-normal-form` | DE | Second normal form: a reference is the key alone | M5-71 | M5-70 |
| `DE_S02_03_monster-hp-computed` | DE | Third normal form: monster hit points are computed | M5-74 | M5-73 |
| `DE_S02_04_descriptions-without-numbers` | DE | Third normal form: a description does not restate its numbers | M5-75 | DE_S02_02_second-normal-form |

**GitHub:** story [#260](https://github.com/SETA1609/riftweave-core/issues/260); tasks [`DE_S02_01_one-property-casing` #282](https://github.com/SETA1609/riftweave-core/issues/282), [`DE_S02_02_second-normal-form` #283](https://github.com/SETA1609/riftweave-core/issues/283), [`DE_S02_03_monster-hp-computed` #284](https://github.com/SETA1609/riftweave-core/issues/284), [`DE_S02_04_descriptions-without-numbers` #285](https://github.com/SETA1609/riftweave-core/issues/285).

**Done when:** every issue above is done, and validate.py and test_integrity.py enforce the normal forms and the casing on every collection.
