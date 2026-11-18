"""Suggested builds (docs/modules/builds/README.md, M5-47): every build is legal from creation to
level 30 at the lowest luck roll, and the checker catches each kind of break."""

import copy
import json
import sys
import unittest
from pathlib import Path

_SCRIPTS = Path(__file__).resolve().parent
sys.path.insert(0, str(_SCRIPTS))

import build_reference  # noqa: E402

DATA = _SCRIPTS.parent / "data"


def builds():
    return json.loads((DATA / "builds" / "core.json").read_text())["builds"]


class TestBuilds(unittest.TestCase):
    def test_every_build_is_legal_at_every_level(self):
        for build in builds():
            with self.subTest(build=build["key"]):
                problems, sheet = build_reference.play(build)
                self.assertEqual(problems, [])
                self.assertEqual(sheet["level"], 30)

    def test_the_melee_path_has_its_three_sub_paths(self):
        melee = sorted(b["sub_path"] for b in builds() if b["path"] == "melee")
        self.assertEqual(melee, ["duelist", "shield_bearer", "two_handed"])

    def test_the_magic_path_has_its_three_sub_paths_and_the_magic_swordsman(self):
        magic = sorted(b["sub_path"] for b in builds() if b["path"] == "magic")
        self.assertEqual(magic, ["blaster", "summoner", "support"])
        self.assertIn("magic_swordsman", [b["sub_path"] for b in builds() if b["path"] == "hybrid"])

    def test_the_ranged_path_has_its_three_sub_paths_and_the_last_two_hybrids(self):
        ranged = sorted(b["sub_path"] for b in builds() if b["path"] == "ranged")
        self.assertEqual(ranged, ["arbalist", "archer", "gunslinger"])
        hybrids = sorted(b["sub_path"] for b in builds() if b["path"] == "hybrid")
        self.assertEqual(hybrids, ["arcane_trickster", "magic_swordsman", "ranger"])

    def test_the_main_page_lists_all_twelve_builds(self):
        self.assertEqual(len(builds()), 12)
        page = (_SCRIPTS.parents[1] / "docs" / "modules" / "builds" / "README.md").read_text()
        for build in builds():
            with self.subTest(build=build["key"]):
                self.assertIn(f"| {build['label']} |", page)

    def test_a_builds_spells_are_of_a_color_whose_skill_it_raises(self):
        spells = {row["key"]: row for row in json.loads((_SCRIPTS.parent / "data" / "spells" / "core.json").read_text())["spells"]}
        for build in builds():
            raised = {p["skill"].rsplit("/", 1)[-1] for p in build["skill_priorities"]}
            for ref in build.get("spells", []):
                with self.subTest(build=build["key"], spell=ref):
                    self.assertIn(f"{spells[ref.rsplit('/', 1)[-1]]['color']}_magic", raised)

    def test_every_build_has_a_page_that_names_it(self):
        pages = {"melee": "melee.md", "magic": "magic.md", "ranged": "ranged.md", "hybrid": "hybrids.md"}
        docs = _SCRIPTS.parents[1] / "docs" / "modules" / "builds"
        for build in builds():
            with self.subTest(build=build["key"]):
                page = (docs / pages[build["path"]]).read_text()
                self.assertIn(f"## {build['label']}", page)
                self.assertIn("suggestion", page)

    def test_keys_are_unique(self):
        keys = [b["key"] for b in builds()]
        self.assertEqual(len(keys), len(set(keys)))


class TestTheChecker(unittest.TestCase):
    """Each kind of break is caught, on a copy of a legal build."""

    def broken(self, change):
        build = copy.deepcopy(next(b for b in builds() if b["key"] == "shield_bearer"))
        change(build)
        problems, _ = build_reference.play(build)
        return problems

    def test_a_perk_before_its_level(self):
        problems = self.broken(lambda b: b["perks_by_level"].update({"2": "core:feature/bulwark", "13": "core:feature/hardened_hide"}))
        self.assertTrue(any(p.startswith("level 2: bulwark unmet") for p in problems), problems)

    def test_a_perk_past_its_ranks(self):
        problems = self.broken(lambda b: b["perks_by_level"].update({"30": "core:feature/hardened_hide"}))
        self.assertTrue(any("'needs': 'rank'" in p for p in problems), problems)

    def test_a_missing_level(self):
        problems = self.broken(lambda b: b["perks_by_level"].pop("17"))
        self.assertIn("level 17: no perk", problems)

    def test_a_point_buy_that_is_not_ten_points(self):
        problems = self.broken(lambda b: b["attributes"].update({"str": 10}))
        self.assertIn("attributes: 4 each plus 10 points, none over 10", problems)

    def test_a_level_perk_as_a_creation_perk(self):
        problems = self.broken(lambda b: b.update(creation_perks=["core:feature/hardened_hide"]))
        self.assertIn("core:feature/hardened_hide is not a creation perk", problems)

    def test_a_skill_prerequisite_the_points_never_reach(self):
        problems = self.broken(lambda b: b.update(skill_priorities=[p for p in b["skill_priorities"] if p["skill"] != "core:skill/shields"]))
        self.assertTrue(any("shield_training unmet" in p for p in problems), problems)


if __name__ == "__main__":
    unittest.main()
