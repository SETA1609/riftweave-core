// Package rest applies one rest occasion (docs/core/rest.md): each resource is refilled,
// partly refilled, or left alone, and timed conditions are cleared. The occasions and their
// fractions are rows in the data.
package rest

import (
	"fmt"
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

// Resources are the four numbers a rest changes.
type Resources struct {
	HP      int64 `json:"hp"`
	MP      int64 `json:"mp"`
	Stamina int64 `json:"stamina"`
	AP      int64 `json:"ap"`
}

// Input is the `rest` command's input.
type Input struct {
	Occasion  string              `json:"occasion"`
	Actor     actor.Actor         `json:"actor"`
	Modifiers []modifier.Modifier `json:"modifiers"`
	Current   Resources           `json:"current"`
	Max       struct {
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
func Run(rs *data.Ruleset, in Input) (Result, error) {
	occasion, ok := rs.Rest.Get(in.Occasion)
	if !ok {
		return Result{}, fmt.Errorf("rest occasion %q is not defined", in.Occasion)
	}
	stats, err := derived.Compute(rs, in.Actor, in.Modifiers)
	if err != nil {
		return Result{}, err
	}
	maximum := func(resource string) (int64, error) {
		if resource == "stamina" {
			return in.Max.Stamina, nil
		}
		i, ok := rs.Derived.Index(resource)
		if !ok {
			return 0, fmt.Errorf("rest resource %q is not defined", resource)
		}
		return stats[i], nil
	}
	res := Result{Resources: in.Current}
	conditions := in.Conditions
	for _, row := range occasion.Recovery {
		if row.Resource == "conditions" {
			if row.Mode == "clear" {
				conditions = slices.DeleteFunc(slices.Clone(conditions), func(c Condition) bool {
					return slices.Contains(row.Durations, c.Duration)
				})
			}
			continue
		}
		var slot *int64
		switch row.Resource {
		case "hp":
			slot = &res.HP
		case "mp":
			slot = &res.MP
		case "stamina":
			slot = &res.Stamina
		case "ap":
			slot = &res.AP
		default:
			return Result{}, fmt.Errorf("rest resource %q is not defined", row.Resource)
		}
		limit, err := maximum(row.Resource)
		if err != nil {
			return Result{}, err
		}
		switch row.Mode {
		case "full":
			*slot = limit
		case "fraction":
			*slot = min(limit, *slot+limit*row.Numerator/row.Denominator)
		}
	}
	res.Conditions = make([]string, len(conditions))
	for i, c := range conditions {
		res.Conditions[i] = c.Key
	}
	return res, nil
}
