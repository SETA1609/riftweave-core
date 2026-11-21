# GD_S19_04_ready-and-opportunity-attacks Ready an attack on reserved AP, and opportunity attacks in the pipeline

Date: 2026-11-21

- **Decision:** an opportunity attack and a readied attack are one mechanism, an attack out of turn paid from held AP, run by `attack_sequence` as the actions `opportunity` and `readied`; the rule is [combat.md#attacks-out-of-turn](../core/combat.md#attacks-out-of-turn), the `ready` row is in [actions.md](../core/actions.md#costs).
- **Owner decisions kept** (2026-11-19): opportunity attacks are a base rule that perks extend; counters stay perks (Riposte); both attacks are paid from held AP at the loadout's attack cost, which replaces modules/combat.md's flat 3 AP from the carried pool.
- **Nothing spent when it does not fire:** a readied attack whose trigger never happens, an opportunity attack a Disengage prevents, and one with too little held all spend 0, so the AP stays held and keeps its guard and carry. The result says why (`reason`). [V 2026-11-21, `attack_sequence` vectors in Python and Go]
- **The guard drop is in the result** (`guard_lost`, the cost times `guard_per_ap`), and the `actions` command's `held_spent` already lowers the guard and the carry the same way (GD_S19_05).
- **Phalanx** lowers only an opportunity attack's cost (`opportunity_attack_cost`), never below 0; a readied attack costs the loadout's attack AP. `opportunity_attack_cost` is no longer pending; Flowing Step stays pending, since no kernel step tracks a turn's first move.
- **What the kernel does not track:** reach and who left it, one readied attack at a time, and the readied attack ending when the actor moves or is knocked down. The caller (a session or a table) says what happened (`event`, `disengaged`); the command decides whether it fires and what it costs. [A]
- **Open:** ranged overwatch waits for GD_S17_05; the triggers that perks may add (casting, drinking, or standing up in reach) are [A].
