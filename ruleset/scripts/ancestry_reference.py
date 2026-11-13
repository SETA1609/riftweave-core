#!/usr/bin/env python3
"""What a race, its traits, and a background emit, as contract modifiers
(docs/modules/race.md#5a-what-a-race-emits, docs/modules/backgrounds.md).

abilityModifiers and skill_bonuses are the compact forms of `add` modifiers. A subrace emits its
parent's modifiers and traits, then its own; a kin emits only its own; a template is never played.
"""

import json
from pathlib import Path

DATA_DIR = Path(__file__).resolve().parent.parent / "data"
INNATE = "core:origin/innate"
TRAINED = "core:origin/trained"


def load_data(rel_path):
    return json.loads((DATA_DIR / rel_path).read_text())


def _rows(rel, collection):
    return {row["key"]: row for row in load_data(rel)[collection]}


def _own(race, features):
    """One race row's own modifiers: its abilityModifiers, then its traits'."""
    source = f"race:{race['key']}"
    mods = [
        {"target": attr, "operation": "add", "value": value, "source": source, "origin": INNATE}
        for attr, value in race.get("abilityModifiers", {}).items()
    ]
    for trait in race.get("traits", []):
        mods.extend(dict(m) for m in features[trait["id"].rsplit("/", 1)[-1]].get("modifiers", []))
    return mods


def race_modifiers(key):
    races = _rows("races/core.json", "races")
    features = _rows("features/core.json", "features")
    race = races[key]
    role = race["lineage"]["role"]
    if role == "template":
        raise ValueError(f"{key} is a template and is never played")
    if role == "subrace":
        parent = races[race["lineage"]["parentRace"].rsplit("/", 1)[-1]]
        return _own(parent, features) + _own(race, features)
    return _own(race, features)


def background_modifiers(key):
    background = _rows("backgrounds/core.json", "backgrounds")[key]
    source = f"background:{key}"
    return [
        {"target": b["skill"].rsplit("/", 1)[-1], "operation": "add", "value": b["bonus"], "source": source, "origin": TRAINED}
        for b in background.get("skill_bonuses", [])
    ]


def ancestry_modifiers(race, background=None):
    """The modifiers a character starts with from its race and, if given, its background."""
    mods = race_modifiers(race)
    if background is not None:
        mods += background_modifiers(background)
    return mods
