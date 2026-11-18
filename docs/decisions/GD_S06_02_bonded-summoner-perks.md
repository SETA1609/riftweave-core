# GD_S06_02_bonded-summoner-perks Bonded summoner perks: an active summon empowers its caster

Date: 2026-11-19

- **Decision:** the bonded summoner's perks emit ordinary modifiers with a new duration, `while_summoning`: live only while the holder has an active summon (`caster.summons`, minions excluded). The cast drops them otherwise. Three perks: Kindred Focus (+10 to every magic skill), Shared Wellspring (`spell_cost` x0.8), and Elemental Kinship (+4 `bonded_element_damage` on a damage packet whose type matches an active summon's element template). [A, the issue's starting points kept; the owner did not settle them]
- **"While a summon is active" is a duration, not a new field:** `duration` already says when a modifier ends (`while_applied`, `while_equipped`), so the condition is one more value, read by the one consumer that knows the summons, the cast. The kernel gains no notion of a bond. [A]
- **Strongest only:** each perk is one source with one modifier per target, so two summons give each bonus once without a rule of its own. [A]
- **The element match** needs the summons' element templates, which `caster.summons` (group refs) does not carry, so the cast takes `caster.summon_templates` (template refs) beside it. [A]
- **Every magic skill:** Kindred Focus lists one `add 10` per `<color>_magic` (nine modifiers), which meets the perk skill-bonus rule (two skills or more, at least +10). [A]
- **Numbers** (+10, x0.8, +4, levels 6, 12, 18) are [A] until a playtest runs a bonded summoner against a plain caster.
- **Verified:** the 8 bonded casting vectors pass in Python (420 tests) and Go (`go test ./...`); `validate.py` passes 52 of 52 files. [V 2026-11-19]
- **A vector fixed before the data step:** two vectors first expected a failed check's margin as -5; a failed check reports `margin: null` (docs/core/resolution.md), so the tests commit was corrected, not the rule.
- **Differences from the issue:** `docs/core/casting.md`, `docs/core/vectors.md`, `ruleset/schemas/schema.json`, `backend/contract/casting.json`, Go (`internal/casting`, `internal/data`), and the regenerated `perk-ladders.md` are files the issue did not list; `test_perks.py` needed no change, since its catalogue checks cover the three perks.
- **Branch:** made from `GD_S06_01_bound-weapon-materials` (the owner asked for the S06 issues one after another), so this pull request carries that one's commits until it merges.
