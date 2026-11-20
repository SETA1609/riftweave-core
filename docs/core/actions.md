# Actions

Author: SETA1609

Modified: 2026-11-21

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

Dodge and block are stances, declared on your turn and in force until your next ([combat.md](combat.md#reading-the-steps)). A fighting stance a perk unlocks is another thing, a way of holding the weapons: it costs nothing at the start of your turn, 1 AP later in it, and stands together with dodge or block ([combat.md](combat.md#fighting-stances)). Dodge spends 3 and doubles your evasion rate. Block is a hold on this pool, not a reaction slot: declare it, spend 2, and the unused 2 comes back if nobody swings; while it holds, a block band of half your `shields` skill stops physical hits. Aimed is not a second attack. It adds 2 to the attack cost.

Dash is not an action. Double move is spending 1 AP twice.

## How far a move goes

```
squares per move = floor(speed / speed_divisor)      speed_divisor = 2 (move row)
```

A race's `speed` is not a distance per round any more; it says how far each `move` goes. At today's speeds of 5 to 7, one move covers 2 or 3 squares, and a pool of 4 spent on moving covers 8 to 12, close to Fallout's 7 to 10 hexes for the same budget. The divisor is the `speed_divisor` on the `move` row of `actions/core.json`, so a module changes it without code. Divinity: Original Sin 1 and 2 price movement the same way: an AP buys a distance that grows with speed. It differs here in that speed comes from the race, not from an attribute. [V 2026-11-21, `actions` vectors]

What raises the price of a move changes its AP, never its squares. Encumbrance and armor emit `add move +1` ([combat.md §8](../modules/combat.md#8-encumbrance-and-armor-action-costs)); difficult terrain doubles the cost ([combat.md §5](../modules/combat.md#5-cover-flanking-terrain-prone)). An encumbered human (speed 6) pays 2 AP for each move and still covers 3 squares. [V 2026-11-21, `actions` vectors]

The `actions` command takes `speed` and returns `squares`, the squares its moves cover; without a speed it returns no `squares`. A real-time client moves at `floor(speed / 2)` squares per `seconds_per_turn / pool` seconds, so a pool of 4 walks a human 3 squares every 1.25 seconds. [A]

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
events    = listed with the effects (disarm), for the caller to apply
save      = with a save_bonus: on a hit, a maneuver attempt the defender saves against
            (#the-maneuver-attempt); the effects and events land only on a failed save
needs     = one of its ranges (any when it lists none), and a weapon with one of its
            weapon_properties (any weapon when it lists none)
```

| Maneuver | AP | Attack | Damage | Needs | On a hit | Unlocked by |
| --- | --- | --- | --- | --- | --- | --- |
| Heavy Swing | 4 (attack + 1) | −10 | ×1.5 | melee | staggered (4 s) | Wind-Up (STR 5) |
| Quick Strike | 2 (attack − 1) | 0 | ×0.5 | melee, a light weapon | — | Snap Cut (AGI 5) |
| Trip | 3 (attack) | −10 | ×0.5 | melee | knockdown (5 s, prone) on a failed save | Low Sweep (AGI 5, a melee skill 30) |
| Disarm | 3 (attack) | −15 | ×0.5 | melee | the `disarm` event (drops its weapon) on a failed save | Bind and Wrench (STR 5, a melee skill 30) |
| Tripping Shot | 3 (attack) | −10 | ×0.5 | ranged | knockdown (5 s, prone) on a failed save | Low Aim (PER 5, a ranged skill 30) |
| Disarming Shot | 3 (attack) | −15 | ×0.5 | ranged | the `disarm` event on a failed save | Knuckle Shot (PER 5, a ranged skill 30) |

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
  A monster may also hold the unlock and rider perks in its `perks`, as a player does: its
  maneuvers are the ones it lists plus the ones its perks unlock (`monster_sheet` reports both,
  GD_S05_04).
- **Trip and disarm** (GD_S05_04) trade half the damage and a harder roll for a control effect,
  in a melee line and a ranged line. Disarm is the same `disarm` event as the arms called shot
  (`combat/situations.json`). The attack resolves as a normal attack first; a hit then makes the
  [maneuver attempt](#the-maneuver-attempt).

## The maneuver attempt

A trip or disarm is a normal attack, then a maneuver attempt the defender saves against
(GD_S05_04). A maneuver with `save_bonus`, or a [rider](#rider-perks), makes one on a hit that is
not evaded; a miss, an evade, or a block makes none.

```
attribute = the weapon's attribute, the attack's damage.ability (AGI for a blade, STR for a
            mace, PER for a bow); an attempt without one is an error
DC        = attribute × maneuver_save_multiplier (10, combat/pipeline.json)
            + the maneuver's save_bonus (20 for trip and disarm; a rider adds none)
target    = max(100 − DC, 0): the defender's d100 save; the caller rolls it
```

- **The result** reports `save` as `{attribute, dc, target}` with the `effects` and `events`. The
  caller rolls the save and applies them only when it fails. The kernel does not roll it, as it
  does not roll the attack. [A]
- **The attacker's attribute + mods.** The attributes the caller passes are the attacker's
  modified ones, so a buff to AGI raises a blade's DC. The maneuver's `save_bonus` is what its
  −10 or −15 on the attack and its half damage buy: a dedicated maneuver lands more often than a
  rider. No perk or situation modifies the DC or the save yet. [A]
- **It replaces the clear margin.** The first draft let a rider ride only on a hit that cleared
  the bands by 20. A save does that job better: every hit makes the attempt, and the attacker's
  attribute against the defender's save decides it, so the margin rule is gone. [A]
- **Heavy Swing and Quick Strike** have no `save_bonus`: their effects land on any hit, as before.

## Rider perks

A rider perk puts a maneuver's attempt on the holder's plain attacks (GD_S05_04). It names the
maneuver in `rider` and needs the maneuver's unlock perk. Each range has a trip line and a disarm
line, and the two lines of a range exclude each other: with the trip line you cannot take the
disarm line, and the other way round ([progression.md](progression.md#excluded-perks)). The melee
lines do not exclude the ranged ones. [A]

| Line | Unlock perk | Rider perk | Rides | Excludes |
| --- | --- | --- | --- | --- |
| Melee trip | Low Sweep (AGI 5, a melee skill 30) | Hooking Blows (level 8, AGI 6) | Trip | the melee disarm line |
| Melee disarm | Bind and Wrench (STR 5, a melee skill 30) | Wrenching Blows (level 8, STR 6) | Disarm | the melee trip line |
| Ranged trip | Low Aim (PER 5, a ranged skill 30) | Hamstring Shots (level 8, PER 6) | Tripping Shot | the ranged disarm line |
| Ranged disarm | Knuckle Shot (PER 5, a ranged skill 30) | Stripping Shots (level 8, PER 6) | Disarming Shot | the ranged trip line |

- **A plain attack only.** The `combat` input's `attacker.perks` names the held perks. With no
  `maneuver`, a held rider whose maneuver fits the attack rides: the attack keeps its own AP,
  attack, and damage, and a hit makes the attempt with no save bonus. The result reports `rider`
  and, on a hit, `save`, `effects`, and `events`. An attack performed as a maneuver uses only
  that maneuver. A melee rider stays off a ranged shot, and a ranged rider off a melee blow.A Boss Strike shield bash carries no rider: a shield has no weapon attribute for the attempt (GD_S05_03). [A]
- **Two riders that fit one attack are an error.** The exclusions keep it from happening to a
  built actor; `combat` still fails on it, and `monster_sheet` and `validate.py` fail a monster
  or template that holds two perks that exclude each other.
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
