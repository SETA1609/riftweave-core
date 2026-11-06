// Package wuxing is the one interaction function (docs/modules/wuxing.md). Combat, crafting,
// and the vectors call Interact; nothing else reads the cycle edges. A directed pair is
// found in at most one cycle, so a pair the matrix does not list is `unrelated`.
package wuxing

import (
	"fmt"

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

var unrelated = Result{"unrelated", "none", 1, "none"}

// Run is the `wuxing` command.
func Run(rs *data.Ruleset, in Input) (Result, error) {
	if in.ActorPhase == nil || in.RecipientPhase == nil || *in.ActorPhase == "" || *in.RecipientPhase == "" {
		return unrelated, nil
	}
	if *in.ActorPhase == *in.RecipientPhase {
		return Result{"same_phase", "identity", 1, "none"}, nil
	}
	w := &rs.Wuxing
	from, to := w.Phase(*in.ActorPhase), w.Phase(*in.RecipientPhase)
	if from < 0 || to < 0 {
		return unrelated, nil
	}
	for i := range w.Cycles {
		c := &w.Cycles[i]
		if !c.Has(from, to) {
			continue
		}
		if c.Gate != nil && in.ActorMagnitude != nil && in.RecipientMagnitude != nil && !(*in.ActorMagnitude > *in.RecipientMagnitude) {
			// The gate failed: the named cycle applies to the reversed pair, so the side it
			// modifies flips with it.
			fallback, ok := w.Cycle(c.Gate.OtherwiseCycle)
			if !ok {
				return Result{}, fmt.Errorf("cycle %q names the unknown cycle %q", c.ID, c.Gate.OtherwiseCycle)
			}
			if !fallback.Has(to, from) {
				return unrelated, nil
			}
			affects := "recipient"
			if fallback.Affects == "recipient" {
				affects = "actor"
			}
			return Result{fallback.ID, fallback.Interaction, fallback.Multiplier, affects}, nil
		}
		return Result{c.ID, c.Interaction, c.Multiplier, c.Affects}, nil
	}
	return unrelated, nil
}
