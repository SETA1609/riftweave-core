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
| `skill_start` | `5 + attribute * 2` | a skill's value before any point is spent on it |
| `skill_group_cost` | `members_per_step` 2 | a group point's price on a grouped sheet (below) |
| `monster_skill_growth` | `per_level` 17, `categories` combat, magic | what a monster's unwritten fighting skills gain per level ([actor-sheet.md](actor-sheet.md)) |

`450` is the normal-difficulty constant. A campaign tunes it by overriding the row, not by changing the formula. [A]

## Rules

1. **Level from XP.** A character's level is the highest `N`, up to `max_level`, with `xp_to_reach(N) <= xp`. XP beyond the cap changes nothing.
2. **Level-up.** Each level gained, `start_level + 1` to `max_level`, grants one skill-point pool and `perks_per_level` perks. A new character also gets a pool at `start_level`; with the cap at 30 that is 30 pools in a career. [A]
3. **The pool.** `skill_points` is a formula tree ([formula.md](formula.md)). The one random term is the `luck_roll`, an input, so a vector pins it. The reference does not roll.
4. **Spending.** One rating point costs `skill_cost.normal` points, or `skill_cost.tag` if the skill is tagged. A skill never goes past `skill_cap`: points that would buy rating above the cap are not spent. Points that cannot buy a whole rating point (one left over on a normal skill) stay in the pool. [A]
5. **Tag ratio.** The ratio is 2:1 in favour of tag skills, as in `docs/modules/progression.md`. `AGENTS.md` words it as "+2 per point instead of +1", which is the same ratio on a different scale; the data uses cost, so a point is a point. [V 2026-11-02, read both]

## Two sheets, one rule set

A sheet is how a character's skills are written down. Both sheets below read the same
`skill_start`, the same costs, and the same cap; neither is a second rule set. (M5-49)

### Track only what you raise

A sheet writes only the skills the character has raised. Any other skill reads its starting
value, `skill_start` of its `associatedAbility`, as a monster's unwritten skill already does
([actor-sheet.md](actor-sheet.md)). A written skill reads what is written. So a newcomer's sheet
holds their tag skills and the few they spend on, and a game that tracks all 44 skills gets the
same numbers. [V 2026-11-12, `progression` vectors, op `read_skill`]

### The grouped sheet (variant)

A table may play with fewer, broader skills. The `skill_groups` collection
(`ruleset/data/skill_groups/core.json`) puts every skill in exactly one group. A tracked group
(Melee, Ranged, Defense, one group per magic color, Stealth, Social, Knowledge, Wilds, Craft,
Athletics, Medicine) holds the ranks the character bought for it. An untracked group (Armor:
the four armor skills) holds none.

1. **Reading.** On a grouped sheet a member skill reads its own `skill_start` plus its group's
   ranks. A skill in an untracked group reads its `skill_start` alone. Each member keeps its own
   attribute, so a Melee rank raises blades (STR) and piercing (AGI) by the same point. [V 2026-11-12, `progression` vectors, op `read_skill`]
2. **Spending.** One rank of a group costs the skill cost, `normal` or `tag`, times
   `ceil(members / members_per_step)`. With `members_per_step` 2, a one- or two-skill group costs what one
   skill does, Melee (4) costs twice that, and Craft (7) four times. A group point buys
   several skills at once and none of them alone, so the price stays below the members' sum. [A]
   Tagging works on groups: a grouped sheet tags 3 groups.
3. **The cap.** A group rank is not bought when it would take its highest member past
   `skill_cap`; the caller passes that member's value as `skill`. [V 2026-11-12, `progression` vectors, op `spend` with `group`]
4. **An untracked group cannot be bought.** Spending on one is refused. [V 2026-11-12]

A character moves from a grouped sheet to a full one by writing each tracked member's current
value as its own skill. [A]

## Resource growth

Each level-up after the first adds a luck roll to each resource pool, on top of its row's `per_level_flat`: HP `4 + hp_roll`, mana `2 + mp_roll`, stamina `2 + sp_roll`. Each roll is an input from 0 to `pool_roll_max`, one per pool, so a vector pins it. A roll out of range is refused. `level_up` returns each pool's `gain`. The caller adds the rolls to the actor's stored `growth`, which the derived function reads ([derived.md](derived.md)).

## Not in the data

XP awards are guidance to designers, not a rule.
