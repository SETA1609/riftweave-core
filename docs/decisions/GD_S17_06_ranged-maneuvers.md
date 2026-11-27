# GD_S17_06_ranged-maneuvers Ranged maneuvers: the split nock, the held breath, pinning fire, and the trigger watch

Date: 2026-11-24

- **Decision:** five ranged maneuvers and four unlock perks ([actions.md#ranged-maneuvers](../core/actions.md#ranged-maneuvers)), and the unlock rule enforced for the first time.
  - **The unlock check:** `combat`, `attack_sequence`, and the new `overwatch` refuse a maneuver no given perk unlocks (`<key> needs a perk that unlocks it`). The check runs only when the input names the attacker's `perks` or a monster's `maneuvers`; an input that names neither is not checked, so callers that do not send perks keep working. [A] One combat vector that held a rider perk and made Heavy Swing now also holds Wind-Up. [V 2026-11-24]
  - **Which command:** a maneuver row's `command` (`combat` when left out) says where it runs; another command refuses it (`split_nock is an attack_sequence maneuver`). Split Nock makes two strikes, so it is an `attack_sequence` maneuver; Pinning Fire and Trigger Watch wait on what foes do, so they are `overwatch`'s. [A]
  - **Tiers:** `unlocks_at_rank` on a feature opens an unlock at a later rank: Heads Down (ranks 2, `rank_levels` 8 and 14) unlocks Pinning Shot at 1 and Pinning Fire at 2. `combat` and `attack_sequence` read perk refs at rank 1; `overwatch` takes held perks with ranks. [V 2026-11-24]
  - **Held Breath:** modifies `aimed`, so it costs the attack's AP plus aimed's plus 1 (6 with a bow). It needs `attacker.range_normal` 16 or more and no move this turn (`moved`), and its `crit_window` +5 widens the attacker's window for that attack (a roll of 8 crits at LCK 0). [V 2026-11-24]
  - **Pinning Shot:** the new `pinned` effect (maneuver channel, magnitude 2 AP, 6 s) on a hit that is not evaded; the caller applies it. [A]
  - **Pinning Fire:** spends all the AP left (at least the attack + 1) and its 3 rounds when laid; at most 3 strikes, the first move of each foe in the zone; an attack from the zone comes back as `zone_attacks` at -20. Two hand guns lay it though they load; any other `loading` weapon is refused. [A]
  - **Trigger Watch:** reserves the attack's AP from held AP; the first event in range draws the shot (one round), otherwise the AP comes back (`ap_returned`) and no round is spent. [V 2026-11-24]
  - **overwatch** resolves no dice: it returns the strikes with their attack add for the caller's `combat` calls. [A]
- **Differences from the issue:**
  - Pinning Fire's "a repeating crossbow" is covered by the property rule (it has no `loading`); "two loaded weapons" is read as two hand guns.
  - Split Nock at two targets "within 2 units" is the caller's check; the command takes `targets`. [A]
  - Held Breath with `steady_aim` (5 AP) is not modelled in `combat`, which reads no AP modifiers. [U]
- **Not done here:** no web page for overwatch yet (the command and its MCP tool come from the contract).
