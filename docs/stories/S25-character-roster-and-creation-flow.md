# S25 Saved characters, the profile, and a step-by-step creation flow

Format, ids, and roles: [README.md](README.md#ids), [ROLES.md](../issues/ROLES.md).

## S25 Saved characters, the profile, and a step-by-step creation flow

**Goal:** A player makes a character through a flow of pages, one step or a group of steps at a time, in the order the core gives (who: name and sex; race; background; traits; attributes; tag skills; perks; review). The character is saved to a roster that tracks every saved character, as the bestiary tracks monsters, and a single player selects one as the profile: the current character every play page starts from.

**Issues:**

| Issue | Role | Title | Was | Needs |
| --- | --- | --- | --- | --- |
| `PO_S25_01_file-character-roster-issues` | PO | File the character roster, profile, and creation flow issues | — | — |
| `GD_S25_02_sex-and-starting-perks` | GD | Sex and the starting perk as creation rules | — | — |
| `BE_S25_03_creation-flow-steps` | BE | The creation command leads a draft step by step (backend) | — | GD_S25_02_sex-and-starting-perks |
| `BE_S25_04_current-character-profile` | BE | The profile: the current character of a single player (backend) | — | — |
| `FE_S25_05_creation-flow-page` | FE | Character creation as a step-by-step flow (frontend) | — | BE_S25_03_creation-flow-steps, FE_S08_12_back-on-every-screen |
| `FE_S25_06_character-roster-page` | FE | The character roster: saved characters tracked like the bestiary (frontend) | — | BE_S25_04_current-character-profile |
| `FE_S25_07_current-character-profile-page` | FE | Play as the current character (frontend) | — | BE_S25_04_current-character-profile, FE_S25_06_character-roster-page |

**GitHub:** story [#492](https://github.com/SETA1609/riftweave-core/issues/492); tasks [`PO_S25_01_file-character-roster-issues` #485](https://github.com/SETA1609/riftweave-core/issues/485), [`GD_S25_02_sex-and-starting-perks` #486](https://github.com/SETA1609/riftweave-core/issues/486), [`BE_S25_03_creation-flow-steps` #487](https://github.com/SETA1609/riftweave-core/issues/487), [`BE_S25_04_current-character-profile` #488](https://github.com/SETA1609/riftweave-core/issues/488), [`FE_S25_05_creation-flow-page` #489](https://github.com/SETA1609/riftweave-core/issues/489), [`FE_S25_06_character-roster-page` #490](https://github.com/SETA1609/riftweave-core/issues/490), [`FE_S25_07_current-character-profile-page` #491](https://github.com/SETA1609/riftweave-core/issues/491).

**Done when:** every issue above is done, and in the web app a player makes a character step by step, finds it in the roster, plays as it, and switches to another character from the header, all offline.
