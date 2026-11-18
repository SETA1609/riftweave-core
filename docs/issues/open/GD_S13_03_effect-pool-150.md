#### GD_S13_03_effect-pool-150 Effect pool to about 150
- Status: todo
- Priority: medium
- Needs: PO_S01_01_licensing-audit-tag-v1, M3-06
- Was: M6-04
- Do: Players build spells from the shared effect pool, so the pool sets how many spells exist, not the written spell list. Grow `effects/core.json` from 84 to about 150, covering every color's schools and the five phases, with each effect's `channels` and `craftable` set (M3-06). Add the effects M6-05 and M6-06 need before their spells.
- Keys: new `core:effect/<key>` rows
- Files: ~5 (ruleset/data/effects/core.json, ruleset/schemas/effect.schema.json if a field is needed, ruleset/tests/vectors/, docs/modules/magic.md)
- See: docs/modules/magic.md, docs/core/casting.md
- Done when: about 150 effects validate; every color and every phase has effects; no effect is unused or unreserved; `validate.py` and the tests pass.
