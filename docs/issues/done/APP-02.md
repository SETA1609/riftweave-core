#### APP-02 Go core: the kernel in Go, every vector in CI
- Status: done
- Merged: #83, 2026-11-06
- Priority: medium
- Needs: APP-01
- Do: A Go module in `backend/` that loads `ruleset/data/` (and module data), evaluates formula trees, and implements every command in the contract. It reads coefficients from the data, never from constants. A Go test runs every file in `ruleset/tests/vectors/` and fails on any disagreement, and CI runs it on every PR. It starts with the commands that exist and grows with the rules: from then on a rule issue that adds or changes a function updates the Python reference, the Go core, and the contract in the same pull request; CONTRIBUTING says so.
- Keys: Go module `backend/`; package per kernel area (formula, resolution, combat, ...)
- Files: ~15 (backend/, .github/workflows/, CONTRIBUTING.md, docs/backend/README.md)
- See: docs/core/vectors.md, backend/contract/
- Done when: `go test ./...` passes every vector the Python reference passes; CI runs it; a data change that breaks a rule fails both.
