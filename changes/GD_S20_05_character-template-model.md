### GD_S20_05_character-template-model Character templates: a reusable change that adds attributes, skills, pools, and perks
- Added: the `character_templates` collection and schema (slot, attributes, skills, innate modifiers, granted perks, creature type), with the Revenant and Ghoul as examples, and its `validate.py` gate.
- Added: `derived`'s `templates` input, one template per slot, and a stored actor's `templates` with how each came (`granted` or `outside_rules`).
