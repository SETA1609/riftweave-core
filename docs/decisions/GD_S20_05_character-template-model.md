# GD_S20_05_character-template-model Character templates: a reusable change that adds attributes, skills, pools, and perks

Date: 2026-11-24

- **Decision:** a new collection, `character_templates` ([templates.md](../core/templates.md)), with its own schema; not the monster template table with an `applies_to`.
  - **Row:** `slot` (`undeath`, `lycanthropy`), optional `prerequisite`, `attributes` and `skills` added, `modifiers` (pools, resistances, drawbacks; every one `core:origin/innate`, checked by `validate.py`), granted `perks` at a rank, optional `tags`, and a `creature_type`. [V 2026-11-24, `test_character_templates.py`]
  - **Applied by `derived`:** its new `templates` input adds the attributes, then joins the modifiers and the granted perks' modifiers (each perk's own origin). The stored actor keeps `templates` (`template`, `how`), never folded into its attributes, and a session's `derived` step fills them from the actor. Without them the sheet is what it was. [V 2026-11-24, `derived_examples.json` in Python and through the Go package]
  - **One per slot:** `derived` refuses two templates in one slot (`revenant and ghoul share the undeath slot`). [V 2026-11-24]
  - **Recorded:** `how` is `granted` or `outside_rules`; it changes nothing on the sheet. [V 2026-11-24, `stored_actor.json`]
  - **Monster templates stay separate**: monster templates are composed into a monster row in its element, power, and nature slots and never change its creature type, while a character template is held at run time and comes off. [A]
  - **Example rows:** `revenant` and `ghoul`, both undeath, so the slot rule has two real rows. The lich, the vampire (GD_S20_06), and the werebeasts (GD_S20_07) are their own issues. Their numbers are [A].
- **Differences from the issue:**
  - **Tags:** the 2NF gate refuses a template's tags that repeat its creature type's (`undead`), so `tags` is optional and the example rows carry none; the schema changed in the data step for it.
  - **Skills added** are data and validated, but no command reads a template's skills yet; `progression`'s `read_skill` takes them in BE_S20_08 when a saved character takes a template on. [U]
  - The stored actor does not check the slot rule itself: the take-on and remove commands are BE_S20_08's, which keeps this PR off `stored.Run`, open in #478.
- **Open:** does a template cost XP or a perk pick [U]; the creature type and the prerequisite are data that no command reads yet [U].
