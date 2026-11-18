#!/usr/bin/env python3
"""The perk catalogue's content floor (docs/modules/progression.md#perks-and-traits).

Run with:
    python -m unittest ruleset/scripts/test_perks.py
"""

import json
import unittest
from pathlib import Path

FEATURES = Path(__file__).resolve().parent.parent / "data" / "features" / "core.json"
ABILITIES = ("str", "per", "end", "int", "wil", "agi", "cha", "lck")


def perks():
    return [f for f in json.loads(FEATURES.read_text())["features"] if f["type"] == "perk"]


def features():
    return json.loads(FEATURES.read_text())["features"]


def creation_traits():
    return [f for f in features() if f["type"] == "creation"]


def level(perk):
    return perk.get("prerequisite", {}).get("level", 1)


def prerequisite_refs(perk):
    """The prerequisite perks' refs, each a held perk { perk, rank } (M3-14, M5-70)."""
    return [item["perk"] for item in perk.get("prerequisite", {}).get("perks", [])]


def emitted(perk):
    """The (target, operation) pairs a perk's modifiers name, and what it unlocks."""
    return frozenset((m["target"], m["operation"]) for m in perk.get("modifiers", [])) | frozenset(
        ("unlocks", ref) for ref in perk.get("unlocks", [])
    )


class TestPerkCatalogue(unittest.TestCase):
    def test_at_least_sixty_perks(self):
        self.assertGreaterEqual(len(perks()), 60)

    def test_every_perk_emits_trained_modifiers_from_itself(self):
        """A perk changes a number or unlocks something (a maneuver, a template, a material);
        an unlock-only perk needs no modifier to justify it (docs/core/actions.md#maneuvers)."""
        for perk in perks():
            with self.subTest(perk=perk["key"]):
                self.assertTrue(perk.get("modifiers") or perk.get("unlocks"), "a perk emits modifiers or unlocks something")
                for modifier in perk.get("modifiers", []):
                    self.assertEqual(modifier["source"], f"feature:{perk['key']}")
                    self.assertEqual(modifier["origin"], "core:origin/trained")

    def test_prerequisite_perks_are_keys_of_perks(self):
        keys = {perk["key"] for perk in perks()}
        for perk in perks():
            for ref in prerequisite_refs(perk):
                with self.subTest(perk=perk["key"], needs=ref):
                    self.assertTrue(ref.startswith("core:feature/"))
                    self.assertIn(ref.removeprefix("core:feature/"), keys)

    def test_a_prerequisite_perk_is_reachable_at_a_lower_or_equal_level(self):
        by_key = {perk["key"]: perk for perk in perks()}
        for perk in perks():
            for ref in prerequisite_refs(perk):
                with self.subTest(perk=perk["key"], needs=ref):
                    self.assertLessEqual(level(by_key[ref.removeprefix("core:feature/")]), level(perk))

    def test_every_attribute_gates_a_perk(self):
        gated = {a for perk in perks() for a in perk.get("prerequisite", {}).get("abilities", {})}
        self.assertEqual(gated, set(ABILITIES))

    def test_combat_magic_and_social_each_have_perks(self):
        counts = {}
        for perk in perks():
            counts[perk["category"]] = counts.get(perk["category"], 0) + 1
        for category in ("combat", "magic", "social"):
            with self.subTest(category=category):
                self.assertGreaterEqual(counts.get(category, 0), 8)

    def test_no_two_perks_at_one_level_emit_the_same_targets(self):
        """A perk is dominated when another at its level does the same thing; the catalogue
        avoids it by giving each perk at one level its own set of targets and ops."""
        seen = {}
        for perk in perks():
            key = (level(perk), emitted(perk))
            with self.subTest(perk=perk["key"]):
                self.assertNotIn(key, seen, f"same level and targets as {seen.get(key)}")
            seen[key] = perk["key"]


class TestOneStatementOfAFeature(unittest.TestCase):
    """M5-35: a feature states its mechanics once, in `modifiers` (docs/modules/progression.md#what-a-perk-emits)."""

    def test_no_feature_has_an_effects_list(self):
        for feature in features():
            with self.subTest(feature=feature["key"]):
                self.assertNotIn("effects", feature)

    def test_every_creation_trait_emits_trained_modifiers_from_itself(self):
        self.assertGreaterEqual(len(creation_traits()), 10)
        for trait in creation_traits():
            with self.subTest(trait=trait["key"]):
                self.assertTrue(trait.get("modifiers"), "a creation trait emits modifiers")
                for modifier in trait["modifiers"]:
                    self.assertEqual(modifier["source"], f"feature:{trait['key']}")
                    self.assertEqual(modifier["origin"], "core:origin/trained")

    def test_every_creation_trait_has_an_upside_and_a_downside(self):
        """A trait is a tradeoff: at least one modifier helps and one hinders or restricts."""
        for trait in creation_traits():
            with self.subTest(trait=trait["key"]):
                self.assertGreaterEqual(len(trait["modifiers"]), 2)

    def test_gifted_adds_one_to_every_attribute(self):
        gifted = next(f for f in features() if f["key"] == "gifted")
        adds = {m["target"]: m["value"] for m in gifted["modifiers"] if m["operation"] == "add"}
        self.assertEqual({ability: adds.get(ability) for ability in ABILITIES}, dict.fromkeys(ABILITIES, 1))

    def test_a_phobia_grants_a_flag_that_frightens(self):
        for key, flag in (("undead_phobia", "afraid_of_undead"), ("beast_phobia", "afraid_of_beasts"),
                          ("arcane_phobia", "afraid_of_constructs")):
            trait = next(f for f in features() if f["key"] == key)
            with self.subTest(trait=key):
                self.assertIn((flag, "grant"), emitted(trait))


class TestSummonLimitPerks(unittest.TestCase):
    """M3-20: Retinue and Host raise the number of active summons (docs/core/casting.md#summons)."""

    def test_retinue_at_14_and_host_at_28_each_add_one_summon_and_host_needs_retinue(self):
        rows = {perk["key"]: perk for perk in perks()}
        retinue, host = rows["summon_retinue"], rows["summon_host"]
        self.assertEqual((retinue["prerequisite"]["level"], host["prerequisite"]["level"]), (14, 28))
        self.assertIn("core:feature/summon_retinue", prerequisite_refs(host))
        for perk in (retinue, host):
            self.assertIn(("summon_limit", "add", 1), {(m["target"], m["operation"], m["value"]) for m in perk["modifiers"]})



class TestSharePerks(unittest.TestCase):
    """M3-14: the bard and skald perks lend CHA to allies (docs/core/share.md)."""

    def test_the_ladder_starts_at_leader_and_each_share_perk_needs_it_at_a_higher_level(self):
        rows = {perk["key"]: perk for perk in perks()}
        sharing = [perk for perk in perks() if "share" in perk and perk["prerequisite"]["level"] == 18]
        self.assertEqual({perk["key"] for perk in sharing}, {"bardic_inspiration", "skalds_war_chant"})
        for perk in sharing:
            with self.subTest(perk=perk["key"]):
                self.assertIn("core:feature/leader", prerequisite_refs(perk))
                self.assertGreater(level(perk), level(rows["leader"]))
                self.assertGreaterEqual(perk["prerequisite"]["abilities"]["cha"], rows["leader"]["prerequisite"]["abilities"]["cha"])
                self.assertEqual(perk["ranks"], 2)

    def test_the_bard_leans_on_magic_and_the_skald_on_weapons(self):
        rows = {perk["key"]: perk for perk in perks()}
        bard, skald = rows["bardic_inspiration"], rows["skalds_war_chant"]
        self.assertEqual(bard["share"]["bonus"], "core:summon_bonus/magic")
        self.assertEqual(skald["share"]["bonus"], "core:summon_bonus/physical")
        self.assertTrue(all(ref.endswith("_magic") for ref in bard["prerequisite"]["skills"][0]["any_of"]))
        self.assertFalse(any(ref.endswith("_magic") for ref in skald["prerequisite"]["skills"][0]["any_of"]))

    def test_each_class_has_its_own_social_skill_at_both_tiers(self):
        rows = {perk["key"]: perk for perk in perks()}
        for key, skill, minimum in (("bardic_inspiration", "persuasion", 30), ("skalds_war_chant", "intimidation", 30),
                                    ("bardic_anthem", "persuasion", 60), ("skalds_saga", "intimidation", 60)):
            with self.subTest(perk=key):
                needs = {(r["any_of"][0], r["min"]) for r in rows[key]["prerequisite"]["skills"] if len(r["any_of"]) == 1}
                self.assertIn((f"core:skill/{skill}", minimum), needs)

    def test_tier_two_needs_tier_one_at_rank_two_and_lends_it_all_twice_as_far_for_longer(self):
        rows = {perk["key"]: perk for perk in perks()}
        for top, base in (("bardic_anthem", "bardic_inspiration"), ("skalds_saga", "skalds_war_chant")):
            with self.subTest(perk=top):
                self.assertEqual(level(rows[top]), 28)
                self.assertIn({"perk": f"core:feature/{base}", "rank": 2}, rows[top]["prerequisite"]["perks"])
                self.assertEqual(rows[top]["share"]["bonus"], rows[base]["share"]["bonus"])
                self.assertEqual(rows[top]["share"]["rank_share"], [1])
                self.assertEqual(rows[top]["share"]["range"], 2 * rows[base]["share"]["range"])
                self.assertGreater(rows[top]["share"]["turns"], rows[base]["share"]["turns"])

if __name__ == "__main__":
    unittest.main()
