# Printouts

Modified: 2026-11-23

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
| `sheet` | `actor`, a stored actor, and `state` | attributes; every derived stat with the actor's modifiers; each raised skill (one the sheet writes, as M5-49's light sheet allows) with its **target**, the skill plus the actor's modifiers on it, so a player only checks `roll <= target`; perks with their rank; known spells |
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

## Printing a character

A sheet prints the character as it is now, or fresh (BE_S09_03). `state` picks which:

- **`current`** (the default) is the session's state: next to each maximum, the HP, mana, and
  stamina left (`data-current="hp"`), the held conditions with their stacks
  (`data-condition="frightened"`), the held modifiers with their source and how long they have
  run, and the gear with its quantity, material, and the slot it is worn in
  (`data-item="core:weapon/shortsword"`). Every skill target takes the perks and the held
  modifiers, as a session step reads them ([sessions.md](sessions.md#what-the-actor-brings-to-a-step-app-13)).
- **`fresh`** is the same character for a new table: its perks, skills, spells, and gear stay; its
  conditions and held modifiers go, and the HP, mana, and stamina left are the maximums.

A condition's own modifiers reach a printed number only as a session holds them, in `modifiers`;
the sheet does not apply a condition a second time. [A]

**The human CLI** prints from the character list: `w` on a character asks for a file (the
character's name with `.html` by default; add ` --fresh` after the path for a fresh sheet), runs
`print` on the character as saved, and says where it wrote the page. Open the file in a browser to
print it or save it as PDF.
[V 2026-11-21, `print_test.go` in core and in `cmd/riftweave`]

**The browser** prints from the sheet (the character in play) and from each row of the character
list (the character as saved): **Print** and **Print fresh** run `print` with the stored actor and
`state`, and open the page in a new tab with the browser's print dialog, so a PDF is the browser's
own and the core in the page prints offline (`printPage` and `sheetPrint` in
`frontend/src/app/files.ts`, FE_S09_04). [V 2026-11-23, `e2e/print.spec.ts` over HTTP and
WebAssembly: the wounded fixture prints 41 of 71 HP, Blades 34, and burning; fresh prints 71 and no
condition]
