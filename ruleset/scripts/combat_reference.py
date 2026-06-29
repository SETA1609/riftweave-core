#!/usr/bin/env python3
"""Combat resolution reference resolver.

Simulates single attacks using crafted items as the primary input,
resolving base weapons, phase interactions, Wuxing cycles,
enchantments, and temporary coatings.

Usage:
    python ruleset/scripts/combat_reference.py [--seed N] [--verbose]

Run with --verbose to see detailed per-step resolution.
"""

import argparse
import json
import random
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
DATA_DIR = ROOT / "data"
SCHEMA_DIR = ROOT / "schemas"


def load_json(rel_path):
    fp = ROOT / rel_path
    if not fp.exists():
        raise FileNotFoundError(f"Data file not found: {fp}")
    return json.loads(fp.read_text())


def find_by_key(collection, key):
    for item in collection:
        if item.get("key") == key:
            return item
    return None


def d100():
    return random.randint(1, 100)


def find_cycle(cycles, from_phase, to_phase):
    """Find a Wuxing cycle where from_phase acts upon to_phase."""
    for cycle in cycles:
        for edge in cycle["edges"]:
            if edge["from"] == from_phase and edge["to"] == to_phase:
                return (
                    cycle["id"],
                    cycle["interaction"],
                    cycle["magnitudeMultiplier"],
                    cycle["affects"],
                )
    return None


def get_interaction(cycles, item_phase, effect_phase):
    """Look up the Wuxing interaction between item phase and effect phase."""
    if item_phase == effect_phase:
        return ("same_phase", "identity", 1.0, "none")
    result = find_cycle(cycles, item_phase, effect_phase)
    if result:
        return result
    result = find_cycle(cycles, effect_phase, item_phase)
    if result:
        cycle_id, interaction, mult, affects = result
        return (f"reverse_{cycle_id}", f"reverse_{interaction}", 1.0 / mult, affects)
    return ("unrelated", "none", 1.0, "none")


class CombatResolver:
    """Minimal combat resolution using crafted items as primary input."""

    def __init__(self, seed=None, verbose=False):
        if seed is not None:
            random.seed(seed)
        self.verbose = verbose

        self.weapons_data = load_json("data/equipment/weapons.json")["equipment"]
        self.armor_data = load_json("data/equipment/armor.json")["equipment"]
        self.skills_data = load_json("data/skills/core.json")["skills"]
        self.effects_data = load_json("data/effects/core.json")["effects"]
        self.damage_types_data = load_json("data/combat/damage_types.json")[
            "damage_types"
        ]
        self.resolution = load_json("data/combat/base_resolution.json")["resolution"]
        self.crafted_items = load_json("data/crafting/crafted_items.json")[
            "crafted_items"
        ]
        self.wuxing = load_json("data/wuxing/core.json")
        self.cycles = self.wuxing["cycles"]
        self.conditions_data = load_json("data/conditions/core.json")["conditions"]

        self.crit = self.resolution["critical"]
        self.quality = self.resolution["hitQuality"]
        self.defense_cfg = self.resolution["defense"]

        # Condition combat modifiers are built dynamically from
        # conditions/core.json → combat object in each entry.
        # See _build_combat_modifiers() below.
        self._build_condition_mappings()
        self._build_combat_modifiers()

        # TODO: Extract ConditionManager class
        # The condition tracking + mechanical impact methods below are a
        # natural seam for extraction. A dedicated ConditionManager would
        # own _build_condition_mappings, _build_combat_modifiers,
        # reset/apply/dispel/advance, and all _get_condition_* /
        # _can_take_action / _is_auto_crit_target methods.
        # CombatResolver would delegate to it via self.conditions.apply(...)
        # etc.  For now, inline is fine — extraction is ~50 lines of
        # boilerplate and doesn't change behavior.

        self.reset_conditions()

    def reset_conditions(self):
        self.active_conditions = {}
        self.simulation_time = 0

    def get_skill(self, skill_key):
        s = find_by_key(self.skills_data, skill_key)
        if not s:
            raise ValueError(f"Unknown skill: {skill_key}")
        return s

    def get_weapon(self, weapon_key):
        w = find_by_key(self.weapons_data, weapon_key)
        if not w:
            raise ValueError(f"Unknown weapon: {weapon_key}")
        return w

    def get_crafted_item(self, key):
        ci = find_by_key(self.crafted_items, key)
        if not ci:
            raise ValueError(f"Unknown crafted item: {key}")
        return ci

    def get_effect(self, effect_id):
        for e in self.effects_data:
            if e["id"] == effect_id:
                return e
        return None

    def get_condition(self, condition_key):
        return find_by_key(self.conditions_data, condition_key)

    def _build_condition_mappings(self):
        self.effect_to_conditions = {}
        self.effect_to_dispels = {}
        for cond in self.conditions_data:
            for eid in cond.get("appliedBy", []):
                self.effect_to_conditions.setdefault(eid, []).append(
                    (cond["key"], cond)
                )
            for eid in cond.get("removedBy", []):
                self.effect_to_dispels.setdefault(eid, []).append((cond["key"], cond))

    def _build_combat_modifiers(self):
        """Build combat modifier maps from conditions/core.json → combat object.
        Replaces previously hardcoded dictionaries with data-driven lookups.
        """
        self.attack_penalties = {}
        self.defense_bonuses = {}
        self.cannot_act = set()
        self.auto_crit_targets = set()
        self.attack_penalty_per_stack = {}

        for cond in self.conditions_data:
            cmb = cond.get("combat")
            if not cmb:
                continue
            key = cond["key"]
            ap = cmb.get("attack_penalty")
            if ap:
                self.attack_penalties[key] = ap
            db = cmb.get("defense_bonus")
            if db:
                self.defense_bonuses[key] = db
            if cmb.get("prevents_action"):
                self.cannot_act.add(key)
            if cmb.get("auto_crit_when_target"):
                self.auto_crit_targets.add(key)
            aps = cmb.get("attack_penalty_per_stack")
            if aps:
                self.attack_penalty_per_stack[key] = aps

    def _get_duration(self, effect_id, dur):
        if dur and dur > 0:
            return dur
        eff = self.get_effect(effect_id)
        if eff:
            return eff.get("defaultDuration", 0)
        return 0

    def _get_duration_type(self, effect_id, dur):
        eff = self.get_effect(effect_id)
        if eff:
            tags = eff.get("tags", [])
            if "disease" in tags:
                return "permanent"
            e_dtype = eff.get("defaultDurationType")
            if e_dtype == "unlimited":
                return "permanent"
            if dur and dur > 0:
                return "timed"
            if eff.get("defaultDuration", 0) > 0:
                return "timed"
        if dur and dur > 0:
            return "timed"
        return "until_dispelled"

    def _list_active_conditions(self):
        if not self.active_conditions:
            self.log("  Active conditions: none")
            return
        self.log(f"  Active conditions ({len(self.active_conditions)}):")
        for ck, entry in self.active_conditions.items():
            dur = entry["remaining_duration"]
            dtype = entry["duration_type"]
            if dtype == "permanent":
                dstr = "permanent"
            elif dtype == "until_dispelled":
                dstr = "until_dispelled"
            else:
                dstr = f"{dur}s remaining"
            self.log(
                f"    [{entry['label']}] ({dstr}, source: {entry['source_label']})"
            )

    def _check_immunity(self, condition_key, source_effect_id):
        """Stub: check if the subject is immune or resistant to this condition.
        TODO: wire into resist effect (id 19) with parameter matching condition_key.
        A future full implementation should check:
          - self.active_effects (resist id 19 with parameter=condition_key at 100%)
          - Racial traits granting immunity via appliedBy on resist effect
        """
        if self.verbose:
            self.log(
                f"  [IMMUNITY CHECK] {condition_key} — no immunity system wired yet"
            )
        return False

    def apply_condition(
        self,
        condition_key,
        source_effect_id,
        duration=0,
        source_label="effect",
        parameter=None,
    ):
        cond = self.get_condition(condition_key)
        if not cond:
            self.log(f"  Unknown condition: {condition_key}")
            return False

        # Immunity check (stub — always passes for now)
        if self._check_immunity(condition_key, source_effect_id):
            self.log(f"  -> Condition [{cond['label']}] resisted (immune)")
            return False

        stacking = cond.get("stacking", False)
        max_stacks = cond.get("maxStacks", 1)

        if condition_key in self.active_conditions:
            if stacking:
                current_stacks = self.active_conditions[condition_key].get("stacks", 1)
                new_stacks = min(current_stacks + 1, max_stacks)
                self.active_conditions[condition_key]["stacks"] = new_stacks
                if new_stacks > current_stacks:
                    self.active_conditions[condition_key]["remaining_duration"] = (
                        self._get_duration(source_effect_id, duration)
                    )
                self.log(
                    f"  -> Condition [{cond['label']}] stacked"
                    f" {current_stacks} -> {new_stacks}/{max_stacks}"
                )
                return True
            else:
                self.log(f"  -> Condition [{cond['label']}] already active, refreshing")
                self.active_conditions[condition_key]["remaining_duration"] = (
                    self._get_duration(source_effect_id, duration)
                )
                return True

        dur = self._get_duration(source_effect_id, duration)
        dtype = self._get_duration_type(source_effect_id, dur)
        remaining = dur if dtype == "timed" else 0

        entry = {
            "key": condition_key,
            "label": cond["label"],
            "applied_by_effect": source_effect_id,
            "duration_type": dtype,
            "remaining_duration": remaining,
            "applied_at": self.simulation_time,
            "source_label": source_label,
            "stacking": stacking,
            "stacks": 1 if stacking else None,
        }
        self.active_conditions[condition_key] = entry

        if dtype == "permanent":
            dstr = "permanent"
        elif dtype == "until_dispelled":
            dstr = "until_dispelled"
        else:
            dstr = f"{remaining}s"
        self.log(f"  -> Condition [{cond['label']}] applied (duration: {dstr})")
        return True

    def try_dispel_condition(
        self, condition_key, dispelling_effect_id, source_label="effect"
    ):
        cond = self.get_condition(condition_key)
        if not cond:
            return False
        if condition_key not in self.active_conditions:
            return False
        removed_by = cond.get("removedBy", [])
        if dispelling_effect_id in removed_by:
            label = self.active_conditions[condition_key]["label"]
            del self.active_conditions[condition_key]
            self.log(f"  -> Condition [{label}] dispelled by {source_label}")
            return True
        return False

    def _advance_time(self, seconds):
        self.simulation_time += seconds
        expired = []
        for ck, entry in self.active_conditions.items():
            if entry["duration_type"] != "timed":
                continue
            entry["remaining_duration"] -= seconds
            if entry["remaining_duration"] <= 0:
                expired.append(ck)
        for ck in expired:
            label = self.active_conditions[ck]["label"]
            del self.active_conditions[ck]
            self.log(f"  -> Condition [{label}] expired (duration elapsed)")

    def _process_effect_conditions(self, eff, source_label, parameter=None):
        """Dispatch condition dispels then applies for an effect.
        Args:
            eff: effect dict from the registry
            source_label: human-readable source tag for logging
            parameter: optional string from the appliedEffect shape; used for
                       targeted dispel (cure with parameter='poisoned' only
                       dispels poisoned, not all conditions in removedBy)
        """
        if not eff:
            return
        eid = eff["id"]
        for ckey, cond in self.effect_to_dispels.get(eid, []):
            # Parameter filtering: cure (id 20) with parameter only dispels
            # matching conditions or "all"
            if parameter is not None and eid == 20:
                if parameter != "all" and parameter != ckey:
                    continue
            # restore_resource (id 7) dispels bleeding/poisoned as a side
            # effect; parameter about resource type, not condition key,
            # so no filtering needed for it
            self.try_dispel_condition(ckey, eid, source_label)
        for ckey, cond in self.effect_to_conditions.get(eid, []):
            default_dur = eff.get("defaultDuration", 0)
            self.apply_condition(ckey, eid, default_dur, source_label, parameter)

    # --- Condition Mechanical Impact Methods ---

    def _get_condition_attack_modifier(self):
        """Return total attack roll penalty/bonus from active conditions.
        Data source: conditions/core.json → combat.attack_penalty / attack_penalty_per_stack
        """
        modifier = 0
        detail = []
        for ck, entry in self.active_conditions.items():
            flat_penalty = self.attack_penalties.get(ck, 0)
            if flat_penalty:
                detail.append(f"{ck}: {flat_penalty}")
                modifier += flat_penalty
            per_stack = self.attack_penalty_per_stack.get(ck, 0)
            if per_stack and entry.get("stacking"):
                stacks = min(entry.get("stacks", 0), 6)
                sp = stacks * per_stack
                if sp:
                    detail.append(f"{ck} x{stacks}: {sp}")
                    modifier += sp
        if detail:
            self.log(f"  Condition attack modifier: {' + '.join(detail)} = {modifier}")
        return modifier

    def _get_condition_defense_bonus(self):
        """Return bonus TO attack rolls made AGAINST the subject.
        Data source: conditions/core.json → combat.defense_bonus
        """
        bonus = 0
        detail = []
        for ck, entry in self.active_conditions.items():
            b = self.defense_bonuses.get(ck, 0)
            if b:
                detail.append(f"{ck}: +{b} to attackers")
                bonus += b
        if detail:
            self.log(f"  Condition defense bonus: {', '.join(detail)}")
        return bonus

    def _can_take_action(self):
        """Return False if any active condition prevents the subject from acting.
        Data source: conditions/core.json → combat.prevents_action
        """
        for ck in self.active_conditions:
            if ck in self.cannot_act:
                label = self.active_conditions[ck]["label"]
                self.log(f"  -> Cannot act: [{label}] prevents all actions")
                return False
        return True

    def _is_auto_crit_target(self):
        """Return True if active conditions make the subject auto-crit by melee.
        Data source: conditions/core.json → combat.auto_crit_when_target
        """
        for ck in self.active_conditions:
            if ck in self.auto_crit_targets:
                return True
        return False

    def _get_condition_intensity(self, key):
        """Return stack count for a condition (0 if not active)."""
        entry = self.active_conditions.get(key)
        if not entry:
            return 0
        if entry.get("stacking"):
            return entry.get("stacks", 1)
        return 1

    # --- End Condition Methods ---

    def log(self, msg):
        if self.verbose:
            print(f"  {msg}")

    def resolve_attack(
        self,
        crafted_item_key: str,
        attacker_skill_value: int = 0,
        defender_armor_dr: int = 0,
        attacker_lck: int = 5,
        attack_modifier: int = 0,
        defender_evasion: int = 0,
    ):
        """Resolve a single attack using a crafted item as primary input."""
        self.reset_conditions()
        item = self.get_crafted_item(crafted_item_key)
        base_weapon = self.get_weapon(item["base_key"])
        wspec = base_weapon["weapon"]
        governing_skill_key = wspec.get("skill", "blades")

        effective_phase = item.get("phase")
        phase_source = "material (default)"
        if item.get("engraving_jewel"):
            effective_phase = item["engraving_jewel"]["phase"]
            phase_source = f"engraving_jewel ({item['engraving_jewel']['gem_key']})"

        effective_attack_bonus = item.get("attack_bonus", 0)
        eff_mag_mult = item.get("effect_magnitude_mult", 1.0)

        self.log(f"Crafted item: {item['label']} ({item['key']})")
        self.log(f"Base weapon: {base_weapon['label']} ({base_weapon['key']})")
        self.log(f"Effective phase: {effective_phase} (from {phase_source})")
        self.log(f"Attack bonus: +{effective_attack_bonus}")
        self.log(f"Effect magnitude mult: {eff_mag_mult}")
        self.log(f"Governing skill: {governing_skill_key} = {attacker_skill_value}")

        condition_attack_mod = self._get_condition_attack_modifier()
        base_target = (
            attacker_skill_value
            + attack_modifier
            + effective_attack_bonus
            + condition_attack_mod
        )
        luck_bonus = attacker_lck // 2
        target = base_target + luck_bonus

        cond_mod_str = f"{condition_attack_mod:+d}" if condition_attack_mod else "0"
        self.log(
            f"Target number: {attacker_skill_value} + {attack_modifier} (mod) "
            f"+ {effective_attack_bonus} (atk bonus)"
            f" + {cond_mod_str} (conditions)"
            f" + {luck_bonus} (LCK/2) = {target}"
        )
        self.log(f"Defender armor DR: {defender_armor_dr}")

        roll = d100()
        margin = target - roll
        self.log(f"d100 roll: {roll}, margin: {margin}")

        if roll > target:
            self.log("RESULT: Miss")
            return {
                "hit": False,
                "crafted_item_key": crafted_item_key,
                "roll": roll,
                "target": target,
                "margin": margin,
                "hit_quality": "miss",
                "damage_dealt": 0,
                "effects_applied": [],
                "coatings_consumed": [],
                "conditions_active": {},
                "condition_attack_modifier": condition_attack_mod,
                "condition_defense_bonus": self._get_condition_defense_bonus(),
                "can_act": self._can_take_action(),
            }

        natural_crit_range = self.crit["naturalCritRange"]
        effective_crit_range = natural_crit_range + (attacker_lck // 2)
        is_natural_crit = roll <= effective_crit_range

        is_critical = False
        if is_natural_crit:
            confirm_roll = d100()
            confirm_target = target + (attacker_lck // 2)
            is_critical = confirm_roll <= confirm_target
            self.log(
                f"Natural crit (roll {roll} <= {effective_crit_range})! "
                f"Confirmation: {confirm_roll} vs {confirm_target} "
                f"-> {'CRITICAL' if is_critical else 'normal hit'}"
            )
            if not is_critical:
                self.log("  (confirmation failed, treated as normal hit)")

        if is_critical:
            hit_quality = "critical"
        elif margin >= self.quality["solidMarginThreshold"]:
            hit_quality = "solid"
        else:
            hit_quality = "glancing"
        self.log(f"Hit quality: {hit_quality} (margin {margin})")

        damage_dice = wspec["damage"]["dice"]
        damage_type = wspec["damage"]["type"]

        m = re.match(r"(\d+)d(\d+)([+-]\d+)?", damage_dice)
        if not m:
            raise ValueError(f"Cannot parse dice: {damage_dice}")
        num_dice = int(m.group(1))
        die_size = int(m.group(2))
        flat_mod = int(m.group(3)) if m.group(3) else 0

        if is_critical:
            raw_damage = (
                sum(random.randint(1, die_size) for _ in range(num_dice * 2)) + flat_mod
            )
            self.log(
                f"Critical damage: {num_dice * 2}d{die_size}+{flat_mod} = {raw_damage}"
            )
        else:
            raw_damage = (
                sum(random.randint(1, die_size) for _ in range(num_dice)) + flat_mod
            )
            self.log(f"Damage roll: {damage_dice} = {raw_damage}")

        dt_entry = find_by_key(self.damage_types_data, damage_type)
        armor_applies = False
        if dt_entry:
            armor_applies = "armor_dr_applies" in dt_entry.get("tags", [])
        effective_dr = defender_armor_dr if armor_applies else 0
        net_damage = max(0, raw_damage - effective_dr)

        self.log(f"Damage type: {damage_type} (armor applies: {armor_applies})")
        self.log(f"Raw: {raw_damage} - DR: {effective_dr} = Net: {net_damage}")

        effects_applied = []
        on_hit = wspec.get("on_hit_effects", [])
        for ae in on_hit:
            eff = self.get_effect(ae["effect"])
            if eff:
                applied = {
                    "effect_id": ae["effect"],
                    "effect_key": eff.get("key", "unknown"),
                    "magnitude": ae.get("magnitude", 0),
                    "duration": ae.get("duration", 0),
                    "source": "base_weapon",
                }
                effects_applied.append(applied)
                self.log(
                    f"Weapon on-hit: {eff['label']} (mag {ae.get('magnitude', 0)}, "
                    f"dur {ae.get('duration', 0)})"
                )
                self._process_effect_conditions(
                    eff, "base weapon on-hit", ae.get("parameter")
                )

        item_on_hit = item.get("on_hit_effects", [])
        for ae in item_on_hit:
            eff = self.get_effect(ae["effect"])
            if eff:
                applied = {
                    "effect_id": ae["effect"],
                    "effect_key": eff.get("key", "unknown"),
                    "magnitude": ae.get("magnitude", 0),
                    "duration": ae.get("duration", 0),
                    "source": "crafted_item",
                }
                effects_applied.append(applied)
                self.log(
                    f"Crafted item on-hit: {eff['label']} (mag {ae.get('magnitude', 0)}, "
                    f"dur {ae.get('duration', 0)})"
                )
                self._process_effect_conditions(
                    eff, "crafted item on-hit", ae.get("parameter")
                )

        enchantments = item.get("enchantments", [])
        for ae in enchantments:
            eff = self.get_effect(ae["effect"])
            if eff:
                mag = ae.get("magnitude", 0)
                dur = ae.get("duration", 0)
                modified_mag = mag * eff_mag_mult if mag else 0

                eff_phase = eff.get("phase")
                wuxing_msg = ""
                final_mag = modified_mag
                if eff_phase and effective_phase:
                    cycle_id, interaction_name, mult, affects = get_interaction(
                        self.cycles, effective_phase, eff_phase
                    )
                    if interaction_name == "identity":
                        wuxing_msg = f"(same phase {effective_phase}: identity x{mult})"
                    elif interaction_name != "none":
                        final_mag = modified_mag * mult
                        wuxing_msg = (
                            f"({interaction_name}: {effective_phase}"
                            f" -> {eff_phase} = x{mult})"
                        )
                    else:
                        wuxing_msg = (
                            f"(unrelated: {effective_phase} vs {eff_phase} = x{mult})"
                        )
                elif eff_phase and not effective_phase:
                    wuxing_msg = (
                        f"(item has no phase; effect has {eff_phase}, no interaction)"
                    )
                elif effective_phase and not eff_phase:
                    wuxing_msg = (
                        f"(effect has no phase; item is {effective_phase},"
                        f" no interaction)"
                    )
                else:
                    wuxing_msg = "(neither has phase)"

                applied = {
                    "effect_id": ae["effect"],
                    "effect_key": eff.get("key", "unknown"),
                    "magnitude": modified_mag,
                    "final_magnitude": final_mag,
                    "duration": dur,
                    "source": "enchantment",
                    "wuxing": wuxing_msg,
                }
                effects_applied.append(applied)

                self.log(
                    f"Enchantment: {eff['label']} (base mag {mag}"
                    + (
                        f", x{eff_mag_mult} material mult -> {modified_mag}"
                        if eff_mag_mult != 1.0
                        else ""
                    )
                    + f")"
                )
                if wuxing_msg:
                    self.log(f"  Wuxing: {wuxing_msg} -> effective mag={final_mag}")
                self._process_effect_conditions(eff, "enchantment", ae.get("parameter"))

        coatings_consumed = []
        coatings = item.get("coatings", [])
        for c in coatings:
            eff = self.get_effect(c["effect"])
            if eff:
                mag = c.get("magnitude", 0)
                uses = c.get("uses_left", 1)

                eff_phase = eff.get("phase")
                wuxing_msg = ""
                final_mag = mag
                if eff_phase and effective_phase:
                    cycle_id, interaction_name, mult, affects = get_interaction(
                        self.cycles, effective_phase, eff_phase
                    )
                    if interaction_name == "identity":
                        wuxing_msg = f"(same phase {effective_phase}: identity x{mult})"
                    elif interaction_name != "none":
                        final_mag = mag * mult
                        wuxing_msg = (
                            f"({interaction_name}: {effective_phase}"
                            f" -> {eff_phase} = x{mult})"
                        )
                    else:
                        wuxing_msg = (
                            f"(unrelated: {effective_phase} vs {eff_phase} = x{mult})"
                        )
                elif eff_phase and not effective_phase:
                    wuxing_msg = (
                        f"(item has no phase; effect has {eff_phase}, no interaction)"
                    )
                elif effective_phase and not eff_phase:
                    wuxing_msg = (
                        f"(effect has no phase; item is {effective_phase},"
                        f" no interaction)"
                    )
                else:
                    wuxing_msg = "(neither has phase)"

                applied = {
                    "effect_id": c["effect"],
                    "effect_key": eff.get("key", "unknown"),
                    "magnitude": mag,
                    "final_magnitude": final_mag,
                    "duration": c.get("duration", 0),
                    "uses_left": uses,
                    "source": "coating",
                    "wuxing": wuxing_msg,
                }
                effects_applied.append(applied)
                coatings_consumed.append(applied)

                self.log(f"Coating: {eff['label']} (mag {mag}, uses {uses})")
                if wuxing_msg:
                    self.log(f"  Wuxing: {wuxing_msg} -> effective mag={final_mag}")
                self.log(f"  -> Coating consumed (uses left: {uses - 1})")
                self._process_effect_conditions(eff, "coating", c.get("parameter"))

        # --- Condition mechanical impact post-resolution ---
        # If this attack applied a target-auto-crit condition (paralyzed,
        # unconscious) and the hit wasn't already critical, upgrade it.
        if self._is_auto_crit_target() and not is_critical:
            is_critical = True
            hit_quality = "critical"
            self.log(
                "  -> AUTO-CRIT: target is paralyzed/unconscious,"
                " attack upgraded to critical!"
            )
            raw_damage = (
                sum(random.randint(1, die_size) for _ in range(num_dice * 2)) + flat_mod
            )
            self.log(
                f"  Re-rolled as critical: {num_dice * 2}d{die_size}+{flat_mod}"
                f" = {raw_damage}"
            )
            net_damage = max(0, raw_damage - effective_dr)
            self.log(f"  Net damage after auto-crit upgrade: {net_damage}")

        can_act = self._can_take_action()
        cond_def_bonus = self._get_condition_defense_bonus()
        if not can_act:
            self.log(
                "  [CONDITION] Subject cannot act (incapacitating condition active)"
            )

        return {
            "hit": True,
            "crafted_item_key": crafted_item_key,
            "roll": roll,
            "target": target,
            "margin": margin,
            "hit_quality": hit_quality,
            "is_critical": is_critical,
            "raw_damage": raw_damage,
            "armor_dr_applied": effective_dr,
            "damage_dealt": net_damage,
            "damage_type": damage_type,
            "effective_phase": effective_phase,
            "attack_bonus_applied": effective_attack_bonus,
            "effects_applied": effects_applied,
            "coatings_consumed": coatings_consumed,
            "conditions_active": dict(self.active_conditions),
            "condition_attack_modifier": condition_attack_mod,
            "condition_defense_bonus": cond_def_bonus,
            "can_act": can_act,
        }

    def demo_condition_interactions(self):
        self.reset_conditions()
        print(f"\n{'=' * 60}")
        print("  Condition Interaction Demo (Active Tracking + Mechanics)")
        print(f"{'=' * 60}")
        print("")
        print("  --- Phase 1: Apply multiple conditions ---")
        self.apply_condition("bleeding", 58, source_label="bleed weapon")
        self.apply_condition("burning", 57, source_label="fire enchantment")
        self.apply_condition("slowed", 59, source_label="frost trap")
        self._list_active_conditions()
        print("")
        print("  --- Phase 2: Exhaustion stacking ---")
        self.apply_condition("exhaustion", 40, source_label="blight")
        self.apply_condition("exhaustion", 40, source_label="blight")
        self.apply_condition("exhaustion", 40, source_label="blight")
        self._list_active_conditions()
        intensity = self._get_condition_intensity("exhaustion")
        atk_mod = self._get_condition_attack_modifier()
        print(f"  Exhaustion intensity: {intensity}/6 | Attack modifier: {atk_mod}")
        self.log("  Adding 2 more exhaustion stacks...")
        self.apply_condition("exhaustion", 40, source_label="blight")
        self.apply_condition("exhaustion", 40, source_label="blight")
        self._list_active_conditions()
        intensity = self._get_condition_intensity("exhaustion")
        atk_mod = self._get_condition_attack_modifier()
        print(f"  Exhaustion intensity: {intensity}/6 | Attack modifier: {atk_mod}")
        print("")
        print("  --- Phase 3: Opposite-effect dispels ---")
        self.log("  Healing (restore_resource) dispels bleeding:")
        self.try_dispel_condition("bleeding", 7, "restore_resource")
        self.log("  Haste (haste_attack) dispels slowed:")
        self.try_dispel_condition("slowed", 70, "haste_attack")
        self._list_active_conditions()
        print("")
        print("  --- Phase 4: Cure dispel ---")
        self.log("  Cure effect dispels burning:")
        self.try_dispel_condition("burning", 20, "cure")
        self._list_active_conditions()
        print("")
        print("  --- Phase 5: Apply timed condition + advance time ---")
        self.apply_condition(
            "paralyzed", 43, duration=8, source_label="paralytic poison"
        )
        self._list_active_conditions()
        print("  Mechanical impact check:")
        print(f"    Can act: {self._can_take_action()}")
        print(f"    Auto-crit target: {self._is_auto_crit_target()}")
        self.log("  Advancing time by 6 seconds...")
        self._advance_time(6)
        self._list_active_conditions()
        self.log("  Advancing time by 3 seconds...")
        self._advance_time(3)
        self._list_active_conditions()
        print("")
        print("  --- Phase 6: Condition attack/defense modifiers ---")
        self.reset_conditions()
        self.apply_condition("prone", 53, duration=5, source_label="knockdown")
        self.apply_condition("restrained", 54, duration=10, source_label="net")
        self._list_active_conditions()
        atk_mod = self._get_condition_attack_modifier()
        def_bonus = self._get_condition_defense_bonus()
        print(f"    Attack modifier: {atk_mod} (prone -20, restrained -20)")
        print(f"    Defense bonus vs subject: +{def_bonus} (restrained +10)")
        print()
        print("  --- Phase 7: Prone not removed by generic cure ---")
        self.log("  Cure (id 20) tries to dispel prone:")
        result = self.try_dispel_condition("prone", 20, "cure")
        self.log(f"  Result: {'dispelled' if result else 'not removed'}")
        self.log("  Prone requires the 'stand' action, not cure.")
        self._list_active_conditions()
        print(f"{'=' * 60}")

    def simulate_example(self, label, **kwargs):
        print(f"\n{'=' * 60}")
        print(f"  {label}")
        print(f"{'=' * 60}")
        result = self.resolve_attack(**kwargs)
        print(f"\n  Result: {'HIT' if result['hit'] else 'MISS'}")
        if result["hit"]:
            qual = result["hit_quality"].upper()
            print(f"  Quality: {qual}")
            print(f"  Phase: {result['effective_phase']}")
            print(
                f"  Damage: {result['damage_dealt']} ({result['damage_type']})"
                f" [+{result['attack_bonus_applied']} atk bonus]"
            )
            if result["effects_applied"]:
                for e in result["effects_applied"]:
                    mag_str = f"final mag={e.get('final_magnitude', e['magnitude'])}"
                    print(
                        f"  Effect [{e['source']}]: {e['effect_key']}"
                        f" ({mag_str}, dur: {e['duration']}s)"
                    )
            active = result.get("conditions_active", {})
            if active:
                print(f"  Conditions active after hit:")
                for ck, entry in active.items():
                    dur = entry["remaining_duration"]
                    dtype = entry["duration_type"]
                    if dtype == "permanent":
                        dstr = "permanent"
                    elif dtype == "until_dispelled":
                        dstr = "until dispelled"
                    else:
                        dstr = f"{dur}s"
                    print(f"    [{entry['label']}] ({dstr})")
        print(f"{'=' * 60}")


def main():
    parser = argparse.ArgumentParser(description="Riftweave Combat Reference Resolver")
    parser.add_argument(
        "--seed", type=int, default=None, help="Random seed for reproducibility"
    )
    parser.add_argument(
        "--verbose", "-v", action="store_true", help="Detailed step-by-step output"
    )
    args = parser.parse_args()

    resolver = CombatResolver(seed=args.seed, verbose=args.verbose)

    resolver.simulate_example(
        "Example 1: Steel Longsword (crafted)",
        crafted_item_key="steel_longsword",
        attacker_skill_value=72,
        defender_armor_dr=4,
        attacker_lck=5,
    )

    resolver.simulate_example(
        "Example 2: Ruby-engraved Steel Longsword (phase overwrite + burn)",
        crafted_item_key="ruby_engraved_steel_longsword",
        attacker_skill_value=75,
        defender_armor_dr=5,
        attacker_lck=5,
    )

    resolver.simulate_example(
        "Example 3: Poisoned Obsidian Dagger (coatings + high attack bonus)",
        crafted_item_key="poisoned_obsidian_dagger",
        attacker_skill_value=82,
        defender_armor_dr=2,
        attacker_lck=6,
    )

    resolver.simulate_example(
        "Example 4: Sapphire-engraved Elven Bow (frost enchantment + Wuxing)",
        crafted_item_key="sapphire_engraved_elven_bow",
        attacker_skill_value=78,
        defender_armor_dr=6,
        attacker_lck=5,
    )

    resolver.simulate_example(
        "Example 5: Condition Application via Poisoned Dagger",
        crafted_item_key="poisoned_obsidian_dagger",
        attacker_skill_value=85,
        defender_armor_dr=2,
        attacker_lck=6,
    )

    if args.verbose:
        resolver.demo_condition_interactions()

    print("\nDone. Run with --seed N for reproducible rolls, --verbose for details.\n")


if __name__ == "__main__":
    main()
