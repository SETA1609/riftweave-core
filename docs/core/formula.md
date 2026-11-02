# Formula

Author: SETA1609

A formula is a JSON tree. No loops. Integer only. Rounding is named.

```
ap = 2 + floor(AGI / 3)

{ "op": "add", "args": [ { "lit": 2 }, { "op": "div", "args": [ { "attr": "agi" }, { "lit": 3 } ] } ] }
```

Python and Zig implement the same evaluate(). The tree is the authority. The prose is a reading of it.

## Leaves

- `{ "lit": n }` an integer.
- `{ "attr": "agi" }` a current attribute.
- `{ "var": "level" }` a named input the caller supplies, such as `level` or `luck_roll`. A formula that reads a var names it in the data that holds it. [A]
