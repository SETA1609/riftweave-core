# Balance bands

The guideline bands of [`modules/combat.md` §10](../modules/combat.md#10-balance-guidelines) as data, and a report that lists monsters and gear outside them. The bands and the checks are in `ruleset/data/balance/core.json` (`balance.schema.json`); the report is `ruleset/scripts/balance_report.py`. [V 2026-11-06]

## Bands

One row per tier: its levels, the expected damage of one hit before DR, and the DR of a worn set per armor category (light, medium, heavy). The weapon dice band (1d6 to 2d10) and the hits-to-kill bands (standard 4 to 6, boss 8 to 12, minion 1 to 2) are rows too.

## Checks

Each check is a row with a `mode`:

| check | reads | in band when |
| --- | --- | --- |
| `weapon_dice` | every weapon's damage dice | the mean is within the weapon dice band |
| `armor_piece_dr` | every armor piece's `drBase` | it does not pass the highest worn-set DR of its category at any tier |
| `worn_set_dr` | the best piece per slot, by category (a set may use lighter pieces), shields left out | the sum is within the category's worn-set band at some tier |
| `monster_damage` | each monster's strongest attack `amount` or `damage_*` ability magnitude | within the expected damage of its level's tier |
| `monster_hits_to_kill` | each monster's hit points over the middle of its tier's expected damage | within the hits-to-kill band of its `role` (standard, boss, minion) |

- **`gate`** fails CI on a finding; **`warn`** lists it and passes. A check starts as `warn` while the data is known to be outside it, and becomes `gate` when the data is fixed, so the move from warning to gate is one word in data. [A]
- **Accepted.** A finding the owner accepts is a row in `accepted` with its check, its subject, and a reason. It is listed and never fails. A light weapon (dagger, sling) below the dice band is one. [A]
- The dice mean is `n * (s + 1) / 2 + bonus`. A monster with no `role` is read as standard. Both monster checks are gates since the bestiary pass (M4-03). [V 2026-11-06]

`python ruleset/scripts/balance_report.py` prints the findings; `--gate` exits non-zero on a finding of a `gate` check. CI runs it with `--gate`, and `test_balance.py` runs the same rule.
