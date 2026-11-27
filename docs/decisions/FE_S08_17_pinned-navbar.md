# FE_S08_17_pinned-navbar A navbar pinned on top, with a burger menu on mobile

Date: 2026-11-27

- **Decision:** `rw-navbar` replaces the shell's header: the name home (kept as the page's `h1`), the language select, the theme toggle, and the GitHub link; sticky on top; under 768px the three items fold into a burger dropdown [V 2026-11-27, e2e at 1024px and 375px over both transports].
- **The package:** `@fortawesome/free-brands-svg-icons` 7.3.1, pinned exactly, the same version as the solid set (published 2026-07-15) [V 2026-11-27, `npm view`]. Font Awesome's free icons are CC BY 4.0; the attribution is the license comment the package keeps in its files, and the bundle keeps third-party license comments [A: to confirm in the licensing audit, PO_S01_01]. GitHub's logo is used only to link to the repository [A].
- **Escape:** closes the dropdown and stops there while it was open, so with FE_S08_12's screen header (Esc goes back) one Escape does one thing [V 2026-11-27, unit spec]. The header listens on the document; the navbar on its own element, which has focus after the burger is clicked [A].
- **Sticky, not fixed**, so nothing hides under the bar and no padding has to match its height [A].
- **Later, its own issue:** a donations link in the same navbar (from the issue).
