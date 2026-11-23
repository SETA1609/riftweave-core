# S22 Visual design system

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S22 Visual design system

**Goal:** The Angular app takes the look of the owner's mockups ([design guide](../design/README.md)): the guide's tokens and bundled fonts in plain CSS, shared components for the colors, phases, roll targets, and meters, a side navigation and a mobile tab bar, and the sheet, spell editor, bestiary, and print pages laid out as in the mockups. Every number still comes from the core.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `FE_S22_01_design-guide-and-issues` | FE | A design guide from the mockups, and its issues | — | — |
| `FE_S22_02_design-tokens-and-fonts` | FE | The design guide's tokens and fonts in styles.css | — | FE_S22_01_design-guide-and-issues |
| `FE_S22_03_color-and-phase-badges` | FE | Badges for the nine colors and the five phases | — | FE_S22_02_design-tokens-and-fonts |
| `FE_S22_04_roll-target-and-meters` | FE | The roll target, the resource meter, and the stepper | — | FE_S22_02_design-tokens-and-fonts |
| `FE_S22_05_app-shell-navigation` | FE | A side navigation on desktop and a tab bar on mobile | — | FE_S22_02_design-tokens-and-fonts, FE_S08_17_pinned-navbar |
| `FE_S22_06_character-sheet-design` | FE | The character sheet in the design guide's layout | — | FE_S22_03_color-and-phase-badges, FE_S22_04_roll-target-and-meters, FE_S22_05_app-shell-navigation |
| `FE_S22_07_spell-editor-design` | FE | The spell editor in the design guide's layout | — | FE_S09_02_spell-editor-page, FE_S22_03_color-and-phase-badges, FE_S22_04_roll-target-and-meters |
| `FE_S22_08_bestiary-design` | FE | The bestiary and its stat block in the design guide's layout | — | FE_S08_06_bestiary-page, FE_S22_03_color-and-phase-badges, FE_S22_04_roll-target-and-meters |
| `FE_S22_09_print-layouts-design` | FE | Ink-light A4 print for the sheet and the stat block | — | FE_S09_04_print-character-page, FE_S09_08_custom-monster-page, FE_S22_02_design-tokens-and-fonts |

**Done when:** every issue above is done; the app renders in the guide's look in dark and light mode with the network off; the sheet, spell editor, and bestiary follow their mockups' layout at 1280px and 375px; the sheet and a monster print on one A4 page each.
