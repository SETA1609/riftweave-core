#!/usr/bin/env python3
"""An agent-played alpha run (M5-18): play builds from creation to level 30 through the agent CLI.

Every step goes through the `session` command of the `riftweave` binary, so the saved session is a
real run of the core and replays to the same actor (docs/backend/sessions.md). The driver holds no
rule: it picks the choices a player would make (a perk, where skill points go, which spell to cast,
which monster to fight), draws the rolls from a seeded generator, and puts each roll in the step's
input. What a session cannot do yet (buy skill ranks, take a perk, learn a spell, damage a monster
with a spell) it does through dev mode, and it records each such step as a finding.

Usage (from the repository root, with the binary built):
    (cd backend && go build -o ../cache/riftweave ./cmd/riftweave)
    python ruleset/scripts/playtest_driver.py --binary cache/riftweave --out docs/playtests

It writes one session per build to <out>/sessions/<build>.json and the tallies to
<out>/report.json. The same seed gives the same sessions.
"""

import argparse
import json
import random
import re
import subprocess
import sys
from collections import Counter
from pathlib import Path

_SCRIPTS = Path(__file__).resolve().parent
sys.path.insert(0, str(_SCRIPTS))

import monster_reference  # noqa: E402
import validate  # noqa: E402

DATA = _SCRIPTS.parent / "data"
ABILITIES = ("str", "per", "end", "int", "wil", "agi", "cha", "lck")
MAX_LEVEL = 30
ROUNDS_PER_FIGHT = 12

# A build is what a player would decide at creation and keep to: attributes (4 each plus 10
# points), tag skills, where points go, which perks come first, and what it fights with.
BUILDS = [
    {"key": "melee_blades", "label": "Blades fighter", "race": "core:race/dwarf",
     "attributes": {"str": 7, "per": 5, "end": 7, "int": 4, "wil": 4, "agi": 6, "cha": 4, "lck": 5},
     "tags": ["blades", "evasion", "heavy_armor"], "weapon": ("1d8", "slashing", "str"), "attack": "blades",
     "spend": ["blades", "evasion", "heavy_armor", "athletics"], "colors": [],
     "perk_order": ["toughness", "power_attack", "alert", "shield_wall", "heavy_armor_expert"]},
    {"key": "ranged_bows", "label": "Archer", "race": "core:race/elf",
     "attributes": {"str": 4, "per": 8, "end": 5, "int": 4, "wil": 4, "agi": 8, "cha": 4, "lck": 5},
     "tags": ["bows", "evasion", "stealth"], "weapon": ("1d8", "piercing", "per"), "attack": "bows",
     "spend": ["bows", "evasion", "stealth", "light_armor"], "colors": [],
     "perk_order": ["deadeye", "alert", "toughness", "ghost"]},
    {"key": "caster_red", "label": "Red mage", "race": "core:race/human",
     "attributes": {"str": 4, "per": 5, "end": 5, "int": 9, "wil": 6, "agi": 5, "cha": 4, "lck": 4},
     "tags": ["red_magic", "evasion", "lore"], "weapon": ("1d4", "piercing", "agi"), "attack": "red_magic",
     "spend": ["red_magic", "evasion", "lore"], "colors": ["red"],
     "perk_order": ["pyromancer", "toughness", "alert"]},
    {"key": "caster_yellow_indigo", "label": "Yellow and indigo mage", "race": "core:race/human",
     "attributes": {"str": 4, "per": 5, "end": 5, "int": 9, "wil": 7, "agi": 5, "cha": 4, "lck": 3},
     "tags": ["yellow_magic", "indigo_magic", "evasion"], "weapon": ("1d4", "piercing", "agi"),
     "attack": "yellow_magic", "spend": ["yellow_magic", "indigo_magic", "evasion"], "colors": ["yellow", "indigo"],
     "perk_order": ["toughness", "alert"]},
    {"key": "caster_blue_green", "label": "Blue and green mage", "race": "core:race/human",
     "attributes": {"str": 4, "per": 5, "end": 6, "int": 9, "wil": 6, "agi": 5, "cha": 4, "lck": 3},
     "tags": ["blue_magic", "green_magic", "evasion"], "weapon": ("1d4", "piercing", "agi"),
     "attack": "blue_magic", "spend": ["blue_magic", "green_magic", "evasion"], "colors": ["blue", "green"],
     "perk_order": ["toughness", "alert"]},
    {"key": "caster_white_violet", "label": "White and violet mage", "race": "core:race/human",
     "attributes": {"str": 4, "per": 5, "end": 6, "int": 8, "wil": 7, "agi": 5, "cha": 4, "lck": 3},
     "tags": ["white_magic", "violet_magic", "evasion"], "weapon": ("1d6", "bludgeoning", "str"),
     "attack": "white_magic", "spend": ["white_magic", "violet_magic", "evasion"], "colors": ["white", "violet"],
     "perk_order": ["toughness", "alert"]},
    {"key": "caster_black_orange", "label": "Black and orange mage", "race": "core:race/human",
     "attributes": {"str": 4, "per": 5, "end": 6, "int": 9, "wil": 6, "agi": 5, "cha": 4, "lck": 3},
     "tags": ["black_magic", "orange_magic", "evasion"], "weapon": ("1d4", "piercing", "agi"),
     "attack": "black_magic", "spend": ["black_magic", "orange_magic", "evasion"], "colors": ["black", "orange"],
     "perk_order": ["toughness", "alert"]},
    {"key": "hybrid_blades_red", "label": "Spellblade", "race": "core:race/human",
     "attributes": {"str": 6, "per": 5, "end": 6, "int": 7, "wil": 4, "agi": 5, "cha": 4, "lck": 5},
     "tags": ["blades", "red_magic", "evasion"], "weapon": ("1d8", "slashing", "str"), "attack": "blades",
     "spend": ["blades", "red_magic", "evasion", "medium_armor"], "colors": ["red"],
     "perk_order": ["power_attack", "toughness", "pyromancer", "alert"]},
]


def load(rel, collection):
    return json.loads((DATA / rel).read_text())[collection]


def key_of(ref):
    return ref.rsplit("/", 1)[-1]


class Binary:
    """The agent CLI. Each call is one process, as an agent would make it."""

    def __init__(self, path):
        self.path = str(path)
        self.calls = 0
        # The progression command's operation key, `op` before M5-46 and `operation` after it.
        described = json.loads(subprocess.run([self.path, "describe", "progression"], capture_output=True,
                                               text=True, check=True).stdout)
        self.operation = "operation" if "operation" in described["input_schema"]["properties"] else "op"

    def run(self, command, payload, dev=False):
        self.calls += 1
        args = [self.path] + (["--dev"] if dev else []) + [command, "-"]
        done = subprocess.run(args, input=json.dumps(payload), capture_output=True, text=True)
        if done.returncode != 0:
            raise RuntimeError(f"{command}: {done.stderr.strip()}")
        return json.loads(done.stdout)


class Run:
    def __init__(self, build, binary, seed):
        self.build, self.cli = build, binary
        self.rng = random.Random(f"{seed}:{build['key']}")
        self.features = {row["key"]: row for row in load("features/core.json", "features")}
        self.skills = {row["key"]: row for row in load("skills/core.json", "skills")}
        self.spells = load("spells/core.json", "spells")
        self.taken = Counter()
        self.findings = Counter()
        self.fights = {"won": 0, "lost": 0, "fled": 0, "rounds": 0}
        self.attacks = {"player": Counter(), "monster": Counter(), "player_damage": 0, "monster_damage": 0}
        self.capped_at = {}  # skill -> the level it first reached the cap
        self.session = None

    # --- session plumbing -------------------------------------------------------------------
    def step(self, command, input, as_=None):
        step = {"command": command, "input": input}
        if as_:
            step["as"] = as_
        out = self.cli.run("session", {"operation": "step", "session": self.session, "step": step}, dev=True)
        self.session = out["session"]
        return out.get("result")

    def dev(self, operation, **fields):
        self.findings[f"dev:{operation}"] += 1
        out = self.cli.run("dev", {"operation": operation, "session": self.session, **fields}, dev=True)
        self.session = out["session"]

    @property
    def actor(self):
        return self.session["actor"]

    def skill(self, key):
        written = self.actor["skills"].get(key)
        if written is not None:
            return written
        return 5 + 2 * self.actor["attributes"][self.skills[key]["associatedAbility"]]

    # --- creation ---------------------------------------------------------------------------
    def start(self):
        attributes = dict(self.build["attributes"])
        derived = self.cli.run("derived", {"actor": {"attributes": attributes, "level": 1}})
        skills = {key: 5 + 2 * attributes[self.skills[key]["associatedAbility"]] for key in self.build["tags"]}
        actor = {"attributes": attributes, "level": 1, "xp": 0, "skills": skills, "tags": list(self.build["tags"]),
                 "growth": {"hp": 0, "mp": 0, "sp": 0},
                 "resources": {"hp": derived["hp"], "mp": derived["mp"], "sp": derived["sp"], "ap": derived["ap"]},
                 "conditions": [], "modifiers": [], "flags": [], "inventory": [], "equipped": {}}
        self.session = self.cli.run("session", {"operation": "start", "start": actor})["session"]
        self.spend_pool(first=True)

    # --- level-up ---------------------------------------------------------------------------
    def spend_pool(self, first=False):
        lck = self.actor["attributes"]["lck"]
        highest = lck // 2
        rolls = {"luck_roll": self.rng.randint(0, lck)}
        rolls.update({f"{pool}_roll": 0 if first else self.rng.randint(0, highest) for pool in ("hp", "mp", "sp")})
        gained = self.step("progression", {self.cli.operation: "level_up", **rolls})
        points = gained["skill_points"]
        order = self.build["spend"]
        index = 0
        while points >= 1 and index < 4 * len(order):
            key = order[index % len(order)]
            index += 1
            share = max(points // 2, 1) if index <= len(order) else points
            spent = self.step("progression", {self.cli.operation: "spend", "skill": self.skill(key),
                                              "tagged": key in self.actor["tags"], "points": share})
            if spent["ranks"]:
                self.dev("skill", key=key, add=spent["ranks"])
                points -= spent["spent"]
                if self.skill(key) >= 100:
                    self.capped_at.setdefault(key, self.actor["level"])
            if spent["ranks"] == 0 and share == points:
                continue

    def eligible(self, key):
        perk = self.features.get(key)
        if perk is None or perk["type"] != "perk":
            return False
        if self.actor.get("perks", {}).get(key, 0) >= perk.get("ranks", 1):
            return False
        need = perk.get("prerequisite", {})
        if self.actor["level"] < need.get("level", 1):
            return False
        if any(self.actor["attributes"][a] < m for a, m in need.get("abilities", {}).items()):
            return False
        for requirement in need.get("skills", []):
            ids = requirement["id"] if isinstance(requirement["id"], list) else [requirement["id"]]
            if not any(self.skill(key_of(i)) >= requirement["min"] for i in ids if key_of(i) in self.skills):
                return False
        return all(key_of(p) in self.actor.get("perks", {}) for p in need.get("perks", []))

    def take_perk(self):
        order = list(self.build["perk_order"])
        rest = sorted(k for k, f in self.features.items() if f["type"] == "perk")
        for key in order + [k for k in rest if k not in order]:
            if self.eligible(key):
                self.dev("perk", key=key)
                self.taken[key] += 1
                return key
        self.findings["no eligible perk at a level-up"] += 1
        return None

    def learn_spells(self):
        known = set(self.actor.get("spells", []))
        for spell in self.spells:
            if spell["color"] not in self.build["colors"] or spell["key"] in known:
                continue
            if spell["level"] <= self.actor["level"] and self.skill(f"{spell['color']}_magic") >= spell["cost"].get("mana", 0):
                self.dev("learn", spell=f"core:spell/{spell['key']}")

    def level_up(self):
        level = self.actor["level"] + 1
        xp = 450 * level * (level - 1)
        self.step("progression", {self.cli.operation: "level", "xp": xp})
        self.spend_pool()
        self.take_perk()
        self.learn_spells()

    # --- fighting ---------------------------------------------------------------------------
    def opponent(self):
        rows = [r for r in monster_reference.raw_monsters() if "familiar" not in r["key"]]
        data = {"monsters": json.loads(json.dumps(rows))}
        validate._resolve_monster_bases(data, [])
        level = self.actor["level"]
        near = [r for r in data["monsters"] if abs(r.get("level", 1) - level) <= 2 and (r.get("attacks") or [{}])[0].get("natural_weapon", (r.get("attacks") or [{}])[0].get("weapon"))]
        if not near:
            self.findings["no monster within 2 levels with a weapon attack"] += 1
            near = sorted(data["monsters"], key=lambda r: abs(r.get("level", 1) - level))[:3]
            near = [r for r in near if (r.get("attacks") or [{}])[0].get("natural_weapon", (r.get("attacks") or [{}])[0].get("weapon"))] or near
        row = self.rng.choice(near)
        attack = (row.get("attacks") or [{}])[0]
        try:
            dice, kind = monster_reference.attack_dice(row, attack), monster_reference.attack_type(attack)
        except (KeyError, TypeError):
            dice, kind = "1d6", "bludgeoning"
            self.findings["monster with no rollable attack"] += 1
        skill = "unarmed" if "natural_weapon" in attack else "blades"
        sheet = self.cli.run("monster_sheet", {"monster": {k: row[k] for k in ("attributes", "level", "hitPoints", "skills") if k in row},
                                                "read_skills": [skill, "evasion"]})
        return {"key": row["key"], "level": row.get("level", 1), "hp": sheet["hp"], "attack": sheet["skills"][skill],
                "evasion": sheet["skills"]["evasion"], "dice": dice, "type": kind, "resist": row.get("resist") or {},
                "lck": row["attributes"].get("lck", 0)}

    def roll(self, dice):
        count, sides = (int(x) for x in re.fullmatch(r"(\d+)d(\d+)", dice).groups())
        return sum(self.rng.randint(1, sides) for _ in range(count))

    def best_spell(self):
        known = set(self.actor.get("spells", []))
        castable = []
        for spell in self.spells:
            damage = [e for e in spell["effects"] if e["effect"].startswith("core:effect/damage_")]
            if spell["key"] in known and damage and spell["cost"].get("mana", 0) <= self.actor["resources"]["mp"]:
                castable.append((damage[0]["magnitude"], spell))
        return max(castable, key=lambda pair: pair[0])[1] if castable else None

    def armor_dr(self):
        return min(1 + self.actor["level"] // 5, 6) if self.build["key"] != "caster_red" else 0

    def fight(self):
        foe = self.opponent()
        hp = foe["hp"]
        self.findings[f"fought level gap {foe['level'] - self.actor['level']:+d}"] += 1
        for _ in range(ROUNDS_PER_FIGHT):
            self.fights["rounds"] += 1
            spell = self.best_spell() if self.build["colors"] else None
            if spell is not None and self.build["attack"].endswith("_magic"):
                cast = self.step("casting", {"spell": spell["key"], "roll": self.rng.randint(1, 100)})
                self.attacks["player"][cast["outcome"]] += 1
                if cast["outcome"] == "cast":
                    damage = sum(e["magnitude"] for e in cast["effects"] if e["effect"].startswith("core:effect/damage_"))
                    kind = next(e["effect"] for e in cast["effects"] if e["effect"].startswith("core:effect/damage_"))
                    dealt = max(damage - foe["resist"].get(key_of(kind).removeprefix("damage_"), 0), 0)
                    self.attacks["player_damage"] += min(dealt, hp)
                    hp = max(hp - dealt, 0)
                    self.findings["spell damage applied by the driver, not a command"] += 1
            else:
                dice, kind, ability = self.build["weapon"]
                hit = self.step("combat", {"attacker": {"skill": self.skill(self.build["attack"]), "roll": self.rng.randint(1, 100)},
                                           "defender": {"evasion_skill": foe["evasion"], "armor_dr": 0, "hp": hp,
                                                        "resist": foe["resist"], "attributes": {"lck": foe["lck"]}},
                                           "damage": {"roll": self.roll(dice), "type": kind, "ability": ability}})
                self.attacks["player"][hit["outcome"]] += 1
                self.attacks["player_damage"] += hp - hit["hp"]
                hp = hit["hp"]
            if hp == 0:
                self.fights["won"] += 1
                return
            back = self.step("combat", {"attacker": {"skill": foe["attack"], "roll": self.rng.randint(1, 100), "attributes": {"lck": foe["lck"]}},
                                        "defender": {"evasion_skill": self.skill("evasion"), "armor_dr": self.armor_dr()},
                                        "damage": {"roll": self.roll(foe["dice"]), "type": foe["type"]}}, as_="defender")
            self.attacks["monster"][back["outcome"]] += 1
            self.attacks["monster_damage"] += back.get("damage", 0) if back["outcome"] == "hit" else 0
            if back["hp"] == 0:
                self.fights["lost"] += 1
                self.dev("resource", pool="hp", set=1)
                return
        self.fights["fled"] += 1

    def rest(self):
        self.step("rest", {"occasion": "long_rest"})

    def play(self):
        self.start()
        self.learn_spells()
        while True:
            self.fight()
            self.rest()
            if self.actor["level"] >= MAX_LEVEL:
                break
            self.level_up()
        replay = self.cli.run("session", {"operation": "replay", "session": self.session}, dev=True)
        return replay["matches"]


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--binary", required=True)
    parser.add_argument("--out", required=True)
    parser.add_argument("--seed", default="m5-18")
    parser.add_argument("--only", nargs="*")
    args = parser.parse_args()
    out = Path(args.out)
    (out / "sessions").mkdir(parents=True, exist_ok=True)
    cli = Binary(args.binary)
    report = {"seed": args.seed, "builds": {}, "perks_taken": Counter(), "findings": Counter()}
    for build in BUILDS:
        if args.only and build["key"] not in args.only:
            continue
        run = Run(build, cli, args.seed)
        matches = run.play()
        (out / "sessions" / f"{build['key']}.json").write_text(json.dumps(run.session, separators=(",", ":")) + "\n")
        actor = run.actor
        report["builds"][build["key"]] = {
            "label": build["label"], "level": actor["level"], "replay_matches": matches, "fights": run.fights,
            "steps": len(run.session["log"]), "skills": actor["skills"], "perks": actor.get("perks", {}),
            "spells": actor.get("spells", []), "findings": dict(run.findings),
            "attacks": {"player": dict(run.attacks["player"]), "monster": dict(run.attacks["monster"]),
                        "player_damage": run.attacks["player_damage"], "monster_damage": run.attacks["monster_damage"]},
            "capped_at": run.capped_at}
        report["perks_taken"].update(run.taken)
        report["findings"].update(run.findings)
        print(f"{build['key']}: level {actor['level']}, replay {matches}, fights {run.fights}, steps {len(run.session['log'])}")
    report["calls"] = cli.calls
    (out / "report.json").write_text(json.dumps(report, indent=2, sort_keys=True) + "\n")


if __name__ == "__main__":
    main()
