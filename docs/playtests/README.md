# Playtests

## M5-18 agent runs (2026-11-12)

The alpha playtest asks for a person to play through the human CLI. These runs are the agent
pass that comes first: `ruleset/scripts/playtest_driver.py` plays eight builds from creation to
level 30 through the agent CLI's `session` command, so every step is a real call into the core
and every roll is in its step's input. The saved sessions replay to the same actor. **They are
agent-played, not human-played**; M5-18 stays open for the human pass
([decision](../decisions/M5-18.md)).

Rerun (from the repository root):

```
(cd backend && go build -o ../cache/riftweave ./cmd/riftweave)
python ruleset/scripts/playtest_driver.py --binary cache/riftweave --out docs/playtests
```

To check one saved session, run `cache/riftweave --dev session` on
`{"operation": "replay", "session": <the file's JSON>}`; `matches` is true for all eight.

The seed (`m5-18`) fixes every roll, so a rerun on the same data writes the same files.

### What a run does

Per level: one fight against a monster within two levels (the nearest if none), up to 12 rounds,
a long rest, then the level-up: `progression level` with the next level's XP, `level_up` with
rolled luck and pool rolls, the pool spent on the build's skills in order, the first eligible perk
from the build's list (then any eligible perk), and the spells of its colors it can cast. A caster
casts its strongest affordable damage spell and falls back to its weapon.

### Results

| Build | Won / fled / lost | Monster hit rate | Skill at the cap, by level | Level-ups with no eligible perk | Steps | Replays |
| --- | --- | --- | --- | --- | --- | --- |
| Black and orange mage (`caster_black_orange`) | 17 / 13 / 0 | 29% | black_magic 6, orange_magic 8, evasion 10 | 6 of 29 | 976 | yes |
| Blue and green mage (`caster_blue_green`) | 23 / 7 / 0 | 25% | blue_magic 5, green_magic 8, evasion 10 | 11 of 29 | 849 | yes |
| Red mage (`caster_red`) | 25 / 5 / 0 | 30% | red_magic 6, evasion 8, lore 10 | 13 of 29 | 720 | yes |
| White and violet mage (`caster_white_violet`) | 28 / 2 / 0 | 26% | white_magic 6, violet_magic 9, evasion 11 | 10 of 29 | 799 | yes |
| Yellow and indigo mage (`caster_yellow_indigo`) | 20 / 10 / 0 | 26% | yellow_magic 5, indigo_magic 8, evasion 10 | 12 of 29 | 856 | yes |
| Spellblade (`hybrid_blades_red`) | 10 / 20 / 0 | 27% | blades 7, red_magic 10, evasion 13, medium_armor 22 | 15 of 29 | 1128 | yes |
| Blades fighter (`melee_blades`) | 10 / 20 / 0 | 24% | blades 8, evasion 13, heavy_armor 17, athletics 29 | 16 of 29 | 1092 | yes |
| Archer (`ranged_bows`) | 13 / 17 / 0 | 32% | bows 8, evasion 13, stealth 17, light_armor 29 | 11 of 29 | 1082 | yes |

### Findings

| Finding | Filed as |
| --- | --- |
| A session cannot buy a skill rank, take a perk, or learn a spell; every run did all three through dev mode. | [APP-12](../issues/open/APP-12.md) |
| A session never applies the actor's perks, held modifiers, race, or gear to a command, so no perk changed any result in these runs. | [APP-13](../issues/open/APP-13.md) |
| Monsters attack and evade with their starting skill at every level and won no fight. | [M3-17](../issues/open/M3-17.md) |
| No command applies a cast's damage to a target; the driver did it by hand. Orange, green, and violet have no damaging spell. | [M3-18](../issues/open/M3-18.md) |
| Main skills reach 100 by level 5 to 8, and weapon damage does not grow while monster HP does. | [M5-50](../issues/open/M5-50.md) |
| 38 of 88 perks were never taken, and 94 of 232 level-ups had no eligible perk for the build. | [M5-19](../issues/open/M5-19.md), through the perk table in the decision record |
| The XP for a level comes from the driver, not from the fights: XP awards are GM guidance ([progression.md](../core/progression.md#not-in-the-data)). | decision, no change |

No run produced a result that contradicts a vector, so no new vector came from these runs.
