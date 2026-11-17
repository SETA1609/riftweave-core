"""release_check.py (M5-62, CONTRIBUTING.md#releasing): a tag is released only when it is the
version, the changelog has its section, and every removed key is marked breaking."""

import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import release_check  # noqa: E402

CHANGELOG = """# Changelog

## v0.2.0

### M5-90 Rename a perk
- Removed (breaking): `old_perk` is now `new_perk`.
- Added: `spare_perk`.

## v0.1.0

### M5-62 One project version
- Added: `VERSION`.
"""


class TestReleaseCheck(unittest.TestCase):
    def test_a_matching_tag_with_its_section_and_marked_removals_passes(self):
        self.assertEqual(release_check.check("v0.2.0", "0.2.0", CHANGELOG, [("features", "old_perk")]), [])

    def test_a_tag_that_is_not_the_version_fails(self):
        problems = release_check.check("v0.3.0", "0.2.0", CHANGELOG, [])
        self.assertTrue(any("not v0.2.0" in problem for problem in problems), problems)

    def test_a_tag_without_its_changelog_section_fails(self):
        problems = release_check.check("v0.3.0", "0.3.0", CHANGELOG, [])
        self.assertEqual(problems, ["CHANGELOG.md has no '## v0.3.0' section"])

    def test_a_removed_key_no_breaking_bullet_names_fails(self):
        problems = release_check.check("v0.2.0", "0.2.0", CHANGELOG, [("spells", "fire_bolt")])
        self.assertEqual(problems, ["spells/fire_bolt was removed, but no (breaking) bullet of '## v0.2.0' names it"])

    def test_a_removed_key_named_only_in_a_bullet_that_is_not_breaking_fails(self):
        self.assertEqual(len(release_check.check("v0.2.0", "0.2.0", CHANGELOG, [("features", "spare_perk")])), 1)

    def test_the_section_is_the_release_notes(self):
        section = release_check.changelog_section(CHANGELOG, "v0.1.0")
        self.assertEqual(section.strip(), "### M5-62 One project version\n- Added: `VERSION`.")

    def test_the_repo_version_is_a_semantic_version(self):
        version = (release_check.REPO_ROOT / "VERSION").read_text()
        self.assertRegex(version, r"^\d+\.\d+\.\d+\n$")


if __name__ == "__main__":
    unittest.main()
