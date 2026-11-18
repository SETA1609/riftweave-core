# S05 Melee combat

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S05 Melee combat

**Goal:** Attacks follow the loadout, and melee, stances, defence, and maneuvers have their perk ladders.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `GD_S05_01_loadout-attacks-melee-ladders` | GD | Attacks by loadout, the jab, the full attack, and the melee perk ladders | M3-28 | M3-27, M5-67 |
| `GD_S05_02_fighting-stances` | GD | Fighting stances: toggles a perk unlocks | M3-29 | M3-27, GD_S05_01_loadout-attacks-melee-ladders |
| `GD_S05_03_defense-perk-ladders` | GD | Defense perk ladders: shields, shield bash, evasion, no armor, and armor on casting | M3-30 | M3-27, GD_S05_02_fighting-stances, M5-47 |
| `GD_S05_04_trip-and-disarm` | GD | Trip and disarm: maneuvers a perk unlocks, and rider perks that put one on every attack | M3-31 | M3-24, GD_S05_01_loadout-attacks-melee-ladders |

**Done when:** every issue above is done, and the combat vectors cover the jab, the full attack, stances, shields, evasion, trip, and disarm.
