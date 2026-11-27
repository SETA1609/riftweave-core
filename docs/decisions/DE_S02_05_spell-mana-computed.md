# DE_S02_05_spell-mana-computed Third normal form: a spell's mana is computed, and every reader shows the number

Date: 2026-11-23

- **Decision:** no spell row or override stores `cost.mana`; every reader computes it with the spellcraft cost formula ([magic.md#cost-formula](../modules/magic.md#cost-formula)), as DE_S02_03 did for monster HP. `cost.reagents` stays: reagents are primary facts.
  - **The gate:** the schema's `cost` holds only `reagents` and refuses other fields, so a row or an override that writes `mana` fails `validate.py`. The spell author refuses it too (a `cost` problem at `/cost/mana`, with the formula's number), right number or not. [V 2026-11-23, `test_spells.py`, the author test]
  - **The readers:** Python's cast (`casting_reference`) and `balance_report` price from `spell_cost`. In Go, `spellcraft.SpellMana` is the formula on the effects, a summon block priced as the summon effect it replaces; the cast, the spell card, and the compendium read it. [V 2026-11-23, all 100 casting vectors in Go]
  - **Every reader shows the number:** the compendium's spell table fills `cost.mana` on each composed row (core and module rows alike), and the spell card prints it, so no one prices a spell by hand. [V 2026-11-23, the catalog test for every core spell, and fire_bolt 6, greater_unlock 30, summon_familiar 18, raise_thrall 36; the print test runs in CI]
- **Data:** all 75 rows lost `cost.mana`; the 74 with no reagents have no `cost` block at all.
- **Breaking:** a file or module that writes a spell's `cost.mana` no longer validates; drop the field. The author's old "mana differs" warning is now a refusal.
- **Not done here:** the spell cards of BE_S09_05 and the SRD's spell list (S21) are not built yet; their issues read the same `SpellMana`.
