# Checks

A check is one call of the [resolution function](resolution.md) against one skill. Rows live in `ruleset/data/checks/core.json`, schema `check.schema.json`. A row names a common use of a skill, so a game, a GM, a hazard, or a recipe can point at one shared key and not restate which skill rolls.

Fallout lists what each skill is for next to the skill. Riftweave keeps that list as rows, so a tool can read it. **[U]**

## Row

```
key          snake_case, unique in the collection
label        display name
description  what the check does, in one sentence
skill        core:skill/<key>, the skill that rolls
opposed_by   core:skill/<key>, optional: the skill the other side rolls
```

A row has no difficulty. Difficulty is a bonus or penalty on the target, as in [resolution.md](resolution.md). A row has no damage either. An unarmed strike or a dodge names its skill here; what follows a hit stays in `docs/modules/combat.md`.

`opposed_by` only names the skill the other side rolls. How two results compare is not specified yet. **[U]** It is an open question in [decisions.md](../decisions.md).

## Every skill has a consumer

A skill nothing rolls is dead weight on the sheet. Each skill in `ruleset/data/skills/core.json` needs a governing attribute (`associatedAbility`, required by the schema) and at least one **consumer**:

| Consumer | Where | Skill it names |
| --- | --- | --- |
| weapon | equipment `weapon.skill` | the ref |
| armor | equipment `armor.category` | `<category>_armor` |
| spell color | spell or effect `color` | `<color>_magic` |
| check | check `skill` or `opposed_by`, recipe `skill` | the ref |
| perk | feature `prerequisite.skills`, or a feature effect on `fortify_skill`, `drain_skill`, or `damage_skill` | the ref or `parameter` |

An effect's `color` counts because it is the default school when the effect is cast (`effect.schema.json`). A spell's `fortify_skill` parameter counts as a spell consumer. **[A]**

Background `skill_bonuses` and `suggested_tag_skills` are references, not consumers: a bonus to a skill nothing rolls is still dead weight. A skill with references but no consumer is an orphan, the same as a skill with no references.

`ruleset/scripts/skill_audit.py` lists every skill with its consumers and references, and fails on an orphan. It also fails when the **Seeds skills** lines in [attributes.md](../modules/attributes.md) differ from `associatedAbility` in the data, and when a `core:skill/` ref names no skill. `test_skill_audit.py` runs it against the repository, so `./scripts/test.sh` fails on an orphan.

On 2026-11-02 the audit found 18 orphans among 44 skills. Each got a check row, and `orange_magic` got its first effect, `bound_weapon`; none was removed. It now reports 44 skills, 0 problems. **[V 2026-11-02]**

A bare skill key in a `parameter` (`"animal_handling"`, one perk today) is counted so it is not lost, but the format is `core:skill/<key>`, as in [keys.md](keys.md).
