# DE_S02_04_descriptions-without-numbers Third normal form: a description does not restate its numbers

Date: 2026-11-20

- **Decision:** option (a), a gate, as the issue's [A] proposed: a feature description may restate its row's numbers, but `validate.py` (`check_description_numbers`, gate 9 in `docs/core/integrity.md`) fails one that says a number the row does not hold. Placeholders (option b) were not chosen: they would make the JSON and its diffs harder to read and need a fill step in every consumer. [A]
- **What counts as the row's number:** any number the row holds outside `label` and `description`; an `add` value times a rank up to `ranks`; a `multiply` change as a percent at a rank up to `ranks` (0.8 is "20%"); a tier or rank ordinal up to `ranks`; a dice string only when the row holds it. Signs are ignored. [A]
- **Applied to `features/core.json`:** 16 of 171 descriptions failed [V 2026-11-20]. All 16 now pass, each rewritten to name where the number lives:
  - `circle_cut`, `follow_up`, `press_on`, `twin_rhythm` restated the attack-sequence penalties (-10, -20, -30, -10 in `attack_sequences/core.json`); they now name the penalty.
  - `tempered_conjuring`, `glassfire_conjuring` restated the material's weapon attack (5, 6); they now name steel's and obsidian's.
  - `catman_claws` restated the claw row's 1d6; it now names the row's dice.
  - `black_elf_sunless_sight` said "extends to 24 units" while the row adds 12 to darkvision: a drifted total. It now says "12 units farther".
  - `boss_strike` said "DR + 1d4", but no table holds 1d4: the bash adds the caller's bash roll (`docs/core/combat.md#the-shield-bash`). It now says "the bash roll".
  - `guardian` said "+10" to an ally's evasion band; no table holds that 10 (only the `guardian` modifier target's description). It now says the band widens. [A]
  - `commanding_summons_magic`, `commanding_summons_physical`, `bardic_inspiration`, `skalds_war_chant`, `bardic_anthem`, `skalds_saga` said "never below 0"; they now say "never negative". The "CHA - 5" next to it passes only because the row's +5 Persuasion or Intimidation matches by coincidence.
- **The issue's count:** it found 79 of 124 descriptions that restate a modifier value. Most of those restate their own value, which this rule allows; the 16 above are the ones that say a number the row does not hold.
- **Next collections** (counted with the same rule, [V 2026-11-20]): `modifier_targets` 35 of 137, `abilities` 8 of 8, `effects` 7 of 89, `conditions` 2 of 25, `stances` 2 of 8, `checks` 2 of 44, `summon_bonuses` 2 of 2, `crafted_items` 1 of 6, `origins` 1 of 13. `spells` and `equipment` pass as they are. Monster descriptions live under `overrides` and are not counted yet. Add each to `DESCRIPTION_NUMBER_COLLECTIONS` in `validate.py` when its rows are fixed.
- **Open:** a number that matches some row value by coincidence passes (the "CHA - 5" case, Boss Strike's "2 AP", which is `shield_bash.stagger_ap`). The gate catches drift, not every restatement. [A] Guardian's +10 and the bash roll's size have no data home; if they are rules, they need a row of their own (a later issue). [U]
- **Risk:** none for consumers: no key changes, only prose.
