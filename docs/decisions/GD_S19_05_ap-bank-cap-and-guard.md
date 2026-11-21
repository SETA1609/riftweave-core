# GD_S19_05_ap-bank-cap-and-guard Carry a little AP, and turn the rest into guard

Date: 2026-11-21

- **Decision:** AP left at the end of a turn is held: +5 `evasion` per AP (`pool.guard_per_ap`), spendable on an opportunity or readied attack, and at most 2 of it (`pool.carry`) join the next pool. `pool.cap` 10 still bounds a turn's AP. The rule lives in [actions.md](../core/actions.md#held-ap). [V 2026-11-21, `actions` vectors in Python and Go]
- **Why:** carry up to the cap of 10 against a pool of 3 to 5 let two quiet turns bank two casts, a burst a real-time client does not have.
- **Key:** `pool.carry` changes from the string `min(unspent, cap)`, which no code read, to the integer 2. A consumer that read the string breaks. (breaking) [V 2026-11-21, no reader in ruleset/scripts, backend, or frontend]
- **Loadouts:** every loadout attacks at least every second turn at pool 3 and 4 under the new carry, so no `ap` changes ([table](../core/actions.md#held-ap)). [V 2026-11-21]
- **Open:** the guard is returned as a number; the caller adds it to the defender's `evasion` as an `add` modifier. A `held` field on the stored actor, and spending held AP in the combat pipeline, come with `GD_S19_04_ready-and-opportunity-attacks`. [A] The numbers 5 and 2 wait for a playtest. [A]
- **Risk:** +5 per AP at pool 5 is +25 evasion for a turn of doing nothing, which may make turtling stronger than dodge (double evasion) for a high-evasion actor. [A]
