#!/usr/bin/env python3
"""Replay the example characters in ruleset/data/examples/characters.json, level by level.

The replay checks every choice against the rules (point-buy, perk prerequisites and ranks,
the skill-point pool and its costs, the skill cap), then generates two things from the
record: the vectors in ruleset/tests/vectors/characters.json and the tables in
docs/modules/character-creation.md. The tests fail if either differs from what the record
gives; `python ruleset/scripts/test_characters.py --write` regenerates both.

The expected derived stats are worked from the closed forms in docs/modules/progression.md,
not from the reference, so the vectors check the reference too.

Run with:
    python -m unittest ruleset/scripts/test_characters.py
"""

import json
import re
import sys
import unittest
from pathlib import Path

_SCRIPTS = Path(__file__).resolve().parent
sys.path.insert(0, str(_SCRIPTS))

from ancestry_reference import race_modifiers  # noqa: E402

ROOT = _SCRIPTS.parent.parent
DATA = ROOT / "ruleset" / "data"
RECORD = DATA / "examples" / "characters.json"
VECTORS = ROOT / "ruleset" / "tests" / "vectors" / "characters.json"
DOC = ROOT / "docs" / "modules" / "character-creation.md"
ABILITIES = ("str", "per", "end", "int", "wil", "agi", "cha", "lck")
DERIVED = ("hp", "mp", "ap", "carry", "initiative")
SKILL_CAP = 100


def load(rel):
    return json.loads((DATA / rel).read_text())


def key_of(ref):
    return ref.rsplit("/", 1)[-1]


def rows(rel, collection):
    return {row["key"]: row for row in load(rel)[collection]}


class Replay:
    """One character, replayed level by level. `sheets` holds one sheet per level."""

    def __init__(self, character):
        self.character = character
        self.skills_table = rows("skills/core.json", "skills")
        self.features = rows("features/core.json", "features")
        self.sheets = []
        self.replay()

    def fail(self, message):
        raise AssertionError(f"{self.character['key']}: {message}")

    def attributes(self):
        assigned = self.character["assigned"]
        if sorted(assigned) != sorted(ABILITIES):
            self.fail("assigned must name all eight attributes")
        if sum(value - 4 for value in assigned.values()) != 10:
            self.fail("the point-buy starts every attribute at 4 and distributes 10")
        final = dict(assigned)
        for modifier in race_modifiers(key_of(self.character["race"])):
            if modifier["target"] in final and modifier["operation"] == "add":
                final[modifier["target"]] += modifier["value"]
        return final

    def replay(self):
        attributes = self.attributes()
        self.final = attributes
        tags = {key_of(ref) for ref in self.character["tag_skills"]}
        background = rows("backgrounds/core.json", "backgrounds")[key_of(self.character["background"])]
        bonus = {key_of(b["skill"]): b["bonus"] for b in background.get("skill_bonuses", [])}
        skills = {}

        def skill(key):
            if key not in skills:
                ability = self.skills_table[key]["associatedAbility"]
                skills[key] = 5 + 2 * attributes[ability] + bonus.get(key, 0)
            return skills[key]

        for key in sorted(tags | set(bonus)):
            skill(key)
        ranks, gear, spells = {}, [], []
        for expected_level, entry in enumerate(self.character["levels"], start=1):
            level = entry["level"]
            if level != expected_level:
                self.fail(f"levels must run 1, 2, 3, ...; found {level} at {expected_level}")
            if level == 1 and entry["perk"] is not None:
                self.fail("a non-human takes no perk at level 1")
            if level > 1:
                if entry["perk"] is None:
                    self.fail(f"level {level} takes a perk")
                self.take(entry["perk"], level, attributes, skills, ranks)
            if not 0 <= entry["luck_roll"] <= attributes["lck"]:
                self.fail(f"level {level}: luck_roll is outside 0..LCK")
            pool = 5 + 2 * attributes["int"] + entry["luck_roll"]
            spent = 0
            for key, bought in entry["spend"].items():
                spent += bought * (1 if key in tags else 2)
                skills[key] = skill(key) + bought
                if skills[key] > SKILL_CAP:
                    self.fail(f"level {level}: {key} passes the cap of {SKILL_CAP}")
            if spent != pool:
                self.fail(f"level {level}: spends {spent} of a pool of {pool}")
            gear += entry.get("gear", [])
            spells += entry.get("spells", [])
            self.sheets.append({
                "level": level,
                "xp": 450 * level * (level - 1),
                "perk": entry["perk"],
                "ranks": dict(ranks),
                "pool": pool,
                "luck_roll": entry["luck_roll"],
                "skills": {key: skills[key] for key in sorted(tags)},
                "derived": self.derived(attributes, level, ranks),
                "modifiers": self.derived_modifiers(ranks),
                "gear": list(gear),
                "spells": list(spells),
            })

    def take(self, ref, level, attributes, skills, ranks):
        key = key_of(ref)
        perk = self.features.get(key)
        if perk is None or perk["type"] != "perk":
            self.fail(f"level {level}: {ref} is not a perk")
        need = perk.get("prerequisite", {})
        if level < need.get("level", 1):
            self.fail(f"level {level}: {key} needs level {need['level']}")
        for ability, minimum in need.get("abilities", {}).items():
            if attributes[ability] < minimum:
                self.fail(f"level {level}: {key} needs {ability} {minimum}")
        for requirement in need.get("skills", []):
            ids = requirement["id"] if isinstance(requirement["id"], list) else [requirement["id"]]
            if not any(skills.get(key_of(i), 0) >= requirement["min"] for i in ids):
                self.fail(f"level {level}: {key} needs {ids} at {requirement['min']}")
        for other in need.get("perks", []):
            if key_of(other) not in ranks:
                self.fail(f"level {level}: {key} needs {other}")
        if ranks.get(key, 0) >= perk.get("ranks", 1):
            self.fail(f"level {level}: {key} has no rank left")
        ranks[key] = ranks.get(key, 0) + 1

    def derived_modifiers(self, ranks):
        """The perk modifiers on a derived stat, one per perk at value times rank."""
        modifiers = []
        for key, rank in ranks.items():
            for modifier in self.features[key].get("modifiers", []):
                if modifier["target"] in DERIVED and modifier["operation"] == "add":
                    modifiers.append({**modifier, "value": modifier["value"] * rank})
        return modifiers

    def derived(self, attributes, level, ranks):
        """The closed forms of docs/modules/progression.md, plus the perks' adds."""
        values = {
            "hp": 15 + 8 * attributes["end"] + 4 * level,
            "mp": 8 * attributes["int"] + 2 * level,
            "sp": 10 + 5 * attributes["end"] + 2 * level,
            "ap": 2 + attributes["agi"] // 3,
            "carry": 25 + 10 * attributes["str"],
            "initiative": attributes["per"],
        }
        for modifier in self.derived_modifiers(ranks):
            values[modifier["target"]] += modifier["value"]
        return values


def replays():
    return [Replay(character) for character in json.loads(RECORD.read_text())["characters"]]


def vectors(all_replays):
    out = []
    for replay in all_replays:
        name = replay.character["label"]
        for sheet in replay.sheets:
            out.append({
                "name": f"{name} level {sheet['level']}: derived stats",
                "function": "derived",
                "actor": {"attributes": replay.final, "level": sheet["level"]},
                "modifiers": sheet["modifiers"],
                "expect": sheet["derived"],
            })
            out.append({
                "name": f"{name} level {sheet['level']}: skill points",
                "function": "progression",
                "operation": "level_up",
                "actor": {"attributes": replay.final},
                "luck_roll": sheet["luck_roll"],
                "hp_roll": 0,
                "mp_roll": 0,
                "sp_roll": 0,
                "expect": {"skill_points": sheet["pool"], "perks": 1, "gain": {"hp": 4, "mp": 2, "sp": 2}},
            })
    return out


def table(replay):
    tags = list(replay.sheets[0]["skills"])
    head = ["Level", "XP", "Perk taken", "Pool"] + [f"`{t}`" for t in tags] + ["HP", "MP", "AP", "Carry", "Init"]
    lines = ["| " + " | ".join(head) + " |", "|" + " --- |" * len(head)]
    for sheet in replay.sheets:
        perk = "—" if sheet["perk"] is None else f"`{key_of(sheet['perk'])}`"
        if sheet["perk"] is not None and sheet["ranks"][key_of(sheet["perk"])] > 1:
            perk += f" (rank {sheet['ranks'][key_of(sheet['perk'])]})"
        cells = [str(sheet["level"]), str(sheet["xp"]), perk, str(sheet["pool"])]
        cells += [str(sheet["skills"][t]) for t in tags]
        cells += [str(sheet["derived"][d]) for d in DERIVED]
        lines.append("| " + " | ".join(cells) + " |")
    last = replay.sheets[-1]
    lines.append("")
    lines.append("Gear by level 10: " + ", ".join(f"`{key_of(g)}`" for g in last["gear"]) + ".")
    spells = ", ".join(f"`{key_of(s)}`" for s in last["spells"]) or "none"
    lines.append(f"Spells by level 10: {spells}.")
    return "\n".join(lines)


def marker(key):
    return re.compile(rf"(<!-- characters:{key} -->\n)(.*?)(\n?<!-- /characters:{key} -->)", re.S)


def write():
    all_replays = replays()
    VECTORS.write_text(json.dumps(vectors(all_replays), indent=2) + "\n")
    text = DOC.read_text()
    for replay in all_replays:
        key = replay.character["key"]
        closing = f"\n<!-- /characters:{key} -->"
        text, count = marker(key).subn(lambda m: m.group(1) + table(replay) + closing, text)
        if count != 1:
            raise SystemExit(f"{DOC}: no <!-- characters:{key} --> section")
    DOC.write_text(text)


class TestExampleCharacters(unittest.TestCase):
    def setUp(self):
        self.replays = replays()

    def test_two_characters_one_martial_one_caster_to_level_10(self):
        self.assertEqual(sorted(r.character["concept"] for r in self.replays), ["caster", "martial"])
        for replay in self.replays:
            self.assertEqual(replay.sheets[-1]["level"], 10)

    def test_the_vectors_are_the_record(self):
        self.assertEqual(json.loads(VECTORS.read_text()), vectors(self.replays), "run with --write")

    def test_the_doc_tables_are_the_record(self):
        text = DOC.read_text()
        for replay in self.replays:
            with self.subTest(character=replay.character["key"]):
                match = marker(replay.character["key"]).search(text)
                self.assertIsNotNone(match, "the doc has no section for this character")
                assert match is not None
                self.assertEqual(match.group(2), table(replay), "run with --write")


if __name__ == "__main__":
    if sys.argv[1:] == ["--write"]:
        write()
    else:
        unittest.main()
