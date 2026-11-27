# FE_S22_02_design-tokens-and-fonts The design guide's tokens and fonts in styles.css

Date: 2026-11-27

- **Decision:** `src/styles.css` carries the guide's tokens, light and dark, under the same `data-theme` rules as before; the names the components use (`--body`, `--body-bg`, `--border`, `--surface`, `--primary`, `--space-1` … `--space-4`, `--success`, `--info`) stay and point at the guide's values, so no screen's own CSS changed [V 2026-11-27, unit and e2e suites unchanged but for the colors they read].
- **Contrast:** every text token on every surface, and every text on its own fill, reaches 4.5:1; the strong keyline reaches 3:1 [V 2026-11-27, `tokens.spec.ts` reads the values from the file]. Two light values did not and were darkened by the least that passes, in the guide too: primary `#0f766e` → `#0f746c` (4.44 → 4.56 on `--surface-container-high`), tertiary `#b45309` → `#a94e08` (4.07 → 4.51). The guide has no success, info, or tertiary fill; those alert fills are ours [A], checked the same way.
- **Fonts:** `@fontsource/space-grotesk` and `@fontsource/jetbrains-mono` 5.3.0 (published 2026-07-19), pinned exactly, OFL-1.1 [V 2026-11-27, `npm view` and the packages' LICENSE files], listed in NOTICE. Only the Latin subset (Basic Latin and Latin-1, which hold German and Spanish) and the weights the guide uses: Space Grotesk 600 and 700, JetBrains Mono 400, 600, and 700 [A].
- **Offline:** the service worker prefetches `/media/*.woff2` with the app, so the fonts draw with the network off [V 2026-11-27, Pages e2e offline].
- **Type:** body text and data are mono with tabular figures; `h1`–`h3` are the heading font. The page keeps its 64rem cap unless a screen asks for the wide layout (`data-layout="wide"` on any element) [A].
- **Shared classes:** 4px radius, 1px keylines, no shadows; focus is a hard 2px primary outline; the ghost button's keyline turns secondary on hover; buttons are at least 44px tall under 768px; table and form labels are small uppercase mono.
- **Tests changed:** the mode e2e read the old Bootstrap background colors; they now read the guide's.
