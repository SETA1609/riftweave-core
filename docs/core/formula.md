# Formula

Author: SETA1609

A formula is a JSON tree. No loops. Integer only. Rounding is named.

```
ap = 2 + floor(AGI / 3)

{ "op": "add", "args": [ { "lit": 2 }, { "op": "div", "args": [ { "attr": "agi" }, { "lit": 3 } ] } ] }
```

Python and Zig implement the same evaluate(). The tree is the authority. The prose is a reading of it.
