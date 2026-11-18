# PO_S08_09_file-compendium-creation-form-issues File the compendium, back button, and creation form issues

Date: 2026-11-19

- **Decision:** file a compendium (`BE_S08_10` the `catalog` command and the CLI screen, `FE_S08_11` the tables page), a Back button on every screen (`FE_S08_12`), and character creation as a CV-like form (`GD_S07_07` the creation numbers as data, `BE_S07_08` the `creation` command, `FE_S07_09` the form). Owner request, 2026-11-19.
- **In S07 and S08, not a new story:** S08 already holds the item lookup and the bestiary, which the compendium generalises, and S07 the character pages, which the form feeds. The item lookup and the bestiary now need `catalog`; the build issues need the creation command and the form, so a build fills the same draft a player fills by hand.
- **4 + 10 kept:** the owner asked for "all attributes at 5 and 10 more"; the rule is 4 plus 10 (docs/modules/character-creation.md#2-assign-attributes), and every suggested build is built on it. Asked, the owner kept 4 + 10 (2026-11-19). The form only moves those 10 points, between 4 and 10 before race.
- **Why a command and a GD issue first:** no contract command lists a collection or checks a creation draft, and the creation numbers are prose only [V 2026-11-19, `backend/contract/`, `ruleset/data/`]. The form draws from `creation rules`, so it holds no rule.
- **Back goes to the parent menu** [A], not browser history, so a screen opened by its URL still has a way back; it matches the CLI's Esc (docs/backend/app-layout.md).
- **Open:** whether the compendium's columns live in the backend or in the schemas (`BE_S08_10`); whether the Back card of the menu grid moves into the shared header (`FE_S08_12`); the CLI already goes back with Esc on every screen [U]. `FE_S07_09` logs the export through the logger of `FE_S18_02_frontend-build-environments` once that issue is filed and done, and with `console.info` until then.
