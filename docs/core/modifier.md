# Modifier

Author: SETA1609

One payload. Conditions, features, and applied effects emit this row. They do not invent a second shape.

```
target   skill, attribute, ap, dr
op       add | mul | set | grant | resist
value    number
source   perk:quick, condition:blinded
stack    none | add
```

Order is add, then mul. Two `set` on one target: last in load order wins. Same source does not stack. Different sources add. This is the departure from 3.5, where typed bonuses stack by name. Here the source is the type.

`resist` is a modifier on a damage type, not a second table.
