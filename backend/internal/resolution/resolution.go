// Package resolution is the d100 roll-under check (docs/core/resolution.md). Combat and
// the barter check call Check; nothing else decides a success, a crit, or a fumble.
package resolution

import (
	"fmt"
	"math"

	"github.com/SETA1609/riftweave-core/backend/internal/actor"
	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/formula"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
)

// Input is the `resolution` command's input.
type Input struct {
	Actor     actor.Actor         `json:"actor"`
	Skill     string              `json:"skill"`
	Modifiers []modifier.Modifier `json:"modifiers"`
	Roll      int64               `json:"roll"`
}

// Result is one check. Margin is nil on a failure.
type Result struct {
	Target  int64    `json:"target"`
	Success bool     `json:"success"`
	Margin  *int64   `json:"margin"`
	Tags    []string `json:"tags"`
}

// Bonus is the signed sum of every `add` modifier's value. Resolution sees bonuses and
// penalties as one list.
func Bonus(modifiers []modifier.Modifier) int64 {
	var sum float64 = 0.0
	for _, current := range modifiers {
		if current.Operation == modifier.Add {
			sum += current.Value
		}
	}
	return int64(math.Floor(sum))
}

// Check rolls one check: the skill plus the bonus is the target, capped if the data caps it.
func Check(rules *data.Resolution, skill, luck, bonus, roll int64) Result {
	var target int64 = skill + bonus
	if rules.TargetCap != nil {
		target = min(target, *rules.TargetCap)
	}
	var success bool = roll <= target && roll < rules.FumbleAt
	var criticalWindow int64 = rules.CriticalWindow + formula.FloorDivide(luck, rules.CriticalWindowPerTwoLuck)
	var result Result = Result{Target: target, Success: success, Tags: []string{}}
	if success {
		var margin int64 = target - roll
		result.Margin = &margin
		if roll <= criticalWindow {
			result.Tags = append(result.Tags, "crit")
		}
	}
	if roll >= rules.FumbleAt {
		result.Tags = append(result.Tags, "fumble")
	}
	return result
}

// Run is the `resolution` command.
func Run(ruleset *data.Ruleset, input Input) (Result, error) {
	var skill string = input.Skill
	if skill == "" {
		if len(input.Actor.Skills) != 1 {
			return Result{}, fmt.Errorf("actor has several skills; name one with `skill`")
		}
		for name := range input.Actor.Skills {
			skill = name
		}
	}
	value, ok := input.Actor.Skills[skill]
	if !ok {
		return Result{}, fmt.Errorf("actor has no skill %q", skill)
	}
	return Check(&ruleset.Resolution, value, input.Actor.Attributes.Get(actor.Luck), Bonus(input.Modifiers), input.Roll), nil
}
