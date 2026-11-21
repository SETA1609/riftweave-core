### GD_S19_04_ready-and-opportunity-attacks Ready an attack on reserved AP, and opportunity attacks in the pipeline
- Added: the `ready` action holds the loadout's attack cost for a trigger, `enters_reach` or `acts_in_reach` (`readyTrigger` in schema.json, `triggers` on an action row).
- Added: `attack_sequence` runs `opportunity` and `readied`, one attack of the loadout paid from held AP; Disengage, a trigger that did not happen, or too little held means no attack and nothing spent, and the result gives `fired`, `reason`, and `guard_lost`.
- Changed: Phalanx's `opportunity_attack_cost` is no longer pending; docs/modules/combat.md section 2 pays an opportunity attack from held AP at the loadout's cost, not 3 AP from the carried pool.
