# PO_S19_01_file-turn-economy-issues File the turn-based action economy issues

Date: 2026-11-19

- **Decision:** a new story, S19 Turn-based action economy, with nine GD issues taken from the owner's comparison of our AP system with Fallout 1 and 2, Fallout Tactics, Divinity: Original Sin 1 and 2, and the Divinity board game (2026-11-19):
  - movement distance per AP (`GD_S19_02`), swap and draw costs (`GD_S19_03`), and a readied attack with opportunity attacks (`GD_S19_04`);
  - a carry cap with guard (`GD_S19_05`) and a partial pool at the start of a fight (`GD_S19_06`);
  - AP perks, traits, and conditions (`GD_S19_07`) and spell AP by level (`GD_S19_08`);
  - one initiative tie rule (`GD_S19_09`) and side initiative for the table (`GD_S19_10`).
- **Why a story:** the mechanics are sound (one cost per action, a stamina conversion, a refundable block, AP restores that arrive next turn), but the economy is coarse. With a pool of 3 to 5 [V 2026-11-19, `docs/core/actions.md#pool`] and attacks of 2 to 5 AP [V 2026-11-19, `attack_sequences/core.json`], most turns are one attack or one spell plus perhaps one move. At pool 3, two loadouts cannot attack without banking AP first. The nine issues together give a turn several priced choices.
- **Priority:**
  - movement distance is high: the battle grid (`BE_S10_05`) and the action menu need it, and today a move covers no set distance while a race's `speed` is per round [V 2026-11-19];
  - swap and draw, and ready with opportunity attacks, are medium-high: since GD_S05_01 a loadout change matters, and opportunity attacks are written in docs/modules/combat.md §2 but not built;
  - the side initiative module is low.
- **New `Needs` edges:**
  - `BE_S10_05_battle-grid-steps` needs `GD_S19_02` (the unit of a move);
  - `BE_S08_07_encounter-tracker-cli` needs `GD_S19_09`: it breaks ties by AGI, LCK, then the GM, while docs/modules/combat.md §7 says AGI, then a d100 roll-off [V 2026-11-19].
- **Roles:** each issue is GD (spec, data, vectors, Python reference). Go runs every vector file, so an issue's Go side lands in the same pull request, as GD_S05_01 did, unless it needs a new contract command.
- **Numbers are proposals** [A], to be settled in each issue's spec step and a playtest: a move covers `floor(speed / 2)` squares; carry pool + 2 and +5 evasion per further unspent AP; a fight starts on pool − 1; an Action Boy-like perk +1 AP at level 12 and up; a cast costs 2, 3, or 4 AP by spell level.
- **GitHub:** a story issue `S19: Turn-based action economy` with the ten issues as sub-issues, as for S17 and S18. The repo stays the source of truth.
- **Open:**
  - Does unspent AP become guard (`GD_S19_05`) or keep carrying up to the old cap of 10? The two pull against each other, and the owner chooses in the spec step.
  - Does an opportunity attack cost AP (`GD_S19_04`, with Phalanx's pending `opportunity_attack_cost`)?
