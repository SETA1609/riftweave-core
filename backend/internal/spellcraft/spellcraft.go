// Package spellcraft checks spell knowledge (docs/core/casting.md#spellcraft): a crafted
// spell against the effects the caster knows, the craftable effects a learned spell teaches,
// and one Research attempt. The roll is an input.
package spellcraft

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"

	"github.com/SETA1609/riftweave-core/backend/internal/actor"
	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
	"github.com/SETA1609/riftweave-core/backend/internal/resolution"
)

// Caster is what the caster knows and can roll.
type Caster struct {
	Skills     map[string]int64 `json:"skills"`
	Known      []string         `json:"known"` // effect refs
	Attributes actor.Attributes `json:"attributes"`
}

// Recipe is a crafted spell: one color and its applied effects.
type Recipe struct {
	Color   string               `json:"color"`
	Effects []data.AppliedEffect `json:"effects"`
}

// Input is the `spellcraft` command's input. Spell is a key for learn and a recipe for craft.
type Input struct {
	Operation string              `json:"operation"`
	Spell     json.RawMessage     `json:"spell"`
	Effect    string              `json:"effect"`
	Color     string              `json:"color"`
	Modifiers []modifier.Modifier `json:"modifiers"`
	Caster    Caster              `json:"caster"`
	Roll      int64               `json:"roll"`
}

// Rejection is one effect a recipe may not carry, and why.
type Rejection struct {
	Effect string `json:"effect"`
	Reason string `json:"reason"`
}

// CraftResult is a checked recipe. The price and thresholds are nil when an effect failed.
type CraftResult struct {
	Mana           *int64      `json:"mana"`
	CastThreshold  *int64      `json:"cast_threshold"`
	CraftThreshold *int64      `json:"craft_threshold"`
	Rejected       []Rejection `json:"rejected"`
	Valid          bool        `json:"valid"`
	SkillOK        bool        `json:"skill_ok"`
}

// LearnResult is what a learned spell taught.
type LearnResult struct {
	Taught []string `json:"taught"`
	Known  []string `json:"known"`
}

// ResearchResult is one Research attempt.
type ResearchResult struct {
	Check   *resolution.Result `json:"check"`
	Reason  *string            `json:"reason"`
	Cost    *data.Cost         `json:"cost"`
	Outcome string             `json:"outcome"`
	Known   []string           `json:"known"`
	Days    int64              `json:"days"`
}

// Run is the `spellcraft` command.
func Run(ruleset *data.Ruleset, input Input) (any, error) {
	switch input.Operation {
	case "craft":
		var recipe Recipe
		if err := json.Unmarshal(input.Spell, &recipe); err != nil {
			return nil, fmt.Errorf("craft needs a recipe: %w", err)
		}
		return craft(ruleset, &input.Caster, &recipe)
	case "learn":
		var key string
		if err := json.Unmarshal(input.Spell, &key); err != nil {
			return nil, fmt.Errorf("learn needs a spell key: %w", err)
		}
		return learn(ruleset, &input.Caster, key)
	case "research":
		return research(ruleset, &input)
	}
	return nil, &data.NotDefinedError{Kind: "spellcraft operation", Key: input.Operation}
}

// roundHalfUp rounds to the nearest integer, a half away from zero.
func roundHalfUp(value float64) int64 {
	return int64(math.Floor(math.Round(value*1e6)/1e6 + 0.5))
}

// isCraftable reports whether a recipe, Research, or a learned spell may hold the effect.
func isCraftable(effect *data.Effect) bool {
	return slices.Contains(effect.Channels, "spell") && (effect.Craftable == nil || *effect.Craftable)
}

// rejection is why one applied effect cannot go into a recipe of color, or "".
func rejection(effect *data.Effect, applied *data.AppliedEffect, color string, known []string) string {
	switch {
	case !slices.Contains(effect.Channels, "spell"):
		return "channel"
	case effect.Craftable != nil && !*effect.Craftable:
		return "uncraftable"
	case applied.Parameter != "" && slices.Contains(effect.UncraftableParameters, applied.Parameter):
		return "excluded_parameter"
	case !slices.Contains(known, applied.Effect):
		return "unknown"
	case effect.Color != "" && effect.Color != color:
		return "off_color"
	}
	return ""
}

// spellCost is a spell's mana from its effects alone (docs/modules/magic.md#cost-formula).
func spellCost(ruleset *data.Ruleset, effects []data.AppliedEffect) (int64, error) {
	var rule *data.SpellCost = &ruleset.Casting.SpellCost
	var total float64
	for index := range effects {
		var applied *data.AppliedEffect = &effects[index]
		effect, err := ruleset.Effects.Require("effect", applied.Effect)
		if err != nil {
			return 0, err
		}
		multiplier, ok := rule.TargetMultiplier[applied.Target]
		if !ok {
			return 0, &data.NotDefinedError{Kind: "target", Key: applied.Target}
		}
		var magnitude float64 = applied.Magnitude
		if effect.MagnitudeUnit == "" || effect.MagnitudeUnit == "seconds" || magnitude == 0 {
			magnitude = 1
		}
		var duration float64 = max(applied.Duration, 1)
		var cost float64 = rule.MagnitudeFactor * effect.BaseCost * magnitude * duration
		if applied.Target == "area" {
			cost += rule.AreaFactor * effect.BaseCost * applied.Radius
		}
		total += cost * multiplier
	}
	return int64(math.Ceil(math.Round(total*1e6) / 1e6)), nil
}

func craft(ruleset *data.Ruleset, caster *Caster, recipe *Recipe) (CraftResult, error) {
	var result CraftResult = CraftResult{Rejected: []Rejection{}}
	for index := range recipe.Effects {
		var applied *data.AppliedEffect = &recipe.Effects[index]
		effect, err := ruleset.Effects.Require("effect", applied.Effect)
		if err != nil {
			return CraftResult{}, err
		}
		if reason := rejection(effect, applied, recipe.Color, caster.Known); reason != "" {
			result.Rejected = append(result.Rejected, Rejection{Effect: applied.Effect, Reason: reason})
		}
	}
	if len(result.Rejected) > 0 {
		return result, nil
	}
	mana, err := spellCost(ruleset, recipe.Effects)
	if err != nil {
		return CraftResult{}, err
	}
	var craftThreshold int64 = roundHalfUp(float64(mana) * 1.25)
	result.Mana, result.CastThreshold, result.CraftThreshold = &mana, &mana, &craftThreshold
	result.SkillOK = caster.Skills[recipe.Color+"_magic"] >= craftThreshold
	result.Valid = result.SkillOK
	return result, nil
}

// sortedUnion is known plus added, sorted, without repeats.
func sortedUnion(known []string, added ...string) []string {
	var union []string = append(append([]string{}, known...), added...)
	slices.Sort(union)
	return slices.Compact(union)
}

func learn(ruleset *data.Ruleset, caster *Caster, key string) (LearnResult, error) {
	spell, err := ruleset.Spells.Require("spell", key)
	if err != nil {
		return LearnResult{}, err
	}
	var taught []string = []string{}
	for index := range spell.Effects {
		var ref string = spell.Effects[index].Effect
		effect, err := ruleset.Effects.Require("effect", ref)
		if err != nil {
			return LearnResult{}, err
		}
		if isCraftable(effect) && !slices.Contains(caster.Known, ref) && !slices.Contains(taught, ref) {
			taught = append(taught, ref)
		}
	}
	return LearnResult{Taught: taught, Known: sortedUnion(caster.Known, taught...)}, nil
}

func research(ruleset *data.Ruleset, input *Input) (ResearchResult, error) {
	effect, err := ruleset.Effects.Require("effect", input.Effect)
	if err != nil {
		return ResearchResult{}, err
	}
	var known []string = sortedUnion(input.Caster.Known)
	refuse := func(reason string) (ResearchResult, error) {
		return ResearchResult{Outcome: "refused", Reason: &reason, Known: known}, nil
	}
	if !isCraftable(effect) {
		return refuse("uncraftable")
	}
	if slices.Contains(input.Caster.Known, input.Effect) {
		return refuse("known")
	}
	var color string = effect.Color
	if color == "" {
		color = input.Color
	}
	if color == "" {
		return refuse("no_color")
	}
	var skill string = color + "_magic"
	var value int64 = input.Caster.Skills[skill]
	cheapest, err := spellCost(ruleset, []data.AppliedEffect{{Effect: input.Effect, Magnitude: 1, Target: "touch"}})
	if err != nil {
		return ResearchResult{}, err
	}
	if value < roundHalfUp(float64(cheapest)*1.25) {
		return refuse("skill")
	}
	bonus, err := resolution.Bonus(ruleset.Origins, modifier.Target(input.Modifiers, skill))
	if err != nil {
		return ResearchResult{}, err
	}
	var check resolution.Result = resolution.Check(&ruleset.Resolution, value, input.Caster.Attributes.Get(actor.Luck), bonus, input.Roll)
	var cost data.Cost = ruleset.Casting.Research.Cost
	var result ResearchResult = ResearchResult{Check: &check, Outcome: "failed", Known: known, Days: ruleset.Casting.Research.Days, Cost: &cost}
	if check.Success {
		result.Outcome, result.Known = "learned", sortedUnion(known, input.Effect)
	}
	return result, nil
}
