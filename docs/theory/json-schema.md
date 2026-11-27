# JSON Schema, and why the ruleset is checked by it

Modified: 2026-11-27

Background for the rule that every data file names a schema and passes it. It covers what JSON
Schema is, why a ruleset that ships as data needs it, how a validator applies a schema and follows
a reference, how this repo wires that up, where draft-07 behaves in ways that surprise, and what a
schema cannot check. This page explains; the rules for schemas and data are in
[AGENTS.md](../../AGENTS.md), [the data model](../core/data-model.md), and
[the integrity checks](../core/integrity.md).

The specification, the `jsonschema` and `referencing` documentation, and the json-schema.org
pages were read from copies under `cache/docs/`, fetched with `devops-tools fetch-docs` at the
commits named in [Sources](#sources). Every claim about this repo's code was read in its files
on 2026-11-27, and the four behaviours in [Where draft-07 surprises](#where-draft-07-surprises)
were run against the installed `jsonschema` 4.26.0 that day. [V 2026-11-27]

## What JSON Schema is

The draft-07 core specification opens with its purpose:

> JSON Schema is a JSON media type for defining the structure of JSON data. JSON Schema is
> intended to define validation, documentation, hyperlink navigation, and interaction control of
> JSON data.

Three words carry the rest of this page.

- **Instance:** the JSON being checked. Here, a data file such as `ruleset/data/races/core.json`.
- **Schema:** a JSON document that describes instances. *"JSON Schemas are themselves JSON
  documents."* Here, a file such as `ruleset/schemas/race.schema.json`.
- **Keyword:** one property of a schema object, such as `type`, `required`, or `$ref`. A schema
  is a set of keywords, and each keyword has one defined meaning.

A schema does nothing on its own. The json-schema.org overview puts it plainly: *"validating a
JSON instance against a schema requires a JSON Schema validator."* In this repo the validator is
the Python `jsonschema` library, driven by `ruleset/scripts/validate.py`.

JSON Schema is published in drafts. Every schema here declares draft-07
(`http://json-schema.org/draft-07/schema#`), published in 2018 as `draft-handrews-json-schema-01`
(core) and `draft-handrews-json-schema-validation-01` (validation). The current version is
2020-12, and the json-schema.org overview recommends *"using the newest version"*. The installed
`jsonschema` still supports draft-07, through `Draft7Validator`. [V 2026-11-27]

## Why the ruleset uses it

Riftweave ships no engine. The deliverable is the data, and several programs read it: the Python
validator and reference functions, the Go core, the Angular app through the API and WebAssembly,
and any engine or virtual tabletop someone builds later. That makes the shape of the data a
contract between programs that do not share code. JSON Schema states that contract once, in a
form every one of those languages has a validator for.

What that buys, each with the reason it matters here:

- **The data is checked before anyone plays it.** CI runs `validate.py` on every change to
  `ruleset/`. A misspelled field name, a value outside its `enum`, or a missing `phase` on a
  playable race fails the pull request instead of failing at a table or in an engine.
- **New mechanics stay declarative.** The project's rule is *declarative data first*: a new
  mechanic is a schema extension plus data, not code. The schema is where the extension is
  written down, so the change can be reviewed as a change to the contract.
- **`additionalProperties: false` turns every new field into a decision.** A field cannot appear
  in the data until the schema names it, so nothing undocumented reaches the readers. The cost is
  that every new field is two edits, the schema and then the data; that is the point.
- **One vocabulary, defined once.** `ruleset/schemas/schema.json` holds the shared definitions
  (`ability`, `diceExpression`, `color`, `phase`, and the rest). A collection schema refers to
  them instead of copying them, so an attribute or a school changes in one place.
- **The schema documents the data.** Properties carry a `description`, so the schema is the
  field reference a person reads, not only a check a machine runs. The json-schema.org overview
  lists *"machine and human readable documentation"* among the reasons to use it.
- **Other artefacts are generated from it.** The backend's command contract in
  `backend/contract/` is also JSON Schema. The API server builds its OpenAPI document from those
  files at start-up, and the frontend's `openapi.d.ts` types are generated from that document, so
  the browser's types follow the contract instead of being written by hand.

Why draft-07 rather than 2020-12 is not recorded. Draft-07 has the keywords the ruleset needs
(`if`/`then`/`else`, `const`, `propertyNames`, `definitions`), and moving would change how `$ref`
and `definitions` read across every schema. [A]

## How a validator applies a schema

The validation specification describes the process as a walk from the root:

> Validation begins by applying the root schema to the complete instance document. From there,
> various keywords are used to determine which additional subschemas are applied to either the
> current location, or a child location.

Keywords fall into three kinds, and knowing which kind a keyword is explains most surprises.

- **Assertions** check the instance where they stand: `type`, `enum`, `const`, `minimum`,
  `maxLength`, `pattern`, `required`, `minItems`. *"Assertion keywords that are absent never
  restrict validation."* A schema with no `required` requires nothing.
- **Applicators** check nothing themselves; they choose subschemas and combine their results.
  `properties`, `items`, and `additionalProperties` apply a subschema to child values. `allOf`,
  `anyOf`, `oneOf`, and `not` apply subschemas to the same value and combine the results with
  boolean logic. `if`, `then`, and `else` apply one subschema and let its result pick another.
- **Annotations** describe the instance without deciding whether it is valid: `title`,
  `description`, `default`, `examples`. `format` is an annotation that an implementation may also
  treat as an assertion.

Keywords are independent of each other, with two named exceptions: *"additionalProperties, whose
behavior is defined in terms of properties and patternProperties; and additionalItems, whose
behavior is defined in terms of items."*

The conditional keywords are how this repo writes "this field is required when that one has this
value". `race.schema.json` uses them for lineage: if `role` is `subrace` or `kin`, then
`parent_race` is required. The specification makes them local on purpose: *"an if in one branch
of an allOf MUST NOT have an impact on a then or else in another branch"*, and *"when if is not
present, both then and else MUST be entirely ignored."*

## How references work

A schema may point at another schema, or at a part of one, with `$ref`. Three ideas from the core
specification govern it.

**Every schema has a URI, and `$ref` is resolved against it.**

> The "$id" keyword defines a URI for the schema, and the base URI that other URI references
> within the schema are resolved against.

With no `$id`, the base URI is *"the URI at which it was found, or a suitable substitute URI if
none is known."* A `$ref` of `schema.json#/definitions/ability` therefore means "the file
`schema.json` next to me, then the `ability` entry inside its `definitions`".

**The part after `#` is a JSON Pointer** (RFC 6901) or a plain name that some `$id` declared. A
pointer is a path of keys from the root of the document: `#/definitions/raceRef`. The
`definitions` keyword exists to hold reusable subschemas at such a path, and *"does not directly
affect the validation result"*.

**A URI is a name, not a download.**

> The URI is not a network locator, only an identifier. A schema need not be downloadable from
> the address if it is a network-addressable URL, and implementations SHOULD NOT assume they
> should perform a network operation when they encounter a network-addressable URI.

And: *"JSON Schema implementations SHOULD understand ahead of time which schemas they will be
using, and the URIs that identify them."* The `referencing` library, which `jsonschema` uses for
`$ref`, follows this. Its predecessor, `RefResolver`, *"did automatically retrieve remote
references (against the recommendation of the spec, and in a way which therefore could lead to
questionable security concerns when combined with untrusted schemas)"*; `referencing.Registry`
*"does not do so."* A registry is the set of schemas a validator knows, keyed by URI, plus an
optional `retrieve` function for a URI that is not in it.

## How this repo wires it

`validate.py` follows the pattern the `referencing` documentation recommends for *"a set of fixed
or static schemas in a few files"*: read them all into memory first.

1. **The store.** Every `ruleset/schemas/*.schema.json` is read into a dict keyed by its
   `file://` URI.
2. **The registry.** The store becomes a `referencing.Registry`, with each schema as a draft-07
   resource under its file URI. Its `retrieve` function loads a schema that is not in the store
   from disk, and only for a `file://` URI. That is how `schema.json` is reached: its name has no
   `.schema` part, so the glob does not load it, and the first `$ref` to it loads it. Nothing is
   ever fetched over the network.
3. **The root.** A collection schema has no `$id`, so `schema_validator` gives it its file URI as
   `$id` before building the validator. Its relative `$ref`s then resolve against its own
   directory, exactly as the specification's base-URI rule says.
4. **One validator per schema.** A `Draft7Validator` is built once per schema and reused for
   every data file and every test vector that names it, so each reference is resolved once.
5. **The data file's `$schema`.** Each data file names its schema by a relative path
   (`"$schema": "../../schemas/race.schema.json"`), which `validate.py` resolves from the file's
   directory and looks up in the store.

That last step is a convention of this repo, not of the specification. In the specification,
`$schema` belongs in a schema and names its meta-schema: *"The "$schema" keyword is both used as a
JSON Schema version identifier and the location of a resource which is itself a JSON Schema,
which describes any schema written for this particular version."* Its value *"MUST be a URI
(containing a scheme)"*, and *"It MUST NOT appear in subschemas."* A data file is an instance, not
a schema, so to a validator its `$schema` is an ordinary property; every collection schema
declares it as `{ "type": "string" }` so that `additionalProperties: false` lets it through. The
schemas themselves use `$schema` the specification's way, naming draft-07.

`schema.json` declares `$id: https://riftweave.dev/ruleset/schemas/schema.json`. By the rule
above, that is a name and not an address: nothing needs to be served there, and the validator
never tries.

The Go server reads the same files differently. It does not validate with them; it turns the
command contract in `backend/contract/` into an OpenAPI document, rewriting each `$ref` to point
inside that document. It accepts a `$ref` into the contract's own `definitions`, `common.json`,
or `schema.json`, and refuses any other, so the contract cannot depend on a schema the API would
not publish. `test_contract.py` validates every test vector's input and every reference result
against those contract schemas, with the same `schema_validator`.

## Where draft-07 surprises

Four behaviours that follow from the specification and are easy to get wrong. Each was checked
against the installed library. [V 2026-11-27]

**Keywords beside `$ref` are ignored.** The draft-07 core specification: *"An object schema with
a "$ref" property MUST be interpreted as a "$ref" reference. … All other properties in a "$ref"
object MUST be ignored."* So `{ "$ref": "#/definitions/x", "minimum": 1 }` checks only `x`; the
`minimum` does nothing and raises no error. A `description` beside a `$ref` is harmless, since it
only documents. To add a constraint to a referenced schema, wrap both in `allOf`:
`{ "allOf": [ { "$ref": "#/definitions/x" }, { "minimum": 1 } ] }`. (2020-12 changed this:
`$ref` there applies alongside its siblings. Moving drafts would change what such schemas mean.)

**`additionalProperties` sees only its own `properties`.** It applies *"only to the child values
of instance names that do not match any names in "properties", and do not match any regular
expression in "patternProperties""*, and those are the `properties` of the same schema object. A
property declared inside an `allOf` branch, or in a schema reached by `$ref`, does not count. An
object with `additionalProperties: false` beside an `allOf` that adds properties rejects those
properties. List every property in the object that closes it; use `allOf` to add constraints,
not fields.

**`format` is not checked.** The validation specification: *"Implementations MAY support the
"format" keyword as a validation assertion."* The `jsonschema` documentation: *"By default, as
per the specification, no validation is enforced."* It checks formats only when a
`FormatChecker` is passed to the validator, and `validate.py` passes none. A `"format": "date"`
here would document a date and check nothing; a constraint that must hold is a `pattern`, an
`enum`, or a check in `validate.py`.

**`default` fills in nothing.** `default` is an annotation: it *"can be used to supply a default
JSON value associated with a particular schema"*, and the specification places no duty on a
validator to write it into the instance. The `jsonschema` FAQ: *"the specification does not
require that default actually do anything"*, and *"none of the other keywords modify the instance
either."* The library leaves the instance as it is. A default in
a schema documents what a reader should assume when the field is absent. The rule that applies
it lives in the reader: the Go core, the Python reference, or the rules page that defines the
field.

## What a schema cannot check

A schema sees one instance at a time. That limits it in three ways that this repo fills in with
code.

- **References between files.** A schema can require that a field matches
  `^core:effect/[a-z0-9_]+$`; it cannot know whether `core:effect/damage_fire` exists in
  `effects/core.json`. `check_references()` in `validate.py` resolves every namespaced reference
  and fails on one that points nowhere ([integrity](../core/integrity.md)).
- **Rules that span rows.** That an effect used by an ingredient declares the `ingredient`
  channel is a fact about two files. `validate.py` enforces it through `CHANNEL_REQUIREMENTS`.
- **Data assembled from parts.** An override entry for a piece of equipment, a monster, or a
  spell is only complete after its base is merged in. `validate.py` merges the bases into the
  file before it validates it against its schema, so the schema sees whole rows, and it skips the
  `bases/` directories as standalone files.

So validation here has two layers: the schema checks the shape of each file, and `validate.py`
checks what the files say about each other. Both must pass. A data file with no `$schema` is
skipped with a warning, not failed, which is why every data file must declare one.

## How to write schemas here

Each rule with the reason for it.

- **Declare draft-07 in every schema, and validate with `Draft7Validator`.** A schema means what
  its draft says; mixing drafts changes what `$ref`, `definitions`, and `items` mean.
- **Close every object with `additionalProperties: false`.** An unknown field is then an error,
  not silent data no reader understands.
- **Put shared types in `schema.json` and `$ref` them.** A copied definition drifts; a referenced
  one cannot.
- **Put nothing but annotations beside a `$ref`.** A constraint there is ignored. Use `allOf`.
- **Declare every property in the object that closes it.** `allOf` adds constraints, not fields.
- **Write a required-when rule as `if`/`then`.** The condition stays next to the fields it
  governs, and it reads the way the rule is stated.
- **Use `pattern` or `enum` for a constraint that must hold, not `format`.** `format` is not
  checked here.
- **Do not count on `default`.** Say in the rules page what an absent field means, and apply it in
  the reader.
- **Keep every `$ref` local.** Refer only to files in `ruleset/schemas/` (or, for the command
  contract, to `backend/contract/` and `schema.json`). The registry knows only local files, and
  the server publishes only those.
- **Write a `description` on every property.** The schema is the field reference people read.

## Sources

- JSON Schema draft-07, `json-schema-org/json-schema-spec` at `1afc34b65ead` (the tag
  `draft-handrews-json-schema-01`):
  - `jsonschema-core.xml`, *JSON Schema: A Media Type for Describing JSON Documents*: the
    introduction and overview, fragment identifiers, the `$schema` keyword, base URI and `$id`,
    and `$ref` (loading and dereferencing).
  - `jsonschema-validation.xml`, *JSON Schema Validation: A Vocabulary for Structural Validation
    of JSON*: applicability and keyword independence, assertions, `additionalProperties`, the
    conditional keywords, `format`, `definitions`, and `default`.
  - `schema.json`, the draft-07 meta-schema.
- `python-jsonschema/jsonschema` at `a7277432b0f7` (tag `v4.26.0`, the installed version):
  `docs/validate.rst` ("Validating Formats"), `docs/referencing.rst` (the registry, filesystem
  retrieval, HTTP retrieval, and the differences from `RefResolver`), and `docs/faq.rst` (format
  and `default`).
- `python-jsonschema/referencing` at `944ed5a20bc5` (tag `v0.37.0`): `docs/intro.rst`.
- `json-schema-org/website` at `1a415f9d36c3`: `pages/overview/what-is-jsonschema.md`,
  `pages/understanding-json-schema/structuring.md`, and
  `pages/understanding-json-schema/reference/` (`object.md`, `schema.md`, `conditionals.md`).
  Published at https://json-schema.org/overview/what-is-jsonschema and
  https://json-schema.org/understanding-json-schema.
- This repo: `ruleset/scripts/validate.py` (`load_store`, `_retrieve`, `schema_registry`,
  `schema_validator`, `check_references`), `ruleset/schemas/schema.json`,
  `ruleset/schemas/race.schema.json`, `ruleset/scripts/test_contract.py`, and
  `backend/cmd/riftweave-server/openapi.go`.
