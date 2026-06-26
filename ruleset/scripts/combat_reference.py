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

        self.crit = self.resolution["critical"]
        self.quality = self.resolution["hitQuality"]
        self.defense_cfg = self.resolution["defense"]

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

        base_target = attacker_skill_value + attack_modifier + effective_attack_bonus
        luck_bonus = attacker_lck // 2
        target = base_target + luck_bonus

        self.log(
            f"Target number: {attacker_skill_value} + {attack_modifier} (mod) "
            f"+ {effective_attack_bonus} (atk bonus) + {luck_bonus} (LCK/2) = {target}"
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
        }

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

    print("\nDone. Run with --seed N for reproducible rolls, --verbose for details.\n")


if __name__ == "__main__":
    main()
