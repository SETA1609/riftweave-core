# S17 Ranged combat

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S17 Ranged combat

**Goal:** Ranged weapons spend ammunition, the bow, the crossbow, the gun, and thrown weapons each have their own perk ladder, and the ranged builds take them to level 30.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `PO_S17_01_file-ranged-combat-issues` | PO | File the ranged combat issues | — | — |
| `GD_S17_02_ammunition-items-and-reloading` | GD | Ammunition items, the reload action, and recovering spent ammunition | — | — |
| `GD_S17_03_thrown-grenades` | GD | Grenades: thrown flasks with a fixed area damage | — | — |
| `GD_S17_04_elemental-bolts-and-bullets` | GD | Elemental bolts and bullets: a grenade's burst at the weapon's damage | — | GD_S17_02_ammunition-items-and-reloading, GD_S17_03_thrown-grenades |
| `GD_S17_05_ranged-full-attack` | GD | The full attack and the extra-strike perks with ranged weapons | — | GD_S05_01_loadout-attacks-melee-ladders, GD_S17_02_ammunition-items-and-reloading |
| `GD_S17_06_ranged-maneuvers` | GD | Ranged maneuvers: the split nock, the held breath, pinning fire, and the trigger watch | — | GD_S05_03_defense-perk-ladders, GD_S17_02_ammunition-items-and-reloading |
| `GD_S17_07_bow-perk-ladders` | GD | Bow perk ladders: the longbow and the shortbow | — | GD_S17_05_ranged-full-attack, GD_S17_06_ranged-maneuvers |
| `GD_S17_08_firearm-perk-ladders` | GD | Firearm perk ladders: pistols, long guns, and the blunderbuss | — | GD_S17_04_elemental-bolts-and-bullets, GD_S17_07_bow-perk-ladders |
| `GD_S17_09_crossbow-thrown-perk-ladders` | GD | Crossbow and thrown-weapon perk ladders | — | GD_S17_08_firearm-perk-ladders |
| `GD_S17_10_ammo-scavenging-perks` | GD | Ammunition scavenging perks: more ammunition from the field and from loot | — | GD_S17_09_crossbow-thrown-perk-ladders |
| `GD_S17_11_bound-ranged-weapons-ammo` | GD | Bound crossbow, pistol, and rifle, and the Bound Ammunition spell | — | GD_S06_01_bound-weapon-materials, GD_S17_04_elemental-bolts-and-bullets, GD_S17_10_ammo-scavenging-perks |
| `GD_S17_12_ranged-builds-new-ladders` | GD | The ranged builds take the ranged ladders, and the skirmisher and the thrower | — | M5-61, GD_S17_13_ranger-switch-mixed-ladders |
| `GD_S17_13_ranger-switch-mixed-ladders` | GD | Ranger ladders: switching between bow and blade, and the mixed hand | — | GD_S17_11_bound-ranged-weapons-ammo |

**GitHub:** story [#274](https://github.com/SETA1609/riftweave-core/issues/274); tasks [`PO_S17_01_file-ranged-combat-issues` #358](https://github.com/SETA1609/riftweave-core/issues/358), [`GD_S17_02_ammunition-items-and-reloading` #359](https://github.com/SETA1609/riftweave-core/issues/359), [`GD_S17_03_thrown-grenades` #360](https://github.com/SETA1609/riftweave-core/issues/360), [`GD_S17_04_elemental-bolts-and-bullets` #361](https://github.com/SETA1609/riftweave-core/issues/361), [`GD_S17_05_ranged-full-attack` #362](https://github.com/SETA1609/riftweave-core/issues/362), [`GD_S17_06_ranged-maneuvers` #363](https://github.com/SETA1609/riftweave-core/issues/363), [`GD_S17_07_bow-perk-ladders` #364](https://github.com/SETA1609/riftweave-core/issues/364), [`GD_S17_08_firearm-perk-ladders` #365](https://github.com/SETA1609/riftweave-core/issues/365), [`GD_S17_09_crossbow-thrown-perk-ladders` #366](https://github.com/SETA1609/riftweave-core/issues/366), [`GD_S17_10_ammo-scavenging-perks` #367](https://github.com/SETA1609/riftweave-core/issues/367), [`GD_S17_11_bound-ranged-weapons-ammo` #368](https://github.com/SETA1609/riftweave-core/issues/368), [`GD_S17_12_ranged-builds-new-ladders` #369](https://github.com/SETA1609/riftweave-core/issues/369), [`GD_S17_13_ranger-switch-mixed-ladders` #370](https://github.com/SETA1609/riftweave-core/issues/370).

**Done when:** every issue above is done, the combat vectors cover a shot that spends ammunition, a reload, a grenade, an elemental bolt's burst, a ranged full attack, a mixed-hand attack, a weapon swap, and each ranged maneuver, and every ranged build passes the legality test from creation to level 30 without a filler perk from the magic table.
