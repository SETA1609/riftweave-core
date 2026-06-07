#!/usr/bin/env python3
"""Validate all JSON data files against their $schema references."""

import json
import os
import sys
import warnings
from pathlib import Path

from jsonschema import validate, ValidationError, RefResolver

ROOT = Path(__file__).resolve().parent.parent
DATA_DIR = ROOT / "data"
SCHEMA_DIR = ROOT / "schemas"


def load_store():
    store = {}
    for f in sorted(SCHEMA_DIR.glob("*.schema.json")):
        store[f.as_uri()] = json.loads(f.read_text())
    return store


def collect_data_files():
    for dirpath, _, filenames in os.walk(DATA_DIR):
        for f in filenames:
            if f.endswith(".json"):
                yield Path(dirpath) / f


def main():
    store = load_store()
    data_files = list(collect_data_files())
    total = passed = failed = 0

    for fp in data_files:
        data = json.loads(fp.read_text())
        schema_ref = data.get("$schema")
        if not schema_ref:
            rel = fp.relative_to(ROOT)
            print(f"  \u26a0  {rel}: no $schema field, skipping")
            continue

        schema_path = (fp.parent / schema_ref).resolve()
        schema_uri = schema_path.as_uri()

        if schema_uri not in store:
            rel = fp.relative_to(ROOT)
            print(f"  \u2717  {rel}: schema '{schema_path.name}' not in store")
            failed += 1
            continue

        total += 1
        schema = store[schema_uri]

        resolver = RefResolver(
            base_uri=schema_uri,
            referrer=schema,
            store=store,
        )

        try:
            validate(data, schema, resolver=resolver)
            print(f"  \u2713  {fp.relative_to(ROOT)}")
            passed += 1
        except ValidationError as e:
            rel = fp.relative_to(ROOT)
            print(f"  \u2717  {rel}")
            path = (
                "/" + "/".join(str(p) for p in e.absolute_path)
                if e.absolute_path
                else "/"
            )
            print(f"       {path} {e.message}")
            failed += 1

    print(f"\n{total} file(s): {passed} passed, {failed} failed")
    return 1 if failed else 0


if __name__ == "__main__":
    with warnings.catch_warnings():
        warnings.filterwarnings("ignore", category=DeprecationWarning)
        sys.exit(main())
