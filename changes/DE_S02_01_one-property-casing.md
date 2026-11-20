### DE_S02_01_one-property-casing One property casing
- Changed: every schema property and data key is `snake_case`; 35 camelCase properties are renamed (`drBase` to `dr_base`, `parentRace` to `parent_race`, `valueFactor` to `value_factor`, ...) in schemas, data, modules, the reference code and the Go `json` tags **(breaking)**.
- Added: `validate.py` `check_property_casing` fails a camelCase schema property or data key (`CONTRIBUTING.md#naming`).
