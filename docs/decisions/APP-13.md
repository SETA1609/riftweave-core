# APP-13 A session step applies the actor's perks, held modifiers, and gear

Date: 2026-11-14

- **Decision:** the session emits a perk's own `modifiers` at the held rank, by the rule in `docs/modules/progression.md` (an add times the rank, a multiply as `1 + (m − 1) × rank`, a grant or set once), then appends the held modifiers. No new stacking: the commands stack them by origin as they already do. [V 2026-11-14, `go test ./core`]
- **Decision:** `derived`, `casting`, and the defender's side of `combat` read each modifier's target themselves, so they get every modifier. The attacker's side of `combat` passes every modifier in `modifiers` to the attack roll without reading its target, so the session gives it only those on the attack's skill; otherwise Toughness's `hp` would raise the to-hit target by 30, which a first run of the vector showed. [V 2026-11-14]
- **Decision:** the attack's skill is the first equipped weapon's `skill` by slot name, since the stored actor names no slots, else `unarmed`, read as `read_skill` reads it. Armor DR is the sum of each equipped armor piece's `drBase` and its material's `armor.defense`, the first default material when the piece names none (`docs/modules/armor.md`). [A]
- **Decision:** a field the step names still wins, so every earlier playthrough vector, which names its skill, DR, and modifiers, is unchanged. [V 2026-11-14]
- **Why:** found by the M5-18 agent runs: no perk changed any result, Toughness included.
- **Open:** the stored actor has no race or background, so their modifiers are not filled. Armor's evasion and speed penalties, layering (`docs/modules/armor-layering.md`), and an enchanted item's added effects are not applied here. [A]
- **Risk:** the playtest driver's runs now get their perks' modifiers on `derived`, so a rerun of it will differ from the M5-18 report.
