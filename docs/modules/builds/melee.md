# Melee builds

Modified: 2026-11-19

Three suggested builds for fighting at arm's length. They are suggestions: nothing here is gated, and any of them can pick up a bow or learn a spell ([README](README.md)). The rows are in [`ruleset/data/builds/core.json`](../../../ruleset/data/builds/core.json); each is checked from creation to level 30 at the lowest luck roll.

All three lean on the same core: the attack roll and the defense bands of [combat.md](../../core/combat.md) (M3-27), the weapons of [weapons.md](../weapons.md), and armor DR from [armor.md](../armor.md). All three take **Gifted** (+1 to every attribute, ×0.9 skill points) at creation: a melee build needs several attributes at 6 to 8 for its perks more than it needs points in many skills.

The numbers below are each build's skills at level 30 on the lowest rolls. A real character with average luck rolls, a background, and a race has more.

## Shield-bearer

A blade, a shield, and heavy plate: holds the line, blocks blows outright, and outlasts anything that cannot get past the shield.

- **Attributes:** STR 7, PER 6, END 7, WIL 5, AGI 5, INT 4, CHA 4, LCK 4; with Gifted and Beast Phobia (+1 STR), STR 9, END 8, PER 7, WIL 6.
- **Tag skills:** Blades, Shields, Heavy Armor.
- **Mechanics:** the carried shield's DR adds to armor DR on every physical hit, with no action; the **block stance** (2 AP, refunded if nobody swings) adds a block band of half the Shields skill that stops a hit outright ([combat.md](../../core/combat.md#reading-the-steps)). Heavy armor and a shield stack DR, so low-tier physical damage stops at the armor.
- **Level by level:**
  - **2–6, the shield.** Hardened Hide, then the tier 1 of each shield ladder ([combat.md](../../core/combat.md#block-and-dodge-perks)): Shield Training (+2 shield DR), Shield Wall (+3 block band), Boss Strike (the shield joins the attack as a bash), and **Spell Ward** at 6 (a magic hit in the block band loses END × 2 % of its damage).
  - **7–16, the wall.** Tier 2 of Shield Training, Shield Wall, and Boss Strike at 7 to 9, **Guardian** at 10, **Bulwark** at 12 on Shield Training tier 2, then tier 3 of Shield Training, Shield Wall, and Boss Strike at 13 to 15, and Guardian tier 2 at 16.
  - **17–24, the veteran.** Heavy Armor Expert, Spell Ward tier 2 at 18, Boss Strike tier 4 at 19, Guardian tier 3 at 20, then Full Weight, Carry Through, Hardened Hide, and Spell Ward tier 3 at 24.
  - **25–30.** Unstoppable, Catch Your Breath, Berserker, Survivor, Ever Watchful, Intimidating Presence. The tiered ladders fill the gap the melee perk table had here ([decision record](../../decisions/M5-47.md)), as GD_S05_03 planned.
- **At level 30:** Shields 100, Blades 100, Heavy Armor 75, Intimidation 60, Evasion 50.
- **Outside the build:** the evasion band is narrow in heavy armor and a shield, so area spells hurt; Spell Ward cuts a magic hit in the block band by at most half. Guardian is a pending perk: the kernel has no allies yet. A bow or crossbow works at its untrained skill. Dropping the shield for a two-hander keeps every perk except the block ladder's.

## Two-handed fighter

A maul or a warhammer in both hands and heavy armor: no shield, every blow a big one, a charge that breaks the enemy line, and a presence that makes it waver first.

- **Attributes:** STR 8, END 7, PER 6, WIL 5, INT 4, AGI 4, CHA 4, LCK 4; with Gifted and Beast Phobia, STR 10, END 8, PER 7, WIL 6.
- **Tag skills:** Blunt, Heavy Armor, Intimidation.
- **Mechanics:** the attribute modifier adds STR − 5 to every blow before the crit doubles it ([attributes.md](../attributes.md#the-attribute-modifier)), so STR 10 is +5 a hit; Crushing Blow and Carry Through are the blunt and two-handed perks; Intimidation feeds Intimidating Presence and Menacing. With no shield the defense is DR and the evasion band only.
- **Level by level:**
  - **2–9, the swing.** Hardened Hide, Full Weight (Blades 40 is the one off-tag skill the build buys early), Forward Guard, a second Hardened Hide, Follow Up, **Carry Through** at 7, **Crushing Blow** at 8, the third Hardened Hide.
  - **10–19, the sweep.** **Long Guard** at 10, Brutal Charge, **Reaping Arc** at 12, Catch Your Breath, Heavy Armor Expert, Berserker, Intimidating Presence, Iron Stomach, **Circle Cut** at 18, **Unstoppable** at 19.
  - **20–24, the presence.** **Menacing** once Intimidation is 60, Ever Watchful, Pack Mule twice, **Executioner** at 24.
  - **25–30.** Brawler, Shield Training, Survivor, Night Eyes, Streetwise, Meditative; see the gaps.
- **At level 30:** Blunt 100, Heavy Armor 100, Intimidation 63, Blades 40, and 30 to 40 in the filler skills.
- **Outside the build:** a sword works at Blades 40; a shield works with Shield Training's +2, though without the other shield ladders. Ranged attacks and magic are untrained.

## Fencer

A blade in light armor and no shield: slips the blow instead of taking it, and answers a dodge with a riposte.

- **Attributes:** AGI 8, STR 6, PER 6, END 6, INT 4, WIL 4, CHA 4, LCK 4; with Gifted and Small Frame (+1 AGI), AGI 10, STR 7, PER 7, END 7.
- **Tag skills:** Blades, Evasion, Light Armor.
- **Mechanics:** the evasion band is half the Evasion skill and the **dodge stance** (3 AP, 2 with Sidestep) doubles it, so a fencer at Evasion 100 evades every success with a margin under 100 while dodging; **Riposte** turns an evade in the band's lowest rolls into a free attack. Light armor keeps the band wide ([combat.md](../../core/combat.md#reading-the-steps)). Stealth gives the opening strike (Backstab, Assassin).
- **Level by level:**
  - **2–7, the footwork.** Sidestep, Ever Watchful, Low Guard, Full Weight, Follow Up, **Armored Evasion** at 7 (tier 1, +5 evasion).
  - **8–13, the blade.** Hardened Hide, Easy Stride, Fleet Footed, Carry Through, **Fencer's Measure** at 12 (Blades 60), **Riposte** at 13 (Evasion 60).
  - **14–26, the shadow.** Armored Evasion tier 2, Catch Your Breath, Berserker, Ready Hand, Press On at 18, Armored Evasion tier 3 at 19, Precise Thrust (Piercing 50, a rapier's skill), Backstab and Ghost as Stealth reaches 50, Hardened Hide twice, Night Eyes, **Assassin** at 26 (Stealth 70).
  - **27–30.** Iron Stomach, Streetwise, Pack Mule, Brawler; see the gaps.
- **At level 30:** Evasion 100, Blades 99, Stealth 70, Light Armor 50, Piercing 50.
- **Outside the build:** a fencer has little HP and DR, so an instant spell, which meets no band, and a crit, which always hits, are what kill one. A shield works at the untrained Shields skill. A bow works at its untrained skill; the ranged builds are in [ranged.md](ranged.md).
