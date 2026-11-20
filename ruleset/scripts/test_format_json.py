"""The one JSON format (CONTRIBUTING.md#json-format, M5-51)."""
import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

SCRIPTS = Path(__file__).resolve().parent
ROOT = SCRIPTS.parent.parent
sys.path.insert(0, str(SCRIPTS))

import format_json  # noqa: E402


def fmt(text, width=100):
    return format_json.format_text(text, width=width)


class TestLayout(unittest.TestCase):
    def test_a_short_object_stays_on_one_line_with_spaces_inside_its_braces(self):
        self.assertEqual(fmt('{"target":"hp","op":"add","value":15}'), '{ "target": "hp", "op": "add", "value": 15 }\n')

    def test_a_short_array_stays_on_one_line_without_spaces_inside_its_brackets(self):
        self.assertEqual(fmt('{"tags": ["fire",   "cold"]}'), '{ "tags": ["fire", "cold"] }\n')

    def test_empty_containers_are_written_closed(self):
        self.assertEqual(fmt('{"a": {}, "b": []}'), '{ "a": {}, "b": [] }\n')

    def test_what_does_not_fit_puts_one_member_per_line_with_a_two_space_indent(self):
        text = json.dumps({"key": "a" * 40, "label": "b" * 40, "kind": "number"})
        self.assertEqual(fmt(text), '{\n  "key": "' + "a" * 40 + '",\n  "label": "' + "b" * 40 + '",\n  "kind": "number"\n}\n')

    def test_the_width_counts_the_indent_the_key_and_the_comma(self):
        # {"rows": [row, row], "pad": long} breaks; each row then sits at indent 4 with a comma.
        row = '{ "key": "' + "x" * 20 + '", "value": 1 }'
        text = '{"rows": [' + row + ", " + row + '], "pad": "' + "y" * 200 + '"}'
        fits = len("    " + row + ",")
        self.assertIn("    " + row + ",", fmt(text, width=fits).splitlines())
        broken = fmt(text, width=fits - 1).splitlines()
        self.assertNotIn("    " + row + ",", broken)
        self.assertIn('      "key": "' + "x" * 20 + '",', broken)

    def test_the_rule_applies_again_inside_a_broken_container(self):
        text = json.dumps({"rows": [{"key": "a", "value": 1}, {"key": "b", "value": 2}], "description": "d" * 120})
        self.assertEqual(
            fmt(text),
            '{\n  "rows": [{ "key": "a", "value": 1 }, { "key": "b", "value": 2 }],\n  "description": "' + "d" * 120 + '"\n}\n',
        )

    def test_a_long_string_is_not_wrapped(self):
        text = json.dumps({"description": "word " * 40})
        out = fmt(text)
        self.assertEqual(json.loads(out), json.loads(text))
        self.assertEqual(len(out.splitlines()), 3)


class TestNothingIsRewritten(unittest.TestCase):
    def test_key_order_is_kept(self):
        self.assertEqual(list(json.loads(fmt('{"z": 1, "a": 2, "m": 3}'))), ["z", "a", "m"])

    def test_numbers_keep_their_spelling(self):
        self.assertEqual(fmt('{"a": 1.0, "b": 0.50, "c": 1e3, "d": -0}'), '{ "a": 1.0, "b": 0.50, "c": 1e3, "d": -0 }\n')

    def test_utf8_stays_utf8(self):
        self.assertEqual(fmt('{"label": "Fire — Wuxing 火"}'), '{ "label": "Fire — Wuxing 火" }\n')

    def test_escapes_a_string_needs_are_kept(self):
        self.assertEqual(json.loads(fmt(r'{"a": "quote \" and \\ and \n"}')), {"a": 'quote " and \\ and \n'})

    def test_the_file_ends_with_one_newline(self):
        self.assertTrue(fmt('{"a": 1}\n\n\n').endswith("}\n"))
        self.assertFalse(fmt('{"a": 1}\n\n\n').endswith("\n\n"))


class TestEveryFile(unittest.TestCase):
    """Run over every JSON file in the repo, so M5-52 can format them without changing a value."""

    # npm writes the lockfile and never reads our layout, so it is not ours to format (OPS_S00_06).
    NOT_OURS = {"frontend/package-lock.json"}

    @classmethod
    def setUpClass(cls):
        tracked = subprocess.run(["git", "-C", str(ROOT), "ls-files", "*.json"], capture_output=True, text=True, check=True)
        cls.files = [ROOT / name for name in tracked.stdout.split() if name not in cls.NOT_OURS and (ROOT / name).is_file()]
        # Format each file once and keep it, so the two tests share one pass (OPS_S00_06).
        cls.formatted = {path: fmt(path.read_text(encoding="utf-8")) for path in cls.files}

    def test_formatting_changes_no_value(self):
        for path in self.files:
            with self.subTest(path=str(path.relative_to(ROOT))):
                self.assertEqual(json.loads(self.formatted[path]), json.loads(path.read_text(encoding="utf-8")))

    def test_formatting_twice_changes_nothing(self):
        for path in self.files:
            with self.subTest(path=str(path.relative_to(ROOT))):
                once = self.formatted[path]
                self.assertEqual(fmt(once), once)


class TestCommandLine(unittest.TestCase):
    def run_tool(self, *args, cwd):
        return subprocess.run([sys.executable, str(SCRIPTS / "format_json.py"), *args], capture_output=True, text=True, cwd=cwd)

    def test_check_lists_an_unformatted_file_and_fails_without_writing_it(self):
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "a.json"
            path.write_text('{"a":1}')
            done = self.run_tool("--check", str(path), cwd=tmp)
            self.assertEqual(done.returncode, 1)
            self.assertIn("a.json", done.stdout)
            self.assertEqual(path.read_text(), '{"a":1}')

    def test_writing_formats_the_file_and_then_the_check_passes(self):
        with tempfile.TemporaryDirectory() as tmp:
            folder = Path(tmp) / "data"
            folder.mkdir()
            (folder / "a.json").write_text('{"a":1}')
            self.assertEqual(self.run_tool(str(folder), cwd=tmp).returncode, 0)
            self.assertEqual((folder / "a.json").read_text(), '{ "a": 1 }\n')
            self.assertEqual(self.run_tool("--check", str(folder), cwd=tmp).returncode, 0)

    def test_a_file_that_is_not_json_fails_with_its_name(self):
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "broken.json"
            path.write_text('{"a": }')
            done = self.run_tool("--check", str(path), cwd=tmp)
            self.assertEqual(done.returncode, 2)
            self.assertIn("broken.json", done.stderr)

    def test_without_paths_it_reads_the_paths_file_and_an_empty_list_passes(self):
        done = self.run_tool("--check", cwd=ROOT)
        self.assertEqual(done.returncode, 0, done.stdout + done.stderr)


if __name__ == "__main__":
    unittest.main()
