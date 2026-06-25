#!/usr/bin/env python3
"""Combat resolution reference resolver.

Simulates single attacks using the ruleset data (weapons, armor, effects, skills)
and the combat resolution parameters from base_resolution.json.

This is a *reference implementation* — a quick validation tool for designers
and a starting point for engine developers. It is NOT a full combat engine.

Usage:
    python ruleset/scripts/combat_reference.py [--seed N] [--verbose]

Run with --verbose to see detailed per-step resolution.
"""

import argparse
import json
import random
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


class CombatResolver:
    """Minimal combat resolution using ruleset data only."""

    def __init__(self, seed=None, verbose=False):
        if seed is not None:
            random.seed(seed)
        self.verbose = verbose

        # Load all needed data
        self.weapons_data = load_json("data/equipment/weapons.json")["equipment"]
        self.armor_data = load_json("data/equipment/armor.json")["equipment"]
        self.skills_data = load_json("data/skills/core.json")["skills"]
        self.effects_data = load_json("data/effects/core.json")["effects"]
        self.damage_types_data = load_json("data/combat/damage_types.json")[
            "damage_types"
        ]
        self.resolution = load_json("data/combat/base_resolution.json")["resolution"]

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
        weapon_key,
        attacker_skill_value,
        defender_armor_dr,
        attacker_lck=5,
        attack_modifier=0,
        defender_evasion=0,
    ):
        """Resolve a single attack and return the detailed result."""
        weapon = self.get_weapon(weapon_key)
        wspec = weapon["weapon"]
        governing_skill_key = wspec.get("skill", "blades")

        # 1. Calculate target number
        base_target = attacker_skill_value + attack_modifier
        luck_bonus = attacker_lck // 2
        target = base_target + luck_bonus

        self.log(f"Weapon: {weapon['label']} ({weapon['key']})")
        self.log(f"Skill: {governing_skill_key} = {attacker_skill_value}")
        self.log(
            f"Target number: {attacker_skill_value} + {attack_modifier} (mod) + {luck_bonus} (LCK/2) = {target}"
        )
        self.log(f"Defender armor DR: {defender_armor_dr}")

        # 2. Roll
        roll = d100()
        margin = target - roll
        self.log(f"d100 roll: {roll}, margin: {margin}")

        if roll > target:
            self.log("RESULT: Miss")
            return {
                "hit": False,
                "roll": roll,
                "target": target,
                "margin": margin,
                "hit_quality": "miss",
                "damage_dealt": 0,
                "effects_applied": [],
            }

        # 3. Hit quality
        natural_crit_range = self.crit["naturalCritRange"]
        effective_crit_range = natural_crit_range + (attacker_lck // 2)
        is_natural_crit = roll <= effective_crit_range

        # Confirm critical
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

        # Hit quality based on margin
        if is_critical:
            hit_quality = "critical"
        elif margin >= self.quality["solidMarginThreshold"]:
            hit_quality = "solid"
        else:
            hit_quality = "glancing"
        self.log(f"Hit quality: {hit_quality} (margin {margin})")

        # 4. Damage calculation
        damage_dice = wspec["damage"]["dice"]
        damage_type = wspec["damage"]["type"]

        # Parse dice expression e.g. "1d8", "2d6+3"
        import re

        m = re.match(r"(\d+)d(\d+)([+-]\d+)?", damage_dice)
        if not m:
            raise ValueError(f"Cannot parse dice: {damage_dice}")
        num_dice = int(m.group(1))
        die_size = int(m.group(2))
        flat_mod = int(m.group(3)) if m.group(3) else 0

        if is_critical:
            # Double the dice (most common method)
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

        # Apply armor DR (physical only)
        dt_entry = find_by_key(self.damage_types_data, damage_type)
        armor_applies = False
        if dt_entry:
            armor_applies = "armor_dr_applies" in dt_entry.get("tags", [])
        effective_dr = defender_armor_dr if armor_applies else 0
        net_damage = max(0, raw_damage - effective_dr)

        self.log(f"Damage type: {damage_type} (armor applies: {armor_applies})")
        self.log(f"Raw: {raw_damage} - DR: {effective_dr} = Net: {net_damage}")

        # 5. Collect on-hit effects from weapon
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
                }
                effects_applied.append(applied)
                self.log(
                    f"On-hit effect: {eff['label']} (magnitude {ae.get('magnitude', 0)}, duration {ae.get('duration', 0)})"
                )

        return {
            "hit": True,
            "roll": roll,
            "target": target,
            "margin": margin,
            "hit_quality": hit_quality,
            "is_critical": is_critical,
            "raw_damage": raw_damage,
            "armor_dr_applied": effective_dr,
            "damage_dealt": net_damage,
            "damage_type": damage_type,
            "effects_applied": effects_applied,
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
            print(f"  Damage: {result['damage_dealt']} ({result['damage_type']})")
            if result["effects_applied"]:
                for e in result["effects_applied"]:
                    print(
                        f"  Effect: {e['effect_key']} (mag: {e['magnitude']}, dur: {e['duration']}s)"
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

    # Example 1: Knight with longsword vs light-armored bandit
    resolver.simulate_example(
        "Example 1: Sir Aldric (Blades 72, LCK 4) vs Bandit (DR 2)",
        weapon_key="longsword",
        attacker_skill_value=72,
        defender_armor_dr=2,
        attacker_lck=4,
        attack_modifier=0,
    )

    # Example 2: Archer with longbow vs medium armor
    resolver.simulate_example(
        "Example 2: Valeria (Bows 78, LCK 6) vs Chainmail Guard (DR 8)",
        weapon_key="longbow",
        attacker_skill_value=78,
        defender_armor_dr=8,
        attacker_lck=6,
        attack_modifier=0,
    )

    # Example 3: Rogue with rapier vs heavy plate
    resolver.simulate_example(
        "Example 3: Kestrel (Piercing 68, LCK 7) vs Plate Knight (DR 22)",
        weapon_key="rapier",
        attacker_skill_value=68,
        defender_armor_dr=22,
        attacker_lck=7,
        attack_modifier=0,
    )

    # Example 4: Power attack (high modifier) with greatsword
    resolver.simulate_example(
        "Example 4: Dorn (Blunt 68, LCK 4, Power Attack) vs Goblin (DR 1)",
        weapon_key="warhammer",
        attacker_skill_value=68,
        defender_armor_dr=1,
        attacker_lck=4,
        attack_modifier=10,
    )

    # Example 5: High-skill expert with dagger against unarmored target
    resolver.simulate_example(
        "Example 5: Assassin (Piercing 95, LCK 10) vs Unarmored Mage (DR 0)",
        weapon_key="dagger",
        attacker_skill_value=95,
        defender_armor_dr=0,
        attacker_lck=10,
        attack_modifier=0,
    )

    print("\nDone. Run with --seed N for reproducible rolls, --verbose for details.\n")


if __name__ == "__main__":
    main()
