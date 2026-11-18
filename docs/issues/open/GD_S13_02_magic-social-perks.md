#### GD_S13_02_magic-social-perks Magic, social, and general perks: about 70 more
- Status: todo
- Priority: medium
- Needs: PO_S01_01_licensing-audit-tag-v1, GD_S01_03_perk-throughput-veto
- Was: M6-03
- Do: The second half of the perk expansion (M6-02 is the first): about 70 perks in `magic`, `social`, `general`, `survival`, `knowledge`, and `utility` (53 today), so each magic color, each social and knowledge skill, and the summoner and support builds have a perk line from level 2 to 30. Same rules as M6-02: start from M5-47's gaps and M5-19's verdicts, modifiers on registered targets, bonuses in multiples of 5.
- Keys: new `core:feature/<key>` perks
- Files: ~5 (ruleset/data/features/core.json, ruleset/data/modifier_targets/core.json, ruleset/tests/vectors/, docs/modules/progression.md, docs/decisions/GD_S13_02_magic-social-perks.md)
- See: docs/modules/progression.md, docs/modules/magic.md, docs/issues/done/M5-47.md
- Done when: about 125 perks in these categories validate; each color has at least 4 magic perks; `validate.py` and the tests pass.
