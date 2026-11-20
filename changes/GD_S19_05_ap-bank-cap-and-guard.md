### GD_S19_05_ap-bank-cap-and-guard Carry a little AP, and turn the rest into guard
- Changed: `pool.carry` is the integer 2, the most held AP that carries into the next turn, not the string `min(unspent, cap)` (breaking).
- Added: `pool.guard_per_ap` 5: each held AP gives +5 evasion until the next turn.
- Added: the `actions` command takes `carried` and `held_spent` and returns `available`, `held`, `guard`, `carry`, and `next_pool`.
