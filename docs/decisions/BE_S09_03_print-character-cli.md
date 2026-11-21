# BE_S09_03_print-character-cli Print the current character from both apps (backend)

Date: 2026-11-21

- **Decision:** `print` takes `state: current | fresh` for a sheet, `current` by default, so a page that printed before still prints; the rules are in [print.md](../backend/print.md#printing-a-character).
- **What fresh drops:** the conditions and the held modifiers, and it fills HP, mana, and stamina to their maximums. Perks, skills, spells, and gear stay, since they are the character, not the session. [A]
- **A condition is not applied twice:** a printed target takes the perks and the held modifiers, as a session step does; a condition is listed by name and stacks. A condition's effect on a number reaches the sheet only as the session holds it. [A]
- **The CLI key is `w`:** `p` already adds a character to the party. The flag ` --fresh` after the path follows Import's ` --accept-loss`.
- **The action is on the character list only.** The human CLI has no character sheet screen yet; the creation screen ends in a stored actor, and the list is where a saved character lives.
- **Risk:** `contract_diagram.py`'s Modified line and `docs/backend/contract-diagram.md` also change in BE_S08_10; whichever merges second regenerates the page (`python ruleset/scripts/contract_diagram.py --write`).
