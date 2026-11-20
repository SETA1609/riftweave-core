# Ranged builds

Modified: 2026-11-19

Three suggested builds for fighting from a distance. They are suggestions: nothing here is gated, and any of them can draw a blade or learn a spell ([README](README.md)). The rows are in [`ruleset/data/builds/core.json`](../../../ruleset/data/builds/core.json); each is checked from creation to level 30 at the lowest luck roll.

All three lean on the same core: the attack roll and the defense bands of [combat.md](../../core/combat.md) (M3-27), the ranged weapons and their split skills (`bows`, `crossbows`, `guns`, `throwing_weapons`) in [weapons.md](../weapons.md), and the action pool of [actions.md](../../core/actions.md#pool). An attack costs 3 AP; an **aimed** attack adds 2, and **Steady Aim** takes 1 off that. All three take **Gap Finder** at level 2 (ranged attacks crit more often) and **Sniper** at 18 (range penalties halved, PER 8), the two perks the ranged table has for every ranged skill. All three take **Gifted** at creation.

The numbers below are each build's skills at level 30 on the lowest rolls. A real character with average luck rolls, a background, and a race has more.

## Archer

A longbow and light armor: picks targets off from far away, keeps moving, and is gone before the enemy closes.

- **Attributes:** PER 8, AGI 7, STR 5, END 5, INT 5, WIL 4, CHA 4, LCK 4; with Gifted and Small Frame (+1 AGI), PER 9, AGI 9, STR 6, END 6, INT 6.
- **Tag skills:** Bows, Evasion, Light Armor.
- **Mechanics:** a longbow rolls 1d8 and is `heavy` and `two_handed` ([weapons.md](../weapons.md)), so there is no shield: the defense is the evasion band, doubled by the **dodge stance**, and the archer's answer to a foe that closes is to move away (1 AP a move). Stealth gives the opening shot (Backstab, Assassin). Small Frame carries 25% less, so arrows and gear stay light.
- **Level by level:**
  - **2–9, the draw.** Gap Finder, Sidestep, Ever Watchful, Easy Stride, **Steady Aim** at 6, Armored Evasion, Fleet-Footed, Low Guard.
  - **10–18, the hunter.** Night Eyes, Hardened Hide, Ghost and Backstab as Stealth reaches 50, **Riposte** at 14 (Evasion 60), Tracker, two more Hardened Hide, **Sniper** at 18.
  - **19–26, the shadow.** Ready Hand, Streetwise, Haggler, Locksmith, Light Fingers, Wind-Up, **Assassin** at 25 (Stealth 70), Snap Cut.
  - **27–30.** Arcane Reservoir three times and Mailed Casting: the ranged and general tables run out here (see the gaps in the [decision record](../../decisions/M5-61.md)).
- **At level 30:** Bows 100, Evasion 80, Light Armor 70, Stealth 70, Survival 40.
- **Outside the build:** a dagger works at the untrained Blades skill; a crossbow or a gun at its own untrained skill, though Gap Finder and Sniper still apply. A shield cannot be carried with a bow.

## Arbalist

A hand crossbow, a shield, and medium armor: every bolt finds a gap, and the line holds while the next one is wound.

- **Attributes:** PER 8, END 7, STR 6, WIL 5, INT 4, AGI 4, CHA 4, LCK 4; with Gifted and Beast Phobia (+1 STR), PER 9, STR 8, END 8, WIL 6.
- **Tag skills:** Crossbows, Medium Armor, Shields.
- **Mechanics:** the hand crossbow rolls 1d6 and is `light` and `loading` ([weapons.md](../weapons.md)); **Fast Reload** at 8 takes 1 AP off reloading. Being `light`, it leaves a hand for a shield (the two-handed crossbow, 1d8, does not): the carried shield's DR adds to armor DR on every physical hit, and the **block stance** (2 AP, refunded if nobody swings) adds a block band of half the Shields skill. **Spell Ward** and **Bulwark** carry the shield-bearer's block ladder into a ranged build.
- **Level by level:**
  - **2–9, the bolt and the shield.** Gap Finder, Shield Training, Hardened Hide twice, Shield Training tier 2 at 6 (Bulwark needs it), **Steady Aim** at 7, **Fast Reload** at 8, Ever Watchful.
  - **10–18, the wall.** **Spell Ward** at 10, Iron Stomach, Catch Your Breath, Berserker, **Bulwark** at 14 (Shields 70), Pack Mule twice, Night Eyes, **Sniper** at 18.
  - **19–30.** Survivor, Streetwise, Beast Friend, Field Medic, Meditative, Haggler, Wind-Up, Snap Cut, Shield Wall, Intimidating Presence, Mana Flow, Empath: general perks, since the crossbow line ends at Sniper (see the gaps).
- **At level 30:** Crossbows 100, Shields 78, Medium Armor 65, and 30 to 40 in the filler skills.
- **Outside the build:** AGI 5 keeps the action pool at 3 AP (2 + AGI / 3, [actions.md](../../core/actions.md#pool)): one shot a turn, or a block with the unspent AP carried, so the arbalist trades moving for standing behind the shield. A sword works at the untrained Blades skill and keeps the shield.

## Gunslinger

A pistol in each holster and nothing heavier than leather: fires from the hip, slips the return fire, and wins the duel before the dust settles.

- **Attributes:** PER 8, AGI 8, END 5, INT 5, STR 4, WIL 4, CHA 4, LCK 4; with Gifted and Small Frame, AGI 10, PER 9, END 6, INT 6.
- **Tag skills:** Guns, Evasion, Light Armor.
- **Mechanics:** a flintlock pistol rolls 1d10 and is `light` and `loading` ([weapons.md](../weapons.md)). **Hip Shot** at 12 makes guns crit on 5 more, on top of Gap Finder, so the gunslinger crits more than any other ranged build. AGI 10 gives a pool of 5 AP, with the archer's AGI 9 the largest of the three. The defense is the fencer's: the evasion band, the dodge stance, and **Riposte** at 13. The build skips the creation perk Fast Shot, which forbids aimed shots.
- **Level by level:**
  - **2–9, the draw.** Gap Finder, Sidestep, Ready Hand, Low Guard, **Steady Aim** at 6, Armored Evasion, Easy Stride, Fleet-Footed.
  - **10–18, the duel.** Streetwise, Hardened Hide, **Hip Shot** at 12 (Guns 60), **Riposte** at 13 (Evasion 60), two more Hardened Hide, Ever Watchful, Night Eyes, **Sniper** at 18.
  - **19–30.** Haggler, Locksmith, Wind-Up, Snap Cut, Light Fingers, Field Medic, Empath, and Arcane Reservoir, Mailed Casting, and Quick Caster as fillers (see the gaps).
- **At level 30:** Guns 100, Evasion 80, Light Armor 70, and 40 in Intimidation, Deception, Lockpick, Sleight of Hand, Medicine, and Insight.
- **Outside the build:** STR 5 makes a two-handed gun or a heavy weapon a poor fit. A bow or a crossbow works at its untrained skill with Gap Finder and Sniper behind it.
