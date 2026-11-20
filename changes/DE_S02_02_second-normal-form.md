### DE_S02_02_second-normal-form Second normal form: a reference is the key alone
- Added: integrity gate 8: `validate.py` fails a field beside a `core:` reference that repeats the same field of the row it names (`docs/core/integrity.md#8-a-reference-is-the-key-alone`).
- Removed: `label` in race `traits` (now `{ id }`), and a crafted item's `phase` and `engraving_jewel.phase`; the phase is the gem's, else the material's **(breaking)**.
- Changed: a monster attack's `label` is optional and defaults to its weapon's label, and its `phase` defaults to a carried weapon's; 19 copied attack labels and 2 copied phases are removed, and the `print` monster stat block reads the weapon row's label **(breaking)**.
