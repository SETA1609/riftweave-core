# Formula

Author: SETA1609

A formula is a JSON tree. No loops. Integer only. Rounding is named.

```
ap = 2 + floor(AGI / 3)

{ "operation": "add", "arguments": [ { "literal": 2 }, { "operation": "divide", "arguments": [ { "attribute": "agi" }, { "literal": 3 } ] } ] }
```

Python and Zig implement the same evaluate(). The tree is the authority. The prose is a reading of it.

## Leaves

- `{ "literal": n }` an integer.
- `{ "attribute": "agi" }` a current attribute.
- `{ "variable": "level" }` a named input the caller supplies, such as `level` or `luck_roll`. A formula that reads a variable names it in the data that holds it. [A]

## Operations

`{ "operation": <name>, "arguments": [ ... ] }`, with two or more arguments:

- `add` the sum.
- `subtract` the first minus the rest.
- `multiply` the product.
- `divide` the first over the second, floored.

The names are spelled out, as every key is ([CONTRIBUTING.md](../../CONTRIBUTING.md#naming)): the tree used `op`, `args`, `lit`, `attr`, `var`, `sub`, `mul`, and `div` until M5-46. [V 2026-11-12, `formula` vectors in Python and Go]
