#!/usr/bin/env python3
"""Python reference for the `creation` command (BE_S07_08, docs/backend/creation.md): the
creation rules an app draws its form from, a check of a draft step by step, and the stored
actor a complete draft becomes.

A draft holds any steps in any order: `name`, `race`, `points` (spent per attribute, negative
when lowered under the start), `background`, `creation_perks`, and `tag_skills`. The numbers
are progression's `creation` block and `tag_skills` (GD_S07_07); race traits and creation
perks that `add` to `tag_skills` or `creation_perks` raise those limits.
"""

import json
from pathlib import Path

DATA_DIR = Path(__file__).resolve().parent.parent / "data"
ABILITIES = ("str", "per", "end", "int", "wil", "agi", "cha", "lck")
STEPS = ("name", "race", "attributes", "background", "creation_perks", "tag_skills")


def load_data(rel_path):
    return json.loads((DATA_DIR / rel_path).read_text())


def _rows(rel, collection):
    return {row["key"]: row for row in load_data(rel)[collection]}


def _key(ref, category):
    prefix = f"core:{category}/"
    return ref[len(prefix):] if isinstance(ref, str) and ref.startswith(prefix) else None


def playable_races(races):
    """Every race a player may pick, in data order: all but templates."""
    return [race for race in races.values() if race["lineage"]["role"] != "template"]


def race_parts(race, races):
    """A race's ability modifiers summed with its parent's, and its traits, parent's first."""
    rows = [race]
    if race["lineage"]["role"] == "subrace":
        rows.insert(0, races[race["lineage"]["parent_race"].rsplit("/", 1)[-1]])
    modifiers, traits = {}, []
    for row in rows:
        for ability, value in row.get("ability_modifiers", {}).items():
            modifiers[ability] = modifiers.get(ability, 0) + value
        traits += [trait["id"] for trait in row.get("traits", [])]
    return modifiers, traits


def run_rules(vector):  # noqa: ARG001  (every runner takes the vector)
    progression = load_data("progression/core.json")["progression"]
    races = _rows("races/core.json", "races")
    features = _rows("features/core.json", "features")
    creation = dict(progression["creation"], tag_skills=progression["tag_skills"])
    race_rows = []
    for race in playable_races(races):
        modifiers, traits = race_parts(race, races)
        race_rows.append({"race": f"core:race/{race['key']}", "label": race["label"], "ability_modifiers": modifiers, "traits": traits})
    return {
        "creation": creation,
        "abilities": list(ABILITIES),
        "races": race_rows,
        "backgrounds": [
            {"background": f"core:background/{row['key']}", "label": row["label"], "skill_bonuses": row.get("skill_bonuses", []),
             "suggested_tag_skills": row.get("suggested_tag_skills", [])}
            for row in load_data("backgrounds/core.json")["backgrounds"]
        ],
        "creation_perks": [
            {"perk": f"core:feature/{row['key']}", "label": row["label"], "description": row["description"]}
            for row in features.values() if row["type"] == "creation"
        ],
        "skills": [
            {"skill": f"core:skill/{row['key']}", "label": row["label"], "ability": row["associated_ability"]}
            for row in load_data("skills/core.json")["skills"]
        ],
    }


def _chosen(refs, category, kind, check, problems, step):
    """The keys of a list of refs; each unknown, unfitting, or repeated ref is a problem."""
    keys = []
    for ref in refs:
        key = _key(ref, category)
        if key is None or not check(key):
            problems.append({"step": step, "message": f'{kind} "{ref}" is not defined'})
        elif key in keys:
            problems.append({"step": step, "message": f'{kind} "{ref}" is chosen twice'})
        else:
            keys.append(key)
    return keys


def run_check(vector):
    from kernel_reference import apply_modifiers, emit_perk, evaluate_formula, run_derived  # imports this module

    draft = vector.get("draft", {})
    progression = load_data("progression/core.json")["progression"]
    rules = progression["creation"]
    races = _rows("races/core.json", "races")
    features = _rows("features/core.json", "features")
    backgrounds = _rows("backgrounds/core.json", "backgrounds")
    skills = load_data("skills/core.json")["skills"]
    problems = []

    if not str(draft.get("name", "")).strip():
        problems.append({"step": "name", "message": "a name is needed"})

    race_modifiers, traits = {}, []
    if "race" not in draft:
        problems.append({"step": "race", "message": "a race is needed"})
    else:
        key = _key(draft["race"], "race")
        if key not in races:
            problems.append({"step": "race", "message": f'race "{draft["race"]}" is not defined'})
        elif races[key]["lineage"]["role"] == "template":
            problems.append({"step": "race", "message": f'race "{draft["race"]}" is a template and is never played'})
        else:
            race_modifiers, traits = race_parts(races[key], races)

    points = draft.get("points", {})
    for ability in points:
        if ability not in ABILITIES:
            problems.append({"step": "attributes", "message": f'attribute "{ability}" is not one of the eight'})
    base = {ability: rules["attribute_start"] + points.get(ability, 0) for ability in ABILITIES}
    in_range = True
    for ability in ABILITIES:
        if base[ability] > rules["attribute_max"]:
            problems.append({"step": "attributes", "message": f"{ability} is {base[ability]}, over the maximum of {rules['attribute_max']}"})
            in_range = False
        elif base[ability] < rules["attribute_min"]:
            problems.append({"step": "attributes", "message": f"{ability} is {base[ability]}, under the minimum of {rules['attribute_min']}"})
            in_range = False
    spent = sum(points.get(ability, 0) for ability in ABILITIES)
    left = max(rules["attribute_points"] - spent, 0)
    if in_range and spent > rules["attribute_points"]:
        problems.append({"step": "attributes", "message": f"{spent} points spent, more than the {rules['attribute_points']} to distribute"})
    elif in_range and left > 0:
        problems.append({"step": "attributes", "message": f"{left} point{'s' if left != 1 else ''} left"})

    background = None
    if "background" not in draft:
        problems.append({"step": "background", "message": "a background is needed"})
    else:
        background = _key(draft["background"], "background")
        if background not in backgrounds:
            problems.append({"step": "background", "message": f'background "{draft["background"]}" is not defined'})
            background = None

    perk_problems = []
    perks = _chosen(draft.get("creation_perks", []), "feature", "feature", lambda key: key in features, perk_problems, "creation_perks")
    for key in list(perks):
        if features[key]["type"] != "creation":
            perk_problems.append({"step": "creation_perks", "message": f'feature "core:feature/{key}" is not a creation perk'})
            perks.remove(key)

    # The modifiers the character starts with: its race traits and creation perks, at rank 1.
    held = [key.rsplit("/", 1)[-1] for key in traits] + perks
    modifiers = [m for key in held for m in emit_perk(features[key], 1)]
    trait_modifiers = [m for key in traits for m in emit_perk(features[key.rsplit("/", 1)[-1]], 1)]
    most_perks = apply_modifiers(rules["creation_perks_max"], [m for m in trait_modifiers if m["target"] == "creation_perks"])
    if len(perks) > most_perks:
        perk_problems.append({"step": "creation_perks", "message": f"at most {most_perks} creation perks, {len(perks)} chosen"})
    problems += perk_problems

    known = {row["key"] for row in skills}
    tag_problems = []
    tags = _chosen(draft.get("tag_skills", []), "skill", "skill", lambda key: key in known, tag_problems, "tag_skills")
    needed = apply_modifiers(progression["tag_skills"], [m for m in modifiers if m["target"] == "tag_skills"])
    if len(draft.get("tag_skills", [])) != needed:
        tag_problems.append({"step": "tag_skills", "message": f"{needed} tag skills are needed, {len(draft.get('tag_skills', []))} chosen"})
    problems += tag_problems

    stored = {ability: base[ability] + race_modifiers.get(ability, 0) for ability in ABILITIES}
    attributes = {ability: apply_modifiers(stored[ability], [m for m in modifiers if m["target"] == ability]) for ability in ABILITIES}
    bonuses = {}
    if background is not None:
        for bonus in backgrounds[background].get("skill_bonuses", []):
            bonuses[bonus["skill"].rsplit("/", 1)[-1]] = bonus["bonus"]
    starting = {
        row["key"]: evaluate_formula(progression["skill_start"], {}, {"attribute": attributes[row["associated_ability"]]}) + bonuses.get(row["key"], 0)
        for row in skills
    }
    derived = run_derived({"actor": {"attributes": attributes, "level": progression["start_level"]}, "modifiers": modifiers})
    problems.sort(key=lambda problem: STEPS.index(problem["step"]))
    return {
        "complete": not problems,
        "problems": problems,
        "points_left": left,
        "attributes": attributes,
        "skills": starting,
        "derived": derived,
        # What finish stores; not part of the check's result.
        "_stored": {"attributes": stored, "held": held, "tags": tags},
    }


def run_finish(vector):
    checked = run_check(vector)
    if not checked["complete"]:
        return {"error": "the draft is not complete: " + "; ".join(problem["message"] for problem in checked["problems"])}
    stored, derived = checked["_stored"], checked["derived"]
    actor = {
        "attributes": stored["attributes"],
        "level": load_data("progression/core.json")["progression"]["start_level"],
        "xp": 0,
        "skills": checked["skills"],
        "tags": stored["tags"],
        "growth": {"hp": 0, "mp": 0, "sp": 0},
        "resources": {pool: derived[pool] for pool in ("hp", "mp", "sp", "ap")},
        "conditions": [],
        "modifiers": [],
        "flags": [],
        "perks": {key: 1 for key in stored["held"]},
        "spells": [],
        "known_effects": [],
        "inventory": [],
        "equipped": {},
    }
    return {"name": vector["draft"]["name"].strip(), "actor": actor}


def run_creation(vector):
    operation = vector.get("operation")
    if operation == "rules":
        return run_rules(vector)
    if operation == "check":
        result = run_check(vector)
        result.pop("_stored")
        return result
    if operation == "finish":
        return run_finish(vector)
    return {"error": f'creation operation "{operation}" is not defined'}
