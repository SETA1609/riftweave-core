# GD_S19_08_spell-ap-by-level A minor spell costs less AP than a major one

Date: 2026-11-21

- **Decision:** a cast's AP comes from the spell's level: 1 to 5 cost 2, 6 to 20 cost 3, 21 and up cost 4, after the caster's modifiers on `cast`, never below 1. The bands are `cast_ap` in `ruleset/data/casting/core.json`; the `casting` command returns `ap` and `stamina`. The rule lives in [actions.md](../core/actions.md#cast-ap-by-level). [V 2026-11-21, `casting` vectors in Python and Go]
- **Why:** every cast was 3 AP, so at a pool of 3 to 5 a minor buff took the whole turn and a caster at a table cast only damage.
- **Choice:** of the issue's two options, the level bands; the Quick Casting perk (1 AP off up to a mana cost) was not added, and the existing Quick Caster (`add cast -1`) stays. The level already sizes a spell, and a mana threshold would be a second scale. [A]
- **Open:** the band edges (5 and 20) and the AP values wait for a playtest at pool 3 and 4. [A] With Quick Caster, every level 1 to 5 spell costs 1 AP; at pool 5 that is a cast and a full attack. [A]
- **Risk:** spells whose level was set for the dispel rule, not for size, may land in a band that does not fit their weight (Cleanse at level 10 costs 3). [A]
