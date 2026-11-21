### FE_S11_01_frontend-i18next Localize the frontend with i18next
- Added: the frontend speaks English, German, and Spanish through i18next 26.7.2, with the locale files in `frontend/src/locales/` bundled into the app; a language select in the header switches every localized string without a reload, sets `<html lang>`, and is saved.
- Added: the app shell, the menus and Back, the theme toggle, the page titles, and the Settings, Modules, and Network screens use locale keys; the play and character screens follow in FE_S11_04_frontend-screen-strings.
- Added: `i18n.spec.ts` fails on a key missing from a language, a menu entry without a key, or a literal string in a template.
