# Hybrid builds

Modified: 2026-11-18

Builds that mix two paths. They are suggestions: nothing here is gated, and a hybrid can lean on either half as the fight needs ([README](README.md)). The rows are in [`ruleset/data/builds/core.json`](../../../ruleset/data/builds/core.json); each is checked from creation to level 30 at the lowest luck roll.

A hybrid pays for its reach: its skill points are spread over two paths, so it reaches each top skill later than a pure build, and its perks come from two tables.

## Magic swordsman

A blade in one hand and fire in the other, in light armor: closes to cut, and burns what will not come close. Melee plus magic.

- **Attributes:** STR 7, INT 7, END 6, AGI 6, the rest 4; with Gifted and Iron Allergy (+1 INT), STR 8, INT 9, END 7, AGI 7.
- **Tag skills:** Blades, Red Magic, Light Armor.
- **Spells:** Fire Bolt, Frost Lance, Fire Burst, Flame Shield, Chain Lightning, Meteor.
- **Mechanics:** an attack and a cast cost 3 AP each, and AGI 7 gives a pool of 4 ([actions.md](../../core/actions.md#pool)), so a turn is one or the other plus a move. **Quick Caster** at 18 makes a cast 2 AP: a cast and two moves, or a cast now and the unspent AP carried, so an attack and a cast land in one turn every other turn. Light armor keeps the casting penalty small, and **Battle Mage** halves it. The blade half is the duelist's (Full Weight, Carry Through, **Duelist**), the fire half the blaster's (**Pyromancer**, Ward Breaker, **Archmage**).
- **Level by level:**
  - **2–9.** Full Weight (Blades 40), Arcane Reservoir twice, Hardened Hide, **Carry Through** at 6, **Pyromancer** at 7 and 9, Easy Stride (Light Armor 50).
  - **10–19.** Sidestep, Hardened Hide, **Battle Mage** at 12, **Duelist** at 13 (Blades 60), Catch Your Breath, Berserker, Ward Breaker, Armored Evasion, **Quick Caster** at 18, the third Hardened Hide.
  - **20–30.** **Executioner** at 24 and **Archmage** at 25; the rest are general perks (Quick Learner, Polyglot, Tinkerer, Streetwise, Iron Stomach, Fleet Footed, Pack Mule, Scholar, Field Medic).
- **At level 30:** Blades 100, Red Magic 100, Light Armor 80, Evasion 50, and 40 in the filler skills.
- **Outside the build:** a shield works at the untrained Shields skill but blocks the free hand a cast wants; a bow works at its untrained skill. Mana runs out before stamina does, so a long fight ends on the blade.

## Arcane trickster

A shortbow, yellow magic, and the shadows: blinds and blurs, charms the guard, and shoots whoever is still looking. Ranged plus magic.

- **Attributes:** AGI 8, PER 7, INT 7, the rest 4; with Gifted and Iron Allergy (+1 INT), AGI 9, INT 9, PER 8.
- **Tag skills:** Bows, Yellow Magic, Stealth.
- **Spells:** Light, Blinding Flash, Blur, Charm Person, Veil, Psychic Scream.
- **Mechanics:** a cast and an attack cost 3 AP each, and AGI 9 gives a pool of 5 ([actions.md](../../core/actions.md#pool)); **Quick Caster** at 18 makes a cast 2 AP, so a cast and a plain shot fit in one turn. **Steady Aim** makes an aimed shot 4 AP. The opening is a spell (Blinding Flash, Blur) or the dark (Ghost, Backstab, **Assassin**), then the bow. Iron Allergy suits a bow of wood or elven wood. The bow half is the archer's (Gap Finder, Steady Aim), the yellow half the illusionist's (**Illusionist**, Ward Breaker).
- **Level by level:**
  - **2–10.** Gap Finder, Arcane Reservoir three times, Ever Watchful, **Steady Aim** at 7, **Illusionist** at 8, Ghost, Backstab.
  - **11–19.** Sidestep, Ward Breaker, Fleet-Footed, Hardened Hide three times, Armored Evasion, **Quick Caster** at 18, **Assassin** at 19 (Stealth 70).
  - **20–30.** General perks: Ready Hand, Night Eyes, Streetwise, Locksmith, Quick Learner, Light Fingers, Polyglot, Field Medic, Tinkerer, Scholar, Empath.
- **At level 30:** Bows 100, Yellow Magic 100, Stealth 70, Evasion 50, and 40 in the filler skills.
- **Outside the build:** END 5 and light gear mean little HP and DR, so a trickster caught in the open is in trouble; the answer is Blur and the dodge stance. A dagger works at the untrained Blades skill.

## Ranger

A longbow, a blade at the hip, and the wild for a home: tracks the quarry, shoots it at range, and finishes it at close quarters. Melee plus ranged.

- **Attributes:** PER 7, STR 6, END 6, AGI 6, WIL 5, the rest 4; with Gifted and Leather Allergy (+1 END), PER 8, END 8, STR 7, AGI 7, WIL 6.
- **Tag skills:** Bows, Blades, Survival.
- **Mechanics:** the bow and the blade cost 3 AP an attack each, so the ranger switches by the range of the fight, not by the turn. Leather Allergy rules out most light armor, so the ranger wears medium armor at a low skill. The bow half is the archer's (Gap Finder, Steady Aim, **Sniper**), the blade half the duelist's (Full Weight, Carry Through, **Duelist**), and Survival feeds Tracker and Beast Friend.
- **Level by level:**
  - **2–10.** Gap Finder, Hardened Hide, Full Weight, **Tracker** at 5, **Steady Aim** at 6, Night Eyes, **Carry Through** at 8, two more Hardened Hide.
  - **11–20.** Beast Friend, Ever Watchful, Berserker, Catch Your Breath, Iron Stomach, Sidestep, Ready Hand, **Sniper** at 18, **Duelist** at 19 (Blades 60), Survivor.
  - **21–30.** General perks: Pack Mule twice, Fleet-Footed, Streetwise, Meditative, Field Medic, Wind-Up, Snap Cut, Empath, Acrobat.
- **At level 30:** Bows 89, Survival 70, Blades 60, and 40 in Animal Handling, Medium Armor, Evasion, Medicine, and Insight. Two top skills share the points, so neither reaches 100.
- **Outside the build:** a shield works at the untrained Shields skill with the blade, but not with the bow. Magic is untrained.
