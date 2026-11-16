"""Suggested builds (docs/modules/builds/README.md): play a build from creation to level 30 and
say where it breaks a rule.

A build is a plan, not a played character: it rolls nothing. It is checked at the lowest luck
roll (0) every level, so a build that is legal here is legal for any roll. Race, background,
and gear add on top of it and only help; they are not read.
"""

import json
import sys
from pathlib import Path

_SCRIPTS = Path(__file__).resolve().parent
sys.path.insert(0, str(_SCRIPTS))

import kernel_reference  # noqa: E402

DATA = _SCRIPTS.parent / "data"
ABILITIES = ("str", "per", "end", "int", "wil", "agi", "cha", "lck")


def _rows(rel, collection):
    return {row["key"]: row for row in json.loads((DATA / rel).read_text())[collection]}


def _key(ref):
    return ref.rsplit("/", 1)[-1]


def attributes_after_creation(build, features=None):
    """The point-buy attributes plus every creation perk's attribute adds."""
    features = features or _rows("features/core.json", "features")
    attributes = dict(build["attributes"])
    for ref in build.get("creation_perks", []):
        for modifier in features[_key(ref)].get("modifiers", []):
            if modifier["target"] in ABILITIES and modifier["operation"] == "add":
                attributes[modifier["target"]] += modifier["value"]
    return attributes


def play(build):
    """Play the build to max_level. Returns (problems, sheet): every rule it breaks, in order, and
    the sheet at the last level (level, attributes, skills, perks held)."""
    features = _rows("features/core.json", "features")
    rules = json.loads((DATA / "progression/core.json").read_text())["progression"]
    creation_rules = rules["tag_skills"]
    problems = []

    # Creation: 4 in every attribute plus 10 points, none over 10 (character-creation.md step 2).
    assigned = build["attributes"]
    if sorted(assigned) != sorted(ABILITIES):
        problems.append("attributes must name the eight abilities")
        return problems, None
    if any(not 4 <= value <= 10 for value in assigned.values()) or sum(v - 4 for v in assigned.values()) != 10:
        problems.append("attributes: 4 each plus 10 points, none over 10")
    creation_mods = []
    for ref in build.get("creation_perks", []):
        row = features.get(_key(ref))
        if row is None or row["type"] != "creation":
            problems.append(f"{ref} is not a creation perk")
            continue
        creation_mods += row.get("modifiers", [])
    if len(build.get("creation_perks", [])) > 2:
        problems.append("at most two creation perks")
    attributes = attributes_after_creation(build, features)
    on = lambda target: [m for m in creation_mods if m["target"] == target]
    tag_count = kernel_reference.apply_modifiers(creation_rules, on("tag_skills"))
    tags = [_key(ref) for ref in build["suggested_tag_skills"]]
    if len(tags) != tag_count:
        problems.append(f"suggested_tag_skills: {tag_count} tag skills, not {len(tags)}")

    sheet = {"level": 1, "attributes": attributes, "skills": {}}
    held = {}
    carried = 0

    def read(skill):
        return kernel_reference.read_skill(sheet, skill)["value"]

    def spend(points):
        for priority in build["skill_priorities"]:
            skill = _key(priority["skill"])
            cost = rules["skill_cost"]["tag" if skill in tags else "normal"]
            ceiling = min(priority["to"], rules["skill_cap"])
            value = read(skill)
            while points >= cost and value < ceiling:
                value += 1
                points -= cost
            sheet["skills"][skill] = value
        return points

    for level in range(rules["start_level"], rules["max_level"] + 1):
        sheet["level"] = level
        points = kernel_reference.evaluate_formula(rules["skill_points"], attributes, {"luck_roll": 0})
        points = kernel_reference.apply_modifiers(points, on("skill_point_pool"))
        if level == rules["start_level"]:
            points = kernel_reference.apply_modifiers(points, on("starting_skill_points"))
        carried = spend(carried + max(points, 0))
        if level == rules["start_level"]:
            continue
        ref = build["perks_by_level"].get(str(level))
        if ref is None:
            problems.append(f"level {level}: no perk")
            continue
        taken = kernel_reference.take_perk(sheet, held, _key(ref))
        if "error" in taken:
            problems.append(f"level {level}: {taken['error']}")
        elif not taken["taken"]:
            problems.append(f"level {level}: {_key(ref)} unmet {taken['unmet']}")
        else:
            held[_key(ref)] = taken["rank"]
    extra = sorted(set(build["perks_by_level"]) - {str(n) for n in range(rules["start_level"] + 1, rules["max_level"] + 1)})
    if extra:
        problems.append(f"perks_by_level names levels that grant no perk: {extra}")
    sheet["perks"] = held
    return problems, sheet


if __name__ == "__main__":
    for build in json.loads((DATA / "builds/core.json").read_text())["builds"]:
        found, end = play(build)
        print(build["key"], "legal" if not found else found)
        if end:
            print("   ", {k: v for k, v in end["skills"].items()})
