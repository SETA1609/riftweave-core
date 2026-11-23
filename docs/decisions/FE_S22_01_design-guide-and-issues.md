# FE_S22_01_design-guide-and-issues A design guide from the mockups, and its issues

Date: 2026-11-23

- **Decision:** a new story, S22 Visual design system (owner request, 2026-11-23). The owner's Google Stitch mockups (six screens, Tailwind HTML and screenshots) are kept in `docs/design/mockups/`, and `docs/design/README.md` turns them into a design guide: tokens, type, shapes, components, layouts, and print.
  - **A guide, not a spec** (owner, 2026-11-23): the mockups' invented rules and names (a crit on doubles, a "DC 14" test, fate tokens, "v2.4", school titles) are not copied. A screen shows what the core returns; a panel with no rule behind it is left out.
  - **Tailwind as plain CSS** (owner, 2026-11-23): the mockups' Tailwind classes are rewritten as declarations in each component's `.css`, reading tokens from `styles.css`. No Tailwind, font CDN, or Material Symbols in the app; icons stay Font Awesome (frontend.md, 2026-11-22).
  - **Fonts bundled**, not fetched, so the app stays offline-first.
  - **Module blocks** (owner, 2026-11-23): a block of an optional module shows only when that module is enabled (`AppContext.modules()`), through one directive `*rwIfModule` (`FE_S22_03`); `wuxing` gates every phase block today, and a later module (a weapon triangle, say) is gated the same way. Colors are core and never gated.
  - **Badge names:** `rw-color-badge`, not "school", since the ruleset's word is `color` (`schema.json`).
  - **Order:** tokens first (`FE_S22_02`), then the shared components (`03`, `04`) and the shell (`05`), then the pages. The page issues need the existing issues that build each page (`FE_S09_02`, `FE_S08_06`, `FE_S09_04`, `FE_S09_08`), so S22 restyles them and does not hold them back.
- **Why role FE:** the guide is the browser's UX, which FE owns (ROLES.md); no rule or contract changes.
- **Assumed** [A]:
  - The light palette is ours; the design file gives only the canvas and ink. `FE_S22_02` checks WCAG AA.
  - Space Grotesk and JetBrains Mono are under the SIL OFL [U]; `FE_S22_02` checks and adds them to NOTICE.
  - The existing token names keep their names with the new values, so screens do not break in `FE_S22_02`.
- **File count:** 19 counted text files; the six PNG screenshots have no changed lines.
- **Risk:** the mockups' HTML fetches Tailwind, fonts, and two portrait images from Google when opened in a browser. They are reference only, never served by the app.
