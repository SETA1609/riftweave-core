# Commanding summons

Modified: 2026-11-19

The commander summoner's summons do the damage, and the caster spends their own turn to drive
them (GD_S06_03). There are two ways, and the `command_summons` function works out what each
summon gets from both. The perks are rows in `ruleset/data/features/core.json`, the held command
is a spell in `ruleset/data/spells/core.json`, and the vectors are in
`ruleset/tests/vectors/command_summons.json`.

## Orders

An **order** is the `command` action (`actions/core.json`, 2 AP [A]): the commander spends AP and
names the summons it reaches. A perk with a `command` object can give orders:

```
command.ap      AP the ordered summon adds to its next turn's pool (the pool's cap of 10 holds)
command.reach   "one" (one summon per order) or "all" (every active summon)
command.turns   how long the grant lasts: 1, the summon's next turn
```

Each ordered summon gets one modifier, `add command.ap` on `ap`, with source
`command:<perk key>:<commander name>`, origin `core:origin/morale`, and `duration: turns`. Morale
stacks highest, so two orders on one summon in one turn give the larger, not the sum. The
commander's modifiers on `command` change the order's cost, as a perk changes any action's
([actions.md](actions.md#costs)). In real time the AP is `ap × stamina_per_ap` stamina, the same
rate as every action ([actions.md](actions.md#one-rule-set-two-clocks)). [A]

## The held command

A **held command** is a concentration spell whose effects include `core:effect/command_summons`.
While the commander holds it, every active summon gets `add magnitude` on `attack_roll`, source
`command:<spell key>:<commander name>`, origin `core:origin/morale`, `duration: while_applied`.
It ends with the concentration: a new concentration spell, Summon Elemental included, ends it
([casting.md](casting.md#what-is-core)), and an ended held command gives nothing. A commander who
holds a command leads summons whose spells are not held (familiars, the undead), or recasts.
[A, the one-held-spell rule kept]

## The ladder

| Perk | Level | Needs | Gives |
| --- | --- | --- | --- |
| Field Command | 6 | orange or black magic 40, CHA 5 | orders: one summon, +2 AP on its next turn |
| Sharp Orders | 10 | Field Command | an order costs 1 AP less (`add -1` on `command`) |
| Voice of the Host | 14 | Sharp Orders, orange or black magic 60 | orders: every summon, +2 AP |
| Iron Drill | 18 | Field Command, orange or black magic 65 | orders: one summon, +3 AP |

The held command is **Steady the Host** (orange, level 8, concentration): +5 to every summon's
attack rolls while held. A commander passes the perk they order with, as a bard passes theirs to
`share`. Numbers are [A] until a playtest runs a commander against a bonded summoner.

## Input and result

```
input   commander: { name, perk: <feature key with a command object>, modifiers,
                     concentration: <spell key or null> }
        summons:   the names of the commander's active summons
        order:     optional { targets: [summon names] }, an order given this turn with commander.perk
result  ap:        the order's cost after the commander's modifiers on `command`, 0 without an order
        modifiers: { <summon name>: the modifiers it gets }, every summon listed
```

An order with no `commander.perk`, a perk with no `command` object, a target that is not an
active summon, or more than one target with `reach: "one"` is an error. The kernel has no party: the
caller passes each summon its modifiers, as with `share` ([share.md](share.md#no-party-in-the-kernel)).
