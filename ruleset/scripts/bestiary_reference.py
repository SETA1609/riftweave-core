#!/usr/bin/env python3
"""The bestiary command's reference (docs/backend/bestiary.md, BE_S08_05): what the players know
of each monster, and the bestiary as the GM or a player sees it. Core monsters only; the Go core
also reads the enabled modules."""

import json
import sys
from functools import lru_cache
from pathlib import Path

_SCRIPTS = Path(__file__).resolve().parent
sys.path.insert(0, str(_SCRIPTS))

import validate  # noqa: E402  (base + override resolution)

OPERATION_REVEAL = "reveal"
OPERATION_VIEW = "view"
LEVEL_UNKNOWN = "unknown"
LEVEL_SEEN = "seen"
LEVEL_KNOWN = "known"
LEVEL_RANK = {LEVEL_UNKNOWN: 0, LEVEL_SEEN: 1, LEVEL_KNOWN: 2}
SHOWS_SEEN = "seen"
SHOWS_FULL = "full"
LORE_SKILL = "lore"
MONSTER_PREFIX = "core:monster/"


@lru_cache(maxsize=None)
def _monsters():
    """Every core monster by its ref, resolved through its base."""
    data = json.loads((_SCRIPTS.parent / "data" / "monsters" / "core.json").read_text())
    validate._resolve_monster_bases(data, [])
    return {MONSTER_PREFIX + row["key"]: row for row in data["monsters"]}


def _require_monster(ref):
    row = _monsters().get(ref)
    if row is None:
        raise LookupError(f'monster "{ref}" is not defined')
    return row


def _same_holder(mark, player):
    return mark.get("player") == player


def known_level(knowledge, monster, player):
    """What a player knows of a monster: the higher of the table's mark and their own. For the
    table (no player), only the table's marks count."""
    best = LEVEL_UNKNOWN
    for mark in knowledge:
        if mark["monster"] == monster and ("player" not in mark or mark.get("player") == player):
            if LEVEL_RANK[mark["level"]] > LEVEL_RANK[best]:
                best = mark["level"]
    return best


def _marked(knowledge, monster, level, player):
    """The marks with this holder's mark of the monster set to level, in place if it had one."""
    mark = {"monster": monster, "level": level}
    if player is not None:
        mark["player"] = player
    marks = [dict(existing) for existing in knowledge]
    for index, existing in enumerate(marks):
        if existing["monster"] == monster and _same_holder(existing, player):
            marks[index] = mark
            return marks
    return marks + [mark]


def _reveal(vector):
    from kernel_reference import run_resolution

    knowledge = vector.get("knowledge", [])
    monster = vector["monster"]
    player = vector.get("player")
    _require_monster(monster)
    result = {}
    check = vector.get("check")
    if check is not None:
        if player is None:
            return {"error": "a lore check needs player"}
        rolled = run_resolution({"actor": check["actor"], "skill": LORE_SKILL, "modifiers": check.get("modifiers", []), "roll": check["roll"]})
        result["check"] = rolled
        if rolled["success"]:
            knowledge = _marked(knowledge, monster, LEVEL_KNOWN, player)
    elif "level" not in vector:
        return {"error": "a reveal needs level"}
    else:
        knowledge = _marked(knowledge, monster, vector["level"], player)
    result["knowledge"] = knowledge
    result["level"] = known_level(knowledge, monster, player)
    return result


def _entry(ref, row, level, shows):
    return {"monster": ref, "label": row["label"], "description": row.get("description", ""), "size": row.get("size", ""),
            "knowledge": level, "shows": shows}


def _view(vector):
    if "monsters" not in vector:
        return {"error": "a view needs monsters"}
    knowledge = vector.get("knowledge", [])
    player = vector.get("player")
    only_known = vector.get("players_see_only_known", False)
    entries = []
    for ref in vector["monsters"]:
        row = _require_monster(ref)
        if player is None:
            entry = _entry(ref, row, known_level(knowledge, ref, None), SHOWS_FULL)
            entry["players"] = {mark["player"]: mark["level"] for mark in knowledge if mark["monster"] == ref and "player" in mark}
            entries.append(entry)
            continue
        level = known_level(knowledge, ref, player)
        if not only_known:
            entries.append(_entry(ref, row, level, SHOWS_FULL))
        elif level == LEVEL_SEEN:
            entries.append(_entry(ref, row, level, SHOWS_SEEN))
        elif level == LEVEL_KNOWN:
            entries.append(_entry(ref, row, level, SHOWS_FULL))
    return {"entries": entries}


OPERATIONS = {OPERATION_REVEAL: _reveal, OPERATION_VIEW: _view}


def run_bestiary(vector):
    """The bestiary command (docs/backend/bestiary.md): reveal a mark, or view the bestiary."""
    try:
        return OPERATIONS[vector["operation"]](vector)
    except LookupError as error:
        return {"error": error.args[0]}
