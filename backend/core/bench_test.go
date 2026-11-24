package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// The hot paths' benchmarks (QA_S24_02, story S24): each replays every vector of its command
// through Run, so a refactor PR shows its change with `go test -bench . ./core/ ./internal/data/`
// and tools/benchdiff. A vector that names another function is left out.
var benchmarked = []string{"combat", "attack_sequence", "casting", "actions", "derived_examples"}

func BenchmarkCommand(b *testing.B) {
	opened, err := Open(filepath.Join(repo, "ruleset", "data"))
	if err != nil {
		b.Fatal(err)
	}
	for _, stem := range benchmarked {
		var command string = stem
		if stem == "derived_examples" {
			command = "derived"
		}
		content, err := os.ReadFile(filepath.Join(repo, "ruleset", "tests", "vectors", stem+".json"))
		if err != nil {
			b.Fatal(err)
		}
		var vectors []map[string]any
		if err := json.Unmarshal(content, &vectors); err != nil {
			b.Fatal(err)
		}
		var inputs [][]byte
		for _, vector := range vectors {
			if function, named := vector["function"].(string); named && function != command {
				continue
			}
			for _, key := range []string{"name", "function", "expect"} {
				delete(vector, key)
			}
			body, err := json.Marshal(vector)
			if err != nil {
				b.Fatal(err)
			}
			inputs = append(inputs, body)
		}
		b.Run(command, func(b *testing.B) {
			for b.Loop() {
				for _, input := range inputs {
					_, _ = opened.Run(command, input)
				}
			}
		})
	}
}
