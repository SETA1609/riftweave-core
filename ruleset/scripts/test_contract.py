#!/usr/bin/env python3
"""The backend command contract (docs/backend/README.md) against the vectors and the reference.

Every vector's input must validate against its command's input schema, the reference's
result against the result schema (or the error shape), and every vector function needs a
contract file.
"""

import json
import sys
import unittest
from pathlib import Path

from jsonschema import Draft7Validator, RefResolver

_SCRIPTS = Path(__file__).resolve().parent
sys.path.insert(0, str(_SCRIPTS))

from kernel_reference import RUNNERS  # noqa: E402
from test_vectors import VECTOR_DIR, discover_vectors, vector_function  # noqa: E402

REPO_ROOT = _SCRIPTS.parent.parent
CONTRACT_DIR = REPO_ROOT / "backend" / "contract"
SCHEMA_DIR = REPO_ROOT / "ruleset" / "schemas"
NOT_COMMANDS = {"command.schema.json", "common.json"}
VECTOR_ONLY_FIELDS = {"name", "function", "expect"}


def load_store():
    """Every ruleset schema and contract file by file URI, for cross-file $ref."""
    store = {}
    for path in list(SCHEMA_DIR.glob("*.json")) + list(CONTRACT_DIR.glob("*.json")):
        store[path.resolve().as_uri()] = json.loads(path.read_text())
    return store


def load_commands(contract_dir=CONTRACT_DIR):
    """{name: (path, contract)} for every command file in the contract directory."""
    commands = {}
    for path in sorted(Path(contract_dir).glob("*.json")):
        if path.name not in NOT_COMMANDS:
            contract = json.loads(path.read_text())
            commands[contract["name"]] = (path, contract)
    return commands


def validator(schema, path, store):
    resolver = RefResolver(base_uri=path.resolve().as_uri(), referrer=schema, store=store)
    return Draft7Validator(schema, resolver=resolver)


def vector_input(vector):
    return {k: v for k, v in vector.items() if k not in VECTOR_ONLY_FIELDS}


class TestContract(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.store = load_store()
        cls.commands = load_commands()
        meta_path = CONTRACT_DIR / "command.schema.json"
        cls.meta = validator(json.loads(meta_path.read_text()), meta_path, cls.store)
        common_path = CONTRACT_DIR / "common.json"
        error_shape = json.loads(common_path.read_text())["definitions"]["error"]
        cls.error = validator(error_shape, common_path, cls.store)
        cls.live = {vector_function(p, v) for p, v in discover_vectors(VECTOR_DIR)}

    def test_every_contract_file_matches_the_meta_schema(self):
        self.assertTrue(self.commands, "backend/contract/ has no command files")
        for name, (path, contract) in self.commands.items():
            with self.subTest(command=name):
                self.meta.validate(contract)
                self.assertEqual(path.stem, name, "the file is named after its command")
                Draft7Validator.check_schema(contract["input_schema"])
                Draft7Validator.check_schema(contract["result_schema"])
                # A command whose rule lands in another pull request has no vectors or doc
                # yet; its doc is checked once its vectors exist.
                if name in self.live:
                    self.assertTrue((REPO_ROOT / contract["doc"]).is_file(), contract["doc"])

    def test_every_vector_function_has_a_contract(self):
        functions = {vector_function(p, v) for p, v in discover_vectors(VECTOR_DIR)}
        self.assertEqual(sorted(functions - set(self.commands)), [])

    def test_every_vector_input_matches_its_command(self):
        for path, vector in discover_vectors(VECTOR_DIR):
            name = vector_function(path, vector)
            if name not in self.commands:
                continue
            command_path, contract = self.commands[name]
            check = validator(contract["input_schema"], command_path, self.store)
            with self.subTest(file=path.name, vector=vector.get("name")):
                check.validate(vector_input(vector))

    def test_every_reference_result_matches_its_command(self):
        for path, vector in discover_vectors(VECTOR_DIR):
            name = vector_function(path, vector)
            if name not in self.commands or name not in RUNNERS:
                continue
            command_path, contract = self.commands[name]
            result = RUNNERS[name](vector)
            with self.subTest(file=path.name, vector=vector.get("name")):
                if "error" in result:
                    self.error.validate(result)
                else:
                    validator(contract["result_schema"], command_path, self.store).validate(result)


if __name__ == "__main__":
    unittest.main()
