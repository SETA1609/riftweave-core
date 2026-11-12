# Actions

Author: SETA1609

One budget. Costs live in `ruleset/data/actions/core.json`. A turn spends AP. A real-time client spends stamina for the same actions, at a fixed rate. There is no bonus action and no reaction slot.

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

A perk changes a cost with `add` on the action's key: the Sidestep perk emits `add dodge -1`, and the cost is the row's cost after the actor's modifiers on that key, stacked by [modifier.md](modifier.md), never below 0. [V 2026-11-06]

Cast AP is static. A spell also spends its mana cost. AP does not vary by spell. A petty spark and a heavy bolt both cost 3 AP. The bolt costs more mana. A perk that wants a cheaper cast emits `add` on the cast cost. It does not edit the spell.

Block is a hold on this pool, not a reaction slot. Declare it, spend 2, and the unused 2 comes back if nobody swings. Aimed is not a second attack. It adds 2 to the attack cost.

Dash is not an action. Double move is spending 1 AP twice.

## One rule set, two clocks

There is one set of rules. A table counts turns and AP; a real-time client counts seconds and
stamina. Each action keeps one cost, and the two convert by fixed rates in `actions/core.json`:

```
turns            = ceil(duration / seconds_per_turn)      seconds_per_turn = 5
real-time cost   = AP cost × stamina_per_ap               stamina_per_ap   = 5
AP at a table    = stamina restored / stamina_per_ap, rounded down
```

- **Mana pays for magic in both.** A spell costs its mana and the `cast` AP at a table, its mana
  and the `cast` stamina in real time. An aura is magic, so it costs mana. No spell costs stamina.
- **Physical actions.** At a table an attack spends 3 AP; in real time it spends 15 stamina. A
  real-time client may also show the cost as a cooldown in the same proportion.
- **Durations.** An effect's `duration` is seconds. At a table it lasts
  `ceil(duration / seconds_per_turn)` turns, and a modifier's `turns` and an effect's seconds use
  the same conversion. A tick covers the seconds that remain, so a 12-second regeneration gives
  the same total over 3 turns as over 12 seconds ([modifier.md](modifier.md#effects-over-time)).
- **Restoring stamina at a table.** An effect that restores stamina restores AP instead:
  stamina divided by `stamina_per_ap`, rounded down, with the remainder kept toward the next AP.
  That AP arrives at the start of the actor's next turn, never the current one, and the pool
  stays capped at its `cap` of 10. Without the delay a Stamina Potion (30 stamina is 6 AP, for
  an `item` costing 2) or Vigor (20 stamina is 4 AP, for a `cast` costing 3) would pay for
  itself and the turn would never end. There is no separate one-item-per-turn cap: the delay
  removes the loop, and toxicity (M2-09) limits drinking again.
- **Food** is never eaten in combat (M2-07), so at a table it fills the carried pool before a
  fight.

[A] for `stamina_per_ap` 5. [V 2026-11-12, `actions` and `consume` vectors]
