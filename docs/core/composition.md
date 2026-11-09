# Composition

Many rows are built from parts instead of written out: a monster from a creature type and
templates, a weapon from a base and a material's overrides. There is one way to do it. Any new
composed row (a bound weapon, a summon with its power and nature) uses this order and these
merge rules, not a procedure of its own.

```
row = compose(base, layers, overrides)
1  start     a copy of the base
2  layers    each layer changes the copy, in order (a template)
3  overrides each field of overrides lands on top
4  identity  key, id, and source come from the row itself, never from a part
```

## How an override lands

| Value | Result |
| --- | --- |
| A plain value, or a list | Replaces the field |
| `{ "$extend": [...] }` on a list | Appends to the field |
| `attributes` | Merged per attribute; attributes it does not name keep their value |

A base is what the row starts from: a creature type, another monster, or an equipment base. A
layer knows only how to change a row, so adding a kind of layer adds no merge rule. Composition
merges into one flat row; there is no inheritance at run time, and a consumer reads the flat
row. `ruleset/scripts/validate.py` holds the function, `compose_row`. [V 2026-11-10]

The rules specific to each kind of row stay with it: monsters in
[monsters.md](../modules/monsters.md), and module overrides of a whole row in
[overrides.md](overrides.md).
