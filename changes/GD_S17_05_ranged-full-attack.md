### GD_S17_05_ranged-full-attack The full attack and the extra-strike perks with ranged weapons
- Added: the ranged hand kinds `bow`, `hand_gun`, `long_gun`, and `light_thrown`, and six ranged rows in `attack_sequences/core.json`, from a bow to the mixed hand of a melee weapon and a hand crossbow.
- Added: `attack_sequence`'s `ammunition` per hand and a foe's `in_reach`: a ranged strike spends a round and may go to a foe in range, a loading hand reloads inside a full attack's budget (1 AP more with the other hand full), and the result's `ammunition`, `reloads`, and `stopped`.
- Changed: Follow-Up, Press On, Carry Through, and Reaping Arc work with ranged strikes; a ranged main hand makes no opportunity attack (`reason: ranged`); a melee strike may not aim at a foe out of reach.
