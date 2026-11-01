#!/usr/bin/env python3
"""Run every golden vector in ruleset/tests/vectors/ against the Python reference.

Adding a vector file needs no edit here. The checklist of kernel functions is the
table under "## Checklist" in docs/core/vectors.md.

Run with:
    python -m unittest ruleset/scripts/test_vectors.py
"""

import json
import re
import sys
import tempfile
import unittest
from pathlib import Path

_SCRIPTS = Path(__file__).resolve().parent
sys.path.insert(0, str(_SCRIPTS))

from kernel_reference import RUNNERS, run_vector  # noqa: E402

REPO_ROOT = _SCRIPTS.parent.parent
VECTOR_DIR = REPO_ROOT / "ruleset" / "tests" / "vectors"
VECTORS_DOC = REPO_ROOT / "docs" / "core" / "vectors.md"

CHECKLIST_ROW = re.compile(r"^\|\s*`(?P<function>[a-z_]+)`\s*\|[^|]*\|\s*(?P<status>[^|]+?)\s*\|\s*$")


def discover_vectors(vector_dir):
    """Return (file, vector) pairs for every *.json file in vector_dir, sorted by name."""
    pairs = []
    for path in sorted(Path(vector_dir).glob("*.json")):
        content = json.loads(path.read_text())
        vectors = content if isinstance(content, list) else [content]
        for vector in vectors:
            pairs.append((path, vector))
    return pairs


def vector_function(path, vector):
    """The kernel function a vector checks: its `function` field, or the file name."""
    return vector.get("function", Path(path).stem)


def read_checklist(doc_path):
    """Return {function: status} from the Checklist table in docs/core/vectors.md."""
    text = Path(doc_path).read_text()
    _, _, section = text.partition("## Checklist")
    section = section.split("\n## ", 1)[0]
    checklist = {}
    for line in section.splitlines():
        match = CHECKLIST_ROW.match(line)
        if match:
            checklist[match["function"]] = match["status"]
    return checklist


def is_covered(status):
    return status == "covered"


class TestVectorDiscovery(unittest.TestCase):
    def test_a_new_vector_file_is_discovered_without_a_test_edit(self):
        with tempfile.TemporaryDirectory() as tmp:
            Path(tmp, "formula.json").write_text(
                json.dumps({"name": "lit", "expr": {"lit": 3}, "expect": {"value": 3}})
            )
            Path(tmp, "extra_formulas.json").write_text(
                json.dumps(
                    [{"name": "lit 4", "function": "formula", "expr": {"lit": 4}, "expect": {"value": 4}}]
                )
            )
            pairs = discover_vectors(tmp)
            self.assertEqual([vector["name"] for _, vector in pairs], ["lit 4", "lit"])
            self.assertEqual({vector_function(p, v) for p, v in pairs}, {"formula"})

    def test_the_vector_directory_has_vectors(self):
        self.assertTrue(discover_vectors(VECTOR_DIR))


class TestEveryVector(unittest.TestCase):
    def test_every_vector_matches_the_reference(self):
        for path, vector in discover_vectors(VECTOR_DIR):
            name = vector.get("name", "<no name>")
            with self.subTest(file=path.name, vector=name):
                self.assertIn("name", vector, "a vector needs a name")
                self.assertIn("expect", vector, "a vector needs expect")
                function = vector_function(path, vector)
                self.assertIn(function, RUNNERS, f"no runner for kernel function {function!r}")
                result = run_vector(function, vector)
                for field, expected in vector["expect"].items():
                    self.assertIn(field, result, f"reference result has no field {field!r}")
                    self.assertEqual(result[field], expected, f"field {field!r}")

    def test_vector_names_are_unique_in_each_file(self):
        names_by_file = {}
        for path, vector in discover_vectors(VECTOR_DIR):
            names_by_file.setdefault(path.name, []).append(vector.get("name"))
        for file_name, names in names_by_file.items():
            with self.subTest(file=file_name):
                self.assertEqual(len(names), len(set(names)), "duplicate vector name")


class TestKernelChecklist(unittest.TestCase):
    def setUp(self):
        self.checklist = read_checklist(VECTORS_DOC)
        self.functions_with_vectors = {
            vector_function(path, vector) for path, vector in discover_vectors(VECTOR_DIR)
        }

    def test_checklist_names_the_kernel_functions(self):
        for function in ("resolution", "derived", "formula", "modifiers", "actions", "progression"):
            self.assertIn(function, self.checklist)

    def test_checklist_statuses_are_covered_or_pending(self):
        for function, status in self.checklist.items():
            with self.subTest(function=function):
                self.assertRegex(status, r"^(covered|pending [A-Z][A-Z0-9]*-\d{2})$")

    def test_every_covered_function_has_a_vector_and_a_runner(self):
        for function, status in self.checklist.items():
            if is_covered(status):
                with self.subTest(function=function):
                    self.assertIn(function, self.functions_with_vectors)
                    self.assertIn(function, RUNNERS)

    def test_a_pending_function_with_a_vector_is_marked_covered(self):
        for function, status in self.checklist.items():
            if not is_covered(status):
                with self.subTest(function=function):
                    self.assertNotIn(
                        function,
                        self.functions_with_vectors,
                        "has a vector; set its checklist row to covered",
                    )

    def test_every_vector_function_is_in_the_checklist(self):
        for function in self.functions_with_vectors:
            with self.subTest(function=function):
                self.assertIn(function, self.checklist)


if __name__ == "__main__":
    unittest.main()
