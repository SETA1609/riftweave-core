# Security Policy

## Scope

`riftweave-core` is a **data-only** project: JSON game data validated against JSON
Schemas. The only executable code in the repository is:

- `ruleset/scripts/validate.py` — the schema validation script,
- the `Dockerfile` used to run it, and
- the GitHub Actions workflow under `.github/workflows/`.

There is no network service, no runtime engine, and no handling of untrusted user
input shipped from this repository. The realistic security surface is therefore the
**supply chain and tooling**: the validation script, its Python dependency
(`jsonschema`), the Docker image, and the CI workflow configuration.

## Supported versions

Only the `main` branch is supported. Fixes are applied to `main`; there are no
long-lived release branches at this time.

## Reporting a vulnerability

Please report security issues **privately** — do not open a public issue for
anything you believe may be sensitive.

1. Preferred: open a private report via GitHub's
   [private vulnerability reporting](https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing-information-about-vulnerabilities/privately-reporting-a-security-vulnerability)
   ("Report a vulnerability" under the repository's **Security** tab).
2. Alternatively, email the maintainers at `<security-contact@example.com>`
   *(replace with the project's real security contact)*.

Please include:

- a description of the issue and its impact,
- steps to reproduce (e.g. a malicious data/schema file, a CI configuration
  weakness, or a dependency concern), and
- any suggested remediation.

## What to expect

- **Acknowledgement** within 5 business days.
- An assessment and, where applicable, a fix on `main`.
- Credit in the fix commit or release notes if you would like it.

Because this project ships no runtime binary, most reports will be handled as normal
code/CI fixes rather than coordinated security releases. We will treat genuinely
sensitive reports (e.g. a CI secret exposure or a supply-chain compromise) with the
appropriate urgency and discretion.
