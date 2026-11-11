#### POST-09 Weapon advantage module: the guard cycle
- Status: todo
- Priority: low
- Needs: M3-03, M5-09, M5-20
- Do: A small module, the first example a module author reads: a rock-paper-scissors between the three melee weapon skills, read from the skill every weapon base already names, so the module adds no table and no field on core rows. The cycle: blunt beats blades (weight batters through a parry), blades beat piercing (an edge turns a point aside and cuts a haft), piercing beats blunt (a thrust finds the gap in a slow, heavy swing). One situation row each way: the attacker whose weapon's skill beats the defender's wielded weapon's skill gets a bonus to hit of the order of the size bonus (+10, [A]), and the same penalty the other way. Unarmed, bows, crossbows, guns, throwing weapons, and natural attacks stand outside it. Decide whether swapping weapons costs AP enough that carrying the counter is a choice, not a reflex. It runs alongside Wuxing or alone; a table may load either, both, or neither. The name is our own: no other game's term for it.
- Keys: `ruleset/modules/weapon_advantage/` with its manifest and situation rows keyed by `core:skill/blades`, `core:skill/blunt`, `core:skill/piercing`
- Files: ~5 (ruleset/modules/weapon_advantage/, docs/modules/weapon-advantage.md, ruleset/tests/vectors/)
- See: docs/modules/combat.md (situations), docs/modules/weapons.md, docs/modules/module-system.md
- Done when: the rows name only core skills; vectors cover advantage, disadvantage, the same skill, and an attack outside the cycle (a bow, a fist, a claw); with the module off, combat vectors are unchanged; `module-system.md` links it as the minimal example.
