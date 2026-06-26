# Combat Examples — Worked Resolutions

**Status:** Comprehensive worked examples live in `docs/modules/combat.md` §10 (Worked Combat Examples).
This file is a quick-reference index and supplement for the data-oriented examples.

---

## Existing Examples (in docs/modules/combat.md)

The main combat document contains 5 fully worked examples:

| Example | Title | Type | Focus |
|---------|-------|------|-------|
| 1 | Knight vs. Bandit | TTRPG turn-based | Basic attack, block, OA, DR |
| 2 | Critical Hit with Effects Table | TTRPG turn-based | Confirmation, double damage, disarm |
| 3 | cRPG Party Combat | cRPG tactical | 3 PCs vs 3 goblins, called shots, positioning |
| 4 | Rogue vs. Guard Captain | Video game action-combat | Parry, poison, armor bypass, conditions |
| 5 | Magic Defense Tree | TTRPG | Spell absorption, resistance, elemental phases |

---

## Simulated Examples (Reference Script)

Run `python ruleset/scripts/combat_reference.py --verbose` for randomly-rolled data-driven examples.
These use real data from the JSON files and produce output like:

```
============================================================
  Example 1: Sir Aldric (Blades 72, LCK 4) vs Bandit (DR 2)
============================================================
  Weapon: Longsword
  Skill: blades = 72
  Target number: 72 + 0 (mod) + 2 (LCK/2) = 74
  Defender armor DR: 2
  d100 roll: 47, margin: 27
  Hit quality: solid (margin 27)
  Damage roll: 1d8 = 6
  Damage type: slashing (armor applies: True)
  Raw: 6 - DR: 2 = Net: 4

  Result: HIT
  Quality: SOLID
  Damage: 4 (slashing)
============================================================
```

---

## Data Integration Examples

### Weapon with On-Hit Effect (Flaming Longsword)

A magical longsword that sets targets ablaze would look like:

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
      { "effect": 57, "magnitude": 3, "duration": 10 }
    ]
  }
}
```

On a successful hit, this weapon deals its 1d8 slashing damage AND applies effect id 57 (burn) with magnitude 3 and duration 10 seconds.

### Armor with Resistances (Asbestos-Lined Plate)

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

This piece provides +8 DR vs fire damage but −4 DR vs frost (vulnerability to cold).

---

## See Also

- `docs/modules/combat.md` — Full combat rules with TTRPG + video game dual resolution
- `docs/modules/combat.md#10-worked-combat-examples` — The 5 main worked examples
- `docs/combat/integration.md` — How effects, perks, weapons, and armor compose
- `ruleset/scripts/combat_reference.py` — Run with `--seed N` for reproducible examples
