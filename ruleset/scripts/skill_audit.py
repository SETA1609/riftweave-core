#!/usr/bin/env python3
"""Skill audit: every skill needs a governing attribute and a consumer.

Rule: docs/core/checks.md, "Every skill has a consumer". A consumer is a weapon,
an armor category, a spell or effect color, a check (check row or recipe), or a
perk. Background bonuses are references, not consumers.

Also checks that the "Seeds skills" lines in docs/modules/attributes.md match
each skill's associatedAbility.

Usage:
    python ruleset/scripts/skill_audit.py          # table, exit 1 on a problem
    python ruleset/scripts/skill_audit.py --quiet  # problems only
"""

import json
import re
import sys
from collections import defaultdict
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
DATA_DIR = ROOT / "ruleset" / "data"
ATTRIBUTES_DOC = ROOT / "docs" / "modules" / "attributes.md"

SKILL_REF = re.compile(r"^core:skill/([a-z_][a-z0-9_]*)$")
SKILL_EFFECTS = {
    "core:effect/fortify_skill",
    "core:effect/drain_skill",
    "core:effect/damage_skill",
}
ARMOR_CATEGORIES = {"light", "medium", "heavy"}

# Consumer kinds; anything else found is a reference only.
CONSUMERS = ("weapon", "armor", "spell", "check", "perk")
COLLECTION_KIND = {
    "features": "perk",
    "spells": "spell",
    "effects": "spell",
    "checks": "check",
    "recipes": "check",
    "backgrounds": "background",
}


def _skill_key(value):
    """Return the skill key named by a ref string, or None."""
    if not isinstance(value, str):
        return None
    m = SKILL_REF.match(value)
    return m.group(1) if m else None


def _walk(node, kind, add):
    """Record every skill named anywhere under node."""
    if isinstance(node, dict):
        weapon = node.get("weapon")
        if isinstance(weapon, dict) and _skill_key(weapon.get("skill")):
            add(_skill_key(weapon["skill"]), "weapon")
        armor = node.get("armor")
        if isinstance(armor, dict) and armor.get("category") in ARMOR_CATEGORIES:
            add(f"{armor['category']}_armor", "armor")
        if kind in ("spell", "perk") and isinstance(node.get("color"), str):
            add(f"{node['color']}_magic", "spell")
        target = node.get("target") or node.get("effect")
        if target in SKILL_EFFECTS:
            param = node.get("parameter")
            # A bare key is the legacy form; count it so it is not lost.
            add(_skill_key(param) or param, kind)
        for value in node.values():
            _walk(value, kind, add)
    elif isinstance(node, list):
        for value in node:
            _walk(value, kind, add)
    else:
        key = _skill_key(node)
        if key:
            add(key, kind)


def collect_uses(data_dir=DATA_DIR):
    """Map skill key -> {kind: set of rel paths} over every JSON file."""
    uses = defaultdict(lambda: defaultdict(set))
    skills_file = data_dir / "skills" / "core.json"
    for fp in sorted(data_dir.rglob("*.json")):
        if fp == skills_file:
            continue
        data = json.loads(fp.read_text(encoding="utf-8"))
        rel = fp.relative_to(data_dir).as_posix()

        def add(skill, kind, rel=rel):
            if isinstance(skill, str):
                uses[skill][kind].add(rel)

        if not isinstance(data, dict) or "$schema" not in data:
            # Base templates (equipment/bases, monsters/bases) are one entry.
            _walk(data, "other", add)
            continue
        for coll, value in data.items():
            if coll != "$schema":
                _walk(value, COLLECTION_KIND.get(coll, "other"), add)
    return uses


def load_skills(data_dir=DATA_DIR):
    data = json.loads((data_dir / "skills" / "core.json").read_text(encoding="utf-8"))
    return data["skills"]


def parse_seeds(doc_path=ATTRIBUTES_DOC):
    """Map ability id -> set of skill keys from the attributes doc."""
    seeds = {}
    ability = None
    in_seeds = False
    heading = re.compile(r"^### .*\(([A-Z]{3})\)\s*$")
    for line in doc_path.read_text(encoding="utf-8").splitlines():
        m = heading.match(line)
        if m:
            ability, in_seeds = m.group(1).lower(), False
            seeds[ability] = set()
            continue
        if line.startswith("## "):
            ability = None
        if ability and "**Seeds skills:**" in line:
            in_seeds = True
        elif not line.startswith("  ") or not line.strip():
            in_seeds = False  # a wrapped bullet continues with indented lines
        if in_seeds:
            for token in re.findall(r"`([^`]+)`", line):
                if re.fullmatch(r"[a-z][a-z0-9_]*", token):
                    seeds[ability].add(token)
    return seeds


def audit(skills, uses, seeds):
    """Return (rows, problems). rows: (key, ability, consumers, references)."""
    rows, problems = [], []
    known = {s["key"] for s in skills}
    data_seeds = defaultdict(set)
    for s in skills:
        key = s["key"]
        ability = s.get("associatedAbility")
        if not ability:
            problems.append(f"skill {key}: no associatedAbility")
        else:
            data_seeds[ability].add(key)
        kinds = uses.get(key, {})
        consumers = sorted(k for k in kinds if k in CONSUMERS)
        refs = sorted(k for k in kinds if k not in CONSUMERS)
        rows.append((key, ability or "-", consumers, refs))
        if not consumers:
            why = "no references" if not kinds else "references only: " + ", ".join(refs)
            problems.append(f"orphan skill {key}: {why}")
    for name in sorted(set(uses) - known):
        files = sorted({f for paths in uses[name].values() for f in paths})
        problems.append(f"unknown skill {name!r} referenced in {', '.join(files)}")
    for ability in sorted(set(data_seeds) | set(seeds)):
        doc, data = seeds.get(ability, set()), data_seeds.get(ability, set())
        if doc != data:
            missing = ", ".join(sorted(data - doc)) or "-"
            extra = ", ".join(sorted(doc - data)) or "-"
            problems.append(
                f"attributes.md seeds for {ability}: missing {missing}; not in data {extra}"
            )
    return rows, problems


def main(argv=None):
    argv = sys.argv[1:] if argv is None else argv
    quiet = "--quiet" in argv
    skills = load_skills()
    rows, problems = audit(skills, collect_uses(), parse_seeds())
    if not quiet:
        print(f"{'skill':<18} {'attr':<5} {'consumers':<24} references")
        for key, ability, consumers, refs in rows:
            print(f"{key:<18} {ability:<5} {','.join(consumers) or '-':<24} {','.join(refs) or '-'}")
        print()
    for p in problems:
        print(f"FAIL {p}")
    print(f"{len(skills)} skill(s), {len(problems)} problem(s)")
    return 1 if problems else 0


if __name__ == "__main__":
    sys.exit(main())
