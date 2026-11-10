// Package wuxing is the one interaction function (docs/modules/wuxing.md). Combat, crafting,
// and the vectors call Interact; nothing else reads the cycle edges. A directed pair is
// found in at most one cycle, so a pair the matrix does not list is `unrelated`.
package wuxing

import (
	"fmt"
	"slices"

	"github.com/SETA1609/riftweave-core/backend/internal/data"
)

// Input is the `wuxing` command's input. A nil phase is no phase.
type Input struct {
	ActorPhase         *string  `json:"actor_phase"`
	RecipientPhase     *string  `json:"recipient_phase"`
	ActorMagnitude     *float64 `json:"actor_magnitude"`
	RecipientMagnitude *float64 `json:"recipient_magnitude"`
}

// Result is how the actor's phase meets the recipient's.
type Result struct {
	Cycle       string  `json:"cycle"`
	Interaction string  `json:"interaction"`
	Multiplier  float64 `json:"multiplier"`
	Affects     string  `json:"affects"`
}

var unrelated Result = Result{Cycle: "unrelated", Interaction: "none", Multiplier: 1, Affects: "none"}

// Run is the `wuxing` command.
func Run(ruleset *data.Ruleset, input Input) (Result, error) {
	if !slices.Contains(ruleset.Modules, "wuxing") {
		return Result{}, &data.NotDefinedError{Kind: "module", Key: "wuxing"}
	}
	if input.ActorPhase == nil || input.RecipientPhase == nil || *input.ActorPhase == "" || *input.RecipientPhase == "" {
		return unrelated, nil
	}
	if *input.ActorPhase == *input.RecipientPhase {
		return Result{Cycle: "same_phase", Interaction: "identity", Multiplier: 1, Affects: "none"}, nil
	}
	var wuxing *data.Wuxing = &ruleset.Wuxing
	var from, to int = wuxing.Phase(*input.ActorPhase), wuxing.Phase(*input.RecipientPhase)
	if from < 0 || to < 0 {
		return unrelated, nil
	}
	for index := range wuxing.Cycles {
		var cycle *data.Cycle = &wuxing.Cycles[index]
		if !cycle.Has(from, to) {
			continue
		}
		var bothMagnitudesGiven bool = input.ActorMagnitude != nil && input.RecipientMagnitude != nil
		if cycle.Gate != nil && bothMagnitudesGiven && !(*input.ActorMagnitude > *input.RecipientMagnitude) {
			// The gate failed: the named cycle applies to the reversed pair, so the side it
			// modifies flips with it.
			fallback, ok := wuxing.Cycle(cycle.Gate.OtherwiseCycle)
			if !ok {
				return Result{}, fmt.Errorf("cycle %q names the unknown cycle %q", cycle.Identifier, cycle.Gate.OtherwiseCycle)
			}
			if !fallback.Has(to, from) {
				return unrelated, nil
			}
			var affects string = "recipient"
			if fallback.Affects == "recipient" {
				affects = "actor"
			}
			return Result{Cycle: fallback.Identifier, Interaction: fallback.Interaction, Multiplier: fallback.Multiplier, Affects: affects}, nil
		}
		return Result{Cycle: cycle.Identifier, Interaction: cycle.Interaction, Multiplier: cycle.Multiplier, Affects: cycle.Affects}, nil
	}
	return unrelated, nil
}
