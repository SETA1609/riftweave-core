# FE_S08_12_back-on-every-screen A Back button on every screen

Date: 2026-11-27

- **Decision (the spec step):** one `rw-screen-header` in the app shell, above the router outlet, rather than one in each screen's template. Every route, including screens added later (the bestiary, the encounter tracker), gets the title and Back with no markup of its own, and the screens' templates stay untouched, so this does not conflict with the issues that edit them (FE_S11_04) [V 2026-11-27, e2e].
- **The menus** drop their own Back card and use the header's, so no screen shows two Backs. The CLI keeps Back as the last entry of each menu; `app-layout.md` now says each app's way [A].
- **Where Back goes:** the parent menu from `menu.json` (`pathOf`), not history, so a deep link has a way back [V 2026-11-27, unit spec over every route, e2e by URL].
- **Esc** navigates the same way unless a field or an open dialog has focus [V 2026-11-27, e2e].
- **Unsaved work:** a `CanDeactivate` guard asks through `confirm`; character creation opts in while its form differs from what it was when opened or last created from. The spell editor's draft does not opt in yet [U: whether it should, since its draft is downloaded, not saved]. Closing the tab is not asked [A].
- **Tests changed:** the menu specs no longer expect a Back card; two e2e checks for a heading named Characters now name the card's region, since the header repeats the title.
