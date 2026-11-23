# BE_S08_10_compendium-catalog-cli A compendium of every table to study a build: the catalog command and the CLI screen (backend)

Date: 2026-11-21

- **Decision:** one `catalog` command with three operations (`collections`, `list`, `get`) over fifteen collections; the rules for it are in [api.md](../backend/api.md#catalog), the screen in [app-layout.md](../backend/app-layout.md#compendium).
- **Columns live in the backend**, in the `collections` table of `backend/internal/catalog/catalog.go`, not in the schemas: they are display, the frontend reads them from `collections`, and a schema change would touch every data schema for no rule.
- **Composition is generic:** base, then `data.MergeOverride` of the overrides, then the row's own fields, with the identity fields a row built on a row of its collection does not inherit (a spell's key, label, description, and source, as `readSpells`). A spell's overrides merge deeply here; `readSpells` merges one level. They agree on today's data, where no override nests deeper. [A]
- **Monsters are not fully composed:** validate.py also scales a monster's attributes and damage by its level (growth) and applies its templates. Porting that is a rule, not a read, so the catalog shows the merge only. BE_S08_05 (the bestiary) is where the full composition belongs. [A]
- **Features, not perks:** the collection is named for its table (`features`), labeled "Perks and traits", because the table also holds creation and racial traits.
- **Module refs:** an added row is `module:<id>/<category>/<key>`; an overriding row keeps the core ref, so a reference to it still resolves, and its `source` names the module.
- **The web entry is pending:** `menu.json` lists Compendium with `issue_apps: ["web"]` until FE_S08_11.
- **Risk:** the frontend's `i18n.spec.ts` (FE_S11_01) checks that every menu key has a locale key. Whichever of the two lands second adds `menu.compendium` to `en.json`, `de.json`, and `es.json`.
