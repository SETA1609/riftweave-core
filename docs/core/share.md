# Share

Modified: 2026-11-15

A bard or a skald lends their Charisma to allies (M3-14). The `share` function works out what
one ally gets from the sharers around it. The perks are rows in `ruleset/data/features/core.json`,
the groups they lend to are the rows of `ruleset/data/summons/bonuses.json` that M3-13 uses for
summons, and the vectors are in `ruleset/tests/vectors/share.json`.

## Who shares

A perk with a `share` object lends its holder's attribute modifier to allies:

```
share.bonus     core:summon_bonus/<key>: the group of attributes it lends to, and the
                attribute it reads (cha), with a share per rank (rank_share)
share.range     units: an ally farther than this gets nothing
share.turns     how many turns the bonus lasts after the sharer starts it
share.rank_share  optional: the share per rank for this perk, in place of the group's
```

Starting a share is the `inspire` action (`actions/core.json`), so it costs AP like any action,
and a perk that lowers its cost is a modifier on `inspire`. The sharer does not gain the bonus;
they lend it.

The same groups as summons, so a bard and a summoner reading one CHA mean the same thing:

| Group | Attributes | Perk |
| --- | --- | --- |
| `magic` | INT, WIL | Bardic Inspiration, then Bardic Anthem |
| `physical` | STR, END, AGI | Skald's War Chant, then Skald's Saga |

## Two tiers

Like summoning, where Commanding Summons has two ranks and Retinue and Host widen it later
(M3-19, M3-20), sharing has two tiers. Each class needs its own social skill on top of its
combat one: Persuasion for the bard, Intimidation for the skald.

| Tier | Level | Perk | Needs | Lends | Range | Turns |
| --- | --- | --- | --- | --- | --- | --- |
| 1 | 18 | Bardic Inspiration, Skald's War Chant | Leader, CHA 7, a magic (bard) or weapon (skald) skill 45, Persuasion or Intimidation 30 | half of CHA − 5 at rank 1, all at rank 2 | 6 | 3 |
| 2 | 28 | Bardic Anthem, Skald's Saga | the tier-1 perk at rank 2, CHA 8, Persuasion or Intimidation 60 | all of CHA − 5 (`rank_share` [1]) | 12 | 5 |

A sharer who holds both tiers starts the one they choose; a bard who holds the Anthem passes it
as `perk`. Holding both does not add: the two are one morale bonus. A prerequisite that names a
perk may name its rank (`{ "perk": ..., "rank": 2 }`, [progression.md](../modules/progression.md#perks-and-traits)).

## The value

```
modifier = attribute_modifier(sharer.cha)        # CHA - 5 (formulas/attribute_modifier.json)
value    = floor(max(modifier, 0) * rank_share[rank - 1])   # the perk's rank_share, else the group's
```

At rank 1 a sharer lends half, at rank 2 all of it, exactly as a summon bonus (M3-19). CHA 8
lends +1 at rank 1 and +3 at rank 2. A CHA below 6 lends nothing, and never a penalty.

## What reaches an ally

For each sharer within `range` units, the function emits one modifier per attribute of the
group:

```
target    the attribute (int, wil, str, end, agi)
operation add
value     the value above
origin    core:origin/morale
source    share:<perk key>:<the sharer's name>
duration  turns, with turns = share.turns
```

`morale` stacks `highest` (`origins/core.json`), so two bards add nothing together: on each
attribute the strongest bonus counts, and a weaker one is dropped. After `share.turns` turns
the modifier has expired ([modifier.md](modifier.md#order-for-one-numeric-target)) and the ally gets nothing until a
sharer starts it again. The result lists the live modifiers and, per attribute, the bonus that
counts.

## No party in the kernel

The kernel has no party. A shared bonus is emitted by the sharer and applied to each ally by
whoever knows who is near: a session (APP-04) or the client calls `share` once per ally with the
sharers in range and passes the modifiers into that ally's commands. Range is a distance the
caller measures in units ([contract.md](contract.md#distance)). [A]

## Input and result

```
input   sharers: [{ name, attributes: { cha }, perk: <feature key>, rank, distance }]
        elapsed_turns: turns since the sharers started (default 0)
result  modifiers: the live modifiers on the ally
        bonus: { <attribute>: the value that counts }
```

A perk with no `share` object, or a rank the perk does not have, is an error. [V 2026-11-15,
`share.json` in Python and Go]
