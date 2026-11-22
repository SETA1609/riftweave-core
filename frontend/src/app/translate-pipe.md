# translate-pipe.ts

Modified: 2026-11-22

The `t` pipe (FE_S11_01): `{{ 'menu.back' | t }}`, or `'app.rules' | t: { version: v }` for a key with values. Generated with `ng generate pipe translate`, hence the file name.

- It calls `I18n.translate` ([i18n.md](i18n.md)); an empty key gives an empty string, so an alert with no message stays empty.
- `pure: false`: the text changes with the language while the key stays the same. The language signal `translate` reads re-renders the view, so a switch needs no reload.
- Cost: an impure pipe runs on every check of its view. In this zoneless app a view is checked when one of its signals changes, and each call is one lookup.

References: [frontend.md#languages](../../../docs/backend/frontend.md#languages); Angular [pipes](https://angular.dev/guide/templates/pipes) (pure and impure). [U]
