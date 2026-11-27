# FE_S09_04_print-character-page Print the current character from both apps (frontend)

Date: 2026-11-23

- **Decision:** the browser half of BE_S09_03: **Print** and **Print fresh** on the sheet (the character in play) and on each row of the character list (the character as saved) run `print` with `kind: sheet`, the stored actor, and `state`, and open the page in a new tab with the browser's print dialog ([print.md](../backend/print.md#printing-a-character)). No contract or core change.
- **A new tab, not an iframe:** the page is the command's self-contained HTML, opened as a blob URL, so the reader sees the sheet, the browser saves it as PDF, and it works offline with the core in the page. [V 2026-11-23, `e2e/print.spec.ts` over HTTP and WebAssembly]
- **The session's numbers:** the wounded fixture prints 41 of 71 HP, Blades 34 (its written value, as `read_skill` reads it with no modifier on it), and burning; fresh prints 71 and no condition. The Go side already checks every printed number against the command that computes it (BE_S09_03). [V 2026-11-23]
- **Not localized:** `sheet.html` and `character-list.html` are still on the i18n spec's `PENDING` list; FE_S11_04 moves their words, the new buttons included, to keys. [A]
- **Also fixed:** the missing `menu.compendium` locale key, as BE_S08_05 and FE_S09_02 do; the same line, so the branches merge in any order.
