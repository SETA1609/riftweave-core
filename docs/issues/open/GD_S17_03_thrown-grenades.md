#### GD_S17_03_thrown-grenades Grenades: thrown flasks with a fixed area damage
- Status: review
- Priority: medium
- Needs: —
- Modified: 2026-11-23
- Do: Owner request, 2026-11-18. There are no grenades yet, and the elemental bolts and bullets of GD_S17_04 are defined against them: a grenade bursts for a **fixed** damage, an elemental round bursts for the **weapon's** damage. Add the grenade as a consumable.
  - **The throw:** a grenade is thrown with `throwing_weapons` as an `area` attack at a point in range (`rangeNormal` 6, `rangeLong` 18 [A]); every creature in its radius takes the area roll of [combat.md](../../core/combat.md) (delivery decides the dodge). It costs the `attack` AP, not `item` [A], so it competes with a shot.
  - **Fixed damage:** the grenade's own dice and damage type, with no attribute, skill, material, or weapon perk added. Only the thrower perks of GD_S17_09 change it.
  - **Rows** (names our own; damage types from `combat/damage_types.json`) [A]:

    | Grenade | Damage | Radius | On a hit |
    | --- | --- | --- | --- |
    | Fire Flask | 2d6 fire | 1 | `burn` |
    | Frost Flask | 2d6 frost | 1 | `slow` |
    | Thunder Flask | 2d6 thunder | 2 | `stagger` |
    | Acid Flask | 2d6 acid | 1 | `sunder_armor` |
    | Smoke Pot | none | 2 | the area blocks sight for 3 turns (a new effect) |

    Each rider is an effect in the shared pool, with a channel for grenades that `validate.py` checks (`CHANNEL_REQUIREMENTS`), as it does for `ingredient` and `maneuver`.
  - **Toxicity and stock:** a grenade is not drunk, so it has no toxicity. Merchants stock them by level [A]; crafting them belongs to the alchemy module (GD_S15_02).
- Keys: a `grenade` consumable type (or delivery `thrown`); `fire_flask`, `frost_flask`, `thunder_flask`, `acid_flask`, `smoke_pot`; the grenade channel
- Files: ~10 (docs/core/combat.md, docs/modules/equipment.md, ruleset/schemas/equipment.schema.json, ruleset/schemas/effect.schema.json, ruleset/data/equipment/consumables.json, ruleset/data/effects/core.json, ruleset/scripts/validate.py, ruleset/tests/vectors/combat.json, docs/decisions/GD_S17_03_thrown-grenades.md, changes/GD_S17_03_thrown-grenades.md)
- See: docs/core/combat.md (area delivery), ruleset/data/combat/damage_types.json, docs/issues/open/GD_S15_02_alchemy-module.md
- Done when:
  - A vector throws a Fire Flask at a point: two creatures in the radius each take the area roll and 2d6 fire, a third outside takes nothing, and the thrower's skill changes the hit but not the damage.
  - `validate.py` fails on a grenade whose rider lacks the grenade channel.
  - `validate.py` and the tests pass.
