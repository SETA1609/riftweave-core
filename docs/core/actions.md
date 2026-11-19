# Actions

Author: SETA1609

Modified: 2026-11-19

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
attack 3, by loadout (2 to 5)
jab 2, by loadout (2 or 3, the off hand alone)
full_attack every AP left, at least attack + 1, by loadout
circle_cut every AP left, at least attack + 1 (Circle Cut perk, two-handed weapon)
cast 3
move 1
disengage 1
dodge 3
item 2
block 2, reserved, refund if unused
command 2, an order to summons (commanding.md)
aimed +2 on the attack cost
```

The attack, the jab, the full attack, and the circle cut cost what the loadout's row of `ruleset/data/attack_sequences/core.json` says (`by_loadout`); the `cost` in `actions/core.json` is the cost with one one-handed weapon, for a caller that knows no loadout. The full attack and the circle cut spend every AP left (`spends_all`). The strikes each makes are in [combat.md](combat.md#attacks-by-loadout) (GD_S05_01).

A perk changes a cost with `add` on the action's key: the Sidestep perk emits `add dodge -1`, and the cost is the row's cost after the actor's modifiers on that key, stacked by [modifier.md](modifier.md), never below 0. [V 2026-11-06]

Cast AP is static. A spell also spends its mana cost. AP does not vary by spell. A petty spark and a heavy bolt both cost 3 AP. The bolt costs more mana. A perk that wants a cheaper cast emits `add` on the cast cost. It does not edit the spell.

Dodge and block are stances, declared on your turn and in force until your next ([combat.md](combat.md#reading-the-steps)). Dodge spends 3 and doubles your evasion rate. Block is a hold on this pool, not a reaction slot: declare it, spend 2, and the unused 2 comes back if nobody swings; while it holds, a block band of half your `shields` skill stops physical hits. Aimed is not a second attack. It adds 2 to the attack cost.

Dash is not an action. Double move is spending 1 AP twice.

## Maneuvers

A maneuver is a way of performing an action, today always an attack, for more or fewer AP. It
is a row of `ruleset/data/maneuvers/core.json` (`core:maneuver/<key>`), and the fighter's
counterpart of a spell: it draws its effects from the shared pool (`appliedEffect`), but costs
AP or stamina rather than mana (M3-24).

```
ap        = max(cost of `modifies` + the maneuver's ap, 1)      real time: ap × stamina_per_ap
attack    = the maneuver's attack, an add on the attack target, as a called shot's
damage    = floor(damage × damage_multiplier), after the called shot's multiplier
effects   = listed on a hit that is not evaded, for the caller to apply, as a called shot's riders
needs     = one of its ranges (any when it lists none), and a weapon with one of its
            weapon_properties (any weapon when it lists none)
```

| Maneuver | AP | Attack | Damage | Needs | On a hit | Unlocked by |
| --- | --- | --- | --- | --- | --- | --- |
| Heavy Swing | 4 (attack + 1) | −10 | ×1.5 | melee | staggered (4 s) | Wind-Up (STR 5) |
| Quick Strike | 2 (attack − 1) | 0 | ×0.5 | melee, a light weapon | — | Snap Cut (AGI 5) |

- **AP follows the attack.** The totals are for an attack with one one-handed weapon, 3 AP. The
  maneuver stores only its difference (`ap`), so its AP moves with the loadout's attack AP: a
  Heavy Swing with two one-handed weapons is 5 + 1, a Quick Strike with one light weapon 2 - 1
  (the `combat` input's `loadout`, GD_S05_01).
- **Mana.** A maneuver may name a small `mana` cost, for one with a touch of magic; the two
  starting rows cost none. [A]
- **Effects** must declare the `maneuver` channel, which `validate.py` checks
  (`CHANNEL_REQUIREMENTS`), as it checks `ingredient` and `enchantment`.
- **Every maneuver needs a perk** (owner 2026-11-18). A character may attempt a maneuver only
  while holding a perk that lists it in `unlocks`; nobody knows one untrained. `validate.py`
  fails on a maneuver no perk unlocks. An unlock perk gives nothing else; a perk's skill
  bonus must cover two skills or be at least +10 (M5-78). The passive damage perk (`full_weight`,
  once `power_attack`) stays a passive bonus and does not unlock Heavy Swing. [A]
- **Monsters** list `maneuvers` on their row or template, and need no perk for them, so a troll
  swings hard with the same row a fighter uses ([monsters.md](../modules/monsters.md#maneuvers)).
- **Names** are our own, not feat names from other games (CONTRIBUTING.md).

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
