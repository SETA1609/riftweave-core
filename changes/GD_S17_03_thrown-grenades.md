### GD_S17_03_thrown-grenades Grenades: thrown flasks with a fixed area damage
- Added: the `grenade` item type and five grenades in `equipment/consumables.json`: the Fire, Frost, Thunder, and Acid Flasks and the Smoke Pot, with a consumable's `damage`, `radius`, and throwing ranges for a thrown item.
- Added: `combat`'s `grenade` and `burst_distance`: an area attack with the grenade's own dice and type, the `outside` outcome past its radius, and its riders in `effects` on a hit.
- Added: the `grenade` effect channel, gated by `validate.py`, on `burn`, `slow`, `stagger`, and `sunder_armor`, and the new `smoke_screen` effect. Changed: `sunder_armor` is no longer reserved.
