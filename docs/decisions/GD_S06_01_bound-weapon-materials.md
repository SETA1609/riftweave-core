# GD_S06_01_bound-weapon-materials Bound weapons have a material, perks, and enchants

Date: 2026-11-19

- **Decision:** the cast of a bound weapon spell returns the conjured weapon (`bound_weapon` in the `casting` result): its base, iron by default (`casting/core.json` `bound_weapon`), no phase, and the new `magic` damage type, which has `armor_dr_applies` and is reduced by `resist` `magic`. A coating or an elemental enchantment (a `damage_<type>` effect with the matching channel) sets the damage type and the phase. Perks name a material in `unlocks` and add on-hit effects in `bound_weapon_effects`; the combat pipeline reads the on-hit effects as `attacker.weapon_effects`. [A, spec step]
- **How a perk names a material:** in `unlocks`, which already takes a `core:material` ref for "a bound weapon's material" [V 2026-11-19, `feature.schema.json`], not a modifier, since a modifier holds a number (the issue's open question). `validate.py` already resolves every `unlocks` ref. The on-hit effects are a new field, `bound_weapon_effects`. [A]
- **Where the weapon is resolved:** in `casting`, not a new command: the weapon exists because of the cast, and the caster's perks are already its input's to read. The coating and the enchantment are optional inputs of the same command. [A]
- **New effects:** `bane` (bonus damage against a creature type), `banish_summoned`, and `lingering_damage` (damage over time of a parameter type), each with the `enchantment` channel; `armor_pierce` stops being reserved. The combat pipeline gains `defender.creature_type` and the `summoned` flag to read them. [A]
- **Enchantment over coating:** when both are named, the enchantment's type and phase win; a coating wears off and an enchantment does not. [A]
- **Numbers:** the ladder's levels, skill gates, and magnitudes (steel at 6, armor pierce 25, bane 4, lingering 2 for 10 s, obsidian at 20) are [A] until a playtest compares a bound weapon with a mundane one.
- **Not here:** the bound ranged weapons, the one-handed rule without concentration, and Bound Ammunition are `GD_S17_11`'s.
