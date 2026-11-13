// Package hazard detects, disarms, and triggers one trap or environmental hazard
// (docs/core/hazards.md). Detect, disarm, and avoid are checks: one resolution through a
// check row, with the hazard's difficulty as an add on the target; an arcane hazard rolls
// its color's magic skill. A trigger returns one packet and the effect list; the caller
// applies DR, resist, and conditions. A spell op applies a cast that succeeded. Rolls are
// inputs.
package hazard

import (
	"fmt"
	"slices"

	"github.com/SETA1609/riftweave-core/backend/internal/actor"
	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/formula"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
	"github.com/SETA1609/riftweave-core/backend/internal/resolution"
)

// Input is the `hazard` command's input; `op` says which fields it reads.
type Input struct {
	Roll       *int64              `json:"roll"`
	Hazard     string              `json:"hazard"`
	Spell      string              `json:"spell"`
	Operation  string              `json:"operation"`
	Modifiers  []modifier.Modifier `json:"modifiers"`
	Actor      actor.Actor         `json:"actor"`
	DamageRoll int64               `json:"damage_roll"`
}

// CheckResult is a detect or a disarm: the check, and for a disarm whether it set the
// trap off (a fumble does).
type CheckResult struct {
	Triggered *bool `json:"triggered,omitempty"`
	resolution.Result
}

// Packet is the damage a trigger deals, before the victim's DR and resist.
type Packet struct {
	Type   string `json:"type"`
	Amount int64  `json:"amount"`
}

// TriggerResult is what a hazard does to whoever sets it off.
type TriggerResult struct {
	Packet  *Packet  `json:"packet"`
	Effects []string `json:"effects"`
	Avoided bool     `json:"avoided"`
}

// SpellResult is a cast spell on a hazard: whether it undid it, and the effect that did.
type SpellResult struct {
	Effect   *string `json:"effect"`
	Disarmed bool    `json:"disarmed"`
}

// undoneBy is the effect that undoes a hazard of each nature (docs/core/hazards.md). A
// natural hazard has none.
var undoneBy = map[string]string{"mechanical": "core:effect/disarm_hazard", "arcane": "core:effect/dispel"}

// spell applies a cast that succeeded: an effect that undoes the hazard's nature, at a
// magnitude of the hazard's level or more, disarms it.
func spell(ruleset *data.Ruleset, input *Input, hazard *data.Hazard) (SpellResult, error) {
	cast, err := ruleset.Spells.Require("spell", input.Spell)
	if err != nil {
		return SpellResult{}, err
	}
	wanted, ok := undoneBy[hazard.Nature]
	if !ok {
		return SpellResult{}, nil
	}
	for _, applied := range cast.Effects {
		if applied.Effect == wanted && applied.Magnitude >= float64(hazard.Level) {
			return SpellResult{Disarmed: true, Effect: &wanted}, nil
		}
	}
	return SpellResult{}, nil
}

// check rolls one of the hazard's checks. Only the modifiers on the check's skill count,
// and the difficulty is one more add on it.
func check(ruleset *data.Ruleset, input *Input, hazard *data.Hazard, spec *data.HazardCheck) (resolution.Result, error) {
	if input.Roll == nil {
		return resolution.Result{}, &data.MissingInputError{Subject: "hazard " + input.Operation, Needs: "roll"}
	}
	var modifiers []modifier.Modifier = modifier.Target(input.Modifiers, spec.Skill)
	if spec.Difficulty != 0 {
		modifiers = append(modifiers, modifier.Modifier{
			Target:    spec.Skill,
			Operation: modifier.Add,
			Value:     float64(spec.Difficulty),
			Source:    "hazard:" + hazard.Key,
		})
	}
	return resolution.Run(ruleset, resolution.Input{
		Actor:     input.Actor,
		Skill:     spec.Skill,
		Modifiers: modifiers,
		Roll:      *input.Roll,
	})
}

// Run is the `hazard` command. Its result type depends on `op`.
func Run(ruleset *data.Ruleset, input Input) (any, error) {
	hazard, err := ruleset.Hazards.Require("hazard", input.Hazard)
	if err != nil {
		return nil, err
	}
	switch input.Operation {
	case "detect":
		result, err := check(ruleset, &input, hazard, &hazard.Detect)
		return CheckResult{Result: result}, err
	case "disarm":
		if hazard.Disarm == nil {
			return nil, fmt.Errorf("%s cannot be disarmed", hazard.Key)
		}
		result, err := check(ruleset, &input, hazard, hazard.Disarm)
		var triggered bool = slices.Contains(result.Tags, "fumble")
		return CheckResult{Result: result, Triggered: &triggered}, err
	case "trigger":
		return trigger(ruleset, &input, hazard)
	case "spell":
		return spell(ruleset, &input, hazard)
	}
	return nil, &data.NotDefinedError{Kind: "hazard operation", Key: input.Operation}
}

// trigger is a hazard going off. Without an avoid check, or without a roll for it, the
// victim had no chance. Avoiding a `half` hazard halves the damage, rounded down; avoiding
// any hazard avoids its effects.
func trigger(ruleset *data.Ruleset, input *Input, hazard *data.Hazard) (TriggerResult, error) {
	var avoided, halved bool
	if hazard.Avoid != nil && input.Roll != nil {
		avoid, err := check(ruleset, input, hazard, hazard.Avoid)
		if err != nil {
			return TriggerResult{}, err
		}
		avoided = avoid.Success
		halved = avoided && hazard.Avoid.OnAvoid == "half"
	}
	var result TriggerResult = TriggerResult{Avoided: avoided, Effects: []string{}}
	if hazard.Damage != nil && (!avoided || halved) {
		var amount int64 = input.DamageRoll
		if halved {
			amount = formula.FloorDivide(amount, 2)
		}
		result.Packet = &Packet{Amount: amount, Type: hazard.Damage.Type}
	}
	if !avoided {
		result.Effects = append(result.Effects, hazard.Effects...)
	}
	return result, nil
}
