#### POST-10 Table module: an optional defence roll
- Status: todo
- Priority: high
- Needs: M5-09
- Do: The defender's roll is the slowest part of combat at a table: an attack roll, then an evade or block roll, compared by margin. The table module makes it optional: by default the defender does not roll, and their evasion lowers the attacker's target (the passive defence; the conversion is decided here [A]). Rolling to evade or block becomes a choice that costs AP, as the `dodge` action (3 AP) already does. This is the first issue of the table module, so it adds its manifest. With the module off, combat vectors are unchanged.
- Keys: `ruleset/modules/table/` with its manifest and situation rows
- Files: ~6 (ruleset/modules/table/, docs/modules/table.md, ruleset/tests/vectors/, docs/guide/gm.md)
- See: docs/core/combat.md, docs/core/actions.md, docs/modules/module-system.md
- Done when: with the module on, an attack against a passive defender is one roll, with vectors; an active evade or block still works and costs AP; with the module off, every core vector is unchanged.
