# M3 Combat and magic playable

Roadmap: [M3](../../ROADMAP.md#where-we-are). Format and rules: [README.md](README.md).

## M3 Combat and magic playable

**Goal:** a combat and a casting round can be run at any level using only core rows.

**Spec:** `docs/core/` for kernel rules, `docs/modules/` for module rules.

**Done when:** optional mechanics (cover, called shots, counterspell, and the like) are each either core with vectors or a named module.

Done: —

#### RW-16 Casting procedure: decide scope, then specify
- Status: todo
- Priority: high
- Needs: —
- Do: Magic emits a roll request and an effect list (SET-16). Still undecided: components, concentration, counterspell and dispel, ritual casting, and spell failure from armor. Decide which are core and which are optional modules, and write the core ones as data.
- See: issues.md section 4, docs/modules/magic.md
- Done when: A decision record lists each mechanic as core, module, or cut; core mechanics have rows, formulas, and vectors; `magic.md` matches.

#### RW-17 Spell list to the content floor
- Status: todo
- Priority: high
- Needs: —
- Do: 7 spells exist. 9 colors need a usable list.
- See: ruleset/data/spells/core.json
- Done when: at least 5 spells per color (45 total) across levels 1 to 30; each spell composes effects from the shared pool and respects `channels`; every color has a damage, utility, and defensive option or a documented reason it does not.

#### RW-18 Cover, flanking, terrain, and called shots as modifiers
- Status: todo
- Priority: medium
- Needs: —
- Do: These are prose in `combat.md`. Decide core or optional, then express them as modifier rows so they stack under the one rule.
- See: docs/modules/combat.md
- Done when: each is a row set with vectors, or is moved to a named optional module; no special-case stacking rule.

#### RW-19 Wuxing interaction vectors
- Status: todo
- Priority: medium
- Needs: —
- Do: Wuxing is a tag reader (SET-18). Pin the four cycles and their multipliers with vectors, including the reverse-cycle math the review called fragile.
- See: docs/modules/wuxing.md, ruleset/data/wuxing/core.json
- Done when: vectors for generating, overcoming, weakening, and insulting; reverse-cycle logic exists once.
