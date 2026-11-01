# Factions and reputation module

**Status:** Shipping as an optional module (POST-05). Module id `factions`, directory
`ruleset/modules/factions/`.

A faction is a group the character can stand well or badly with: a city watch, a merchant
guild, a cabal. Reputation is one integer per faction. Standing and rank effects are
contract modifiers, so combat, magic, and dialogue never call this module; they read the
modifiers it emits ([kernel contract](../core/contract.md)).

## Inspiration and how it differs

- **Fallout: New Vegas** tracks fame and infamy per faction and names the result (Vilified
  to Idolized), and a bad enough standing makes a faction shoot on sight **[U]**. Riftweave
  keeps one signed number per faction instead of two counters, and the named result is a
  standing row with modifiers, not a dialogue table.
- **The Elder Scrolls (Morrowind, Oblivion)** guilds have ranks gated by reputation and
  skills, and promotion is something you ask for **[U]**. Riftweave keeps the gate and the
  explicit promotion, gates on reputation only, and leaves quests to the consumer.
- **D&D 3.5 typed bonuses** are the failure this avoids: a faction bonus has no type of its
  own. It is a modifier with a faction source and stacks by the one rule in
  [modifier.md](../core/modifier.md).

## Data

One new collection, `factions`, validated by `ruleset/schemas/faction.schema.json` **[V 2026-11-02]** (`validate.py`
passes the module's `data/factions.json`).
A faction is referenced as `module:factions/faction/<key>`.

```
Faction
  id, key, label, description, source?
  reputation_min        lowest reputation, default -100
  reputation_max        highest reputation, default 100
  reputation_start      reputation of an actor with no entry, default 0
  standings[]           named bands, ascending by min_reputation
    key, label, min_reputation, modifiers[]
  ranks[]               membership ladder, lowest first
    key, label, min_reputation, modifiers[]
  relations[]           how a change here spills onto another faction
    faction, share_percent      (-100..100)
  reputation_events[]   default deltas a consumer may emit
    key, label, delta
```

`modifiers[]` is `schema.json#/definitions/modifier`, the only shared shape. A standing's
modifiers all use `source: faction_standing:<faction key>`; a rank's use
`source: faction_rank:<faction key>`. The schema checks the prefix; the reference code checks
the key, the band order, and that each relation names another faction
(`check_factions` in `faction_reference.py`) **[V 2026-11-02]**. `validate.py` does not run
those checks yet.

## Actor state

The module adds two primary fields to the actor snapshot. They are owned by the actor, like
XP, and only this module's events write them:

```
reputation[faction ref]     integer; missing = reputation_start
faction_ranks[faction ref]  rank key held; missing = not a member
```

Standing, rank eligibility, and the emitted modifiers are functions of this state. They are
never stored: the
reference recomputes them from the snapshot on every call **[V 2026-11-02]**.

## Rules

1. **Standing.** The standing is the last row whose `min_reputation <= reputation`. The first
   row's `min_reputation` equals `reputation_min`, so every value has a standing. Bounds are
   inclusive at the bottom: reputation 10 with a `friendly` row at 10 is friendly.
2. **Change.** A reputation event names a faction and a delta (the row's `delta`, or one the
   consumer supplies). The new value is `clamp(reputation + delta, reputation_min,
   reputation_max)`.
3. **Spillover.** For each `relations` row on the event's faction, the related faction also
   changes by `trunc(delta * share_percent / 100)`, rounding toward zero, then clamps to its
   own bounds. Spillover is one hop: it does not cascade through the related faction's own
   relations, so there are no loops. Spillover uses the unclamped delta.
4. **Rank.** Rank is held, not computed: a consumer joins or promotes the actor by event,
   usually after a quest. The module only answers whether the next rank is allowed: the
   actor holds the rank before it (or nothing, for the first rank) and has
   `reputation >= min_reputation` of the next rank.
5. **Suspension.** A held rank whose `min_reputation` the actor no longer meets is suspended:
   it stays held, but emits no modifiers until reputation recovers. Expulsion is a consumer
   event that removes the rank.
6. **Emission.** For each faction the actor has reputation with or a rank in, the module emits
   the current standing's modifiers and, if not suspended, the held rank's modifiers. Ranks
   do not accumulate: a higher rank's row lists its whole bundle.
7. **Stacking.** The kernel rule, unchanged. Moving between standings replaces the old
   bundle because every standing of one faction shares one source. Two factions are two
   sources and add. Standing and rank of one faction are two sources and add.

Every shipped modifier uses `duration: while_applied`, the value conditions already use: it
lasts while the standing or rank holds **[V 2026-11-02]**.

Hostility is a `grant` of a flag such as `lantern_wardens_hostile`. What a hostile guard
does is the consumer's rule; the module only says the flag is set.

## Example

Lantern Wardens, reputation 0, persuasion 40. A completed contract (`delta` 15) moves the
Wardens to 15, which is `friendly` (+5 persuasion). The Ashen Circle relation has
`share_percent` -50, so the Circle changes by `trunc(15 * -50 / 100) = -7`. Persuasion
against a resolution roll is now 45. The cases are pinned in
`ruleset/tests/vectors/factions.json` (18 cases), and `ruleset/scripts/faction_reference.py`
reproduces them, run by `ruleset/scripts/test_faction_reference.py` **[V 2026-11-02]**.

A faction the actor has no entry for emits nothing, even if its `reputation_start` band has
modifiers. Every shipped faction starts in `neutral`, which has none **[V 2026-11-02]**.

## Out of scope

- Disguises, witnesses, and crime detection that decide when an event fires.
- Prices. A barter modifier waits for the economy issue (M4); until then standings speak
  only to skills and flags.
- Alignment and karma, which stay parked in the POST story.
