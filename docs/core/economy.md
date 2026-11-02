# Economy

Buy, sell, and barter as data. Equipment already carries a `cost` (`quantity` and `unit`) and materials a `valueFactor`; this file adds the denominations, the price formula, the merchant rows, and the barter check. The rows are in `ruleset/data/economy/core.json`, the formula in `ruleset/data/formulas/price.json`, the vectors in `ruleset/tests/vectors/economy.json`. Nothing here needs another module loaded: an item with no material is priced at its base cost. [`docs/modules/factions.md`](../modules/factions.md) leaves a standing-based price modifier for later; it would add to the barter shift below, not replace it.

## Currency

One base unit, the copper piece. Every other denomination is a whole multiple of it.

| unit | copper |
| --- | --- |
| `cp` | 1 |
| `sp` | 10 |
| `gp` | 100 |
| `pp` | 1000 |

Prices are integers in copper. There are no fractions of a copper piece: a price rounds down. [A] The ratios are the common fantasy ones, not taken from a Riftweave source. [U]

## Price

```
base_cp  = cost.quantity * value(cost.unit)
price_cp = floor(base_cp * value_percent * percent / 10000)
```

- `value_percent` is the material's `valueFactor` times 100 (steel `2.0` is `200`), or `100` for an item with no material. The caller reads it from the item; the formula does not look materials up. [A]
- `percent` is the merchant's side of the spread, after barter.
- The formula is a tree ([formula.md](formula.md)) with the vars `base_cp`, `value_percent`, `percent`.

## Merchants

A merchant row has a `buy_percent` (what the player pays, as a percent of the item's price), a `sell_percent` (what the merchant pays the player), and a `stock` list of namespaced item refs with a quantity. `buy_percent` is above 100 and `sell_percent` below, so a round trip always loses money. Stock is data; the rules do not restock. [A]

## Barter

One check, the ordinary d100 function ([resolution.md](resolution.md)) on the skill the data names (`persuasion`, governed by CHA). The result is a shift in percentage points:

```
success   shift = min(max_shift, margin / margin_divisor)       floor
failure   shift = 0
fumble    shift = -fumble_penalty
```

A buyer's `percent` is `buy_percent - shift`; a seller's is `sell_percent + shift`. The data keeps `buy_percent - max_shift >= 100` and `sell_percent + max_shift <= 100`, so barter never beats the item's own price and a bad roll can still cost the player. A crit adds nothing beyond the larger margin. [A]

## Not covered

Haggling over several rounds, merchant wealth and restocking, stolen goods, taxes, and standing-based prices. The coin weight of a purse is not counted against carry weight. [U]
