#### GD_S14_04_gm-encounter-travel-tools GM tools: encounter budgets and travel
- Status: todo
- Priority: medium
- Needs: PO_S01_01_licensing-audit-tag-v1
- Was: M6-10
- Do: A referee needs to build a fight of the right difficulty and move a party across a map. Turn the encounter suggestions in `docs/modules/advancement.md` into an encounter budget per party level and size, as data with a formula, and add travel: pace, distance per day in units, and what a forced march costs. Leave hunger and thirst to the survival module (POST-08).
- Keys: `encounters` and `travel` rows [A]
- Files: ~8 (docs/core/ or docs/modules/encounters.md, docs/modules/travel.md, ruleset/schemas/, ruleset/data/, ruleset/tests/vectors/, docs/guide/gm.md)
- See: docs/modules/advancement.md, docs/guide/gm.md, docs/issues/open/GD_S15_07_survival-needs-module.md
- Done when: an encounter budget and a day of travel each compute from data with vectors; `gm.md` explains both; `validate.py` and the tests pass.
