# DE_S02_03_monster-hp-computed Third normal form: monster hit points are computed

Date: 2026-11-20

- **Decision:** no monster stores HP. The `monster_sheet` command computes it: the player formula (`derived/core.json`, the `hp` row) before perks, times the role's `monster_hp_multiplier` in `progression/core.json` (standard 0.4, boss 0.8, minion 0.2), rounded down and at least 1, then the perks' hp modifiers. `validate.py` (`check_monster_hit_points`) fails `hitPoints`, `hitPointsPerLevel` or `hit_points_multiplier` anywhere in a monster row or its overrides, a creature-type base, a template, or a module's monsters. [V 2026-11-20, `monster_sheet` vectors, `test_bestiary.py`, `go test ./...`]
- **Why a role term:** the plain player formula gives a standard monster about 2.4 times its authored HP, a boss 1.2 times and a minion 4.4 times (medians; the bestiary was written against the hits-to-kill bands of `balance/core.json`, the wolf example rows were not), so keeping the fights the same size by attributes alone would have meant cutting END on nearly every row. The issue allows a creature-type term [A]; the role already chooses the balance band, so the multiplier hangs on the role instead of on the 8 creature types. [A] The three values are fitted to the bestiary: with them, 40 of the 59 rows land within 20% of their authored HP with no change. [V 2026-11-20]
- **The minion template** sets `role: minion` instead of multiplying stored HP by 0.25; its END -2 still lowers the formula. A minion skeleton at level 22 stays under half its row's HP. [V 2026-11-20, `test_spells.py`]
- **Moved by more than 20%, or out of their band, and what was done.** Computed is the new formula on the old attributes. Each change is the smallest END change that keeps HP within 20% of the authored value and inside the role's hits-to-kill band, set in `overrides.attributes.end`. [V 2026-11-20, `balance_report.py --gate`]

  | Monster | Role, level | Authored HP | Computed | END | Final HP | Done |
  |---|---|---|---|---|---|---|
  | bandit | standard 1 | 24 | 17 (-29%) | 3 → 5 | 23 (-4%) | END |
  | skeleton | standard 1 | 22 | 20 (-9%, under band) | 4 → 5 | 23 (+5%) | END |
  | zombie | standard 1 | 30 | 20 (-33%) | 4 → 6 | 26 (-13%) | END |
  | giant_rat | minion 1 | 8 | 10 (+25%) | 4 → 3 | 8 (0%) | END |
  | grove_stalker | standard 3 | 30 | 23 (-23%) | 4 → 5 | 26 (-13%) | END |
  | acid_ooze | standard 4 | 32 | 38 (+19%, over band) | 8 → 6 | 31 (-3%) | END |
  | frost_ooze | standard 4 | 30 | 38 (+27%) | 8 → 6 | 31 (+3%) | END |
  | fire_ooze | standard 5 | 32 | 39 (+22%) | 8 → 6 | 33 (+3%) | END |
  | stone_golem | boss 5 | 60 | 85 (+42%) | 9 → 6 | 66 (+10%) | END |
  | ember_drake | boss 6 | 78 | 69 (-12%, under band) | 6 → 7 | 76 (-3%) | END |
  | ghoul | standard 6 | 42 | 34 (-19%, under band) | 6 → 7 | 38 (-10%) | END |
  | orc_raider | standard 6 | 45 | 31 (-31%) | 5 → 7 | 38 (-16%) | END |
  | iron_sentinel | standard 7 | 55 | 42 (-24%) | 8 → 9 | 46 (-16%) | its END 8 override dropped; it inherits 9 |
  | thunder_ram | standard 7 | 50 | 36 (-28%) | 6 → 8 | 42 (-16%) | END |
  | troll | boss 13 | 100 | 134 (+34%) | 8 → 4 | 109 (+9%) | END |
  | storm_giant_youth | standard 15 | 70 | 55 (-21%) | 8 → 9 | 58 (-17%) | its END 8 override dropped; it inherits 9 |
  | iron_colossus | boss 20 | 150 | 191 (+27%) | 18 → 12 | 152 (+1%) | END |
  | wolf | standard 3 | 59 | 23 (-61%) | 4 | 23 | kept; now in band, accepted entry removed |
  | dog | standard 3 | 51 | 20 (-61%) | 3 | 20 | kept; accepted (3.6 hits, band 4–6) |
  | dire_wolf | standard 3 | 67 | 41 (-39%) | 5 | 41 | kept; accepted (7.5 hits) |
  | undead_horse | standard 4 | 71 | 28 (-61%) | 5 | 28 | kept; now in band, accepted entry removed |
  | demonic_hound | standard 6 | 71 | 28 (-61%) | 4 | 28 | kept; accepted (2.9 hits) |
  | elder_wolf | standard 12 | 151 | 60 (-60%) | 11 | 60 | kept; accepted (6.3 hits) |

  Rows derived from a changed row move with it and stay within 20%: bone_colossus (from skeleton, 105 → 102), bronze_guardian (from stone_golem, 52 → 50), ancient_dragon (from ember_drake, 210 → 216). bandit_captain keeps END 6 and its HP.
- **The wolf example rows** (wolf, dog, dire_wolf, elder_wolf, demonic_hound, undead_horse) took their HP from the player formula on purpose (`example_hp`, docs/modules/monsters.md#generated-monsters), and four of them sat outside the band as accepted findings. They keep their attributes, so the formula they illustrate is unchanged; their HP falls to the standard share (about 60% lower), which brings wolf and undead_horse into the band. [A] Raising their END to stay within 20% would have pushed them far above the band and broken the dog-weaker-than-wolf rule the tests hold.
- **Playtests:** `playtest_driver.py` reran with the same seed; the 8 builds still reach level 30 with no fight lost, and fights won move by at most 2 per build (240 fights). [V 2026-11-20, `docs/playtests/report.json`]
- **Rounding (review, 2026-11-20):** the product is floored with a tolerance of 1e-9 in Go and Python, so a share stored just under its decimal does not lose a point: with a module share of 0.7, a formula of 90 gave 62 instead of 63, and 16 such values under 1000 did (0.35: 7). Today's shares (0.4, 0.8, 0.2) are stored just above their decimals and give the same results either way. [V 2026-11-20, `TestScaledHitPoints`, the `monster_sheet` vectors in Go]
- **Size:** 21 counted files, over the 20 of CONTRIBUTING.md rule 2. The ruleset and backend split of that rule does not work here: the Go tests run the same `monster_sheet` vectors (`backend/core/core_test.go`), so a ruleset PR without the Go change would leave `development` red (rule 3). The 21 are one change: spec, schemas and contracts, vectors, data, both kernels, and the playtest report. [A]
- **Risk:** **breaking**: a client that sent `hitPoints` to `monster_sheet` or `print` must send `role`. `frontend/src/api/openapi.d.ts` still lists `hitPoints` because it is not regenerated (it needs a running server), as in GD_S05_01. The `ECS` example in `docs/modules/monsters.md` now writes `role` instead of `hitPoints`.
- **Open:** whether the multipliers should be per creature type as well as per role (a construct tougher than a beast of the same END) [U]; whether the wolf example rows should be rebalanced in the balance pass rather than stay accepted [U].
