# Melee builds

Modified: 2026-11-18

Three suggested builds for fighting at arm's length. They are suggestions: nothing here is gated, and any of them can pick up a bow or learn a spell ([README](README.md)). The rows are in [`ruleset/data/builds/core.json`](../../../ruleset/data/builds/core.json); each is checked from creation to level 30 at the lowest luck roll.

All three lean on the same core: the attack roll and the defense bands of [combat.md](../../core/combat.md) (M3-27), the weapons of [weapons.md](../weapons.md), and armor DR from [armor.md](../armor.md). All three take **Gifted** (+1 to every attribute, ×0.9 skill points) at creation: a melee build needs several attributes at 6 to 8 for its perks more than it needs points in many skills.

The numbers below are each build's skills at level 30 on the lowest rolls. A real character with average luck rolls, a background, and a race has more.

## Shield-bearer

A blade, a shield, and heavy plate: holds the line, blocks blows outright, and outlasts anything that cannot get past the shield.

- **Attributes:** STR 7, PER 6, END 7, WIL 5, AGI 5, INT 4, CHA 4, LCK 4; with Gifted and Beast Phobia (+1 STR), STR 9, END 8, PER 7, WIL 6.
- **Tag skills:** Blades, Shields, Heavy Armor.
- **Mechanics:** the carried shield's DR adds to armor DR on every physical hit, with no action; the **block stance** (2 AP, refunded if nobody swings) adds a block band of half the Shields skill that stops a hit outright ([combat.md](../../core/combat.md#reading-the-steps)). Heavy armor and a shield stack DR, so low-tier physical damage stops at the armor.
- **Level by level:**
  - **2–7, the shield.** Hardened Hide, then the block ladder as Shields reaches 30 and 50: Shield Training (+2 shield DR), Shield Wall, Spell Ward (the block band also stops magic, at half width) at 6, and Power Attack once Blades is 40.
  - **8–14, the wall.** Cleave, Brutal Charge, the third Hardened Hide, **Bulwark** at 12 (a block in the band's lowest rolls staggers the attacker), Catch Your Breath, and Heavy Armor Expert once Heavy Armor is 60.
  - **15–24, the veteran.** Berserker, Pack Mule, Iron Stomach, **Unstoppable** at 18, Survivor, then Executioner at 24 on Power Attack.
  - **25–30.** Fillers: the melee perk table runs out here (see the gaps in the [decision record](../../decisions/M5-47.md)).
- **At level 30:** Shields 100, Blades 100, Heavy Armor 75, Intimidation 60, Evasion 50.
- **Outside the build:** the evasion band is narrow in heavy armor and a shield, so area spells hurt; there is no answer to magic beyond Spell Ward's half band. A bow or crossbow works at its untrained skill. Dropping the shield for a two-hander keeps every perk except the block ladder's.

## Two-handed fighter

A maul or a warhammer in both hands and heavy armor: no shield, every blow a big one, a charge that breaks the enemy line, and a presence that makes it waver first.

- **Attributes:** STR 8, END 7, PER 6, WIL 5, INT 4, AGI 4, CHA 4, LCK 4; with Gifted and Beast Phobia, STR 10, END 8, PER 7, WIL 6.
- **Tag skills:** Blunt, Heavy Armor, Intimidation.
- **Mechanics:** the attribute modifier adds STR − 5 to every blow before the crit doubles it ([attributes.md](../attributes.md#the-attribute-modifier)), so STR 10 is +5 a hit; Crushing Blow and Cleave are the blunt and two-handed perks; Intimidation feeds Intimidating Presence and Menacing. With no shield the defense is DR and the evasion band only.
- **Level by level:**
  - **2–8, the swing.** Hardened Hide, Power Attack (Blades 40 is the one off-tag skill the build buys early), Ever Watchful, a second Hardened Hide, **Crushing Blow** and **Cleave** at 6 and 7, Brutal Charge.
  - **9–18, the armor.** The third Hardened Hide, Pack Mule twice, Berserker and Catch Your Breath at 12, Heavy Armor Expert, Intimidating Presence, Iron Stomach, Night Eyes, **Unstoppable** at 18.
  - **19–24, the presence.** Streetwise, **Menacing** once Intimidation is 60, Meditative, Brawler twice, **Executioner** at 24.
  - **25–30.** Fillers from survival and social perks (Tracker, Shield Training, Survivor, Haggler, Empath, Mana Flow); see the gaps.
- **At level 30:** Blunt 100, Heavy Armor 100, Intimidation 63, Blades 40, and 30 to 40 in the filler skills.
- **Outside the build:** a sword works at Blades 40; a shield works at Shields 30 with Shield Training's +2, though without the block ladder. Ranged attacks and magic are untrained.

## Duelist

A blade in light armor and no shield: slips the blow instead of taking it, and answers a dodge with a riposte.

- **Attributes:** AGI 8, STR 6, PER 6, END 6, INT 4, WIL 4, CHA 4, LCK 4; with Gifted and Small Frame (+1 AGI), AGI 10, STR 7, PER 7, END 7.
- **Tag skills:** Blades, Evasion, Light Armor.
- **Mechanics:** the evasion band is half the Evasion skill and the **dodge stance** (3 AP, 2 with Sidestep) doubles it, so a duelist at Evasion 100 evades every success with a margin under 100 while dodging; **Riposte** turns an evade in the band's lowest rolls into a free attack. Light armor keeps the band wide ([combat.md](../../core/combat.md#reading-the-steps)). Stealth gives the opening strike (Backstab, Assassin).
- **Level by level:**
  - **2–7, the footwork.** Sidestep, Ever Watchful, Ready Hand, Acrobat, Power Attack, **Armored Evasion** at 7.
  - **8–13, the blade.** Hardened Hide, Easy Stride, Fleet Footed, Cleave, **Duelist** at 12 (Blades 60), **Riposte** at 13 (Evasion 60).
  - **14–26, the shadow.** Catch Your Breath, Berserker, Hardened Hide, Night Eyes, Ghost and Backstab as Stealth reaches 50, Precise Thrust (Piercing 50, a rapier's skill), the third Hardened Hide, Iron Stomach, Streetwise, Brawler twice, **Assassin** at 26 (Stealth 70).
  - **27–30.** Pack Mule twice, Locksmith, Light Fingers; see the gaps.
- **At level 30:** Evasion 100, Blades 99, Stealth 70, Light Armor 50, Piercing 50.
- **Outside the build:** a duelist has little HP and DR, so an instant spell, which meets no band, and a crit, which always hits, are what kill one. A shield works at the untrained Shields skill. A bow works at its untrained skill; the ranged builds are in `ranged.md` (M5-61).
