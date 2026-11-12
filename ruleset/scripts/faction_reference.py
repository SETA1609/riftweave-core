#!/usr/bin/env python3
"""Factions module reference (docs/modules/factions.md).

Reads module data and an actor snapshot; returns standings, reputation changes,
promotion eligibility, and the contract modifiers the module emits. It stores
nothing: standing, eligibility, and modifiers are recomputed from the snapshot.

Usage:
    python ruleset/scripts/faction_reference.py   # runs every vector in modules/factions/tests/vectors.json
"""

import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
DATA_FILE = ROOT / "modules" / "factions" / "data" / "factions.json"
VECTORS = ROOT / "modules" / "factions" / "tests" / "vectors.json"
REF_PREFIX = "module:factions/faction/"


def faction_ref(faction):
    return REF_PREFIX + faction["key"]


def check_factions(factions):
    """Raise ValueError on rules the schema cannot express."""
    refs = {faction_ref(f) for f in factions}
    for f in factions:
        low = f.get("reputation_min", -100)
        high = f.get("reputation_max", 100)
        start = f.get("reputation_start", 0)
        if not low <= start <= high:
            raise ValueError(f"{f['key']}: reputation_start outside bounds")
        mins = [s["min_reputation"] for s in f["standings"]]
        if mins[0] != low or mins != sorted(set(mins)):
            raise ValueError(f"{f['key']}: standings must start at reputation_min and ascend")
        rank_mins = [r["min_reputation"] for r in f.get("ranks", [])]
        if rank_mins != sorted(rank_mins):
            raise ValueError(f"{f['key']}: ranks must not descend in min_reputation")
        for group, prefix in (("standings", "faction_standing"), ("ranks", "faction_rank")):
            for row in f.get(group, []):
                for m in row["modifiers"]:
                    if m["source"] != f"{prefix}:{f['key']}":
                        raise ValueError(f"{f['key']}/{row['key']}: source {m['source']}")
        for rel in f.get("relations", []):
            if rel["faction"] not in refs or rel["faction"] == faction_ref(f):
                raise ValueError(f"{f['key']}: bad relation {rel['faction']}")


def load_factions(path=DATA_FILE):
    factions = json.loads(Path(path).read_text())["factions"]
    check_factions(factions)
    return {faction_ref(f): f for f in factions}


FACTIONS = load_factions()


def reputation_of(actor, ref):
    return actor.get("reputation", {}).get(ref, FACTIONS[ref].get("reputation_start", 0))


def standing(ref, actor):
    """Rule 1: the last band whose min_reputation <= reputation."""
    value = reputation_of(actor, ref)
    current = None
    for row in FACTIONS[ref]["standings"]:
        if row["min_reputation"] <= value:
            current = row
    return current


def _clamp(faction, value):
    return max(faction.get("reputation_min", -100), min(faction.get("reputation_max", 100), value))


def _trunc_percent(delta, percent):
    """trunc(delta * percent / 100), toward zero, integer only."""
    product = delta * percent
    quotient = abs(product) // 100
    return quotient if product >= 0 else -quotient


def apply_reputation_event(actor, event):
    """Rules 2 and 3: clamp, then one-hop spillover from the unclamped delta."""
    ref = event["faction"]
    faction = FACTIONS[ref]
    if "delta" in event:
        delta = event["delta"]
    else:
        delta = next(e["delta"] for e in faction.get("reputation_events", []) if e["key"] == event["event"])
    result = dict(actor.get("reputation", {}))
    result[ref] = _clamp(faction, reputation_of(actor, ref) + delta)
    for rel in faction.get("relations", []):
        other = rel["faction"]
        shift = _trunc_percent(delta, rel["share_percent"])
        result[other] = _clamp(FACTIONS[other], reputation_of(actor, other) + shift)
    return result


def _held_rank_index(ref, actor):
    held = actor.get("faction_ranks", {}).get(ref)
    if held is None:
        return None
    keys = [r["key"] for r in FACTIONS[ref].get("ranks", [])]
    return keys.index(held)


def can_promote(ref, actor):
    """Rule 4: the next rank after the held one, if reputation meets it."""
    ranks = FACTIONS[ref].get("ranks", [])
    index = _held_rank_index(ref, actor)
    next_index = 0 if index is None else index + 1
    if next_index >= len(ranks):
        return None, False
    nxt = ranks[next_index]
    return nxt["key"], reputation_of(actor, ref) >= nxt["min_reputation"]


def emit_modifiers(actor):
    """Rules 5 and 6: standing bundle, plus the held rank's bundle unless suspended."""
    involved = set(actor.get("reputation", {})) | set(actor.get("faction_ranks", {}))
    modifiers, suspended = [], []
    for ref, faction in FACTIONS.items():
        if ref not in involved:
            continue
        modifiers.extend(standing(ref, actor)["modifiers"])
        index = _held_rank_index(ref, actor)
        if index is None:
            continue
        rank = faction["ranks"][index]
        if reputation_of(actor, ref) >= rank["min_reputation"]:
            modifiers.extend(rank["modifiers"])
        else:
            suspended.append(ref)
    return modifiers, suspended


def operation_of(modifier):
    """A modifier's operation, spelled out, whether it names `operation` or the short `op` (M5-46)."""
    operation = modifier.get("operation", modifier.get("op"))
    return {"mul": "multiply"}.get(operation, operation)


def apply_modifiers(base, modifiers):
    """Kernel stacking (docs/core/contract.md): same source does not stack, the
    latest wins; different sources add; multiply after add; set replaces."""
    by_source = {}
    for m in modifiers:
        by_source[(m["source"], operation_of(m))] = m
    kept = list(by_source.values())
    value = base + sum(m["value"] for m in kept if operation_of(m) == "add")
    for m in kept:
        if operation_of(m) == "multiply":
            value = value * m["value"]
    for m in kept:
        if operation_of(m) == "set":
            value = m["value"]
    return value


def skill_target(actor, skill):
    """Resolution target for a skill after this module's modifiers (rule 7)."""
    modifiers, _ = emit_modifiers(actor)
    relevant = [m for m in modifiers if m["target"] == skill]
    return apply_modifiers(actor.get("skills", {}).get(skill, 0), relevant)


def run_vector(case):
    function = case["function"]
    actor = case.get("actor", {})
    if function == "standing":
        return {"standing": standing(case["faction"], actor)["key"]}
    if function == "apply_reputation_event":
        return {"reputation": apply_reputation_event(actor, case["event"])}
    if function == "can_promote":
        next_rank, allowed = can_promote(case["faction"], actor)
        return {"next_rank": next_rank, "allowed": allowed}
    if function == "emit_modifiers":
        modifiers, suspended = emit_modifiers(actor)
        return {"modifiers": modifiers, "suspended_ranks": suspended}
    if function == "skill_target":
        return {"target": skill_target(actor, case["skill"])}
    raise ValueError(f"unknown function {function}")


def main():
    failed = 0
    for case in json.loads(VECTORS.read_text()):
        got = run_vector(case)
        ok = got == case["expect"]
        failed += not ok
        print(f"  {'ok' if ok else 'FAIL'}  {case['name']}")
        if not ok:
            print(f"        expected {case['expect']}\n        got      {got}")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
