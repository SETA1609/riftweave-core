# Magic builds

Modified: 2026-11-19

Three suggested builds for casters. They are suggestions: nothing here is gated, and any of them can pick up a blade or a bow ([README](README.md)). A blaster can still stab with a dagger at its untrained skill, and a summoner can still cast a bolt. The rows are in [`ruleset/data/builds/core.json`](../../../ruleset/data/builds/core.json); each is checked from creation to level 30 at the lowest luck roll.

All three lean on the same core: the color schools and their `<color>_magic` skills, casting resolution, and mana ([magic.md](../magic.md)); a cast is the `cast` action's 3 AP plus the spell's mana ([actions.md](../../core/actions.md)). All three take **Arcane Reservoir** three times (+10 MP a rank) and **Mana Flow** early, because a caster with no mana is a person in robes. Each lists the spells it leans on (`spells`); every spell is a row of [`spells/core.json`](../../../ruleset/data/spells/core.json) in a color the build raises.

The numbers below are each build's skills at level 30 on the lowest rolls. A real character with average luck rolls, a background, and a race has more. Past level 20 the magic perk table runs thin, so all three take general perks there; the gaps are in the [decision record](../../decisions/M5-60.md).

## Blaster

Fire and force from range in light clothes: burns a crowd before it closes, and keeps out of reach of what survives.

- **Attributes:** INT 8, WIL 7, END 6, PER 5, the rest 4; with Gifted and Iron Allergy (+1 INT), INT 10, WIL 8, END 7, PER 6.
- **Tag skills:** Red Magic, Indigo Magic, Evasion.
- **Spells:** Fire Bolt, Frost Lance, Thunderclap, Fire Burst, Force Push, Chain Lightning, Ice Storm, Meteor, Sonic Boom, Inferno.
- **Mechanics:** red magic carries the damage spells, indigo the force and thunder ones that push and knock back. A `single` or `area` spell meets the target's evasion and block bands; an area spell's evaded targets still take half ([combat.md](../../core/combat.md)). Evasion is the blaster's only defense, so it is a tag skill.
- **Level by level:**
  - **2–9, the mana.** Arcane Reservoir three times, Hardened Hide, **Pyromancer** at 5 (Red Magic 50) and its second rank at 9, Mana Flow, **Telekinetic** (+50% Indigo range).
  - **10–18, the craft.** Meditative, Hardened Hide, **Ward Breaker** at 12 (ignores 5 resist), Mailed Casting (×0.75 on the armor and shield penalties, for a hauberk if it wants one), Warded Mind, the third Hardened Hide, Ever Watchful, Scholar, **Elemental Harmony** at 18.
  - **19–30.** **Archmage** at 24 (spells cost 20% less); the rest are general perks (Trap Sense, Quick Learner, Tinkerer, Polyglot, Tracker, Beast Friend, Survivor, Unshaken, Streetwise, Empath, Iron Stomach).
- **At level 30:** Red Magic 100, Indigo Magic 100, Evasion 100, Lore 60, and 40 in the filler skills.
- **Outside the build:** HP and DR are low, so a blow that gets through hurts; a dagger works at its untrained Blades or Piercing. Instant spells from enemies meet no band, so a save is the only answer to them.

## Summoner

Fights through what it calls: a familiar first, then beasts, elementals, and the dead, each stronger as the ladder climbs, while the caster stays behind them.

- **Attributes:** CHA 7, INT 7, WIL 6, END 6, the rest 4; with Gifted and Iron Allergy, CHA 8, INT 9, WIL 7, END 7.
- **Tag skills:** Orange Magic, Black Magic, Evasion.
- **Spells:** Summon Familiar, Bound Blade, Drain Life, Conjured Barrier, Raise Thrall, Conjured Blades, Summon Elemental.
- **Mechanics:** the summon groups and their templates ([magic.md](../magic.md#summon-groups), [monsters.md](../monsters.md#summoned-creatures)). A power perk (Dire, Elder, Ancient) replaces the one before it; a nature perk (Demonic) adds its template; Retinue and Host raise how many summons stay at once. CHA feeds both **Commanding Summons** lines, which add the CHA modifier to what is called.
- **Level by level:**
  - **2–6, the pool.** Arcane Reservoir three times, Hardened Hide, Mana Flow.
  - **7–17, the ladder.** Commanding Summons (magic) twice and (physical) twice, Conjurer at 8 (+50% duration), **Summon Dire** at 9, **Summon Demonic** at 13, Necromancer (one more thrall), Leader, **Summon Retinue** at 16 (two summons), **Summon Elder** at 17.
  - **18–30.** **Summon Ancient** at 24 and **Summon Host** at 28 (three summons, Orange Magic 90); between them general perks (Warded Mind, Tinkerer, Hardened Hide, Quick Learner, Polyglot, Survivor, Beast Friend, Scholar, Unshaken, Meditative, Streetwise).
- **At level 30:** Orange Magic 100, Black Magic 100, Evasion 80, and 40 in the filler skills.
- **Outside the build:** the caster itself is fragile; Bound Blade conjures it a weapon, and Drain Life is a black-magic bolt that heals it. A leader's +5 to allies' attack helps its summons too.

## Support

Heals, wards, and lifts the party: keeps the front line standing with green and white magic, and leads it with a song.

- **Attributes:** CHA 8, WIL 7, INT 6, PER 5, the rest 4; with Gifted and Undead Phobia (+2 WIL), WIL 10, CHA 9, INT 7, PER 6.
- **Tag skills:** Green Magic, White Magic, Persuasion.
- **Spells:** Mend Wounds, Holy Light, Blessing, Cure Poison, Regeneration, Holy Ward, Cleanse, Greater Heal, Mass Heal, Renewal.
- **Mechanics:** green magic restores, white magic wards and smites; **Healer** and **Zealot** add 15% a rank to each. **Leader** (+5 to allies' attack) opens the bard line: **Bardic Inspiration** spends the `inspire` action (2 AP) to add half the CHA modifier to allies, and **Bardic Anthem** the whole modifier at twice the range.
- **Level by level:**
  - **2–7, the pool.** Arcane Reservoir twice, Hardened Hide twice, Mana Flow, Field Medic (Medicine 40).
  - **8–17, the hands.** **Healer** at 8 and 11, **Zealot** at 9 and 12, Empath, Meditative, **Leader** at 14, Unshaken, Warded Mind, the third Arcane Reservoir.
  - **18–30, the voice.** **Bardic Inspiration** at 18 and 19, then general perks (Hardened Hide, Ever Watchful, Quick Learner, Trap Sense, Polyglot, Streetwise, Tinkerer, Haggler), **Bardic Anthem** at 28 (Persuasion 60, CHA 8), Tracker, Beast Friend.
- **At level 30:** Green Magic 100, White Magic 100, Persuasion 80, Medicine 60, and 40 in most filler skills.
- **Outside the build:** Smite and Holy Light are white-magic damage, so a support is not helpless alone; a mace works at the untrained Blunt. END stays at 5, so it has the least HP of the three.
