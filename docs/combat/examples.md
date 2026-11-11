# Combat examples

The worked examples live in [`docs/modules/combat.md`](../modules/combat.md#9-worked-examples). Every step in them is a vector in `ruleset/tests/vectors/combat.json`, so CI checks the numbers on every pull request.

| Example | Vectors | Shows |
| --- | --- | --- |
| 1. Knight against bandit | `example 1a`, `example 1b`, `example 1c` | a hit through a failed evade, a block that leaves the rest to armor, a crit doubling before DR |
| 2. Mage against fire | `example 2a`, `example 2b`, `example 2c` | a single-target spell not dodged, fire resist, an area spell halved by one evade and taken in full by another |

To check one by hand, run its vector through the reference:

```
python -m unittest ruleset/scripts/test_vectors.py
```

## Data shapes

### A weapon with an on-hit effect

```json
{
  "id": 100, "key": "flaming_longsword", "label": "Flaming Longsword",
  "type": "weapon", "cost": { "quantity": 500, "unit": "gp" }, "weight": 3,
  "weapon": {
    "category": "martial", "range": "melee", "length": "normal",
    "attackReach": 1, "skill": "blades",
    "damage": { "dice": "1d8", "type": "slashing" },
    "properties": ["versatile", "magical"],
    "on_hit_effects": [
      { "effect": "core:effect/burn", "magnitude": 3, "duration": 10 }
    ]
  }
}
```

On a hit it deals 1d8 slashing through the pipeline and applies `burn`, which applies the `burning` condition.

### Armor with resistances

```json
{
  "id": 101, "key": "asbestos_plate", "label": "Asbestos-Lined Plate",
  "type": "armor", "cost": { "quantity": 1500, "unit": "gp" }, "weight": 30,
  "armor": {
    "slot": "torso", "layer": "upper", "category": "heavy",
    "drBase": 5, "evasionPenalty": -15, "speedPenalty": -2,
    "strengthRequirement": 8,
    "resistances": { "fire": 8, "frost": -4 }
  }
}
```

It resists 8 fire and takes 4 more frost. In the pipeline these are the resist step's values for those damage types ([`docs/core/combat.md`](../core/combat.md)).

## See also

- [`docs/core/combat.md`](../core/combat.md): the order of one attack
- [`docs/modules/combat.md`](../modules/combat.md): commentary and the worked examples
- [`docs/combat/integration.md`](integration.md): how effects, perks, weapons, and armor compose
