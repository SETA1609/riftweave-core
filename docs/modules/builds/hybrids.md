# Hybrid builds

Modified: 2026-11-18

Builds that mix two paths. They are suggestions: nothing here is gated, and a hybrid can lean on either half as the fight needs ([README](README.md)). The rows are in [`ruleset/data/builds/core.json`](../../../ruleset/data/builds/core.json); each is checked from creation to level 30 at the lowest luck roll. The arcane archer or trickster and the ranger follow in M5-61.

A hybrid pays for its reach: its skill points are spread over two paths, so it reaches each top skill later than a pure build, and its perks come from two tables.

## Magic swordsman

A blade in one hand and fire in the other, in light armor: closes to cut, and burns what will not come close. Melee plus magic.

- **Attributes:** STR 7, INT 7, END 6, AGI 6, the rest 4; with Gifted and Iron Allergy (+1 INT), STR 8, INT 9, END 7, AGI 7.
- **Tag skills:** Blades, Red Magic, Light Armor.
- **Spells:** Fire Bolt, Frost Lance, Fire Burst, Flame Shield, Chain Lightning, Meteor.
- **Mechanics:** an attack and a cast cost 3 AP each, and AGI 7 gives a pool of 4 ([actions.md](../../core/actions.md#pool)), so a turn is one or the other plus a move. **Quick Caster** at 18 makes a cast 2 AP: a cast and two moves, or a cast now and the unspent AP carried, so an attack and a cast land in one turn every other turn. Light armor keeps the casting penalty small, and **Battle Mage** halves it. The blade half is the duelist's (Power Attack, Cleave, **Duelist**), the fire half the blaster's (**Pyromancer**, Ward Breaker, **Archmage**).
- **Level by level:**
  - **2–9.** Power Attack (Blades 40), Arcane Reservoir twice, Hardened Hide, **Cleave** at 6, **Pyromancer** at 7 and 9, Easy Stride (Light Armor 50).
  - **10–19.** Sidestep, Hardened Hide, **Battle Mage** at 12, **Duelist** at 13 (Blades 60), Catch Your Breath, Berserker, Ward Breaker, Armored Evasion, **Quick Caster** at 18, the third Hardened Hide.
  - **20–30.** **Executioner** at 24 and **Archmage** at 25; the rest are general perks (Quick Learner, Polyglot, Tinkerer, Streetwise, Iron Stomach, Fleet Footed, Pack Mule, Scholar, Field Medic).
- **At level 30:** Blades 100, Red Magic 100, Light Armor 80, Evasion 50, and 40 in the filler skills.
- **Outside the build:** a shield works at the untrained Shields skill but blocks the free hand a cast wants; a bow works at its untrained skill. Mana runs out before stamina does, so a long fight ends on the blade.
