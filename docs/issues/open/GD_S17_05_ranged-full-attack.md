#### GD_S17_05_ranged-full-attack The full attack and the extra-strike perks with ranged weapons
- Status: todo
- Priority: medium
- Needs: GD_S05_01_loadout-attacks-melee-ladders, GD_S17_02_ammunition-items-and-reloading
- Modified: 2026-11-18
- Do: Owner request, 2026-11-18: the extra-attack perks work at range as well as in melee. GD_S05_01 defines the attack by loadout, the full attack, and `attack_sequences` for melee loadouts; this issue adds the ranged loadouts and the rules a ranged strike adds.
  - **Ranged rows in `attack_sequences`** [A]:

    | Loadout | Attack AP | Strikes |
    | --- | --- | --- |
    | a bow (any) | 3 | main 0 |
    | a hand crossbow or one pistol, off hand empty or a shield | 2 | main 0 |
    | a two-handed crossbow or a long gun | 3 | main 0 |
    | two pistols or two hand crossbows | 3 | main -10, off -20 |
    | one light thrown weapon in each hand | 3 | main -10, off -20 |
    | a one-handed melee weapon and a pistol, a hand crossbow, or a light thrown weapon (the ranger's mixed hand) | 4 | main -10 (melee, in reach), off -20 (ranged, in range) |

  - **The mixed hand** (owner 2026-11-18, the ranger): the melee strike goes to a foe in reach and the ranged strike to any target in range, so one attack can cut one foe and shoot another. The ranged strike at a foe in reach takes the shooting-into-melee penalty, as any shot does. Reloading a `loading` weapon with the other hand full costs 1 AP more [A]; GD_S17_13's perks soften both.
  - **A strike needs a round:** each strike spends one round (GD_S17_02). A `loading` weapon fires once and then needs a `reload` (2 AP, or less with `fast_reload`) before it strikes again; inside a full attack, a reload spends AP from the full attack's own budget [A]. Two loaded pistols are the gunslinger's opening pair; a bow, which draws as part of the shot, gains the most from extra strikes.
  - **The extra-strike ladder is any weapon** (GD_S05_01 changes its text to say so): `follow_up` and `press_on` add a strike with the main-hand weapon, melee or ranged; a loading weapon's added strike needs a reload first.
  - **Carry on at range:** `carry_through` and `reaping_arc` read "a foe in reach" as "a foe in range" when the strike was ranged [A]; `circle_cut` stays melee.
  - **Aimed and full:** a full attack cannot be aimed; the held-breath maneuver (GD_S17_06) is the aimed shot that spends a turn [A].
- Keys: the ranged rows of `attack_sequences`; the reload-inside-a-full-attack rule; the range reading of `carry_through` and `reaping_arc`
- Files: ~8 (docs/core/combat.md, docs/core/actions.md, ruleset/data/attack_sequences/core.json, ruleset/data/features/core.json (descriptions of `follow_up`, `press_on`, `carry_through`, `reaping_arc`), ruleset/scripts/validate.py, ruleset/tests/vectors/combat.json, docs/decisions/GD_S17_05_ranged-full-attack.md, changes/GD_S17_05_ranged-full-attack.md)
- See: docs/issues/open/GD_S05_01_loadout-attacks-melee-ladders.md (`attack_sequences`, the full attack), docs/core/actions.md#pool
- Done when:
  - A vector runs a full attack with a shortbow and `follow_up`: two strikes, two arrows spent; with `press_on` as well, three.
  - A vector runs a full attack with two loaded pistols: two strikes, then a reload inside the budget, or the full attack ends when the AP runs out.
  - A vector shows `carry_through` sending the next arrow to a second foe in range.
  - A vector runs a mixed-hand attack (shortsword and hand crossbow): the cut at a foe in reach, the bolt at a second foe in range, and the one-handed reload at 1 AP more.
  - `validate.py` and the tests pass.
