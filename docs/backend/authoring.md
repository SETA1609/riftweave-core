# Authoring

Modified: 2026-11-21

The `author` command (APP-10) builds content rows for authors: it checks a draft the way
`validate.py` would, works out what the formulas say about it, and writes a data file the ruleset
accepts. It is a content tool, not player spellcrafting, so a draft may use effects that are not
`craftable` ([magic.md](../modules/magic.md#spell-knowledge)): an authored spell is a stock spell,
the kind a wondrous item may hold. It adds no rule; it calls the spell schema and the spellcraft
cost the vectors pin. `spell` is its first `kind`; other collections follow as their own issues.

## Input and result

```
input   kind       spell
        operation  check (default), load, or effects
        check:     spells: one or more draft rows, as spells/core.json holds them
                   export: optional { shape: core | module, module: <name> }
        load:      spell: a key in spells/core.json
        effects:   color: one of the nine colors
result  check:     spells: per draft { key, row, problems, warnings, cost }
                   export: the data file, and for a module its manifest and the refs to the rows
        load:      row: the spell as written in spells/core.json, ready to edit or copy
        effects:   effects: per effect { ref, row }, the row as effects/core.json writes it
```

**The effect pool** (`effects`, BE_S09_01) is what a spell editor offers once a color is picked:
every effect on that color or on none ([magic.md](../modules/magic.md#the-core-spell-list)) that
lists `spell` in its `channels`, in file order. A color is one a `<color>_magic` skill governs; an
unknown or missing color is refused. Effects from a module are not listed yet. [A]
[V 2026-11-21, `author_test.go`]

## Problems

A problem is what `validate.py` would refuse. Each names the draft's key and where it is:

| Check | Problem |
| --- | --- |
| Schema | the draft in the `spells` wrapper does not match `spell.schema.json` |
| Refs | a `core:<category>/<key>` names no row (the categories `validate.py` indexes); `module:` refs are not checked, as there |
| Channels | an effect a spell uses does not list `spell` in its `channels` |
| Parameters | an effect with a `parameter_category` has a parameter of another kind ([integrity.md](../core/integrity.md)) |

A warning is not a refusal: an authored `cost.mana` that differs from the spellcraft formula is a
warning with both numbers, since a stock spell may cost what its author says.

## Cost

`cost` is what the spellcraft formula gives for the draft's effects
([casting.md](../core/casting.md#spellcraft)): `mana`, `cast_threshold` (`round(mana)`), and
`craft_threshold` (`round(mana x 1.25)`, a half rounding up). It is null when an effect does not
resolve.

## Export

```
core    { "$schema": "../../schemas/spell.schema.json", "spells": [rows] }
        for ruleset/data/spells/<file>.json; each row is then core:spell/<key>
module  the same file with "$schema": "../../../schemas/spell.schema.json", for
        ruleset/modules/<name>/data/spells.json, and the manifest it needs:
        id <name>, extends.collections ["spells"], data ["data/spells.json"];
        each row is then module:<name>/spell/<key>
```

An export with problems is still written, so the author sees the whole file, and the problems
say what `validate.py` will refuse.

## Where it runs

`author` is a command of the contract (`backend/contract/author.json`), so the agent CLI
(`riftweave run author`), the HTTP API (`POST /v1/commands/author`), MCP, and the browser core give
the same result for the same draft. It needs the ruleset schemas: `core.Open` reads them from
`ruleset/schemas` beside the data, and the browser build embeds them: `scripts/build-wasm.sh`
copies `ruleset/schemas` beside the data, and the entry point opens with `core.NewWithSchemas`
(`openRuleset` in `backend/cmd/riftweave-wasm/open.go`). [V 2026-11-21, `call_test.go`, and the
built `riftweave.wasm` run under Node: `fire_bolt` checks with no problems at mana 6] The spell
editor page is FE_S09_02.
