#!/usr/bin/env python3
"""The bestiary content floor (docs/modules/monsters.md#the-floor).

Run with:
    python -m unittest ruleset/scripts/test_bestiary.py
"""

import json
import re
import sys
import unittest
from pathlib import Path

_SCRIPTS = Path(__file__).resolve().parent
sys.path.insert(0, str(_SCRIPTS))

import validate  # noqa: E402  (base + override resolution)


def ref():
    """The monster reference (docs/modules/monsters.md); imported per test so one missing helper fails only its tests."""
    import monster_reference
    return monster_reference

DATA = _SCRIPTS.parent / "data"
# Product identity of other games (docs/modules/monsters.md#names): never a monster name.
PROTECTED_NAMES = (
    "beholder", "mind flayer", "illithid", "displacer", "owlbear", "umber hulk", "carrion crawler",
    "gelatinous cube", "githyanki", "githzerai", "kuo-toa", "slaad", "yuan-ti", "flan", "tonberry",
    "cactuar", "malboro", "moogle", "chocobo",
)


def raw_monsters():
    return json.loads((DATA / "monsters" / "core.json").read_text())["monsters"]


def monsters():
    data = {"monsters": raw_monsters()}
    validate._resolve_monster_bases(data, [])
    return data["monsters"]


def resolve(rows):
    """Resolve a list of rows as validate.py does; fail the test on any error."""
    errors = []
    data = {"monsters": rows}
    validate._resolve_monster_bases(data, errors)
    if errors:
        raise AssertionError(errors)
    return {row["key"]: row for row in data["monsters"]}


def damage_types():
    return [t["key"] for t in json.loads((DATA / "combat" / "damage_types.json").read_text())["damage_types"]]


def tiers():
    return json.loads((DATA / "balance" / "core.json").read_text())["tiers"]


class TestBestiary(unittest.TestCase):
    def test_at_least_forty_monsters_from_level_1_to_30(self):
        rows = monsters()
        self.assertGreaterEqual(len(rows), 40)
        levels = [row["level"] for row in rows]
        self.assertEqual((min(levels), max(levels)), (1, 30))

    def test_every_balance_tier_has_monsters(self):
        levels = [row["level"] for row in monsters()]
        for tier in tiers():
            low, high = tier["levels"]
            with self.subTest(tier=tier["key"]):
                self.assertGreaterEqual(sum(low <= level <= high for level in levels), 5)

    def test_every_damage_type_is_an_attack_and_a_resistance(self):
        rows = monsters()
        attacks = {ref().attack_type(attack) for row in rows for attack in row.get("attacks", [])}
        resisted = {kind for row in rows for kind in row.get("resist", {})}
        for kind in damage_types():
            with self.subTest(damage_type=kind):
                self.assertIn(kind, attacks, "no monster attacks with it")
                self.assertIn(kind, resisted, "no monster resists it")

    def test_every_monster_is_a_base_plus_overrides(self):
        for row in raw_monsters():
            with self.subTest(monster=row["key"]):
                self.assertIn("base", row)
                self.assertIn("overrides", row)

    def test_every_monster_attacks_or_has_a_damaging_ability(self):
        for row in monsters():
            with self.subTest(monster=row["key"]):
                self.assertTrue(row.get("attacks") or ref().damaging_abilities(row))

    def test_families_derive_from_a_base_monster(self):
        rows = {row["key"]: row for row in raw_monsters()}
        families = {
            "ooze": ["acid_ooze", "fire_ooze", "frost_ooze"],
            "giant": ["frost_giant", "storm_giant_youth"],
            "fire_elemental": ["magma_titan"],
            "stone_golem": ["bronze_guardian", "iron_colossus"],
            "treant_sapling": ["elder_treant"],
            "bandit": ["bandit_captain"],
            "ember_drake": ["ancient_dragon"],
            "skeleton": ["bone_colossus"],
            "wolf": ["dire_wolf", "elder_wolf"],
            "dog": ["demonic_hound"],
            "dire_wolf": ["shadow_hound"],
            "storm_hawk": ["thunderbird"],
        }
        for parent, children in families.items():
            for child in children:
                with self.subTest(monster=child):
                    self.assertEqual(rows[child]["base"], f"core:monster/{parent}")

    def test_element_variants_use_a_template(self):
        rows = {row["key"]: row for row in raw_monsters()}
        expected = {"acid_ooze": ["acid"], "fire_ooze": ["fire"], "frost_ooze": ["frost"],
                    "frost_giant": ["frost"], "storm_giant_youth": ["storm"]}
        for key, templates in expected.items():
            with self.subTest(monster=key):
                self.assertEqual(rows[key].get("templates"), templates)

    def test_every_derived_monster_states_its_level(self):
        for row in raw_monsters():
            if row["base"].startswith("core:monster/"):
                with self.subTest(monster=row["key"]):
                    self.assertIn("level", row["overrides"])

    def test_a_template_sets_phase_attacks_resist_tags_and_label(self):
        merged = resolve([
            {"id": 1, "key": "blob", "base": "core:creature_type/ooze", "overrides": {
                "label": "Blob", "level": 2, "resist": {"fire": 2, "acid": 5}, "tags": ["ooze"],
                "attacks": [{"key": "slam", "label": "Slam", "weapon": "core:natural_weapon/slam"}]}},
            {"id": 2, "key": "fire_blob", "base": "core:monster/blob", "templates": ["fire"], "overrides": {"level": 2}},
        ])["fire_blob"]
        self.assertEqual(merged["label"], "Fire Blob")
        self.assertEqual(merged["phase"], "fire")
        self.assertEqual(merged["attacks"][0]["damage"]["type"], "fire")
        self.assertEqual(merged["attacks"][0]["phase"], "fire")
        self.assertEqual(merged["resist"]["acid"], 5)
        self.assertGreater(merged["resist"]["fire"], 2)
        self.assertEqual(merged["tags"].count("ooze"), 1)
        self.assertIn("fire", merged["tags"])

    def test_a_derived_monster_moves_by_its_types_growth(self):
        rows = resolve([
            {"id": 1, "key": "blob", "base": "core:creature_type/ooze", "overrides": {"level": 2}},
            {"id": 2, "key": "big_blob", "base": "core:monster/blob", "overrides": {"level": 6, "label": "Big Blob"}},
        ])
        base = json.loads((DATA / "monsters" / "bases" / "ooze.json").read_text())
        growth = base["growth"]
        self.assertEqual(rows["big_blob"]["hitPoints"], rows["blob"]["hitPoints"] + 4 * growth["hitPointsPerLevel"])
        for ability, rule in growth["attributes"].items():
            with self.subTest(ability=ability):
                expected = max(1, round(rows["blob"]["attributes"][ability] + 4 * rule["perLevel"]))
                self.assertEqual(rows["big_blob"]["attributes"][ability], expected)

    def test_a_missing_or_cyclic_base_monster_or_template_fails(self):
        cases = {
            "missing": [{"id": 1, "key": "a", "base": "core:monster/nobody", "overrides": {"level": 1}}],
            "cycle": [{"id": 1, "key": "a", "base": "core:monster/b", "overrides": {"level": 1}},
                      {"id": 2, "key": "b", "base": "core:monster/a", "overrides": {"level": 1}}],
            "template": [{"id": 1, "key": "a", "base": "core:creature_type/ooze", "templates": ["glitter"], "overrides": {}}],
            "no level": [{"id": 1, "key": "a", "base": "core:creature_type/ooze", "overrides": {}},
                         {"id": 2, "key": "b", "base": "core:monster/a", "overrides": {}}],
        }
        for name, rows in cases.items():
            with self.subTest(case=name):
                errors = []
                validate._resolve_monster_bases({"monsters": rows}, errors)
                self.assertTrue(errors)

    def test_bases_then_variants_then_one_off_monsters(self):
        rows = raw_monsters()
        parent = {row["key"]: row["base"].removeprefix("core:monster/") for row in rows if row["base"].startswith("core:monster/")}
        group = [0 if row["key"] in parent.values() else 1 if row["key"] in parent else 2 for row in rows]
        self.assertEqual(group, sorted(group))

    def test_no_name_from_another_franchise(self):
        for row in monsters():
            text = " ".join([row["key"].replace("_", " "), row["label"], row["description"]]).lower()
            for name in PROTECTED_NAMES:
                with self.subTest(monster=row["key"], name=name):
                    self.assertIsNone(re.search(rf"\b{re.escape(name)}\b", text))



def features():
    return {f["key"]: f for f in json.loads((DATA / "features" / "core.json").read_text())["features"]}


def summon_group(key):
    groups = json.loads((DATA / "summons" / "core.json").read_text())["summons"]
    return [v.removeprefix("core:monster/") for g in groups if g["key"] == key for v in g["variants"]]


class TestTemplatesAndGeneratedMonsters(unittest.TestCase):
    """Power and nature templates, generated monsters, and the example rows (M3-10)."""

    def test_a_dog_stays_weaker_than_a_wolf_with_the_same_template(self):
        dog = ref().generate("core:monster/dog", power="dire", level=3)
        wolf = ref().generate("core:monster/wolf", power="dire", level=3)
        self.assertLess(sum(dog["attributes"].values()), sum(wolf["attributes"].values()))
        self.assertLess(dog["hitPoints"], wolf["hitPoints"])
        for ability, value in dog["attributes"].items():
            with self.subTest(ability=ability):
                self.assertLessEqual(value, wolf["attributes"].get(ability, 0))

    def test_the_power_ladder_needs_the_rung_before(self):
        perks = features()
        self.assertIn("core:feature/summon_dire", perks["summon_elder"]["prerequisite"]["perks"])
        self.assertIn("core:feature/summon_elder", perks["summon_ancient"]["prerequisite"]["perks"])
        self.assertFalse(perks["summon_dire"].get("prerequisite", {}).get("perks"))

    def test_each_power_and_nature_perk_unlocks_its_template(self):
        perks = features()
        for key in ("dire", "elder", "ancient", "demonic", "angelic", "otherworldly"):
            with self.subTest(template=key):
                self.assertEqual(perks[f"summon_{key}"]["unlocks"], [f"core:monster_template/{key}"])

    def test_the_latest_power_and_nature_bought_replace_the_earlier(self):
        bought = ["core:feature/summon_dire", "core:feature/summon_demonic", "core:feature/summon_elder",
                  "core:feature/summon_angelic"]
        self.assertEqual(ref().held_templates(bought), {"power": "elder", "nature": "angelic"})
        self.assertEqual(ref().held_templates([]), {"power": None, "nature": None})

    def test_two_templates_in_one_slot_fail(self):
        cases = {
            "one row": [{"key": "a", "base": "core:creature_type/ooze", "templates": ["fire", "frost"], "overrides": {"level": 2}}],
            "a chain": [{"key": "a", "base": "core:creature_type/beast", "templates": ["dire"], "overrides": {"level": 2}},
                        {"key": "b", "base": "core:monster/a", "templates": ["elder"], "overrides": {"level": 2}}],
        }
        for name, rows in cases.items():
            with self.subTest(case=name):
                errors = []
                validate._resolve_monster_bases({"monsters": rows}, errors)
                self.assertTrue(errors)

    def test_slots_apply_element_power_then_nature_whatever_the_order_listed(self):
        rows = resolve([{"key": "hound", "base": "core:creature_type/beast", "templates": ["demonic", "dire", "fire"],
                         "overrides": {"label": "Hound", "level": 3}}])
        self.assertEqual(rows["hound"]["label"], "Hound")
        generated = ref().generate("core:monster/dog", power="dire", nature="demonic")
        self.assertEqual(generated["label"], "Demonic Dire Dog")

    def test_a_generated_monster_composes_with_no_row_written_for_it(self):
        self.assertNotIn("ancient_otherworldly_ember_drake", {row["key"] for row in raw_monsters()})
        drake = ref().generate("core:monster/ember_drake", power="ancient", nature="otherworldly")
        plain = {row["key"]: row for row in monsters()}["ember_drake"]
        self.assertEqual(drake["label"], "Otherworldly Ancient Ember Drake")
        self.assertEqual(drake["level"], plain["level"])
        self.assertGreater(drake["attributes"]["str"], plain["attributes"]["str"])
        self.assertIn("otherworldly", drake["tags"])

    def test_a_slot_the_row_fills_keeps_the_rows_template(self):
        dire = ref().generate("core:monster/dire_wolf", power="elder")
        self.assertEqual(dire["label"], "Dire Wolf")

    def test_no_template_changes_the_creature_type(self):
        types = {path.stem for path in (DATA / "monsters" / "bases").glob("*.json")}
        templates = json.loads((DATA / "monsters" / "templates.json").read_text())["monster_templates"]
        self.assertFalse(types & {t["key"] for t in templates})
        for template in templates:
            with self.subTest(template=template["key"]):
                self.assertNotIn("base", template)
        self.assertEqual(ref().creature_type("core:monster/undead_horse"), "undead")
        self.assertEqual(ref().creature_type("core:monster/demonic_hound"), "beast")

    def test_dire_wolf_is_wolf_plus_dire(self):
        row = {row["key"]: row for row in raw_monsters()}["dire_wolf"]
        self.assertEqual((row["base"], row["templates"]), ("core:monster/wolf", ["dire"]))

    def test_the_example_rows_take_hp_and_mana_from_the_player_formulas(self):
        from kernel_reference import run_derived, run_monster_sheet
        rows = {row["key"]: row for row in monsters()}
        for key in ("dog", "wolf", "dire_wolf", "elder_wolf", "demonic_hound", "undead_horse"):
            with self.subTest(monster=key):
                row = rows[key]
                self.assertEqual(row["hitPoints"], ref().example_hp(row))
                derived = run_derived({"actor": {"attributes": row["attributes"], "level": row["level"]}})
                self.assertEqual(run_monster_sheet({"monster": row})["mp"], derived["mp"])

    def test_a_skill_missing_from_the_sheet_reads_the_starting_value(self):
        from kernel_reference import run_monster_sheet
        wolf = {row["key"]: row for row in monsters()}["wolf"]
        sheet = run_monster_sheet({"monster": wolf, "read_skills": ["stealth"]})
        self.assertEqual(sheet["skills"]["stealth"], 5 + wolf["attributes"]["agi"] * 2)

    def test_monster_attributes_pass_ten_and_growth_is_not_clamped(self):
        rows = resolve([{"key": "brute", "base": "core:creature_type/beast", "overrides": {"level": 30}}])
        self.assertGreater(rows["brute"]["attributes"]["str"], 10)
        store = validate.load_store()
        schema_uri = (validate.SCHEMA_DIR / "monster.schema.json").resolve().as_uri()
        resolver = validate.RefResolver(base_uri=schema_uri, referrer=store[schema_uri], store=store)
        validate.validate({"monsters": [rows["brute"]]}, store[schema_uri], resolver=resolver)


class TestAttacksAndAbilities(unittest.TestCase):
    """Natural and carried weapons, the attribute modifier, damage growth, and spell-like abilities (M3-10)."""

    def test_no_attack_has_an_amount_and_every_weapon_exists(self):
        natural = ref().natural_weapons()
        for row in monsters():
            for attack in row.get("attacks", []):
                with self.subTest(monster=row["key"], attack=attack["key"]):
                    self.assertNotIn("amount", attack.get("damage", {}))
                    weapon = attack["weapon"]
                    if weapon.startswith("core:natural_weapon/"):
                        self.assertIn(weapon.removeprefix("core:natural_weapon/"), natural)
                    else:
                        self.assertTrue(weapon.startswith("core:weapon/"))
                        self.assertIsNotNone(ref().weapon_row(weapon))
                    self.assertRegex(ref().attack_dice(row, attack), r"^\d+d\d+$")

    def test_a_large_creatures_natural_weapon_rolls_one_step_above_a_medium_ones(self):
        bite = {"key": "bite", "label": "Bite", "weapon": "core:natural_weapon/bite"}
        ladder = [rung for rungs in json.loads((DATA / "combat" / "dice_ladder.json").read_text())["dice_ladder"] for rung in [rungs]]
        rung = {dice: i for i, rungs in enumerate(ladder) for dice in rungs}
        medium = ref().attack_dice({"size": "medium"}, bite)
        large = ref().attack_dice({"size": "large"}, bite)
        self.assertEqual(rung[large], rung[medium] + 1)
        self.assertEqual(rung[ref().attack_dice({"size": "tiny"}, bite)], rung[medium] - 2)

    def test_an_attack_takes_the_modifier_of_its_own_attribute(self):
        claw = {"key": "claw", "label": "Claw", "weapon": "core:natural_weapon/claw", "ability": "agi"}
        stomp = {"key": "stomp", "label": "Stomp", "weapon": "core:natural_weapon/hoof"}
        for strength, agility, claw_modifier, stomp_modifier in ((10, 5, 0, 5), (5, 10, 5, 0), (3, 2, -3, -2)):
            row = {"size": "medium", "attributes": {"str": strength, "agi": agility}}
            with self.subTest(str=strength, agi=agility):
                self.assertEqual(ref().attack_modifier(row, claw), claw_modifier)
                self.assertEqual(ref().attack_modifier(row, stomp), stomp_modifier)

    def test_a_creature_above_its_rows_level_hits_by_damage_per_level(self):
        row = {r["key"]: r for r in monsters()}["fire_elemental"]
        summoned = ref().generate("core:monster/fire_elemental", level=row["level"] + 20)
        growth = json.loads((DATA / "monsters" / "bases" / "elemental.json").read_text())["growth"]["damagePerLevel"]
        step = round(growth * 20)
        self.assertGreater(step, 0)
        for before, after in zip(row["attacks"], summoned["attacks"]):
            self.assertEqual(after.get("damage", {}).get("bonus", 0), before.get("damage", {}).get("bonus", 0) + step)
        for before, after in zip(ref().damaging_abilities(row), ref().damaging_abilities(summoned)):
            self.assertEqual(after["magnitude"], before["magnitude"] + step)

    def test_an_ooze_fights_with_a_touch_ability(self):
        rows = {r["key"]: r for r in monsters()}
        for key, effect in (("ooze", "damage_acid"), ("acid_ooze", "damage_acid"), ("fire_ooze", "damage_fire"), ("frost_ooze", "damage_frost")):
            with self.subTest(monster=key):
                self.assertFalse(rows[key].get("attacks"))
                touches = [a for a in ref().damaging_abilities(rows[key]) if a["target"] == "touch"]
                self.assertEqual([a["effect"] for a in touches], [f"core:effect/{effect}"])

    def test_every_summoned_elemental_has_a_physical_attack_a_bolt_and_a_touch(self):
        rows = {r["key"]: r for r in monsters()}
        physical = {"bludgeoning", "piercing", "slashing"}
        # An element's damage effect, by its template's damage type (shock is a wood effect in the pool,
        # while the storm template is metal, so the phase does not name the element).
        element = {"fire_elemental": "fire", "ice_elemental": "frost", "storm_elemental": "shock",
                   "acid_elemental": "acid", "venom_elemental": "poison"}
        self.assertEqual(sorted(summon_group("elemental")), sorted(element))
        for key in summon_group("elemental"):
            row = rows[key]
            with self.subTest(monster=key):
                self.assertTrue(any(ref().attack_type(a) in physical for a in row["attacks"]))
                abilities = ref().damaging_abilities(row)
                for target in ("ranged", "touch"):
                    hits = [a for a in abilities if a["target"] == target]
                    self.assertTrue(hits, target)
                    self.assertEqual({a["effect"] for a in hits}, {f"core:effect/damage_{element[key]}"}, target)

    def test_a_spell_like_ability_spends_its_mana_and_the_cast_ap(self):
        from kernel_reference import run_monster_sheet
        row = {r["key"]: r for r in monsters()}["fire_elemental"]
        bolt = next(a for a in ref().damaging_abilities(row) if a["target"] == "ranged")
        cost = ref().ability_cost(bolt)
        self.assertGreater(cost, 0)
        mana = run_monster_sheet({"monster": row})["mp"]
        self.assertGreaterEqual(mana, cost)
        self.assertEqual(ref().use_ability(row, bolt, mana), {"mp": mana - cost, "ap": 3})
        self.assertEqual(ref().use_ability(row, bolt, cost - 1), {"refused": "mana"})

    def test_a_creature_whose_only_attack_is_an_ability_pays_no_mana(self):
        row = {r["key"]: r for r in monsters()}["ooze"]
        touch = ref().damaging_abilities(row)[0]
        self.assertEqual(ref().use_ability(row, touch, 0), {"mp": 0, "ap": 3})


class TestMonsterSkillsKeepPace(unittest.TestCase):
    """M3-17: a level-matched monster's fighting skill stays near a player's main skill (docs/core/actor-sheet.md)."""

    def test_a_monster_attack_skill_is_within_15_of_a_player_main_skill_at_every_level(self):
        import kernel_reference
        progression = kernel_reference.load_data("progression/core.json")["progression"]
        average = {"int": 5, "lck": 5, "agi": 5}
        # An average player: INT 5, LCK 5, every level's pool on one tag skill, an average luck roll.
        pool = kernel_reference.evaluate_formula(
            progression["skill_points"], average, {"luck_roll": average["lck"] / 2})
        start = kernel_reference.evaluate_formula(progression["skill_start"], {}, {"attribute": average["agi"]})
        for level in range(1, progression["max_level"] + 1):
            with self.subTest(level=level):
                player = min(progression["skill_cap"], start + pool / progression["skill_cost"]["tag"] * (level - 1))
                monster = kernel_reference.run_monster_sheet(
                    {"monster": {"attributes": average, "level": level}, "read_skills": ["evasion"]})
                self.assertLessEqual(abs(player - monster["skills"]["evasion"]), 15)


class TestNaturalWeaponsAreUnarmed(unittest.TestCase):
    """M3-16: a natural weapon is an unarmed attack (docs/modules/monsters.md#natural-weapons)."""

    CLAW = {"key": "claw", "label": "Claw", "weapon": "core:natural_weapon/claw", "ability": "agi"}

    def test_the_natural_weapons_file_names_unarmed(self):
        self.assertEqual(json.loads((DATA / "natural_weapons" / "core.json").read_text())["skill"], "core:skill/unarmed")

    def test_every_natural_attack_rolls_unarmed_and_a_carried_one_its_weapons_skill(self):
        for row in monsters():
            for attack in row.get("attacks", []):
                with self.subTest(monster=row["key"], attack=attack["key"]):
                    if attack["weapon"].startswith("core:natural_weapon/"):
                        self.assertEqual(ref().attack_skill(attack), "unarmed")
                    else:
                        self.assertEqual(ref().attack_skill(attack), ref().weapon_row(attack["weapon"])["weapon"]["skill"].rsplit("/", 1)[-1])

    def test_an_unarmed_perk_changes_a_natural_weapons_damage(self):
        features = {f["key"]: f for f in json.loads((DATA / "features" / "core.json").read_text())["features"]}
        jaguar = {"size": "medium", "attributes": {"agi": 8}}
        # A claw roll of 4, plus AGI 8's modifier of 3.
        self.assertEqual(ref().attack_damage(jaguar, self.CLAW, 4, []), 7)
        self.assertEqual(ref().attack_damage(jaguar, self.CLAW, 4, features["brawler"]["modifiers"]), 9)
        self.assertEqual(ref().attack_damage(jaguar, self.CLAW, 4, features["brawler"]["modifiers"] + features["iron_fist"]["modifiers"]), 13)
        sword = {"key": "rusty_blade", "label": "Rusty Blade", "weapon": "core:weapon/shortsword"}
        self.assertEqual(ref().attack_damage({"attributes": {"str": 5}}, sword, 4, features["brawler"]["modifiers"]), 4)

    def test_a_creature_with_no_attack_slams(self):
        self.assertEqual(ref().monster_attacks({"attributes": {"str": 6}}),
                         [{"key": "slam", "label": "Slam", "weapon": "core:natural_weapon/slam", "ability": "str"}])
        wolf = {r["key"]: r for r in monsters()}["wolf"]
        self.assertEqual(ref().monster_attacks(wolf), wolf["attacks"])

    def test_catman_claws_names_the_claw_row_and_rolls_its_dice(self):
        features = {f["key"]: f for f in json.loads((DATA / "features" / "core.json").read_text())["features"]}
        claws = features["catman_claws"]
        self.assertEqual(claws["natural_weapon"], "core:natural_weapon/claw")
        self.assertEqual(ref().feature_attack_dice(claws, "medium"), ref().natural_weapons()["claw"]["dice"])
        self.assertEqual(ref().feature_attack_dice(claws, "large"), ref().attack_dice({"size": "large"}, self.CLAW))

    def test_no_feature_writes_its_own_natural_weapon_dice(self):
        for feature in json.loads((DATA / "features" / "core.json").read_text())["features"]:
            if "natural_weapon" in feature:
                with self.subTest(feature=feature["key"]):
                    self.assertIsInstance(feature["natural_weapon"], str)


class TestMonstersTakePerks(unittest.TestCase):
    """M3-15: monster rows and templates hold perks (docs/modules/monsters.md#templates)."""

    def held(self, row):
        return {entry["perk"]: entry.get("rank", 1) for entry in row.get("perks", [])}

    def test_the_example_rows_and_templates_name_their_perks(self):
        rows = {r["key"]: r for r in monsters()}
        self.assertEqual(self.held(rows["troll"]), {"core:feature/toughness": 2})
        self.assertEqual(self.held(rows["bandit_captain"]), {"core:feature/armored_evasion": 1, "core:feature/alert": 1})
        templates = {t["key"]: t for t in json.loads((DATA / "monsters" / "templates.json").read_text())["monster_templates"]}
        self.assertEqual(self.held(templates["dire"]), {"core:feature/toughness": 1})
        self.assertEqual(self.held(templates["elder"]), {"core:feature/second_wind": 1})

    def test_a_templates_perk_adds_to_the_rows_keeping_the_higher_rank(self):
        rows = resolve([
            {"key": "a", "base": "core:creature_type/beast", "templates": ["dire"], "overrides": {"level": 3,
             "perks": {"$extend": [{"perk": "core:feature/alert"}]}}},
            {"key": "b", "base": "core:creature_type/beast", "templates": ["dire"], "overrides": {"level": 3,
             "perks": {"$extend": [{"perk": "core:feature/toughness", "rank": 3}]}}},
        ])
        self.assertEqual(self.held(rows["a"]), {"core:feature/toughness": 1, "core:feature/alert": 1})
        self.assertEqual(self.held(rows["b"]), {"core:feature/toughness": 3})

    def test_a_rank_past_the_perks_ranks_fails_and_a_social_perk_warns(self):
        errors, warnings = validate.check_monster_perks(
            [{"key": "x", "perks": [{"perk": "core:feature/toughness", "rank": 4}, {"perk": "core:feature/diplomat"}]}], [])
        self.assertTrue(any("toughness" in e and "rank 4" in e for e in errors), errors)
        self.assertTrue(any("diplomat" in w and "social" in w for w in warnings), warnings)
        self.assertEqual(validate.check_monster_perks([{"key": "y", "perks": [{"perk": "core:feature/toughness", "rank": 3}]}], []), ([], []))


if __name__ == "__main__":
    unittest.main()
