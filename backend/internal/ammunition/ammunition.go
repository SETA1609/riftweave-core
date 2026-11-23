// Package ammunition is the `ammunition` command (docs/core/combat.md#ammunition, GD_S17_02):
// the rounds recovered after a fight, each kind's share in combat/ammunition.json, rounded down.
package ammunition

import (
	"math"

	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
)

// Operation is what the command does.
type Operation string

// OperationRecover gives back the share of the rounds fired that a fight leaves to pick up.
const OperationRecover Operation = "recover"

// Input is one ammunition call: the rounds fired by kind, and the thrown weapons ("thrown").
type Input struct {
	Operation Operation           `json:"operation"`
	Fired     map[string]int64    `json:"fired"`
	Modifiers []modifier.Modifier `json:"modifiers"`
}

// Output is the rounds recovered by kind.
type Output struct {
	Recovered map[string]int64 `json:"recovered"`
}

// Run runs one ammunition call.
func Run(ruleset *data.Ruleset, input Input) (Output, error) {
	switch input.Operation {
	case OperationRecover:
		var recovered map[string]int64 = make(map[string]int64, len(input.Fired))
		for kind, count := range input.Fired {
			share, known := ruleset.AmmunitionRecovered[kind]
			if !known {
				return Output{}, &data.NotDefinedError{Kind: "ammunition kind", Key: kind}
			}
			recovered[kind] = int64(math.Floor(float64(count)*share + 1e-9))
		}
		return Output{Recovered: recovered}, nil
	}
	return Output{}, &data.NotDefinedError{Kind: "ammunition operation", Key: string(input.Operation)}
}
