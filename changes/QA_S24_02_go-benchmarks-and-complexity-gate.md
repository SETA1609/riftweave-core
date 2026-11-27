### QA_S24_02_go-benchmarks-and-complexity-gate Benchmarks and a complexity gate before the Go refactor
- Added: the Go complexity gate (`backend/tools/complexity`): a function over a complexity of 40 or a nesting depth of 4 fails `go test`, except the 10 in its allow list, which may only shrink.
- Added: benchmarks of the hot commands and of opening the ruleset, `tools/benchdiff`, and the on-request `Go benchmarks` workflow that compares a branch with a base.
