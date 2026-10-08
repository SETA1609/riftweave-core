# Actions

Author: SETA1609

One budget. Costs live in `ruleset/data/actions/core.json`. A turn spends AP. A real-time client maps the same numbers to cooldowns. There is no bonus action and no reaction slot.

## Pool

```
pool  = 2 + floor(AGI / 3)
cap   = 10
carry = min(unspent, cap)
rest  = refill
```

The size matches the derived `ap` function. The cap is why unspent points carry. Without it a module can bank a second turn. A short rest refills the pool. Fallout spends a small pool the same way.

## Costs

```
attack 3
cast 3
move 1
disengage 1
dodge 3
item 2
block 2, reserved, refund if unused
aimed +2 on the attack cost
```

Cast AP is static. A spell also spends its mana cost. AP does not vary by spell. A petty spark and a heavy bolt both cost 3 AP. The bolt costs more mana. A perk that wants a cheaper cast emits `add` on the cast cost. It does not edit the spell.

Block is a hold on this pool, not a reaction slot. Declare it, spend 2, and the unused 2 comes back if nobody swings. Aimed is not a second attack. It adds 2 to the attack cost.

Dash is not an action. Double move is spending 1 AP twice.
