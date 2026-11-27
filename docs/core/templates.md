# Character templates

Owner request, 2026-11-20 (GD_S20_05). A character can take on a **template** later in play, such
as a lich, a vampire, or a werebeast. A template improves the base: it adds attributes, skills,
pools, and perks on top of what the character already is, and it comes off again. It is a row of
`ruleset/data/character_templates/core.json` (`core:character_template/<key>`), with its own
schema, `character_template.schema.json`.

```
template      key, label, description
  slot        one template per slot: undeath (lich, vampire, revenant), lycanthropy (werebeasts)
  prerequisite?  { level, perks[] }, as a perk's
  attributes  { ability: n }   added to the attribute; a missing one counts 0, and the sum may pass 10
  skills      { core:skill/<key>: n }   added to the skill, where a sheet reads it
  modifiers[] modifiers on the derived keys (hp, mp, sp, ap), resistances, and drawbacks, each
              with the innate origin, so origins/core.json's stacking applies
  perks[]     { perk, rank }   granted while the template is held, not counted against perks_per_level
  tags[]      the template's tags (undead, shapechanger)
  creature_type?  core:creature_type/<key>: the character also counts as it (a bane, turn undead,
              and a condition immunity read it)

a held template   { template, how }   how: granted (a GM's story or game event) or outside_rules
```

- **Applied, not baked in:** a held template is kept on the stored actor (`templates`), never
  folded into its attributes. The `derived` command takes `templates` and adds each one's
  attributes, then its modifiers and its perks' modifiers at their ranks, to the sheet. Without
  them the sheet is what it was: removing a template removes everything it gave. [A]
- **One per slot:** two held templates in one slot are refused, by `derived` and when a stored
  actor is checked. [A]
- **Who may take one** (owner, 2026-11-20): by the rules a player cannot pick a template; it is no
  level-up choice, and the GM grants it through a story or a game event (`how: granted`). The app
  has a checkbox, off by default, to take one anyway, for a supposed character (an NPC, a
  one-shot, a what-if build): `how: outside_rules`, so the sheet shows the character was built
  outside the rules. The checkbox lifts only the grant rule; one per slot still holds. [A]
- **Validation:** `validate.py` checks that every ref resolves (skills, perks, the creature type),
  that every modifier is innate, and that each slot is one of the slots above. [A]
- **Cost:** a template costs nothing by itself; a row may carry its price as drawbacks (negative
  modifiers, a vulnerability). Whether one should also cost XP or a perk pick is open. [U]

## Monster templates stay separate

A monster template (`monsters/templates.json`, [monsters.md](../modules/monsters.md#templates)) is
applied when a monster row is composed, in its `element`, `power`, and `nature` slots, and never
changes the monster's creature type. A character template is held at run time and comes off
again. The two keep separate tables: a GM's vampire monster is a monster row with its own numbers,
and a vampire character is a character with the `vampire` template. A later issue may let a monster
row name a character template. [A]
