#!/usr/bin/env python3
"""Crafting / Blacksmithing reference resolver.

Demonstrates:
  - Phase inheritance from material to crafted item
  - Phase overwrite via jewel engraving
  - Material attack/defense bonus propagation
  - Effect magnitude modified by material.effectMagnitude
  - Wuxing cycle interaction between item phase and effect phase
  - Distinction between permanent enchantments and consumable coatings

Usage:
    python ruleset/scripts/crafting_reference.py [--verbose]
"""

import argparse
import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent


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


def find_cycle(cycles, from_phase, to_phase):
    """Find a Wuxing cycle where from_phase acts upon to_phase.
    Returns (cycle_id, interaction, multiplier, affects) or None."""
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
    """Look up the Wuxing interaction between item phase and effect phase.
    Returns (cycle_id, interaction, multiplier, affects) or None."""
    if item_phase == effect_phase:
        return ("same_phase", "identity", 1.0, "none")
    result = find_cycle(cycles, item_phase, effect_phase)
    if result:
        return result
    # TODO: The current reverse logic (1.0 / mult) works for generating/overcoming/weakening.
    #       The "insulting" cycle has special conditions (magnitude check).
    #       Revisit when we add full insulting cycle support.
    result = find_cycle(cycles, effect_phase, item_phase)
    if result:
        cycle_id, interaction, mult, affects = result
        return (f"reverse_{cycle_id}", f"reverse_{interaction}", 1.0 / mult, affects)
    return ("unrelated", "none", 1.0, "none")


class CraftingResolver:
    """Phase propagation, Wuxing interaction, and composition for crafted items."""

    def __init__(self, verbose=False):
        self.verbose = verbose
        self.materials = load_json("data/materials/core.json")["materials"]
        self.gems = load_json("data/gems/core.json")["gems"]
        self.wuxing = load_json("data/wuxing/core.json")
        self.cycles = self.wuxing["cycles"]
        self.effects = load_json("data/effects/core.json")["effects"]
        self.crafted_items = load_json("data/crafting/crafted_items.json")[
            "crafted_items"
        ]

    def log(self, msg):
        if self.verbose:
            print(f"  {msg}")

    def get_material(self, key):
        m = find_by_key(self.materials, key)
        if not m:
            raise ValueError(f"Unknown material: {key}")
        return m

    def get_gem(self, key):
        g = find_by_key(self.gems, key)
        if not g:
            raise ValueError(f"Unknown gem: {key}")
        return g

    def get_effect(self, eid):
        return find_by_key(self.effects, eid) or next(
            (e for e in self.effects if e["id"] == eid), None
        )

    def get_crafted_item(self, key):
        ci = find_by_key(self.crafted_items, key)
        if not ci:
            raise ValueError(f"Unknown crafted item: {key}")
        return ci

    def resolve_crafted_item(self, item_key):
        """Load and resolve a crafted item: material, phase, bonuses."""
        item = self.get_crafted_item(item_key)
        material = self.get_material(item["material_key"])

        self.log(f"Item: {item['label']} ({item['key']})")
        self.log(f"  Base template: {item['base_key']}")
        self.log(f"  Material: {material['label']} (phase={material['phase']})")

        # Phase: inherit from material
        effective_phase = material["phase"]
        phase_source = "material"

        # Check for engraving jewel overwrite
        engraving = item.get("engraving_jewel")
        if engraving:
            gem = self.get_gem(engraving["gem_key"])
            effective_phase = gem["phase"]
            phase_source = f"engraving_jewel ({gem['label']})"
            self.log(f"  Engraving jewel: {gem['label']} (phase={gem['phase']})")
            self.log(
                f"    -> Phase OVERWRITE: {material['phase']} -> {effective_phase} (from {phase_source})"
            )

        item["_effective_phase"] = effective_phase
        item["_phase_source"] = phase_source

        # Bonuses
        attack_bonus = item.get("attack_bonus", 0)
        defense_bonus = item.get("defense_bonus", 0)
        eff_mag_mult = item.get("effect_magnitude_mult", 1.0)

        self.log(f"  Attack bonus: +{attack_bonus}")
        self.log(f"  Defense bonus: +{defense_bonus}")
        self.log(f"  Effect magnitude mult: {eff_mag_mult}")

        # Enchantments (permanent)
        enchants = item.get("enchantments", [])
        if enchants:
            self.log(f"  Enchantments ({len(enchants)} permanent):")
            for ae in enchants:
                eff = self.get_effect(ae.get("effect"))
                eff_label = eff["label"] if eff else f"effect#{ae['effect']}"
                mag = ae.get("magnitude", 0)
                dur = ae.get("duration", 0)

                # Apply material effectMagnitude multiplier
                mag_mult = item["effect_magnitude_mult"]
                modified_mag = mag * mag_mult if mag else 0

                # Wuxing interaction: item phase vs effect phase
                eff_phase = eff.get("phase") if eff else None
                wuxing_msg = ""
                mod_mag = modified_mag
                if eff_phase and effective_phase:
                    cycle_id, interaction_name, mult, affects = get_interaction(
                        self.cycles, effective_phase, eff_phase
                    )
                    if interaction_name == "identity":
                        wuxing_msg = f"(same phase {effective_phase}: identity ×{mult})"
                    elif interaction_name != "none":
                        mod_mag = modified_mag * mult
                        wuxing_msg = f"({interaction_name}: {effective_phase} -> {eff_phase} = ×{mult})"
                    else:
                        wuxing_msg = (
                            f"(unrelated: {effective_phase} vs {eff_phase} = ×{mult})"
                        )
                elif eff_phase and not effective_phase:
                    wuxing_msg = (
                        f"(item has no phase; effect has {eff_phase}, no interaction)"
                    )
                elif effective_phase and not eff_phase:
                    wuxing_msg = f"(effect has no phase; item is {effective_phase}, no interaction)"
                else:
                    wuxing_msg = "(neither has phase)"

                self.log(f"    {eff_label}: mag={mag}")
                if mag_mult != 1.0:
                    self.log(
                        f"      material.effectMagnitude mult: ×{mag_mult} -> {mag * mag_mult}"
                    )
                self.log(f"      Wuxing: {wuxing_msg} -> effective mag={mod_mag}")
        else:
            self.log(f"  Enchantments: none")

        # Coatings (temporary)
        coatings = item.get("coatings", [])
        if coatings:
            self.log(f"  Coatings ({len(coatings)} temporary, consumable):")
            for c in coatings:
                eff = self.get_effect(c["effect"])
                eff_label = eff["label"] if eff else f"effect#{c['effect']}"
                uses = c.get("uses_left", 1)
                mag = c.get("magnitude", 0)

                # Wuxing interaction for coatings too
                eff_phase = eff.get("phase") if eff else None
                wuxing_msg = ""
                mod_mag = mag
                if eff_phase and effective_phase:
                    cycle_id, interaction_name, mult, affects = get_interaction(
                        self.cycles, effective_phase, eff_phase
                    )
                    if interaction_name == "identity":
                        wuxing_msg = f"(same phase {effective_phase}: identity ×{mult})"
                    elif interaction_name != "none":
                        mod_mag = mag * mult
                        wuxing_msg = f"({interaction_name}: {effective_phase} -> {eff_phase} = ×{mult})"
                    else:
                        wuxing_msg = (
                            f"(unrelated: {effective_phase} vs {eff_phase} = ×{mult})"
                        )
                elif eff_phase and not effective_phase:
                    wuxing_msg = (
                        f"(item has no phase; effect has {eff_phase}, no interaction)"
                    )
                elif effective_phase and not eff_phase:
                    wuxing_msg = f"(effect has no phase; item is {effective_phase}, no interaction)"
                else:
                    wuxing_msg = "(neither has phase)"

                self.log(f"    {eff_label}: mag={mag}, uses={uses}")
                self.log(f"      Wuxing: {wuxing_msg} -> effective mag={mod_mag}")
        else:
            self.log(f"  Coatings: none")

        return item

    def demonstrate(self, item_key):
        self.log("")
        print(f"\n{'=' * 65}")
        print(f"  Crafted Item: {self.get_crafted_item(item_key)['label']}")
        print(f"{'=' * 65}")
        self.resolve_crafted_item(item_key)


def main():
    parser = argparse.ArgumentParser(
        description="Riftweave Crafting Reference Resolver"
    )
    parser.add_argument(
        "--verbose", "-v", action="store_true", help="Detailed step-by-step output"
    )
    args = parser.parse_args()

    resolver = CraftingResolver(verbose=args.verbose)

    # 1. Basic steel longsword — phase inheritance from material
    resolver.demonstrate("steel_longsword")

    # 2. Iron plate cuirass — material defense bonus
    resolver.demonstrate("iron_plate_cuirass")

    # 3. Ruby-engraved steel longsword — jewel overwrites phase
    resolver.demonstrate("ruby_engraved_steel_longsword")

    # 4. Sapphire-engraved elven bow — jewel overwrite + frost enchantment via Wuxing
    resolver.demonstrate("sapphire_engraved_elven_bow")

    # 5. Poisoned obsidian dagger — coating (temporary) on weapon
    resolver.demonstrate("poisoned_obsidian_dagger")

    # 6. Enchanted gold amulet — high effectMagnitude + permanent enchantment
    resolver.demonstrate("enchanted_gold_amulet")

    print(f"\n{'=' * 65}")
    print("  Summary: Phase inheritance, jewel overwrite, coating tracking,")
    print("  and Wuxing interaction all demonstrated.")
    print(f"{'=' * 65}\n")


if __name__ == "__main__":
    main()
