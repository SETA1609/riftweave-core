# Magic & Spell Crafting

Modified: 2026-11-19

**Status:** Core (implemented as data + schema) · runtime resolution lives in the engine

This document defines how magic works in Riftweave: how spells are built from
effects, how casting resolves, how players craft their own spells, and how the
**five-phase elemental layer** makes effects interact with each other, with
monsters, and with races. For the character-progression view (skills, point-buy,
attributes) see [`progression.md`](./progression.md). For *permanent magic items*
(enchanting), see [`magic-crafting.md`](./magic-crafting.md) — that is a separate
module.

> **Naming rule.** Engine vocabulary is **color-named** (`red_magic`, `blue_magic`,
> …) and **English-named** for phases (`wood`, `fire`, `earth`, `metal`, `water`).
> Do not introduce school names like "destruction" or non-English phase labels in
> ids or keys.

---

## 1. Two orthogonal axes

Every piece of magic is described by **two independent axes**. Keeping them separate
is the whole design:

| Axis | Field | Answers | Governs |
| --- | --- | --- | --- |
| **Color school** | `color` | *Which skill casts it?* | cost, success chance, power scaling |
| **Element (phase)** | `phase` | *How does it interact?* | amplify / suppress / drain / backlash vs other phased things |

A Red-school **Fire Bolt** (`color: red`, effect phase `fire`) and a Red-school
**Frost Bolt** (`color: red`, effect phase `water`) are trained by the *same*
`red_magic` skill but behave completely differently when they meet other elements.
Color never changes how elements interact; phase never changes which skill you
train.

### The nine color schools

Each color is a normal point-buy skill (`<color>_magic`, 0–100) in
`skills/core.json`. A school is defined by the kind of change its magic makes; the element
of that change is the other axis. The word in brackets is a functional parallel,
commentary only. Spells are listed by level, from `spells/core.json`. "Not in the data
yet" is what the skill's description promises and no effect or spell provides.
[V 2026-11-09]

**`red` (offense).** Harms at range with raw energy: fire, frost, shock, and force. Its
variety is shape, from one bolt to a whole battlefield, and it has no utility spell.
- Spells: Fire Bolt, Force Dart, Frost Lance, Fire Burst, Flame Shield, Chain Lightning,
  Ice Storm, Meteor, Inferno.

**`orange` (conjuration).** Makes what was not there, for as long as the spell lasts:
spirit creatures, elementals, weapons of light, and barriers.
- Spells: Summon Familiar, Bound Blade, Conjured Barrier, Conjured Blades, Bound Bow,
  Summon Elemental, Conjured Legion.

**`yellow` (illusion).** Changes what others see, hear, and believe: light and darkness of
the senses, false friendship, invisibility, and pain that exists only in the mind.
- Spells: Light, Blinding Flash, Blur, Mind Spike, Charm Person, Silence, Veil, Psychic
  Scream, Mass Charm.
- Not in the data yet: calm.

**`green` (restoration).** Mends the mortal body: health, stamina, poison and its cures. It
deals no damage.
- Spells: Mend Wounds, Cure Poison, Vigor, Regeneration, Purity Ward, Greater Heal, Mass
  Heal, Renewal.

**`blue` (alteration).** Changes the physical state of a body or thing: how heavy, how
hard, how fast, or what it is made of. Its damage is matter altered into acid; its
strongest spell turns flesh to stone.
- Spells: Feather Fall, Ward, Corrode, Slow, Stoneskin, Paralyze, Acid Rain, Petrify.
- Not in the data yet: water-walking, transmutation.

**`indigo` (force).** Pushes, blocks, and shakes with unseen force and sound, and works
locks and mechanical traps from a distance.
- Spells: Unlock, Disarm Trap, Thunderclap, Force Push, Force Wall, Greater Unlock, Greater
  Disarm, Sonic Boom.
- Not in the data yet: levitation.

**`violet` (mysticism).** Works on magic itself and on what magic connects: unravels and
absorbs spells, drains mana and will, moves objects by thought.
- Spells: Telekinesis, Dispel Magic, Sap Will, Spell Absorption, Dispel Ward, Mana Drain,
  Greater Dispel.
- Not in the data yet: Mark and Recall (planned effects, M3-06), soul detection and storage.

**`white` (holy).** Divine power that shelters allies and burns the unholy: radiant damage,
blessings, wards, and the cure of disease. Opposed by black.
- Spells: Holy Light, Blessing, Smite, Holy Ward, Cleanse, Sunburst, Divine Shield.
- Not in the data yet: banishing undead and demons, divine-grade healing, sanctify.

**`black` (necromancy).** Death and what it leaves behind: necrotic damage, curses, stolen
strength, armor of bone, and corpses raised to fight. Opposed by white.
- Spells: Drain Life, Curse, Bone Armor, Raise Thrall, Wither, Raise Horde, Death Grip.
- Not in the data yet: soul-corruption.

Where two schools come close, this is the line between them:

| Both touch | Who does what |
| --- | --- |
| Force | Red strikes with it (Force Dart); indigo pushes and blocks with it. |
| Summoning | Orange calls spirits and elementals; black raises the dead. |
| Moving things | Indigo shoves with force; violet lifts by thought. |
| Drains | Violet takes mana and will; black takes life and strength. |
| Cures | Green cures poison; white cures disease. |

### The five phases

`wood · fire · earth · metal · water` — defined in `schema.json#/definitions/phase`,
related through the cycles in [`modules/wuxing/data/wuxing.json`](../../ruleset/modules/wuxing/data/wuxing.json).
`phase` is **optional** on effects: non-elemental effects (invisibility,
telekinesis, summons, pure attribute buffs) carry none and sit out the cycles.

---

## 2. Effects — the atomic pool

A spell does not contain mechanics directly; it **references effects** from the
shared registry [`data/effects/core.json`](../../ruleset/data/effects/core.json)
(schema `effect.schema.json`). The same pool feeds potions, poisons, coatings,
ingredients, food, and enchantments — see [`progression.md`](./progression.md)
§ Effect registry.

An effect declares: `category`, `polarity`, an optional default `color`, an optional
`phase`, a `magnitudeUnit`, and the `channels` allowed to draw it. A spell may only
use an effect whose `channels` include `"spell"`.

---

## 3. Spells — composition

A spell (`spells/core.json`, schema `spell.schema.json`) is **one `color` + one or
more applied effects**:

```json
{
  "id": 1,
  "key": "fire_bolt",
  "label": "Fire Bolt",
  "color": "red",
  "effects": [ { "effect": 1, "magnitude": 8, "target": "ranged" } ],
  "cost": { "mana": 5 }
}
```

Each entry under `effects` is a `schema.json#/definitions/appliedEffect`: an effect
id plus its `magnitude`, `duration`, `target` shape (`self`/`touch`/`ranged`/`area`/
`world`, [casting.md](../core/casting.md#the-world-target)), and `radius` for areas. A spell's **phase(s)** are read from its composed
effects — there is no separate phase field on the spell. A multi-effect spell can
therefore carry multiple phases.

`level` (1–30) places the spell on the level scale; dispel reads it. `concentration`
marks a spell the caster holds. `cost`, `cooldown`, and the `custom` flag round out the
spell. `cost` is the
authored baseline; the engine derives it (§5).

A spell can be written as another spell plus what changes (`base` and `overrides`,
[composition.md](../core/composition.md#a-spell-built-on-a-spell)): `greater_unlock` is
`unlock` at level 15 with a higher mana cost and magnitude. A GM's or a module's custom spell
builds on a core one the same way once a module's spell file can name a core base. [V 2026-11-14 for `greater_unlock`; A for modules]

### The core spell list

`spells/core.json` holds at least five spells per color, 45 or more in all, from level 1 to
30. An authored spell draws only effects on its own color or with no color, so the list
teaches which school casts what; a crafted spell may be off-color only where a game allows it
(§5). [V 2026-11-08]

Every color has a damage, a utility, and a defensive spell, or a reason it has none:

| Color | Damage | Utility | Defensive |
| --- | --- | --- | --- |
| red | fire, frost, shock, force | none: red is the offense school; its range is shape (bolt, burst, storm) | flame shield |
| orange | conjured blades (bleed) | familiar, bound weapons, summoned elemental | conjured barrier |
| yellow | psychic | light, blinding flash, charm, silence, veil | blur |
| green | none: restoration does not harm; life taken is black's | cures, vigor | regeneration, purity ward, heals |
| blue | acid (corrosion) | feather fall, slow, paralyze, petrify | ward, stoneskin |
| indigo | thunder | unlock, disarm trap, force push | force wall |
| violet | mana and will drain | dispel, dispel ward | spell absorption |
| white | radiant | blessing, cleanse | holy ward, divine shield |
| black | necrotic | curse, wither, raise thrall | bone armor |

The damage effects `damage_radiant` (white), `damage_necrotic` (black), `damage_psychic`
(yellow), `damage_thunder` (indigo), and `damage_acid` (blue) join the pool so that every
color but green deals damage of its own; each matches a key in `combat/damage_types.json`.
`disarm_hazard` and `unlock` (indigo) are the spell side of traps and locks
([hazards.md](../core/hazards.md#spells-on-a-hazard)). `test_spells.py` checks the floor,
the colors, the channels, and this table.

Every spell, authored or crafted, costs what §5's formula gives for its effects. A spell's `level`
(1–30) is its power band, which dispel reads; it does not set the cost.

### Summon groups

A caster has one summon at a time unless a perk says otherwise: **Retinue** (level 14) allows two,
and **Host** (level 28, needs Retinue) three, each an `add 1` on `summon_limit`. A new summon past
the limit dismisses the oldest ([casting.md](../core/casting.md#summons)). [A]

A summon spell may exist only once what it summons is data. Its `summon` block names a group,
`core:summon/<key>`, from `summons/core.json`, and a duration (M3-25; it replaces the
`core:effect/summon` entry, [casting.md](../core/casting.md#summons)). A group lists its `variants`
as `core:monster/<key>` refs, and the caster picks one when casting [A]. A single summon calls one
creature; a minion spell calls `count` minions ([casting.md](../core/casting.md#minions)). The
creatures are ordinary monster rows ([monsters.md](./monsters.md#summoned-creatures)).

| Group | Spells | Variants |
| --- | --- | --- |
| `familiar` | Summon Familiar | owl, cat, raven |
| `elemental` | Summon Elemental, Conjured Legion (minions) | fire, ice, storm, acid, venom |
| `undead` | Raise Thrall, Raise Horde (minions) | skeleton, zombie |

A creature appears at its caster's level, whatever the spell's level. The row is the base of a
[derived monster](./monsters.md#derived-monsters) at that level, so its attributes and HP move by
its creature type's growth. The row still sets the stats: of two creatures at the same level, the
weaker row stays weaker. Templates go on top in order: the element, then the caster's power tier,
then the caster's nature. The caster's power and nature are perks with `unlocks`: the power ladder
`summon_dire`, `summon_elder`, `summon_ancient`, each needing the one before, and the natures
`summon_demonic`, `summon_angelic`, `summon_otherworldly`. The latest bought in each slot is the one
the caster holds ([monsters.md](./monsters.md#summoned-creatures)). [V 2026-11-12] `validate.py` refuses a summon block whose `group` is not a group, and
`test_spells.py` fails on a group or variant that does not exist, or on a spell that still names
`core:effect/summon`. [V 2026-11-16]

Minions take none of this: no power or nature, no Commanding Summons bonus, and no summoner
archetype perk. They are the row at the caster's level with the `minion` template, and they do
not count toward the summon limit ([casting.md](../core/casting.md#minions)). [V 2026-11-16, `test_spells.py`]

**Charisma strengthens summons through a perk.** `commanding_summons_magic` and
`commanding_summons_physical` each name a summon bonus, `core:summon_bonus/<key>`, from
`summons/bonuses.json`: `magic` adds to INT and WIL, `physical` to STR, END, and AGI. When the
caster summons, each bonus the caster holds adds the caster's CHA modifier, `CHA − 5`
([attributes.md](./attributes.md#the-attribute-modifier)), to every attribute of its group, after
the power and nature, through the same merge as a template's `attributes`, so END moves an authored
`hitPoints` and the creature's mana and damage follow its new attributes. A CHA of 5 or below adds
nothing; there is no penalty. Without a bonus perk, CHA changes nothing. A caster takes one group
with one perk; the second perk adds the other group. [V 2026-11-12, `test_spells.py`]

Each Commanding Summons perk has two ranks (M3-19). Rank 1 adds half the modifier, rank 2 the
full modifier: the bonus row's `rank_share` is `[0.5, 1]`, and a rank adds
`floor(max(CHA − 5, 0) × rank_share[rank − 1])`. At CHA 8 that is +1, then +3; at CHA 10, +2, then
+5. The caster's perks are refs at rank 1, or `{ "perk": ref, "rank": n }`. [V 2026-11-14, `test_spells.py`]

### The bonded summoner

The first of three summoner archetypes (GD_S06_02; the commander and the hybrid follow): the
caster stays the damage dealer, and an active summon makes their own magic stronger. The perks'
modifiers have the duration `while_summoning`: they hold only while the caster has at least one
active summon, and end when the last one goes ([modifier.md](../core/modifier.md#order-for-one-numeric-target)).
Minions do not count, as for every archetype perk. Each perk is one source, so two or three
summons out give each bonus once. [A]

| Perk | Level | Needs | While a summon is active |
| --- | --- | --- | --- |
| Kindred Focus | 6 | orange or black magic 40 | +10 to every magic skill (the nine `<color>_magic`) |
| Shared Wellspring | 12 | Kindred Focus, orange or black magic 55 | spells cost 20% less mana (`multiply 0.8` on `spell_cost`, after the skill discount) |
| Elemental Kinship | 18 | Kindred Focus, orange or black magic 70 | +4 damage on each damage packet whose type matches an active summon's element (`bonded_element_damage`) |

**An element match** reads the element templates of the caster's active summons (`fire`, `frost`,
`storm`, `acid`, `venom`, M3-10): each template's `damage_type`. A fire elemental out and a Fire
Bolt cast: the bolt's fire packet gains +4; a Frost Bolt gains nothing. The cast takes the
templates as `caster.summon_templates` ([casting.md](../core/casting.md#the-bonded-summoner)). [A]

The summon emits nothing of its own: the perks are the caster's, the session hands their
modifiers to the cast as it hands every perk's, and the cast drops the `while_summoning` ones when
`caster.summons` is empty. The kernel gains no notion of a bond. [A]

---

### The hybrid summoner

A summoner archetype that fights with a weapon beside their summons, like a magic swordsman or
an arcane archer (GD_S06_04). Its perks add to the attributes such a build needs while a summon
is out, and reward the caster and a summon striking one target together. Their modifiers have
the duration `while_summoning`: they hold only while the caster has at least one active summon
(minions do not count), and end when the last one goes
([modifier.md](../core/modifier.md#order-for-one-numeric-target)). [A]

| Perk | Level | Needs | While a summon is active |
| --- | --- | --- | --- |
| Bonded Vanguard (melee) | 6 | orange or black magic 30, Blades, Blunt, or Piercing 30 | +1 STR, +1 END |
| Bonded Marksman (ranged) | 6 | orange or black magic 30, Bows, Crossbows, Guns, or Throwing Weapons 30 | +1 AGI, +1 PER |
| Pincer Strike | 10 | orange or black magic 40 | +10 to attack when you and a summon of yours attack the same target (`beside_summon`, melee or ranged) |
| Kindred Mind | 12 | orange or black magic 50 | +1 INT, +1 WIL |
| Vanguard Ascendant | 18 | Bonded Vanguard | +1 STR, +1 END more |
| Marksman Ascendant | 18 | Bonded Marksman | +1 AGI, +1 PER more |

The melee path is Bonded Vanguard, then Vanguard Ascendant; the ranged path Bonded Marksman, then
Marksman Ascendant; both take Pincer Strike and Kindred Mind. Different perks add, so a hybrid
with both tiers of a path has +2 on each of its attributes. Numbers are [A] until a playtest runs
a hybrid against the spellblade of docs/playtests/.

**Striking together** is the caller's call, as positions always are: the caller names the
`beside_summon` situation when a summon of the attacker attacks or engages the same target this
round ([combat.md](../core/combat.md#situations-and-called-shots)). The row has no modifier of its
own; its `adds` field names `beside_summon_attack`, and the attacker's live modifiers on that
target join the attack only while the row applies.

**A summoned mount, later.** Once a mounted combat module exists, the ladder takes a rung that
makes a summon a mount (a summoned steed that carries its caster) and a mounted bonus beside it;
it is noted here as an extension and not built. [A]

### Bound weapons

A bound weapon spell (`bound_blade`, `bound_bow`) conjures the weapon its `bound_weapon` effect
names. The cast returns the conjured weapon as `bound_weapon` ([casting.md](../core/casting.md#bound-weapons)),
and the attacks it makes run the combat pipeline like any weapon's. The rules are the same for
every bound weapon, Bound Bow included. (GD_S06_01, owner rules of 2026-11-14.)

- **It starts as iron.** The weapon is the base it copies (dice, skill, properties) made of
  `core:material/iron` (tier 2, weapon attack +3). The default is written once, in
  `casting/core.json` under `bound_weapon`, not on each spell.
- **No phase.** The material gives stats only. A bound weapon has no phase, whatever its
  material's, so it sits out the five-phase cycles.
- **`magic` damage.** It deals `magic`, a plain damage type for non-elemental magic, in place of
  its weapon's physical type. `magic` has `armor_dr_applies`, so armor DR applies in full at the
  start, and it can miss, be evaded, and be blocked; `resist` with parameter `magic` reduces it,
  so a mage's magic resistance works against it. `force` does not fit: `bypasses_armor` and
  `unerring` would make a starting bound weapon too strong.
- **A coating or an elemental enchantment** sets the damage type and the phase. It is a
  `core:effect/damage_<type>` effect with the `coating` or the `enchantment` channel; the weapon
  then deals `<type>` and has that effect's phase. An enchantment wins over a coating. [A]
- **Perks upgrade it, one step at a time.** A perk changes the material by listing it in
  `unlocks` (a `core:material/<key>` ref: a modifier holds only a number, so the perk names it),
  or adds on-hit effects in `bound_weapon_effects` (applied effects from the shared pool with
  the `enchantment` channel). The material of the held perk highest on the ladder (prerequisite level, then the
  material's attack) wins, whatever order the perks are listed in; every
  held perk's effects add.

| Perk | Level | Needs | Gives |
| --- | --- | --- | --- |
| Tempered Conjuring | 6 | orange magic 40 | the material is `core:material/steel` (attack +5) |
| Piercing Conjuring | 10 | Tempered Conjuring, orange magic 50 | `armor_pierce` 25: a hit ignores a quarter of armor DR |
| Lingering Conjuring | 12 | Tempered Conjuring, orange magic 55 | `lingering_damage` 2 for 10 s, of the weapon's current damage type |
| Hallowed Conjuring | 14 | Piercing Conjuring, orange magic 60 | `bane` 4 against `undead` |
| Unbinding Edge | 18 | Hallowed Conjuring, orange magic 70 | `banish_summoned`: a hit ends a summoned creature |
| Glassfire Conjuring | 20 | Tempered Conjuring, orange magic 75 | the material is `core:material/obsidian` (attack +6) |

The levels sit 2 to 6 apart along the ladder; every number is [A] until a playtest runs a bound
weapon against a mundane one of the same level.

### The commander summoner

A summoner archetype (GD_S06_03): the summons are the damage dealers, and the caster spends their turn to
drive them, by **orders** (the `command` action, which gives one or every summon AP for its next
turn) and by a **held command** (Steady the Host, a concentration spell that gives every summon
+5 to attack while held). The ladder is Field Command (6), Sharp Orders (10), Voice of the Host
(14), and Iron Drill (18); the rules and the `command_summons` function are in
[commanding.md](../core/commanding.md). [A]

## 4. Casting resolution (d100 roll-under)

The procedure is [docs/core/casting.md](../core/casting.md): one resolution request on
the spell's `<color>_magic` skill, after the cast spends its mana and reagents.

1. **Gate.** A `silenced` caster cannot cast. Too little mana, or a missing reagent,
   means no cast and nothing spent.
2. **Cost.** Mana (`INT × 8 + level × 2` is the pool; it regenerates with WIL) and the
   spell's reagents are spent before the roll, so a failed cast still costs them.
3. **Success.** The target is the skill, plus the armor penalty of the heaviest armor
   worn, plus the modifiers on the skill. Crit and fumble are tags from
   [docs/core/resolution.md](../core/resolution.md). A phase is a tag. It is not
   required to cast.
4. **Effects.** A success emits the spell's effects at their authored magnitudes. The
   margin is in the result; scaling a magnitude by it is not core.
5. **Concentration and dispel.** A `concentration` spell is held, one at a time. A
   `dispel` effect of magnitude N ends the target's standing spells of level N or
   below.
6. **Elemental interaction** (§6) then adjusts the realized magnitude/duration based
   on the target's element and any standing effects.

Counterspell is cut (there is no reaction slot), and ritual casting is a module after
1.0. Casting **never grants skill XP** — magic skills rise by point-buy like every other
skill (advancement, not use-training).

---

## 5. Spell crafting (spellmaking)

Players assemble custom spells by combining effects. A crafted spell is an ordinary
spell entry with **`custom: true`**; it is subject to the same schema and the same
casting rules.

### Rules

- **Single color.** Every effect in a custom spell must be castable under one chosen
  color. (An effect's default `color` is the natural fit; a game may allow
  off-color use at a skill penalty.)
- **Channel gate.** Each chosen effect must list `"spell"` in its `channels`.
- **Skill gate.** The spell is only castable once the matching `<color>_magic` skill
  meets the spell's **complexity threshold** (below). Spellmaking the recipe may
  require a higher threshold than casting an authored equivalent.
- **No new mechanics.** Crafting composes *existing* effects at chosen magnitudes,
  durations, and targets. Inventing a new effect is a content/authoring change to the
  pool, not a player action.

- **No saves.** A crafted spell cannot carry a save: saves are for authored instant spells
  ([casting.md](../core/casting.md#delivery-and-saves)). A crafted spell is avoided only by a dodge or a block.
- **Known effects only.** A crafted spell composes only effects the character knows. The
  `spellcraft` function checks a recipe ([casting.md](../core/casting.md#spellcraft)). [A]

### Spell knowledge

A character knows a set of effects, kept as refs (`core:effect/<key>`). Knowledge comes two ways:

- **Learning a spell** (a scroll, a trainer, a perk) teaches every craftable effect in it. It
  does not teach the target shape: shapes are free and priced by the target multiplier, so
  learning Fire Bolt lets you craft a fire burst. A spell whose effects are all uncraftable
  (Bound Blade, Raise Thrall) teaches nothing for crafting; the spell itself is still known
  and castable. [A]
- **Research**, a downtime action, teaches one craftable effect without a spell: 7 days and
  50 gp at a library or guild, then one `<color>_magic` check on the effect's color (or a color
  the researcher names, for an effect with none). Research is refused when the skill is below
  the effect's craft threshold, the threshold of the cheapest spell that carries it
  (magnitude 1, instant, touch). It is the only downtime activity until M4. The numbers are in
  `casting/core.json` under `research`. [A]

### Uncraftable effects

An effect has `craftable` (default `true`). `false` means only an authored spell carries it, so
no recipe, Research, or learned spell teaches it. Where an effect is craftable for some
parameters only, `uncraftable_parameters` lists the others. [A]

| Effect | Why it is authored only |
| --- | --- |
| `teleport` | Divine Intervention sends the caster to the nearest temple of a deity. |
| `mark`, `recall` | A fixed return point is a world rule, not a recipe. |
| `storage` | Pocket Storage, a pocket inventory, is engine state. |
| `bound_weapon` | Bound weapons are rows a perk ladder upgrades ([Bound weapons](#bound-weapons)). |
| `summon` | Summons are static rows a GM or dev adds (M3-08). |
| `command_summons` | The held command belongs to the commander's ladder ([commanding.md](../core/commanding.md)), not to a recipe. |
| `restore_resource` with `mana` | No crafted spell restores mana, instant or over time, so mana cannot pay for itself. |

Absorb and drain effects (`spell_absorption`, `life_steal`, `mana_leech`, the `drain_*`
effects, `damage_resource`) stay craftable: they take from a target, not from nothing. The
regen family (health and stamina over time) uses `duration` with magnitude per second (M2-05).

Morrowind gives Divine Intervention, Mark, and Recall high base costs (150, 350) and lets them
be spellmade; here they are authored only, and their base costs only price the authored rows.

### Template spells

A spell has `template` (default `false`). A template is a plain, learnable spell with one
effect, no reagents, no save, and no concentration, whose mana cost and cast threshold are
exactly the cost formula's, so a crafted spell reads as a template with other effects. One per
target shape: [A]

| Shape | Template | Target |
| --- | --- | --- |
| bolt | Fire Bolt | `ranged` |
| touch | Holy Light | `touch` |
| burst | Fire Burst | `area` |
| ward | Conjured Barrier | `self` |
| mend | Mend Wounds | `touch`, a heal |

### Cost formula

A spell's mana cost comes from its effects alone: what they do, for how long, and how far
they reach. The character's level never enters it; a level-1 and a level-30 caster pay the
same for the same spell, before their own skill and perks (below). The coefficients are in
`ruleset/data/casting/core.json` (`spell_cost`). [V 2026-11-08]

```
per effect
  m          = magnitude, or 1 for an effect whose magnitude is not a strength
               (magnitudeUnit seconds, or none: a stun's strength is its duration)
  cost       = 0.1 × baseCost × m × duration          duration in seconds, 1 = instant
  cost      += 0.05 × baseCost × radius               area only
  cost      ×= target multiplier                       self 1, touch 1, ranged 1.5, area 2
spell cost   = ceil(sum over the effects)              cost.mana on the row
```

- **`baseCost`** is on each effect row: how much one point of it for one second is worth.
  It carries the difference between a point of fire and a second of paralysis, so no unit
  weights are needed. Light is 0.2, a point of fire 5, a second of paralysis 40.
- **Duration** multiplies, so a pool effect is priced by its total (10 health a second for
  10 seconds costs what 100 at once does) and a buff by how long it holds.
- **Riders** cost what they do: a 6-second blind on a burst is priced like a blind spell.

**Sources.** The shape is Morrowind's: OpenMW's `calcEffectCost` multiplies the mean
magnitude by `0.1 × baseCost` and by the duration, adds `0.05 × area × baseCost`, and
multiplies an effect cast on a target by 1.5; an effect with no magnitude counts it as 1
(checked against OpenMW/openmw `apps/openmw/mwmechanics/spellutil.cpp` at `b8285cc95d19`,
2026-11-08). The per-effect base costs follow its ratios (UESP, Morrowind:Magic Effects,
2026-11-08: Light 0.2, Fire Damage 5, Damage Health 8, Paralyze and Silence 40), scaled to
our magnitudes. The area multiplier is Daggerfall's: Daggerfall Unity's
`ApplyTargetCostMultiplier` is caster and touch 1.0, single target at range 1.5, area around
the caster 2.0, area at range 2.5; we have one area shape and take 2.0 (Interkarma/
daggerfall-unity `Assets/Scripts/Game/Formulas/FormulaHelper.cs` at `4e78e1412b17`,
2026-11-08). Not taken: Daggerfall's magnitude and duration that grow per caster level
(`plus` per `perLevel` levels), Morrowind's `fEffectCostMult`, and its offset of one second
on a player-made spell's duration. [V 2026-11-08]

### The caster's side

What a caster pays is the spell cost after their skill and perks, never their level.
Daggerfall lowers spell points as the skill rises (`gold × (110 − skill) / 400`); here the
skill discount is `<color>_magic / 200`, capped at 0.45, and perks emit modifiers on
`spell_cost` ([paid cost](../core/casting.md#paid-cost)). The cast threshold and the craft threshold read the spell cost before any
discount:

```
castThreshold   = round( spell cost )           // min <color>_magic to cast at all
craftThreshold  = round( spell cost × 1.25 )    // min <color>_magic to author the recipe
```

Uncraftable effects ([above](#uncraftable-effects)) are priced only for their authored
spells.

---

## 6. Elemental interaction — the five-phase layer

This layer is the Wuxing module ([wuxing.md](./wuxing.md)); with the module off, phases are inert tags.

This is the part that makes effects "interact with each other." It is fully defined
in [`modules/wuxing/data/wuxing.json`](../../ruleset/modules/wuxing/data/wuxing.json) (the single source
of truth) and summarised in [`progression.md`](./progression.md) § Elemental
interaction. When a **phased** effect resolves, the engine compares its phase against
the relevant other phase and applies the matching cycle's multiplier:

| Cycle | When it fires | Multiplier (reference) |
| --- | --- | --- |
| **Generating** | acting phase nourishes the other (mother → child) | amplify ×1.5 |
| **Overcoming** | acting phase controls the other (e.g. Water → Fire) | suppress ×0.5 (dispel if overwhelming) |
| **Weakening** | a child phase drains its mother | drain ×0.75 |
| **Insulting** | the normally-controlled phase is over-strong and rebels | backlash ×1.25 (conditional) |

The "other phase" can come from three places:

1. **Effect vs standing effect.** Cast `damage_fire` on a target already burning with
   a Wood effect → Generating amplifies the fire. Cast a Water effect on a standing
   Fire effect → Overcoming suppresses/dispels it.
2. **Effect vs monster element.** Every monster has exactly one `phase`
   (`monster.schema.json`). A Water spell against an `ember_drake` (Fire) lands under
   Overcoming → boosted; a Fire spell against it falls under Weakening/affinity →
   reduced. This is the core of elemental counterplay.
3. **Effect vs race element.** Every playable race has one `phase`
   (`race.schema.json`) — the same rules give player characters innate elemental
   strengths and soft spots. (Abstract `parent`/`template` race entries are not
   played and may omit it; subraces inherit or override their parent's phase, and
   beastfolk kin each declare their own.)

> **Direction convention.** In the cycle data an edge `from → to` is the canonical
> reading "*from* acts upon *to*". The engine maps the **incoming** effect and the
> **standing** effect/target onto `from`/`to` by matching the edge, then applies the
> multiplier to the participant named by the cycle's `affects` field.

### Worked examples

- **Frost Bolt vs Ember Drake.** Frost = `water`, drake = `fire`. Water overcomes
  Fire → Overcoming ×0.5 *suppression of the drake's element*: the bolt bites deep.
- **Fire Bolt vs Ember Drake.** Fire vs Fire, and Fire is *fed by* Wood / *drains
  into* Earth — same-element attacks land at affinity (reduced). Bring Water instead.
- **Combo setup.** Apply a Wood damage-over-time, then a Fire spell: Generating
  ×1.5 — the mother phase feeds the child, and the fire flares.
- **Insult backlash.** Hit a Water-phase `frost_revenant` with an overwhelming Fire
  spell whose magnitude exceeds the target's: the controlled Water *insults* back
  ×1.25 — high-risk, high-reward against a resistant element.

---

## 7. Data & schema map

| Concern | Schema | Data |
| --- | --- | --- |
| Atomic effects (+ `phase`) | `effect.schema.json` | `data/effects/core.json` |
| Spells (+ spellmaking `custom`) | `spell.schema.json` | `data/spells/core.json` |
| Phase vocabulary | `schema.json#/definitions/phase` | — |
| Interaction matrix | `wuxing.schema.json` | `modules/wuxing/data/wuxing.json` |
| Monster elements | `monster.schema.json` | `data/monsters/core.json` |
| Race elements | `race.schema.json` | `data/races/core.json` |
| Color skills | `skill.schema.json` | `data/skills/core.json` |

---

## 8. Open items / follow-ups

- **Engine resolution.** Phases, the matrix, and the elements on every monster/race
  live here; the consuming engine has no five-phase concept yet and must implement
  the resolution in §6. This data **leads** that design.
- **Tuning.** All multipliers, weights, and thresholds above are reference values,
  not balanced numbers.
- **Off-color casting.** Whether (and how steeply) an effect may be cast under a
  non-default color is left to the game.
- **Phasing the rest of the pool.** Consumables and ingredients can opt into the same
  `phase` tags so alchemy reuses the cycles (Generating → potency, Overcoming →
  cures/antidotes). Not yet tagged.
- **Multi-phase spells.** Resolution order when one spell carries several phases is
  engine-defined (suggested: resolve each effect independently against the target).
