"""Module loading, validation, and data merging for Riftweave modules.

A module is a directory under ruleset/modules/<module_id>/ containing:
  - manifest.json   (required: module metadata, schema v1)
  - data/            (optional: JSON data files extending core collections)
  - schemas/         (optional: additional JSON schemas)

Usage:
    from module_loader import discover_modules, load_module, merge_data

    modules = discover_modules()
    for mod in modules:
        load_module(mod)
    merged = merge_data(core_data, modules)
"""

import json
import os
from pathlib import Path
from jsonschema import validate, ValidationError

ROOT = Path(__file__).resolve().parent.parent
MODULES_DIR = ROOT / "modules"
SCHEMA_DIR = ROOT / "schemas"

MANIFEST_SCHEMA_PATH = SCHEMA_DIR / "module.manifest.schema.json"


# --- Data structures ---


class ModuleLoadError(Exception):
    """Raised when a module cannot be loaded or validated."""


class ConflictError(Exception):
    """Raised when two data sources contribute the same entry ID in the same collection."""


class ModuleInfo:
    """Holds loaded module metadata and data."""

    def __init__(self, manifest, module_dir):
        self.id = manifest["id"]
        self.version = manifest["version"]
        self.description = manifest.get("description", "")
        self.author = manifest.get("author", "")
        self.requires = manifest.get("requires", [])
        self.extends = manifest.get("extends", {})
        self.module_dir = Path(module_dir)
        self.manifest = manifest
        self.data = {}  # coll_name -> list of entries
        self.schemas = {}  # file_uri -> schema_dict

    def __repr__(self):
        return f"<ModuleInfo {self.id} v{self.version}>"


# --- Manifest loading ---


def load_manifest_schema():
    """Load and return the module manifest JSON Schema."""
    if not MANIFEST_SCHEMA_PATH.exists():
        raise ModuleLoadError(f"Manifest schema not found: {MANIFEST_SCHEMA_PATH}")
    return json.loads(MANIFEST_SCHEMA_PATH.read_text())


def validate_manifest(manifest, manifest_schema):
    """Validate a manifest dict against the manifest schema. Raises on failure."""
    try:
        validate(manifest, manifest_schema)
    except ValidationError as e:
        raise ModuleLoadError(f"Manifest validation failed: {e.message}")


# --- Module discovery ---


def discover_modules(modules_dir=None):
    """Scan modules_dir for subdirectories containing manifest.json.
    Returns a list of (module_dir, manifest_dict) tuples (not yet loaded).
    """
    if modules_dir is None:
        modules_dir = MODULES_DIR
    modules_dir = Path(modules_dir)

    if not modules_dir.exists():
        return []

    results = []
    for entry in sorted(modules_dir.iterdir()):
        if not entry.is_dir() or entry.name.startswith("."):
            continue
        manifest_path = entry / "manifest.json"
        if not manifest_path.exists():
            continue
        try:
            manifest = json.loads(manifest_path.read_text())
        except json.JSONDecodeError as e:
            raise ModuleLoadError(f"Invalid JSON in {manifest_path}: {e}")
        results.append((entry, manifest))
    return results


# --- Module loading ---


def load_module(module_dir, manifest, manifest_schema=None, schema_store=None):
    """Load a single module: validate manifest, load data and schema files.
    Returns a ModuleInfo instance.
    """
    module_dir = Path(module_dir)

    if manifest_schema is None:
        manifest_schema = load_manifest_schema()

    validate_manifest(manifest, manifest_schema)

    info = ModuleInfo(manifest, module_dir)

    # Load data files
    data_paths = manifest.get("data", [])
    for rel_path in data_paths:
        fp = (module_dir / rel_path).resolve()
        if not fp.exists():
            raise ModuleLoadError(
                f"Module '{info.id}': data file not found: {rel_path}"
            )
        try:
            raw = json.loads(fp.read_text())
        except json.JSONDecodeError as e:
            raise ModuleLoadError(
                f"Module '{info.id}': invalid JSON in {rel_path}: {e}"
            )
        for k, v in raw.items():
            if k.startswith("$") or not isinstance(v, list):
                continue
            info.data.setdefault(k, [])
            info.data[k].extend(v)

    # Load schema files
    schema_paths = manifest.get("schemas", [])
    for rel_path in schema_paths:
        fp = (module_dir / rel_path).resolve()
        if not fp.exists():
            raise ModuleLoadError(
                f"Module '{info.id}': schema file not found: {rel_path}"
            )
        try:
            schema = json.loads(fp.read_text())
        except json.JSONDecodeError as e:
            raise ModuleLoadError(
                f"Module '{info.id}': invalid JSON in {rel_path}: {e}"
            )
        uri = fp.as_uri()
        info.schemas[uri] = schema
        if schema_store is not None:
            schema_store[uri] = schema

    return info


def load_all_modules(modules_dir=None, schema_store=None):
    """Discover and load all modules. Returns a list of ModuleInfo, ordered by discovery."""
    manifest_schema = load_manifest_schema()
    discovered = discover_modules(modules_dir)
    modules = []
    for module_dir, manifest in discovered:
        mod = load_module(module_dir, manifest, manifest_schema, schema_store)
        modules.append(mod)
    return modules


# --- Data merging ---


def build_identity_map(modules):
    """Build a dict mapping coll_name -> set of entry IDs from all modules."""
    identity = {}
    for mod in modules:
        for coll_name, entries in mod.data.items():
            identity.setdefault(coll_name, set())
            for entry in entries:
                if isinstance(entry, dict) and "id" in entry:
                    identity[coll_name].add(entry["id"])
    return identity


def merge_data(core_data, modules):
    """Merge module data into core data in-place.
    core_data is a dict in the same shape as loaded core JSON files:
      { "features": [...], "equipment": [...], ... }

    Returns (merged_data, conflicts) where conflicts is a list of
    (coll_name, entry_id, source_a, source_b) tuples.

    Module entries are appended to the corresponding core collections.
    Conflicts are detected when the same entry ID appears in the same
    collection from more than one source.
    """
    from collections import defaultdict

    # Track which source owns each entry ID in each collection
    # sources[coll_name][entry_id] = module_id or "core"
    sources = defaultdict(dict)
    conflicts = []

    # Register core entries
    for coll_name, entries in core_data.items():
        if coll_name.startswith("$") or not isinstance(entries, list):
            continue
        for entry in entries:
            if isinstance(entry, dict) and "id" in entry:
                sources[coll_name][entry["id"]] = "core"

    # Merge module entries
    for mod in modules:
        for coll_name, entries in mod.data.items():
            if coll_name not in core_data:
                core_data[coll_name] = []
            for entry in entries:
                if not isinstance(entry, dict) or "id" not in entry:
                    core_data[coll_name].append(entry)
                    continue
                eid = entry["id"]
                existing = sources[coll_name].get(eid)
                if existing is not None and existing != mod.id:
                    conflicts.append((coll_name, eid, existing, mod.id))
                else:
                    sources[coll_name][eid] = mod.id
                    core_data[coll_name].append(entry)

    return core_data, conflicts
