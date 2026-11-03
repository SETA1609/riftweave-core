#!/usr/bin/env python3
"""Python reference for the wuxing interaction function (docs/modules/wuxing.md).

The one place that reads the cycles in ruleset/data/wuxing/core.json, including the
reverse cycles (`reverse_of`) and the insulting condition. The vector runner and the demo
scripts (combat_reference.py, crafting_reference.py) all call resolve_interaction.
"""

import json
from pathlib import Path

WUXING_FILE = Path(__file__).resolve().parent.parent / "data" / "wuxing" / "core.json"

NO_INTERACTION = {"cycle": None, "interaction": "none", "affects": None, "multiplier": 1}
SWAPPED_ROLE = {"actor": "recipient", "recipient": "actor"}


def load_cycles(path=WUXING_FILE):
    return json.loads(Path(path).read_text())["cycles"]


def cycle_edges(cycles, cycle):
    """The (from, to) pairs of a cycle; a reverse_of cycle reads its target's edges backwards."""
    if "reverse_of" in cycle:
        target = find_cycle_by_id(cycles, cycle["reverse_of"])
        return [(to_phase, from_phase) for from_phase, to_phase in cycle_edges(cycles, target)]
    return [(edge["from"], edge["to"]) for edge in cycle["edges"]]


def find_cycle_by_id(cycles, cycle_id):
    for cycle in cycles:
        if cycle["id"] == cycle_id:
            return cycle
    raise ValueError(f"wuxing cycle {cycle_id!r} is not defined")


def find_cycle_for_pair(cycles, actor_phase, recipient_phase):
    """The one cycle whose edges contain actor_phase -> recipient_phase, or None."""
    matches = [c for c in cycles if (actor_phase, recipient_phase) in cycle_edges(cycles, c)]
    if len(matches) > 1:
        names = ", ".join(c["id"] for c in matches)
        raise ValueError(f"{actor_phase} -> {recipient_phase} is in more than one cycle: {names}")
    return matches[0] if matches else None


def interaction_result(cycle, affects):
    return {
        "cycle": cycle["id"],
        "interaction": cycle["interaction"],
        "affects": affects,
        "multiplier": cycle["magnitudeMultiplier"],
    }


def resolve_interaction(
    cycles, actor_phase, recipient_phase, actor_magnitude=None, recipient_magnitude=None
):
    """Name the cycle that fires when an actor_phase effect meets a recipient_phase one."""
    if not actor_phase or not recipient_phase or actor_phase == recipient_phase:
        return dict(NO_INTERACTION)
    cycle = find_cycle_for_pair(cycles, actor_phase, recipient_phase)
    if cycle is None:
        return dict(NO_INTERACTION)
    condition = cycle.get("condition")
    if condition is None:
        return interaction_result(cycle, cycle["affects"])
    if condition != "actor_magnitude_exceeds_recipient":
        raise ValueError(f"wuxing condition {condition!r} is not defined")
    if actor_magnitude is None or recipient_magnitude is None:
        return {"error": f"{cycle['id']} needs actor_magnitude and recipient_magnitude"}
    if actor_magnitude > recipient_magnitude:
        return interaction_result(cycle, cycle["affects"])
    fallback = find_cycle_by_id(cycles, cycle["reverse_of"])
    return interaction_result(fallback, SWAPPED_ROLE[fallback["affects"]])


def run_wuxing(vector):
    return resolve_interaction(
        load_cycles(),
        vector.get("actor_phase"),
        vector.get("recipient_phase"),
        vector.get("actor_magnitude"),
        vector.get("recipient_magnitude"),
    )


def interaction_tuple(cycles, item_phase, effect_phase, item_magnitude=None, effect_magnitude=None):
    """(cycle_id, interaction, multiplier, affects) for the log lines of the demo scripts.

    The item's phase is the actor and the effect's the recipient. Same phase is reported as
    `identity`; a conditional cycle without magnitudes is reported as no interaction.
    """
    if item_phase == effect_phase:
        return ("same_phase", "identity", 1, "none")
    result = resolve_interaction(cycles, item_phase, effect_phase, item_magnitude, effect_magnitude)
    if "error" in result:
        return ("needs_magnitudes", "none", 1, "none")
    if result["cycle"] is None:
        return ("unrelated", "none", 1, "none")
    return (result["cycle"], result["interaction"], result["multiplier"], result["affects"])
