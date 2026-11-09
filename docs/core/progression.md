# Progression

The numbers that turn experience into levels, skill points, and perks. They live in `ruleset/data/progression/core.json`; this file reads the tree. [`docs/modules/progression.md`](../modules/progression.md) keeps the commentary (why XP is event-driven, what breakthroughs add) and cites the data for every number. [A]

## What the data holds

| key | value | meaning |
| --- | --- | --- |
| `start_level` | 1 | a new character |
| `max_level` | 30 | level cap |
| `skill_cap` | 100 | highest skill rating the base ruleset allows |
| `xp_to_reach` | `450 * level * (level - 1)` | total XP needed for `level` |
| `skill_points` | `5 + INT * 2 + luck_roll` | points in one level's pool |
| `luck_roll` | `0` to `LCK`, inclusive | the random part of the pool, rolled by the caller |
| `pool_roll_max` | `floor(LCK / 2)` | the highest `hp_roll`, `mp_roll`, or `sp_roll` |
| `perks_per_level` | 1 | perks chosen on each level gained |
| `tag_skills` | 3 | skills tagged at creation |
| `skill_cost` | normal 2, tag 1 | points for +1 rating |

`450` is the normal-difficulty constant. A campaign tunes it by overriding the row, not by changing the formula. [A]

## Rules

1. **Level from XP.** A character's level is the highest `N`, up to `max_level`, with `xp_to_reach(N) <= xp`. XP beyond the cap changes nothing.
2. **Level-up.** Each level gained, `start_level + 1` to `max_level`, grants one skill-point pool and `perks_per_level` perks. A new character also gets a pool at `start_level`; with the cap at 30 that is 30 pools in a career. [A]
3. **The pool.** `skill_points` is a formula tree ([formula.md](formula.md)). The one random term is the `luck_roll`, an input, so a vector pins it. The reference does not roll.
4. **Spending.** One rating point costs `skill_cost.normal` points, or `skill_cost.tag` if the skill is tagged. A skill never goes past `skill_cap`: points that would buy rating above the cap are not spent. Points that cannot buy a whole rating point (one left over on a normal skill) stay in the pool. [A]
5. **Tag ratio.** The ratio is 2:1 in favour of tag skills, as in `docs/modules/progression.md`. `AGENTS.md` words it as "+2 per point instead of +1", which is the same ratio on a different scale; the data uses cost, so a point is a point. [V 2026-11-02, read both]

## Resource growth

Each level-up after the first adds a luck roll to each resource pool, on top of its row's `per_level_flat`: HP `4 + hp_roll`, mana `2 + mp_roll`, stamina `2 + sp_roll`. Each roll is an input from 0 to `pool_roll_max`, one per pool, so a vector pins it. A roll out of range is refused. `level_up` returns each pool's `gain`. The caller adds the rolls to the actor's stored `growth`, which the derived function reads ([derived.md](derived.md)).

## Not in the data

XP awards are guidance to designers, not a rule.
