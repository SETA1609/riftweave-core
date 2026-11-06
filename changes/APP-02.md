### APP-02 Go core: the kernel in Go, every vector in CI
- Added: `backend/`, a Go module (standard library only) that loads `ruleset/data/` into flat tables, compiles formulas to flat programs, and implements all 12 commands; `go test ./...` runs every golden vector through it and checks the command list against `backend/contract/`.
- Changed: CI has a `backend` job (gofmt, vet, vectors in Go); a new kernel function needs a Go command as well as a runner, a checklist row, and a contract file.
