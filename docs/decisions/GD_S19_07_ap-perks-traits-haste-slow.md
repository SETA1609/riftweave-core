# GD_S19_07_ap-perks-traits-haste-slow Perks, creation traits, and conditions that change the AP pool

Date: 2026-11-23

- **Decision:** modifiers on `ap` change the pool, and the pool after them is kept between `pool.floor` (1) and `pool.cap` (10), in the derived `ap` and in the `actions` command's `pool` ([actions.md#pool](../core/actions.md#pool)). [V 2026-11-23, derived and actions vectors in Python and Go]
  - **Battle Tempo** (perk, level 12, one rank): `add ap 1`, the Action Boy shape. `take_perk` refuses it below level 12 and a second time. [V 2026-11-23, progression vectors]
  - **Heavy Handed** (creation perk, the Bruiser shape): `multiply melee_damage 1.25`, `add ap -1`. The ×1.25 matches Full Weight. [A]
  - **Glass Frame** (creation perk, the Glass Cannon shape): `add ap 1`, `multiply hp 0.75`. [A]
  - **Hasted** (condition): `add ap 2`, applied by the new `haste` effect and removed by `slow`, as `slowed` is removed by haste. [A]
- **Slowed is unchanged:** the issue asked for `add ap -2` only if `slowed` did not already change AP, and it does (`action_cost +1`, every action 1 AP more). Read in the spec step. [V 2026-11-23, `conditions/core.json`]
- **The floor is data:** `pool.floor` in `actions/core.json`, beside `pool.cap`; the derived `ap` reads both, so the two pools agree.
- **Vectors:** the derived vector for Heavy Handed carries only its `ap` modifier, because `melee_damage` is still a pending target and the gate keeps pending targets out of golden vectors.
- **Found:** the derived `ap` row and `actions/core.json`'s `pool` both hold the base, the per-point, and the divisor of the pool, a copy left for a normal-form issue.
