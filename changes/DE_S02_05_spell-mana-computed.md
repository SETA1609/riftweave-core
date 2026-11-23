### DE_S02_05_spell-mana-computed Third normal form: a spell's mana is computed, and every reader shows the number
- Removed (breaking): `cost.mana` on spells. No spell row or override stores its mana; a row or a draft that writes it is refused by `validate.py` and the spell author.
- Changed: the cast, the balance report, the compendium's spell table, and the spell card compute a spell's mana with the spellcraft cost formula, a summon block priced as its summon effect.
