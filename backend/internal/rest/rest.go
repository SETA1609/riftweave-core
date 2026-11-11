// Package rest applies one rest occasion (docs/core/rest.md): each resource is refilled,
// partly refilled, or left alone, and timed conditions are cleared. The occasions and their
// fractions are rows in the data.
package rest

import (
	"slices"

	"github.com/SETA1609/riftweave-core/backend/internal/actor"
	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/derived"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
)

// Condition is an active condition and how long it lasts.
type Condition struct {
	Key      string `json:"key"`
	Duration string `json:"duration"`
}

// Resources are the numbers a rest changes. Toxicity is nil when the input names none.
type Resources struct {
	Toxicity     *int64 `json:"toxicity,omitempty"`
	HitPoints    int64  `json:"hp"`
	ManaPoints   int64  `json:"mp"`
	Stamina      int64  `json:"stamina"`
	ActionPoints int64  `json:"ap"`
}

// Input is the `rest` command's input.
type Input struct {
	Occasion  string              `json:"occasion"`
	Actor     actor.Actor         `json:"actor"`
	Modifiers []modifier.Modifier `json:"modifiers"`
	Current   Resources           `json:"current"`
	Maximum   struct {
		Stamina int64 `json:"stamina"`
	} `json:"max"`
	Conditions []Condition `json:"conditions"`
}

// Result is the resources after the rest and the conditions left.
type Result struct {
	Resources
	Conditions []string `json:"conditions"`
}

// Run is the `rest` command. The maximums are the derived stats, and stamina's comes in.
func Run(ruleset *data.Ruleset, input Input) (Result, error) {
	occasion, err := ruleset.Rest.Require("rest occasion", input.Occasion)
	if err != nil {
		return Result{}, err
	}
	stats, err := derived.Compute(ruleset, input.Actor, input.Modifiers)
	if err != nil {
		return Result{}, err
	}
	var maximumOf func(resource string) (int64, error) = func(resource string) (int64, error) {
		if resource == "stamina" {
			return input.Maximum.Stamina, nil
		}
		statIndex, ok := ruleset.Derived.Index(resource)
		if !ok {
			return 0, &data.NotDefinedError{Kind: "rest resource", Key: resource}
		}
		return stats[statIndex], nil
	}
	var result Result = Result{Resources: input.Current}
	var conditions []Condition = input.Conditions
	for _, recovery := range occasion.Recovery {
		if recovery.Resource == "conditions" {
			if recovery.Mode == "clear" {
				conditions = slices.DeleteFunc(slices.Clone(conditions), func(condition Condition) bool {
					return slices.Contains(recovery.Durations, condition.Duration)
				})
			}
			continue
		}
		if recovery.Resource == "toxicity" {
			// empty sets stored toxicity to 0 (docs/core/modifier.md#toxicity).
			if recovery.Mode == "empty" && result.Toxicity != nil {
				var cleared int64
				result.Toxicity = &cleared
			}
			continue
		}
		var resourceValue *int64
		switch recovery.Resource {
		case "hp":
			resourceValue = &result.HitPoints
		case "mp":
			resourceValue = &result.ManaPoints
		case "stamina":
			resourceValue = &result.Stamina
		case "ap":
			resourceValue = &result.ActionPoints
		default:
			return Result{}, &data.NotDefinedError{Kind: "rest resource", Key: recovery.Resource}
		}
		limit, err := maximumOf(recovery.Resource)
		if err != nil {
			return Result{}, err
		}
		switch recovery.Mode {
		case "full":
			*resourceValue = limit
		case "fraction":
			*resourceValue = min(limit, *resourceValue+limit*recovery.Numerator/recovery.Denominator)
		}
	}
	result.Conditions = make([]string, len(conditions))
	for index, condition := range conditions {
		result.Conditions[index] = condition.Key
	}
	return result, nil
}
