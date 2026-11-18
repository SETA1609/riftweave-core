#### GD_S13_01_combat-stealth-perks Combat and stealth perks: about 60 more
- Status: todo
- Priority: medium
- Needs: PO_S01_01_licensing-audit-tag-v1, GD_S01_03_perk-throughput-veto
- Was: M6-02
- Do: Classes and feats both become perks in a classless system, and a character makes 30 perk picks, so breadth matters: the D&D SRD 5.2 offers 12 classes with a subclass each and about 17 feats [A, approximate]. Add about 60 perks in the `combat` and `stealth` categories (35 today), so each weapon skill, armor class, and the stealth skills have a perk line worth taking from level 2 to 30. Start from the gaps M5-47 lists per build and from M5-19's verdicts, so no perk M5-19 cut comes back. Every mechanical bonus is a modifier on a registered target (M5-32), and every new target gets a vector or stays `pending`. Bonuses on a d100 target are multiples of 5 (M6-11).
- Keys: new `core:feature/<key>` perks
- Files: ~5 (ruleset/data/features/core.json, ruleset/data/modifier_targets/core.json, ruleset/tests/vectors/, docs/modules/progression.md, docs/decisions/GD_S13_01_combat-stealth-perks.md)
- See: docs/modules/progression.md, docs/issues/done/M5-47.md, docs/issues/open/GD_S01_03_perk-throughput-veto.md
- Done when: about 95 combat and stealth perks validate; every weapon skill has at least 4 perks; no perk has an unregistered target; `validate.py` and the tests pass.
