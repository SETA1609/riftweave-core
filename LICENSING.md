# Licensing

Riftweave is **dual-licensed** so that both software developers and game/tabletop
creators can use it freely:

| Material | License | File |
| --- | --- | --- |
| Software & tooling | Apache License 2.0 | [`LICENSE`](LICENSE) |
| Game content & documentation | Creative Commons Attribution 4.0 | [`LICENSE-CONTENT`](LICENSE-CONTENT) |

This is the same split model used by many "code + open content" projects (and the
same content license, CC BY 4.0, that the D&D 5.1 SRD uses).

## What falls under which license

**Apache-2.0 (software & tooling):**

- `ruleset/scripts/` — the validation script and any tooling
- `ruleset/schemas/` — the JSON Schema definitions (functional structures)
- `Dockerfile`, `.dockerignore`
- `.github/` — CI workflows and configuration

**CC BY 4.0 (content & documentation):**

- `ruleset/data/` — all game data (attributes, skills, perks, effects, spells,
  items, ancestries, …)
- `docs/` — design and module documentation
- Prose in `README.md`, `CONTRIBUTING.md`, and similar

If a file's classification is ever unclear, treat **machine-executed logic as
Apache-2.0** and **human-readable rules/content as CC BY 4.0**.

## A note on game mechanics

Copyright does not protect game *mechanics* — ideas, systems, and methods of play
are not copyrightable, only their specific expression (text, names, data, and code).
These licenses cover Riftweave's expression; anyone is free to implement the
underlying system independently.

## Trademark

These licenses grant rights to the **content and code**, not to the project's
identity. The "Riftweave" name and any associated logos are not licensed for use in
a way that implies endorsement or that could cause confusion about the source of a
derivative work. Attribution as described in `LICENSE-CONTENT` is encouraged and is
not a trademark use.

## Contributing

Unless you state otherwise, contributions you submit are provided under these same
licenses (Apache-2.0 for code, CC BY 4.0 for content) — inbound matches outbound.
