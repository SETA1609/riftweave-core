# GD_S19_04_ready-and-opportunity-attacks Opportunity attacks as reactions in the pipeline

Date: 2026-11-23

- **Decision:** an opportunity attack is one melee attack of the loadout, made when a foe leaves the actor's reach without disengaging, run by `attack_sequence` as the action `opportunity`. It is a reaction: no AP, no stamina, no guard. The rule is [combat.md#opportunity-attacks](../core/combat.md#opportunity-attacks).
- **Owner decisions kept** (2026-11-19): opportunity attacks are a base rule that perks extend; counters stay perks (Riposte).
- **Reactions** (owner, 2026-11-23): everyone has `reactions_per_round` (1, in `actions/core.json`'s pool) between two of their turns; the caller passes `reactions_used`, and one past the most does not fire (`no_reaction_left`). Watchful Blade grants `reactions_from_agi`, which makes the most max(1, AGI − 5); Phalanx then adds 1. A reaction-built character is a real choice. This replaces paying opportunity attacks from held AP, which this issue first built (2026-11-21).
- **Phalanx** adds a reaction a round (`add reactions_per_round 1`, owner 2026-11-23). It lowered `opportunity_attack_cost` by 1 until opportunity attacks became free; that target is removed.
- **Overwatch is separate** (owner, 2026-11-23): overwatch is the ranged attacker's, held AP fired when an enemy moves within its weapon's range, and spends no reaction. It is GD_S17_06's Trigger Watch, which already specified it; no new issue was filed. The melee `ready` action and its `readyTrigger` vocabulary that this issue first built are removed.
- **Melee only** (owner, 2026-11-23): a ranged loadout makes no opportunity attacks. No ranged loadout exists yet, so GD_S17_05 enforces it when it adds them; GD_S17_13's `ready_steel` stays the perk exception.
- **Nothing spent when it does not fire:** Disengage or no reaction left spends nothing. The result says why (`reason`). [V 2026-11-23, `attack_sequence` vectors in Python and Go]
- **What the kernel does not track:** reach and who left it, and the reactions used. The caller (a session or a table) says what happened (`disengaged`, `reactions_used`). [A]
- **Real time** (owner, 2026-11-23): reactions are the table rule; a real-time engine fires opportunity attacks automatically or leaves them out.
- **Assumed** [A]: Riposte does not count against the reactions (it is a counter of the dodge stance, run by the caller as its own attack); Watchful Blade's tier (level 6, AGI 7, ungated) is a first guess.
- **Open:** the triggers that perks may add (casting, drinking, or standing up in reach) are [A].
