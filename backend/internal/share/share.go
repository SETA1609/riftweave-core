// Package share works out what one ally gets from the bards and skalds near it
// (docs/core/share.md, M3-14). Each sharer within its perk's range lends its attribute
// modifier, scaled by the rank's share, to the perk's group of attributes as a morale
// modifier that lasts the perk's turns. Morale stacks highest, so the strongest counts per
// attribute. The kernel has no party: the caller says who is near and how far.
package share

import (
	"fmt"
	"math"

	"github.com/SETA1609/riftweave-core/backend/internal/actor"
	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/formula"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
)

// Sharer is one bard or skald near the ally.
type Sharer struct {
	Name       string           `json:"name"`
	Attributes actor.Attributes `json:"attributes"`
	Perk       string           `json:"perk"`
	Rank       int64            `json:"rank"`
	Distance   float64          `json:"distance"` // units from the ally
}

// Input is the `share` command's input.
type Input struct {
	Sharers      []Sharer `json:"sharers"`
	ElapsedTurns int64    `json:"elapsed_turns"`
}

// Result is what reaches the ally: the live modifiers and, per attribute, the bonus that counts.
type Result struct {
	Modifiers []modifier.Modifier `json:"modifiers"`
	Bonus     map[string]int64    `json:"bonus"`
}

// Run lends each sharer's bonus to the ally.
func Run(ruleset *data.Ruleset, input Input) (Result, error) {
	var emitted []modifier.Modifier = []modifier.Modifier{}
	for _, sharer := range input.Sharers {
		perk, err := ruleset.Features.Require("feature", sharer.Perk)
		if err != nil {
			return Result{}, err
		}
		if perk.Share == nil {
			return Result{}, &data.MissingInputError{Subject: "share", Needs: "a perk with a share object"}
		}
		group, err := ruleset.SummonBonuses.Require("summon bonus", perk.Share.Bonus)
		if err != nil {
			return Result{}, err
		}
		var rankShare []float64 = group.RankShare
		if perk.Share.RankShare != nil {
			rankShare = perk.Share.RankShare
		}
		var rank int64 = sharer.Rank
		if rank == 0 {
			rank = 1
		}
		if rank < 1 || rank > int64(len(rankShare)) || rank > max(perk.Ranks, 1) {
			return Result{}, &data.NotDefinedError{Kind: "share rank", Key: fmt.Sprint(rank)}
		}
		if sharer.Distance > perk.Share.Range {
			continue
		}
		ability, known := actor.ByKey(group.Attribute)
		if !known {
			return Result{}, &data.NotDefinedError{Kind: "ability", Key: group.Attribute}
		}
		attributeModifier, err := ruleset.AttributeModifier.Evaluate(nil, formula.Variable{Name: "attribute", Value: sharer.Attributes.Get(ability)})
		if err != nil {
			return Result{}, err
		}
		var value float64 = math.Floor(float64(max(attributeModifier, 0)) * rankShare[rank-1])
		if value <= 0 {
			continue
		}
		for _, target := range group.Attributes {
			emitted = append(emitted, modifier.Modifier{
				Target:    target,
				Operation: modifier.Add,
				Value:     value,
				Source:    "share:" + perk.Key + ":" + sharer.Name,
				Origin:    "core:origin/morale",
				Duration:  "turns",
				Turns:     perk.Share.Turns,
			})
		}
	}
	var live []modifier.Modifier = modifier.Live(emitted, input.ElapsedTurns, nil)
	var bonus map[string]int64 = map[string]int64{}
	for _, held := range live {
		if _, done := bonus[held.Target]; done {
			continue
		}
		value, err := modifier.Value(0, modifier.Target(live, held.Target), ruleset.Origins, 0, nil)
		if err != nil {
			return Result{}, err
		}
		bonus[held.Target] = value
	}
	return Result{Modifiers: live, Bonus: bonus}, nil
}
