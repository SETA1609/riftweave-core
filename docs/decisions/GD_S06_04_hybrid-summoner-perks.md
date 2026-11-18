# GD_S06_04_hybrid-summoner-perks Hybrid summoner perks: a spellblade or arcane archer fights beside their summons

Date: 2026-11-19

- **Decision:** the hybrid summoner's perks are attribute bonuses with the `while_summoning` duration of GD_S06_02, in two paths: melee (Bonded Vanguard, then Vanguard Ascendant: STR and END) and ranged (Bonded Marksman, then Marksman Ascendant: AGI and PER), plus Kindred Mind (INT and WIL) for both, and Pincer Strike (+10 to attack beside a summon). [A, the issue left the attributes to the spec]
- **`modifiers` reads `summoning`:** an attribute bonus is read through the generic `modifiers` function, which knew nothing of summons, so it gains an optional `summoning` input; absent is false and drops every `while_summoning` modifier. [A]
- **Striking together is a situation:** positions are the caller's (docs/core/combat.md), so `beside_summon` is a caller-named row of `situations.json` with no modifier of its own; the attacker's modifiers on the new `beside_summon_attack` target add to the attack when it is named. It works at every range, unlike `flanking`, so the ranged path has it too. [A]
- **Mounts** are written in the spec as a later extension (a summoned steed and a mounted bonus) and not built.
- **Numbers** (+1 per attribute per tier, +10, levels 6, 10, 12, 18) are [A] until a playtest.
