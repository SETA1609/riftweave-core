# GD_S17_04_elemental-bolts-and-bullets Elemental bolts and bullets: a grenade's burst at the weapon's damage

Date: 2026-11-24

- **Decision:** an elemental round is an ammunition row with a `burst_radius` ([combat.md#elemental-rounds](../core/combat.md#elemental-rounds)), read by `combat`, so a module's rounds and GD_S17_11's Bound Ammunition reuse it.
  - **Inputs:** `combat` takes `round` (`core:ammunition/<key>`) and `burst` (`distance`, `crit`). A round's `damage_type` replaces the weapon's and its `effects` land on a hit that is not evaded; this makes the two fields GD_S17_02 reserved on the row work for every round. [V 2026-11-24]
  - **The burst** is one `combat` call per creature caught in it, with `burst`: an area roll of its own and the shot's damage roll, which keeps the weapon's attribute modifier and weapon effects. Past the radius it is `outside`, as a grenade's. On a miss the caller sends the target as `burst.distance` 0. [V 2026-11-24, vectors in Python and Go]
  - **Crit:** the shot's crit (`burst.crit`) doubles every burst creature's damage; a burst roll's own crit neither doubles it nor carries the creature through the bands. [A]
  - **Rows:** Fire, Frost, Thunder, and Acid Bolts and Bullets, 5 a bundle, the flasks' riders and radii (thunder 2). A bundle is 120 gold, thunder 180: about three times the flask per round. Bolts weigh 1, bullets 0.5. [A]
  - **Gate:** `validate.py`'s `burst_problems`: only bolts and bullets burst, and a bursting round's riders need the `grenade` channel. [V 2026-11-24]
  - **Phase:** carried by the element's damage type tags (`phase_fire`, `phase_water`, `phase_wood`); acid and thunder have none. No new field. [A]
- **Differences from the issue:**
  - **Damage perks:** combat applies no weapon damage perk today (`melee_damage` is a pending target), so the flask-against-bullet vector shows the bullet's PER modifier: the weapon's bonus adds to the bullet's roll and not to the flask's. Material and perk bonuses ride the damage roll the caller sends. [A]
  - **Availability:** loot tables cannot list an ammunition row (the `item` ref pattern and loot pricing) and merchants have no level gate, so "from merchant and loot level 6" is left for a follow-up; the spec says so. [U]
  - The spec's last paragraph changed in the data step, when the loot table turned out not to fit.
- **Not done here:** the blunderbuss cone (GD_S17_08); a web page for the burst.
