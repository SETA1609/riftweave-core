# Modifier

Author: SETA1609

One payload. Conditions, features, and applied effects emit this row. They do not invent a second shape.

```
target     skill, attribute, ap, dr, a flag, a damage tag
op         add | mul | set | grant | resist
value      number
source     perk:quick, condition:blinded
duration   turns | uses | while_equipped | while_applied | permanent
turns      how many turns a `turns` modifier lasts
stack      none | add
```

This is the departure from 3.5, where typed bonuses stack by name. Here the source is the type. The stacking rule is written once, in [contract.md](contract.md); this file fixes its order so a second implementation gets the same number. The `modifiers` and `conditions` vectors pin every step. [V 2026-11-04]

## Order for one numeric target

```
1. expire     drop a modifier whose source was removed,
              and a `turns` modifier whose turns have elapsed (elapsed >= turns)
2. dedupe     same source, same target, same op: the latest in list order wins
3. add        value = value + sum(add)
4. mul        value = floor(value * product(mul))
5. set        if any set: value = the latest set in list order
```

- **Same source does not stack.** Two `add` from `perk:tough` on `hp` count once; the later one wins, even if it is smaller. A source that touches two targets, or uses two ops on one target, keeps each: the key is source, target, and op. [V 2026-11-04]
- **Different sources add**, and two `mul` from different sources multiply. [V 2026-11-04]
- **`mul` after `add`**, whatever the list order. The product is floored, so `mul 1.5` on 5 is 7. [A] Numbers in the snapshot are integers; the floor keeps them so.
- **`set` is last.** It replaces the value whatever the `add` and `mul`, so a paralyzed actor with an AP bonus still has 0 AP. Two `set` on one target: the latest in list order (load order) wins. [A] The contract says `set` replaces the value "for that source's duration"; expiry (step 1) is what ends it.
- **`grant`** adds the flag named by `target`. It adds no number, and its `value` is ignored. A flag stays while any live source grants it: removing one of two sources that grant `prone` leaves the actor prone. [V 2026-11-04]
- **`resist`** subtracts from a packet whose tag equals `target`: `amount = max(amount - sum(resist), 0)`. Same-source and different-source rules are the ones above. It is the same op the combat pipeline's resist step reads ([combat.md](combat.md)). [V 2026-11-04]
- **Expiry.** `permanent` never expires. `while_applied` and `while_equipped` end when their source is removed. `turns` ends when the elapsed turns reach `turns`. `uses` is counted by the caller, who removes the source when the uses run out; the vectors do not cover it. [A]

## Conditions

A condition is a row in `ruleset/data/conditions/core.json`: a named bundle of modifiers, all with source `condition:<key>`, plus `appliedBy` and `removedBy` effect lists.

1. **An effect event** applies every condition whose `appliedBy` lists it, and removes every condition whose `removedBy` lists it. One effect may do several: `blight` applies `exhaustion` and `diseased`; `cure` removes every condition that lists it.
2. **Apply** adds the condition, or one stack if it is already active and has `maxStacks`. Stacks never pass `maxStacks`. A condition without `maxStacks` applied twice is still one condition: same source does not stack. [V 2026-11-04]
3. **Remove** drops the condition and every modifier with its source, all stacks at once. [A]
4. **A remove event** (`remove: <key>`) ends a condition with no removing effect: `prone` ends when the actor stands up, an action the caller resolves. [A]
5. The actor's values are then the base values with every live condition's modifiers, in the order above.

The `combat` and `effects` fields on a condition row (defense bonus, TTRPG and video-game prose) are not modifiers yet; they are read by nobody and are left to the issues that turn them into rows (M1-01, M3-03). [U]
