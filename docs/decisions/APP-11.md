# APP-11 Printouts generated from the data

Date: 2026-11-14

- **Decision:** one command, `print`, with a `kind` (`sheet`, `spell`, `monster`, `condition`), returning `title` and one self-contained HTML page (inline CSS, `@page`, no script). PDF is the browser's print, so the backend needs no PDF library. [A]
- **Decision:** `print` holds no rule. A sheet's derived stats come from `derived` with the actor's modifiers (the session's `actorModifiers`, APP-13); a skill's target is `progression read_skill` passed through `modifiers` with the modifiers on that skill; a stat block's to-hit and evasion come from `monster_sheet`, with the attack's skill `unarmed` for a natural weapon (M3-16) or the carried weapon's skill. Only the raised skills, the ones a sheet writes, are printed (M5-49). [V 2026-11-14, `print_test.go`]
- **Decision:** every printed number carries a `data-` attribute (`data-target="blades"`, `data-stat="hp"`, `data-to-hit="bite"`), and the test reads the page and compares each with the command that computes it. `html/template` refuses a dynamic attribute name with a hyphen, so the attribute is built in Go from a fixed name and an escaped key. [V 2026-11-14]
- **Decision:** the branch builds on APP-13, which adds the actor's modifiers and a weapon's skill to the Go core; it merges after APP-13. [A]
- **Why:** owner plan, M6-01: printouts should match the rules, so a rule change is a rebuild, not an errata pass.
- **Open:** a natural weapon's dice are not printed, since sizing them is a rule only the Python reference holds; the frontend does not offer printing yet; a spell card lists its effects as their JSON rows, not as prose. [A]
- **Risk:** the stat block takes a composed monster row from the caller, as `monster_sheet` does, so a printed stat block is only as right as the composition the caller ran.
