#!/usr/bin/env python3
"""Tests for validate.py integrity checks.

Run with:
    python -m unittest ruleset/scripts/test_validate.py
    # or
    python ruleset/scripts/test_validate.py
"""

import json
import sys
import tempfile
import unittest
from pathlib import Path

_SCRIPTS = Path(__file__).resolve().parent
sys.path.insert(0, str(_SCRIPTS))

from validate import (
    build_equipment_index,
    build_namespaced_ref_index,
    check_damage_type_vocabulary,
    check_equipment_uniqueness,
    check_attack_sequences,
    check_stances,
    check_rank_levels,
    check_build_tiers,
    check_spell_ward_table,
    check_maneuver_unlocks,
    check_rider_perks,
    check_racial_trait_refs,
    load_canonical_damage_types,
    check_references,
    equipment_ref_valid,
    resolve_namespaced_ref,
)


class TestEquipmentIdUniqueness(unittest.TestCase):
    def test_detects_duplicate_ids_across_equipment_files(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            weapons = root / "data" / "equipment" / "weapons.json"
            armor = root / "data" / "equipment" / "armor.json"
            weapons.parent.mkdir(parents=True, exist_ok=True)

            weapons.write_text(
                json.dumps(
                    {"equipment": [{"id": 22, "key": "shortsword", "type": "weapon"}]}
                )
            )
            armor.write_text(
                json.dumps(
                    {"equipment": [{"id": 22, "key": "padded_tunic", "type": "armor"}]}
                )
            )

            errors = check_equipment_uniqueness([weapons, armor], rel_base=root)

            self.assertEqual(len(errors), 1)
            self.assertIn("DUPLICATE equipment id 22", errors[0])
            self.assertIn("shortsword", errors[0])
            self.assertIn("padded_tunic", errors[0])

    def test_passes_when_ids_are_unique(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            weapons = root / "data" / "equipment" / "weapons.json"
            armor = root / "data" / "equipment" / "armor.json"
            weapons.parent.mkdir(parents=True, exist_ok=True)

            weapons.write_text(
                json.dumps(
                    {"equipment": [{"id": 45, "key": "shortsword", "type": "weapon"}]}
                )
            )
            armor.write_text(
                json.dumps(
                    {"equipment": [{"id": 22, "key": "padded_tunic", "type": "armor"}]}
                )
            )

            errors = check_equipment_uniqueness([weapons, armor], rel_base=root)
            self.assertEqual(errors, [])


class TestEquipmentKeyUniqueness(unittest.TestCase):
    def test_detects_duplicate_keys_across_equipment_files(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            weapons = root / "data" / "equipment" / "weapons.json"
            armor = root / "data" / "equipment" / "armor.json"
            weapons.parent.mkdir(parents=True, exist_ok=True)

            weapons.write_text(
                json.dumps(
                    {"equipment": [{"id": 1, "key": "shared_key", "type": "weapon"}]}
                )
            )
            armor.write_text(
                json.dumps(
                    {"equipment": [{"id": 2, "key": "shared_key", "type": "armor"}]}
                )
            )

            errors = check_equipment_uniqueness([weapons, armor], rel_base=root)

            self.assertEqual(len(errors), 1)
            self.assertIn("DUPLICATE equipment key 'shared_key'", errors[0])


class TestEquipmentRefResolution(unittest.TestCase):
    def setUp(self):
        self._tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self._tmp.cleanup)
        self.root = Path(self._tmp.name)
        weapons = self.root / "data" / "equipment" / "weapons.json"
        weapons.parent.mkdir(parents=True, exist_ok=True)
        weapons.write_text(
            json.dumps(
                {"equipment": [{"id": 45, "key": "shortsword", "type": "weapon"}]}
            )
        )
        self.id_index = {"equipment": {45}}
        self.equipment_index = build_equipment_index([weapons], rel_base=self.root)

    def test_namespaced_key_reference(self):
        self.assertTrue(
            equipment_ref_valid(
                "core:weapon/shortsword", self.id_index, self.equipment_index
            )
        )

    def test_rejects_numeric_reference(self):
        self.assertFalse(equipment_ref_valid(45, self.id_index, self.equipment_index))

    def test_rejects_bare_key_reference(self):
        self.assertFalse(
            equipment_ref_valid("shortsword", self.id_index, self.equipment_index)
        )

    def test_invalid_reference(self):
        self.assertFalse(
            equipment_ref_valid(
                "core:weapon/missing", self.id_index, self.equipment_index
            )
        )


class TestNamespacedRefIndex(unittest.TestCase):
    def test_build_namespaced_ref_index(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            effects = root / "effects.json"
            effects.write_text(
                json.dumps(
                    {
                        "effects": [
                            {"id": 1, "key": "damage_fire"},
                            {"id": 2, "key": "damage_frost"},
                        ]
                    }
                )
            )
            spells = root / "spells.json"
            spells.write_text(
                json.dumps(
                    {
                        "spells": [
                            {"id": 1, "key": "fire_bolt"},
                        ]
                    }
                )
            )
            index = build_namespaced_ref_index([effects, spells])
            self.assertIn("core:effect/damage_fire", index)
            self.assertIn("core:effect/damage_frost", index)
            self.assertIn("core:spell/fire_bolt", index)
            self.assertNotIn("damage_fire", index)
            self.assertNotIn("damage_frost", index)
            self.assertNotIn("fire_bolt", index)
            self.assertEqual(len(index), 3)

    def test_resolve_namespaced_ref_valid(self):
        index = {"core:effect/damage_fire", "core:spell/fire_bolt"}
        self.assertTrue(resolve_namespaced_ref("core:effect/damage_fire", index))
        self.assertTrue(resolve_namespaced_ref("core:spell/fire_bolt", index))
        self.assertFalse(resolve_namespaced_ref("damage_fire", index))
        self.assertFalse(resolve_namespaced_ref("fire_bolt", index))

    def test_resolve_namespaced_ref_invalid(self):
        index = {"core:effect/damage_fire"}
        self.assertFalse(resolve_namespaced_ref("core:effect/missing", index))
        self.assertFalse(resolve_namespaced_ref("core:weapon/shortsword", index))
        self.assertFalse(resolve_namespaced_ref("damage_fire", index))

    def test_resolve_unknown_module_ref(self):
        index = set()
        self.assertTrue(
            resolve_namespaced_ref("module:my-mod/weapon/frost_sword", index)
        )
        self.assertTrue(resolve_namespaced_ref("module:example/effect/custom", index))

    def test_resolve_non_string(self):
        index = set()
        self.assertFalse(resolve_namespaced_ref(42, index))
        self.assertFalse(resolve_namespaced_ref(None, index))


class TestManeuverUnlocks(unittest.TestCase):
    """M3-24: every maneuver needs a perk that unlocks it (docs/core/actions.md#maneuvers)."""

    def test_a_maneuver_no_perk_unlocks_fails(self):
        result = check_maneuver_unlocks([{"key": "shove"}], [{"key": "wind_up", "unlocks": ["core:maneuver/heavy_swing"]}])
        self.assertEqual(len(result), 1)
        self.assertIn("shove", result[0])

    def test_a_maneuver_a_perk_unlocks_passes(self):
        self.assertEqual(check_maneuver_unlocks([{"key": "heavy_swing"}], [{"key": "wind_up", "unlocks": ["core:maneuver/heavy_swing"]}]), [])


class TestRiderPerks(unittest.TestCase):
    """GD_S05_04: a rider perk needs the perk that unlocks its maneuver, and an exclusion is
    listed both ways."""

    UNLOCK = {"key": "low_sweep", "unlocks": ["core:maneuver/trip"]}

    def test_the_data_passes(self):
        features = json.loads((Path(__file__).resolve().parent.parent / "data" / "features" / "core.json").read_text())["features"]
        self.assertEqual(check_rider_perks(features), [])

    def test_a_one_way_exclusion_fails(self):
        other = {"key": "bind", "unlocks": ["core:maneuver/disarm"]}
        errors = check_rider_perks([dict(self.UNLOCK, excludes=["core:feature/bind"]), other])
        self.assertTrue(any("does not exclude it back" in e for e in errors), errors)
        self.assertEqual(check_rider_perks([dict(self.UNLOCK, excludes=["core:feature/bind"]), dict(other, excludes=["core:feature/low_sweep"])]), [])

    def test_a_rider_that_does_not_need_its_unlock_perk_fails(self):
        rider = {"key": "hook", "rider": "core:maneuver/trip", "prerequisite": {"perks": []}}
        errors = check_rider_perks([self.UNLOCK, rider])
        self.assertTrue(any("needs no perk that unlocks it" in e for e in errors), errors)


class TestAttackSequences(unittest.TestCase):
    """GD_S05_01: one attack_sequences row per loadout (docs/core/combat.md#attacks-by-loadout)."""

    ONE = {"key": "one", "main": ["one_handed"], "off": ["empty"], "ap": 3, "strikes": [{"hand": "main", "attack": 0}]}
    TWO = {"key": "two", "main": ["light"], "off": ["light_weapon"], "ap": 3,
           "strikes": [{"hand": "main", "attack": -10}, {"hand": "off", "attack": -20}], "jab": {"ap": 2, "attack": -10}}

    def test_the_data_passes(self):
        rows = json.loads((Path(__file__).resolve().parent.parent / "data" / "attack_sequences" / "core.json").read_text())
        self.assertEqual(check_attack_sequences(rows["attack_sequences"]), [])

    def test_two_rows_for_one_loadout_fail(self):
        errors = check_attack_sequences([self.ONE, {**self.ONE, "key": "again"}])
        self.assertTrue(any("already covered by 'one'" in e for e in errors), errors)

    def test_a_jab_without_an_off_hand_weapon_fails(self):
        errors = check_attack_sequences([{**self.ONE, "jab": {"ap": 2, "attack": -10}}])
        self.assertTrue(any("needs an off-hand weapon" in e for e in errors), errors)

    def test_an_off_hand_weapon_without_a_jab_fails(self):
        errors = check_attack_sequences([{k: v for k, v in self.TWO.items() if k != "jab"}])
        self.assertTrue(any("needs an off-hand strike and a jab" in e for e in errors), errors)


class TestStances(unittest.TestCase):
    """GD_S05_02: a perk unlocks each fighting stance, and its modifiers are martial and while applied."""

    MOD = {"target": "attack_roll", "operation": "add", "value": 10, "source": "stance:lunge",
           "origin": "core:origin/martial", "duration": "while_applied"}
    ROW = {"key": "lunge", "label": "Lunge", "description": "x", "modifiers": [MOD]}
    PERK = {"key": "lunge", "unlocks": ["core:stance/lunge"]}

    def test_the_data_passes(self):
        data = Path(__file__).resolve().parent.parent / "data"
        stances = json.loads((data / "stances" / "core.json").read_text())["stances"]
        features = json.loads((data / "features" / "core.json").read_text())["features"]
        self.assertEqual(check_stances(stances, features), [])

    def test_a_stance_no_perk_unlocks_fails(self):
        errors = check_stances([self.ROW], [])
        self.assertTrue(any("no perk unlocks it" in e for e in errors), errors)

    def test_a_modifier_with_another_origin_source_or_duration_fails(self):
        bad = {**self.MOD, "origin": "core:origin/trained", "source": "feature:lunge", "duration": "instant"}
        errors = check_stances([{**self.ROW, "modifiers": [bad]}], [self.PERK])
        for needed in ("origin core:origin/martial", "source stance:lunge", "duration while_applied"):
            self.assertTrue(any(needed in e for e in errors), (needed, errors))


class TestTieredPerks(unittest.TestCase):
    """GD_S05_03: rank_levels gates each tier, a build takes no tier early, and the spell ward
    table is what its formula gives."""

    DATA = Path(__file__).resolve().parent.parent / "data"
    PERK = {"key": "wall", "ranks": 3, "rank_levels": [1, 6, 12], "prerequisite": {},
            "modifiers": [{"target": "shield_dr", "operation": "add", "value": 2}], "rank_values": {"shield_dr": [2, 3, 4]}}

    def rows(self, name, key):
        return json.loads((self.DATA / name / "core.json").read_text())[key]

    def test_the_data_passes(self):
        features = self.rows("features", "features")
        self.assertEqual(check_rank_levels(features), [])
        self.assertEqual(check_build_tiers(self.rows("builds", "builds"), features), [])
        pipeline = json.loads((self.DATA / "combat" / "pipeline.json").read_text())["pipeline"]
        self.assertEqual(check_spell_ward_table(pipeline), [])

    def test_a_gate_count_a_first_gate_or_a_step_out_of_rule_fails(self):
        errors = check_rank_levels([{**self.PERK, "rank_levels": [1, 6]}])
        self.assertTrue(any("2 gates for 3 ranks" in e for e in errors), errors)
        errors = check_rank_levels([{**self.PERK, "rank_levels": [4, 8, 12]}])
        self.assertTrue(any("not its prerequisite level 1" in e for e in errors), errors)
        errors = check_rank_levels([{**self.PERK, "rank_levels": [1, 3, 12]}])
        self.assertTrue(any("not 3 to 6 levels apart" in e for e in errors), errors)

    def test_rank_values_need_one_value_per_rank_on_a_target_the_perk_emits(self):
        errors = check_rank_levels([{**self.PERK, "rank_values": {"shield_dr": [2, 3], "hp": [1, 2, 3]}}])
        self.assertTrue(any("2 values for 3 ranks" in e for e in errors), errors)
        self.assertTrue(any("names hp" in e for e in errors), errors)

    def test_a_tier_taken_before_its_gate_fails(self):
        build = {"key": "early", "perks_by_level": {"2": "core:feature/wall", "3": "core:feature/wall"}}
        errors = check_build_tiers([build], [self.PERK])
        self.assertTrue(any("tier 2 at level 3, before its gate 6" in e for e in errors), errors)
        build = {"key": "on_time", "perks_by_level": {"2": "core:feature/wall", "6": "core:feature/wall"}}
        self.assertEqual(check_build_tiers([build], [self.PERK]), [])

    def test_a_spell_ward_cell_off_the_formula_fails(self):
        pipeline = json.loads((self.DATA / "combat" / "pipeline.json").read_text())["pipeline"]
        table = [list(row) for row in pipeline["spell_ward_table"]]
        table[8][2] = 54
        errors = check_spell_ward_table({**pipeline, "spell_ward_table": table})
        self.assertTrue(any("END 9 tier 3 reads 54, the formula gives 50" in e for e in errors), errors)


class TestManeuverChannel(unittest.TestCase):
    """M3-24: a maneuver's effects declare the maneuver channel (docs/core/actions.md#maneuvers)."""

    def maneuver(self, effect, modifies="attack"):
        return {"maneuvers": [{"key": "shove", "label": "Shove", "description": "x", "modifies": modifies,
                               "ap": 0, "effects": [{"effect": effect}]}]}

    def test_an_effect_without_the_maneuver_channel_fails(self):
        channels = {"core:effect/knockdown": {"spell", "innate"}}
        result = check_references(self.maneuver("core:effect/knockdown"), "maneuvers", effect_channels=channels)
        self.assertEqual(len(result), 1)
        self.assertIn("maneuver", result[0])

    def test_an_effect_with_the_maneuver_channel_passes(self):
        channels = {"core:effect/stagger": {"spell", "maneuver"}}
        self.assertEqual(check_references(self.maneuver("core:effect/stagger"), "maneuvers", effect_channels=channels), [])

    def test_a_maneuver_of_an_unknown_action_fails(self):
        channels = {"core:effect/stagger": {"maneuver"}}
        result = check_references(self.maneuver("core:effect/stagger", "fly"), "maneuvers", effect_channels=channels)
        self.assertEqual(len(result), 1)
        self.assertIn("not an action", result[0])


class TestCraftedItemsRefValidation(unittest.TestCase):
    def test_rejects_numeric_effect_in_enchantment(self):
        result = check_references(
            {
                "crafted_items": [
                    {
                        "id": 1,
                        "key": "test_item",
                        "enchantments": [{"effect": 999, "magnitude": 5}],
                        "coatings": [],
                    }
                ]
            },
            "crafted_items",
        )
        self.assertEqual(len(result), 1)
        self.assertIn("numeric refs not allowed", result[0])

    def test_detects_missing_effect_in_coating(self):
        result = check_references(
            {
                "crafted_items": [
                    {
                        "id": 1,
                        "key": "test_item",
                        "enchantments": [],
                        "coatings": [
                            {
                                "effect": "core:effect/missing",
                                "magnitude": 3,
                                "uses_left": 2,
                            }
                        ],
                    }
                ]
            },
            "crafted_items",
            namespaced_ref_index={"core:effect/damage_fire"},
        )
        self.assertEqual(len(result), 1)
        self.assertIn("does not resolve", result[0])

    def test_accepts_valid_namespaced_effect_ref(self):
        result = check_references(
            {
                "crafted_items": [
                    {
                        "id": 1,
                        "key": "test_item",
                        "enchantments": [
                            {"effect": "core:effect/damage_fire", "magnitude": 5}
                        ],
                        "coatings": [],
                    }
                ]
            },
            "crafted_items",
            namespaced_ref_index={"core:effect/damage_fire"},
        )
        self.assertEqual(result, [])

    def test_rejects_invalid_namespaced_effect_ref(self):
        result = check_references(
            {
                "crafted_items": [
                    {
                        "id": 1,
                        "key": "test_item",
                        "enchantments": [
                            {"effect": "core:effect/missing", "magnitude": 5}
                        ],
                        "coatings": [],
                    }
                ]
            },
            "crafted_items",
            namespaced_ref_index={"core:effect/damage_fire"},
        )
        self.assertEqual(len(result), 1)
        self.assertIn("does not resolve", result[0])


if __name__ == "__main__":
    unittest.main()


class TestDamageTypeVocabulary(unittest.TestCase):
    def test_rejects_cold_alias_and_accepts_frost(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            weapons = root / "data" / "equipment" / "weapons.json"
            weapons.parent.mkdir(parents=True, exist_ok=True)
            weapons.write_text(
                json.dumps(
                    {
                        "equipment": [
                            {
                                "id": 45,
                                "key": "frost_blade",
                                "overrides": {"weapon": {"damage": {"type": "cold"}}},
                            }
                        ]
                    }
                )
            )
            errors = check_damage_type_vocabulary(
                [weapons], {"frost", "slashing"}, rel_base=root
            )
            self.assertEqual(len(errors), 1)
            self.assertIn("use 'frost'", errors[0])

    def test_live_registry_matches_schema_and_backgrounds_resolve(self):
        root = Path(__file__).resolve().parent.parent
        canonical = load_canonical_damage_types(root / "data")
        self.assertIn("frost", canonical)
        self.assertNotIn("cold", canonical)
        self.assertNotIn("lightning", canonical)
        data_files = list((root / "data").rglob("*.json"))
        # bases are skipped by collect; this check includes them via the helper
        equipment_files = [
            fp for fp in data_files if fp.parent.name == "equipment" and fp.name.endswith(".json")
        ]
        errors = check_damage_type_vocabulary(equipment_files, canonical, rel_base=root)
        self.assertEqual(errors, [])
        from validate import build_equipment_index, equipment_ref_valid
        index = build_equipment_index(equipment_files, rel_base=root)
        backgrounds = json.loads((root / "data" / "backgrounds" / "core.json").read_text())
        for bg in backgrounds["backgrounds"]:
            for entry in bg.get("starting_equipment") or []:
                ref = entry.get("item")
                self.assertTrue(
                    equipment_ref_valid(ref, {}, index),
                    f"{bg.get('key')} item {ref} does not resolve",
                )


class TestRacialTraitRefs(unittest.TestCase):
    def _write(self, root, features, races):
        f = root / "features.json"
        r = root / "races.json"
        f.write_text(json.dumps({"features": features}))
        r.write_text(json.dumps({"races": races}))
        return [f, r]

    def test_race_trait_resolving_to_racial_trait_feature_passes(self):
        with tempfile.TemporaryDirectory() as tmp:
            files = self._write(
                Path(tmp),
                [{"key": "elf_darkvision", "type": "racial_trait"}],
                [{"key": "elf", "traits": [{"id": "core:feature/elf_darkvision"}]}],
            )
            self.assertEqual(check_racial_trait_refs(files), [])

    def test_core_trait_ref_is_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            files = self._write(
                Path(tmp), [], [{"key": "elf", "traits": [{"id": "core:trait/elf_darkvision"}]}]
            )
            errors = check_racial_trait_refs(files)
            self.assertEqual(len(errors), 1)
            self.assertIn("must be a core:feature/<key> ref", errors[0])

    def test_race_trait_pointing_at_a_perk_is_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            files = self._write(
                Path(tmp),
                [{"key": "hardened_hide", "type": "perk"}],
                [{"key": "dwarf", "traits": [{"id": "core:feature/hardened_hide"}]}],
            )
            errors = check_racial_trait_refs(files)
            self.assertEqual(len(errors), 1)
            self.assertIn("not 'racial_trait'", errors[0])

    def test_racial_trait_no_race_lists_is_an_orphan(self):
        with tempfile.TemporaryDirectory() as tmp:
            files = self._write(
                Path(tmp), [{"key": "lonely", "type": "racial_trait"}], []
            )
            errors = check_racial_trait_refs(files)
            self.assertEqual(
                errors, ["racial_trait 'core:feature/lonely' is not listed by any race"]
            )

    def test_module_trait_refs_pass(self):
        with tempfile.TemporaryDirectory() as tmp:
            files = self._write(
                Path(tmp), [], [{"key": "x", "traits": [{"id": "module:m/feature/y"}]}]
            )
            self.assertEqual(check_racial_trait_refs(files), [])

    def test_live_data_has_no_traits_table_and_all_race_traits_are_features(self):
        root = Path(__file__).resolve().parent.parent
        self.assertFalse((root / "data" / "traits").exists())
        self.assertFalse((root / "schemas" / "trait.schema.json").exists())
        files = list((root / "data").rglob("*.json"))
        self.assertEqual(check_racial_trait_refs(files), [])

