### GD_S05_01_loadout-attacks-melee-ladders Attacks by loadout, the jab, the full attack, and the melee perk ladders
- Added: the `attack_sequences` collection (schema and data): per loadout, the attack's AP and its strikes in order, and the off-hand jab; the loadout vocabulary (`mainHand`, `offHand`, `loadout`) in `schema.json`.
- Added: the `attack_sequence` command and kernel function (Python and Go): the attack, the jab, the full attack, and the circle cut, each strike its own `combat` run, with carries, free strikes, and Scissor Cut; `attack_sequence.json` vectors.
- Added: the `jab`, `full_attack`, and `circle_cut` actions; the `full_attack_strikes`, `full_attack_off_strikes`, `mainhand_penalty`, `offhand_penalty`, and `two_weapon_evasion_band` modifier targets and the `carry_through`, `reaping_arc`, `scissor_cut`, and `can_circle_cut` flags.
- Added: the perks Reaping Arc, Circle Cut, Follow-Up, Press On, Off-Hand Drill, Crossed Guard, Twin Rhythm, and Scissor Cut.
- Changed: `combat` takes an optional `loadout` (a maneuver's AP is the loadout's attack AP plus its own) and `defender.loadout` (Crossed Guard widens the band).
- Changed (breaking): `power_attack` is `full_weight` (open to Blades, Blunt, or Piercing 40), `cleave` is `carry_through` (needs Full Weight), the `cleave_on_kill` target is `carry_through`, and the reserved effect `cleave` is `spillover`.
