#!/usr/bin/env python3
"""The one Wuxing interaction function (docs/modules/wuxing.md, "Interaction function").

Combat, crafting, and the golden vectors all call get_interaction(); nothing else reads
the cycle edges. A directed pair is found in exactly one cycle, so there is no reverse
lookup: a pair the matrix does not list is `unrelated`.
"""


def _edge_cycle(cycles, from_phase, to_phase):
    for cycle in cycles:
        for edge in cycle["edges"]:
            if edge["from"] == from_phase and edge["to"] == to_phase:
                return cycle
    return None


def get_interaction(cycles, actor_phase, recipient_phase, actor_magnitude=None, recipient_magnitude=None):
    """Return (cycle_id, interaction, multiplier, affects) for an actor phase acting on a recipient phase."""
    if not actor_phase or not recipient_phase:
        return ("unrelated", "none", 1.0, "none")
    if actor_phase == recipient_phase:
        return ("same_phase", "identity", 1.0, "none")
    cycle = _edge_cycle(cycles, actor_phase, recipient_phase)
    if cycle is None:
        return ("unrelated", "none", 1.0, "none")
    gate = cycle.get("gate")
    if gate and actor_magnitude is not None and recipient_magnitude is not None:
        if not actor_magnitude > recipient_magnitude:
            # The gate failed: the named cycle applies to the reversed pair, so the side it
            # modifies flips with it.
            fallback = next(c for c in cycles if c["id"] == gate["otherwise_cycle"])
            if _edge_cycle([fallback], recipient_phase, actor_phase) is None:
                return ("unrelated", "none", 1.0, "none")
            affects = "actor" if fallback["affects"] == "recipient" else "recipient"
            return (fallback["id"], fallback["interaction"], fallback["magnitudeMultiplier"], affects)
    return (cycle["id"], cycle["interaction"], cycle["magnitudeMultiplier"], cycle["affects"])
