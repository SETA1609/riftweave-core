# Quick start

## The one rule

Almost everything is one roll. Add your skill, your bonuses, and your penalties to get a **target**, then roll d100.

- **Roll the target or lower and you succeed.** Your **margin** is the target minus the roll, so low rolls are better.
- **A crit** is a success rolled at or under your crit window: 5, plus half your LCK rounded down.
- **A 100 always fails**, and it is a fumble, even with a target above 100.

There are no classes. A character is eight attributes, a set of skills, and the perks you pick.

## Make a character

1. **Choose a race.** It adjusts your attributes and gives you traits, a speed, and a size ([races](../modules/race.md)). Common Dwarf and Common Elf are the plain choices.
2. **Assign attributes.** Start all eight at 4 and spend 10 points, with no attribute above 10. Then apply your race's adjustments, which may pass 10.

| Attribute | What it does |
| --- | --- |
| STR | Melee damage, carrying |
| PER | Aim, awareness, acting first |
| END | Hit points, stamina |
| INT | Skill points, mana, the power of magic |
| WIL | Regeneration, resisting magic |
| AGI | Action points, stealth, finesse |
| CHA | Persuasion, prices |
| LCK | Crits, loot, everything a little |

3. **Choose a background.** It gives skill bonuses, gear, and sometimes spells ([backgrounds](../modules/backgrounds.md)).
4. **Take creation perks if you want them.** They are optional and come with a tradeoff.
5. **Pick 3 tag skills.** They cost 1 point per rank, where other skills cost 2.
6. **Set your skills.** Every skill starts at `5 + its attribute × 2`, plus your background's bonus. Then spend a pool of `5 + INT × 2 + a luck roll` points, where the luck roll is any number you roll from 0 up to your LCK. Skills stop at 100.
7. **Work out your numbers.**

| Stat | Formula |
| --- | --- |
| Hit points (HP) | `15 + END × 8 + level × 4` |
| Mana (MP) | `INT × 8 + level × 2` |
| Stamina (SP) | `10 + END × 5 + level × 2` |
| Action points (AP) | `2 + AGI ÷ 3`, rounded down |
| Carry | `25 + STR × 10` |
| Initiative | `PER` |

8. **Take your gear and spells** from your background, and buy more if you have money ([equipment](../modules/equipment.md), [magic](../modules/magic.md)).

**Example.** Mara has STR 7, PER 5, END 6, INT 5, WIL 4, AGI 6, CHA 4, LCK 5 before her race. She has 67 HP, 42 MP, 42 SP, 4 AP, carries 95, and has initiative 5. Blades starts at 19. With a luck roll of 3 she has 18 points to spend, and because Blades is tagged, 10 of them raise it to 29.

## Take a turn

Highest initiative goes first. On your turn you spend AP on any mix of these:

| Action | AP |
| --- | --- |
| Attack | 3 |
| Cast a spell | 3, plus the spell's mana |
| Move one unit | 1 |
| Disengage | 1 |
| Dodge (ready to evade) | 3 |
| Use an item | 2 |
| Block (ready your shield) | 2, refunded if no attack comes |
| Aim an attack | +2 on that attack |
| Stand up | 2 |

Unspent AP carries over, up to 10. There is no bonus action and no reaction slot: if you want to dodge on someone else's turn, you pay for it now. Walking away from an enemy without disengaging lets it take one attack at you, if it has 3 AP left.

## Attack

1. **Roll to hit.** The target is your weapon skill plus any bonuses and penalties. Flanking is +10, half cover −10, and three-quarters cover −20.
2. **The defender may react** if it readied a dodge or a block, and rolls its own skill. It must match or beat your margin. An evade avoids a single-target attack and halves an area attack. A block subtracts the shield's DR.
3. **Roll damage.** Your weapon's dice plus your attribute modifier, which is the attribute minus 5. Melee uses STR, finesse weapons AGI, and each ranged weapon says which it uses.
4. **A crit doubles the damage.** Then subtract armor DR, then resistance.

**Example.** Mara swings a longsword at a bandit with Blades 29 and rolls 12: a hit with margin 17. She rolls 5 on the 1d8 and adds +2 for STR 7, for 7 damage. The bandit's armor stops 1, so it loses 6 HP.

Spells work the same way on the matching magic skill: spend the mana first, then roll. Some spells land with no chance to dodge, and those allow a save instead.

## After the fight

- **End of combat** refills stamina and AP.
- **A short rest** (an hour) restores a quarter of your HP and MP, and refills stamina and AP.
- **A long rest** (a night) restores everything and clears what time would clear ([rest](../core/rest.md)).

Experience comes from the GM. You reach a level at `450 × level × (level − 1)` total XP, so level 2 is at 900. Each level gives a new skill pool and one perk, and adds `4 + a roll` HP, `2 + a roll` MP, and `2 + a roll` SP, where each roll is from 0 up to half your LCK.

## Run an encounter

1. **Set the scene.** Note the ground, the cover, and who can see whom.
2. **Pick opponents** near the party's level from the [bestiary](../modules/monsters.md).
3. **Roll or read initiative**, highest first.
4. **Play turns** until one side falls, flees, or talks.
5. **Give the loot.** Each monster rolls on its own tables.
