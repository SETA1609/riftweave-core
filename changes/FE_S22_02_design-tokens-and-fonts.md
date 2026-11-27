### FE_S22_02_design-tokens-and-fonts The design guide's tokens and fonts in styles.css
- Changed: the web app takes the design guide's look: its dark and light colors, the nine color and five phase hues as tokens, Space Grotesk for headings and JetBrains Mono for text and numbers, 4px radius, keylines with no shadows, and 44px buttons on small screens; every text pair reaches WCAG AA, and two light colors were darkened to get there.
- Added: the dependencies `@fontsource/space-grotesk` and `@fontsource/jetbrains-mono` 5.3.0, bundled so the fonts work offline, with their licenses in NOTICE.
