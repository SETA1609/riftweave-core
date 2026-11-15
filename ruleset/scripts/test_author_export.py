#!/usr/bin/env python3
"""The author command's exports pass validate.py unchanged (docs/backend/authoring.md, APP-10).

The Go test TestTheExportsMatchTheirFixtures keeps ruleset/tests/fixtures/author/*.json equal to
what the command writes. This test puts each export where its `path` says, in a copy of the
ruleset, and runs validate.py there.
"""
import json
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent.parent
FIXTURES = ROOT / "ruleset" / "tests" / "fixtures" / "author"


def validate_with(*exports):
    """Copy the ruleset, add each export at its path, and run validate.py on the copy."""
    with tempfile.TemporaryDirectory() as tmp:
        copy = Path(tmp)
        shutil.copytree(ROOT / "ruleset", copy / "ruleset", ignore=shutil.ignore_patterns("__pycache__", "tests"))
        for export in exports:
            target = copy / export["path"]
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_text(json.dumps(export["file"], indent=2) + "\n")
            if "manifest" in export:
                (target.parent.parent / "manifest.json").write_text(json.dumps(export["manifest"], indent=2) + "\n")
        return subprocess.run([sys.executable, str(copy / "ruleset" / "scripts" / "validate.py")], capture_output=True, text=True)


class TestAuthorExports(unittest.TestCase):
    def test_a_core_export_passes_validate_unchanged(self):
        export = json.loads((FIXTURES / "core.json").read_text())
        self.assertEqual(export["shape"], "core")
        self.assertTrue(export["path"].startswith("ruleset/data/spells/"))
        done = validate_with(export)
        self.assertEqual(done.returncode, 0, done.stdout[-2000:] + done.stderr[-2000:])

    def test_a_module_export_passes_validate_unchanged(self):
        export = json.loads((FIXTURES / "module.json").read_text())
        self.assertEqual(export["shape"], "module")
        self.assertTrue(export["path"].startswith("ruleset/modules/authored/data/"))
        self.assertEqual(export["refs"], ["module:authored/spell/authored_spark"])
        done = validate_with(export)
        self.assertEqual(done.returncode, 0, done.stdout[-2000:] + done.stderr[-2000:])


if __name__ == "__main__":
    unittest.main()
