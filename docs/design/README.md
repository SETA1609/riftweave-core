# Design guide

Modified: 2026-11-23

How the Angular app looks: the colors, type, shapes, components, and layouts every screen
uses (story [S22](../stories/S22-visual-design-system.md)). The guide is adapted from a set of
mockups the owner made with Google Stitch (2026-11-23), named "Tactical Arcane Codex": obsidian
surfaces, teal and cyan accents, amber for hazards, and numbers that never need arithmetic at
the table. How the code is written stays in [frontend.md](../backend/frontend.md#conventions).

## The mockups are a guide, not a spec

The pictures and their HTML are in [mockups/](mockups/). They show the **look and the layout**.
They are not a 1:1 target, and their game content is invented (owner, 2026-11-23):

- **Numbers and rules come from the core.** The mockups invent rules (a crit on doubles, a
  "DC 14 STR test", fate tokens, a "v2.4" engine) and names (Vaelen Voss, the Riftborn
  Dreadstalker, school titles like "Abjuration"). A screen shows what the core returns, as
  [frontend.md](../backend/frontend.md) requires: the crit and fumble tags of
  [resolution.md](../core/resolution.md), the nine colors of `schema.json`, the data's own names.
- **A panel with no rule behind it is left out**, not faked: no fate tokens, no "WASM latency"
  widget, no "arcane strain".
- **Tailwind is a source, not a dependency.** The HTML uses Tailwind's CDN, Google Fonts, and
  Material Symbols. The app uses none of them: the classes are rewritten as plain CSS (below),
  the fonts are bundled, and icons stay Font Awesome. Open the HTML only to read its classes; it
  fetches from the network when opened, and the app never serves it.

| Mockup | Screen | Issue |
| --- | --- | --- |
| [character-sheet-desktop](mockups/character-sheet-desktop.png) ([html](mockups/character-sheet-desktop.html)) | The sheet at 1280px: three panes | FE_S22_06 |
| [character-sheet-mobile](mockups/character-sheet-mobile.png) ([html](mockups/character-sheet-mobile.html)) | The sheet at 360px: one column, a bottom dock | FE_S22_06 |
| [spell-editor](mockups/spell-editor.png) ([html](mockups/spell-editor.html)) | The spell editor: effect library, effect stack, live cost and card | FE_S22_07 |
| [bestiary](mockups/bestiary.png) ([html](mockups/bestiary.html)) | The bestiary: roster and filters, the stat block, the attack resolver | FE_S22_08 |
| [character-sheet-print](mockups/character-sheet-print.png) ([html](mockups/character-sheet-print.html)) | The sheet printed on A4, ink-light | FE_S22_09 |
| [monster-print](mockups/monster-print.png) ([html](mockups/monster-print.html)) | A stat block printed on A4, ink-light | FE_S22_09 |

## From Tailwind to plain CSS

A Tailwind class in a mockup becomes a declaration in the component's own `.css`, reading a token
from `src/styles.css`. Nothing is copied as a class name.

| Tailwind in the mockup | Plain CSS in the app |
| --- | --- |
| `bg-surface-container`, `text-on-surface`, `border-outline-variant` | `background: var(--surface-container)`, `color: var(--on-surface)`, `border: 1px solid var(--outline-variant)` |
| `font-headline-lg` (from the config's `fontSize`) | `font: 600 var(--headline-lg) / 1.2 var(--font-heading)` |
| `font-mono`, `tabular-nums` | `font-family: var(--font-mono)`, `font-variant-numeric: tabular-nums` |
| `grid grid-cols-12 gap-6` | `display: grid; grid-template-columns: repeat(12, 1fr); gap: var(--gutter-desktop)` |
| `md:grid-cols-2`, `lg:col-span-3` | `@media (min-width: 768px) { … }`, `grid-column: span 3` |
| `rounded`, `rounded-lg` | `border-radius: var(--radius)`, `var(--radius-lg)` |
| `bg-primary/10` | `background: color-mix(in srgb, var(--primary) 12%, transparent)` |

## Tokens

The values come from the mockups' design file; the token names are ours. Token issue: FE_S22_02.

### Color, dark (the default look)

| Token | Value | Use |
| --- | --- | --- |
| `--canvas` | `#0b1326` | The page |
| `--surface-container-low` | `#131b2e` | Panels |
| `--surface-container` | `#171f33` | Raised modules inside a panel |
| `--surface-container-high` | `#222a3d` | Overlays, roll windows |
| `--outline-variant` / `--outline` | `#3d4947` / `#879391` | 1px keylines / stronger lines |
| `--on-surface` / `--muted` | `#dae2fd` / `#bcc9c6` | Text / secondary text |
| `--primary` / `--on-primary` | `#6bd8cb` / `#003732` | Main actions, focus, a pass |
| `--secondary` | `#7bd0ff` | Highlights, targets, links |
| `--tertiary` | `#ffb95f` | Hazards, tagged skills, warnings |
| `--danger` / `--danger-bg` | `#ffb4ab` / `#93000a` | A fail, a breach, errors |

### Color, light

The design file only says "alabaster `#F8FAFC` with slate ink `#0F172A`, same hues". The light set
is ours [A]; FE_S22_02 checks it reaches WCAG AA contrast:

| Token | Value |
| --- | --- |
| `--canvas` | `#f8fafc` |
| `--surface-container-low` / `--surface-container` / `--surface-container-high` | `#ffffff` / `#f1f5f9` / `#e2e8f0` |
| `--outline-variant` / `--outline` | `#cbd5e1` / `#64748b` |
| `--on-surface` / `--muted` | `#0f172a` / `#475569` |
| `--primary` / `--on-primary` | `#0f766e` / `#ffffff` |
| `--secondary` | `#0369a1` |
| `--tertiary` | `#b45309` |
| `--danger` / `--danger-bg` | `#b91c1c` / `#fee2e2` |

### The nine colors and the five phases

Each magic color has a hue **and a glyph**, so a badge never relies on color alone. The label is
the color's name from the locale files; the mockups' school titles ("Destruction") are not used.

| Color | Hue | Glyph |
| --- | --- | --- |
| `red` | `#ef4444` | triangle, point down |
| `orange` | `#f97316` | hexagon |
| `yellow` | `#eab308` | diamond |
| `green` | `#10b981` | square |
| `blue` | `#06b6d4` | circle |
| `indigo` | `#6366f1` | octagon |
| `violet` | `#a855f7` | four-point star |
| `white` | `#f8fafc` (`#94a3b8` outline in light mode) | triangle, point up |
| `black` | `#64748b` | plus |

| Phase | Hue |
| --- | --- |
| `wood` | `#059669` |
| `fire` | `#dc2626` |
| `earth` | `#d97706` |
| `metal` | `#94a3b8` |
| `water` | `#0284c7` |

A phase the character does not hold is drawn in `#475569`, dimmed. The phases belong to the
`wuxing` module: without it, nothing phase-related is drawn ([Module blocks](#module-blocks)).

### Type

- **Space Grotesk** for headings: names, section titles, a spell's or a monster's title.
- **JetBrains Mono** for every number and every data field: targets, dice, tables, logs, tags.
  Numbers use `tabular-nums`, so a changing HP value never shifts its box.
- Both are bundled with the app (no font CDN), so the app works offline. Both are under the SIL
  Open Font License [A, FE_S22_02 checks and adds them to NOTICE].

| Token | Size / weight / line | Font |
| --- | --- | --- |
| `--headline-xl` | 2.5rem / 700 / 1.1 (1.875rem under 768px) | heading |
| `--headline-lg` | 1.75rem / 600 / 1.2 (1.375rem under 768px) | heading |
| `--headline-md` | 1.25rem / 600 / 1.3 | heading |
| `--body-lg`, `--body-md`, `--body-sm` | 1rem, 0.875rem, 0.75rem / 400 | mono |
| `--label-lg`, `--label-md`, `--label-sm` | 0.875rem, 0.75rem, 0.625rem / 600–700, uppercase, tracked | mono |

### Shape, space, depth

- **Radius:** 4px (`--radius`) for buttons, inputs, badges, and cards; 8px (`--radius-lg`) for
  dialogs and screen-level overlays. Circles only for phase tokens, counters, and round gauges.
- **Space:** a 4px base: `--space-xs` 0.25rem, `--space-sm` 0.5rem, `--space-md` 0.75rem,
  `--space-lg` 1.25rem, `--space-xl` 2rem; gutters 1rem (1.5rem from 1024px), page margins 1rem
  (2rem from 1024px).
- **Depth without shadows:** each layer is a lighter surface with a 1px keyline. Overlays and roll
  windows add a 1px primary stroke and a soft glow (`0 0 16px` of the primary at 25%). Focus and a
  failed check snap to a hard border (primary, or danger), with no blur.
- No parchment, no bevels, no ornament: dividers are 1px lines.

## Components

Shared components in `src/app/ui/`, each in its own issue:

- **Buttons** (FE_S22_02): primary is a filled primary with a bold mono label; ghost is a
  transparent fill with a keyline that turns secondary on hover. At least 44px tall under 768px.
- **Color badge** `rw-color-badge` and **phase badge** `rw-phase-badge` (FE_S22_03): the glyph, the
  uppercase label, a 12% fill and a 1px border in the hue. **Phase matrix** `rw-phase-matrix`: the
  five phases in a row, each with the character's value, the held ones lit.
- **Roll target** `rw-roll-target` (FE_S22_04): the target large (`≤ 74%`), the base and each
  modifier small beside it, situational modifier buttons, a roll button, and the result tagged
  `CRIT`, `PASS`, `FAIL`, or `FUMBLE` from the core's tags, with the margin.
- **Resource meter** `rw-resource-meter` and **stepper** `rw-stepper` (FE_S22_04): a bar with the
  exact fraction (`48 / 52`), and −/+ buttons of at least 36×36px.

## Module blocks

A block that belongs to an optional module (`ruleset/modules/`) is shown **only when that module
is enabled** (owner, 2026-11-23). Core is always loaded; a module is not.

- **Which modules:** `AppContext.modules()`, the enabled module ids from the settings file
  ([settings.md](../backend/settings.md)). A screen does not read the settings itself.
- **How:** a structural directive, `*rwIfModule="'wuxing'"`, that renders its block only when the
  id is enabled and follows the signal, so toggling a module in the settings adds or removes the
  block without a reload. No block is hidden with CSS: a disabled module's block is never created,
  so it fetches nothing and has no locale keys to show.
- **Layout:** the grid closes up around a missing block; no empty panel, no placeholder.
- **Today:** `wuxing` gates the phase badge, the phase matrix, the phase tabs and filters, phase
  affinities on a stat block, and the phase row on a printed page. **Later:** every new optional
  module that adds a block, for example a weapon triangle, is gated the same way, under its own id.
- **Not gated:** the nine colors are core (`schema.json`), so color badges always show.

## Layouts

| Width | Grid | Shape |
| --- | --- | --- |
| 360–767px | 4 columns, 16px margins | One column. A bottom tab bar for navigation; on the sheet, a sticky dock (64px) with HP, MP, SP, and a quick check |
| 768–1023px | 8 columns, 24px margins | Two columns: the active play column (5) and reference (3) |
| 1024px and up | 12 columns, 32px margins | A side navigation, then three panes: status (3), actions and rolls (6), spells and reference (3) |

- **Navigation** (FE_S22_05): the side navigation and the bottom tab bar list the main menu of
  [app-layout.md](../backend/app-layout.md) from `menu.json`, the same tree the CLI shows. The top
  navbar (FE_S08_17) keeps the name, the language, the theme, and the GitHub link.
- **Print** (FE_S22_09): ink-light black on white, A4, `@media print`. Boxes to tick for HP, MP,
  and SP; skill targets in two tables; a footer with the character, page, and a blank for the
  session date. No color is needed to read it.
