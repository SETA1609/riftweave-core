# Printouts

The `print` command renders a printable page from the data (APP-11), so a printed number is the
number the rules compute and a rule change is a rebuild, not an errata pass. It is an adapter like
`session`: every number on a page comes from another command's result (`derived`,
`progression read_skill`, `modifiers`, `monster_sheet`), never from a formula of its own.

```
input    kind: sheet | spell | monster | condition, and that kind's field below
result   title, html: one self-contained page (inline CSS, no script), ready to print
```

| `kind` | Input | The page shows |
| --- | --- | --- |
| `sheet` | `actor`, a stored actor | attributes; every derived stat with the actor's modifiers; each raised skill (one the sheet writes, as M5-49's light sheet allows) with its **target**, the skill plus the actor's modifiers on it, so a player only checks `roll <= target`; perks with their rank; known spells |
| `spell` | `spell`, a spell key | label, color and governing skill, level, mana cost and reagents, delivery, concentration, save, and each effect with its magnitude, duration, and target |
| `monster` | `monster`, a composed monster row (as `monster_sheet` takes), and `label` | level, attributes, HP, MP, AP, evasion, each attack with its **to-hit** (the skill it rolls with: `unarmed` for a natural weapon, the carried weapon's skill otherwise, read as `monster_sheet` reads it) and dice, and its resistances |
| `condition` | `condition`, a condition key | label, description, what applies and removes it, stacking, and its modifiers |

The actor's modifiers are the ones a session fills ([sessions.md](sessions.md#what-the-actor-brings-to-a-step-app-13)):
perks at their rank, then held modifiers. A target on the sheet is `modifiers` applied to the
skill's value with the modifiers on that skill, the same stacking any command uses. Each printed
number carries a `data-` attribute naming what it is (`data-target="blades"`, `data-stat="hp"`,
`data-to-hit="bite"`), so a test reads the page and compares it with the command that computes
the number. PDF is the browser's print: the page sets `@page` and print CSS. [A]

[V 2026-11-14, `go test ./core`: a sheet's targets and derived stats, a stat block's to-hit and
evasion, and a spell and a condition card each match the command that computes them]

A natural weapon's dice are not printed: they move with the creature's size, a rule the Go core
does not hold yet, so the stat block names the weapon and its type and leaves the dice to the row. [A]
