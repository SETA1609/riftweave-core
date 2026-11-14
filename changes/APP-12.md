### APP-12 Sessions record skill purchases, perks, and learned spells
- Added: `progression take_perk` checks a perk's rank, level, attributes, skills, and prerequisite perks and returns `taken`, `rank`, and `unmet`, in Python and Go.
- Added: session adapters: `progression spend` with a skill `key` adds the bought ranks to that skill, `take_perk` sets the perk's rank, and `spellcraft learn` adds the spell and its effects, so a level-up needs no dev step.
- Changed: the playtest driver plays every build with no dev step.
