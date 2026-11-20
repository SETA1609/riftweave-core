#### GD_S20_07_werebeast-templates Werebeast templates: werewolf, werebear, and others
- Status: todo
- Priority: medium
- Needs: GD_S20_05_character-template-model
- Modified: 2026-11-20
- Do: Owner request, 2026-11-20: the `lycanthropy` slot, written as data in the model of `GD_S20_05_character-template-model`.
  - **The werebeasts:** at least werewolf and werebear, then others (wererat, weretiger, wereboar) as the owner picks.
  - **Each adds** attributes and skills of its beast:
    - a werewolf: AGI and PER, tracking and stealth;
    - a werebear: STR and END, more HP.

    Each also adds natural weapons (claws, a bite), and granted perks such as keen senses and regeneration [A].
  - **Two forms:** a werebeast has a human form and a beast form, and possibly a hybrid. Each form is a set of the template's modifiers that holds while the character is in that form, switched by an action with an AP cost [A]. Whether forms are part of the template model or a separate condition is settled with `GD_S20_05` [U].
  - **Drawbacks:** silver vulnerability (a damage-type or material weakness), and a moon-driven change the GM may call [A].
  - **The numbers:** checked against the balance bands at levels 5, 10, and 20, as the undead templates are.
- Keys: `werewolf`, `werebear`, ... (`slot: lycanthropy`); the forms; the change action
- Files: ~7 (ruleset/data/character_templates/core.json, ruleset/data/features/core.json, ruleset/data/natural_weapons/core.json, ruleset/data/actions/core.json, docs/modules/templates.md, ruleset/tests/vectors/derived_examples.json, changes/GD_S20_07_werebeast-templates.md)
- See: docs/issues/open/GD_S20_05_character-template-model.md, ruleset/data/natural_weapons/core.json, docs/stories/S19-turn-economy.md (the change action's AP)
- Done when: werewolf and werebear validate; `derived` vectors give a sheet in each form; changing form is an action with a cost; the balance report checks them.
