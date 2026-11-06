// Package progression is XP, level, the skill points of a level-up, and spending them on a
// skill (docs/core/progression.md). The curve, the caps, and the costs are in the data.
package progression

import (
	"errors"
	"fmt"

	"github.com/SETA1609/riftweave-core/backend/internal/actor"
	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/formula"
)

// Input is the `progression` command's input; `op` says which fields it reads.
type Input struct {
	Operation  string       `json:"op"`
	Level      *int64       `json:"level"`
	Experience *int64       `json:"xp"`
	Actor      *actor.Actor `json:"actor"`
	LuckRoll   *int64       `json:"luck_roll"`
	Skill      *int64       `json:"skill"`
	Tagged     *bool        `json:"tagged"`
	Points     *int64       `json:"points"`
}

type experienceResult struct {
	Total int64  `json:"xp_total"`
	Step  *int64 `json:"xp_step"`
}

type levelResult struct {
	Level int64 `json:"level"`
}

type levelUpResult struct {
	SkillPoints int64 `json:"skill_points"`
	Perks       int64 `json:"perks"`
}

type spendResult struct {
	Ranks int64 `json:"ranks"`
	Spent int64 `json:"spent"`
	Left  int64 `json:"left"`
}

func requireField(operation, field string, given bool) error {
	if given {
		return nil
	}
	return fmt.Errorf("progression %s needs %s", operation, field)
}

// Run is the `progression` command. Its result type depends on `op`.
func Run(ruleset *data.Ruleset, input Input) (any, error) {
	var rules *data.Progression = &ruleset.Progression
	switch input.Operation {
	case "xp":
		if err := requireField(input.Operation, "level", input.Level != nil); err != nil {
			return nil, err
		}
		total, err := rules.ExperienceToReach.Evaluate(nil, formula.Variable{Name: "level", Value: *input.Level})
		if err != nil {
			return nil, err
		}
		var result experienceResult = experienceResult{Total: total}
		if *input.Level > rules.StartLevel {
			previousTotal, err := rules.ExperienceToReach.Evaluate(nil, formula.Variable{Name: "level", Value: *input.Level - 1})
			if err != nil {
				return nil, err
			}
			var step int64 = total - previousTotal
			result.Step = &step
		}
		return result, nil
	case "level":
		if err := requireField(input.Operation, "xp", input.Experience != nil); err != nil {
			return nil, err
		}
		var level int64 = rules.StartLevel
		for level < rules.MaximumLevel {
			needed, err := rules.ExperienceToReach.Evaluate(nil, formula.Variable{Name: "level", Value: level + 1})
			if err != nil {
				return nil, err
			}
			if needed > *input.Experience {
				break
			}
			level++
		}
		return levelResult{Level: level}, nil
	case "level_up":
		if err := requireField(input.Operation, "actor and luck_roll", input.Actor != nil && input.LuckRoll != nil); err != nil {
			return nil, err
		}
		var attributes actor.Attributes = input.Actor.Attributes
		if !attributes.Has(rules.LuckMaximumAbility) {
			return nil, fmt.Errorf("progression level_up needs the %s attribute", rules.LuckMaximumAbility)
		}
		if *input.LuckRoll < rules.LuckRoll.Minimum || *input.LuckRoll > attributes.Get(rules.LuckMaximumAbility) {
			return nil, errors.New("luck_roll out of range")
		}
		points, err := rules.SkillPoints.Evaluate(&attributes, formula.Variable{Name: "luck_roll", Value: *input.LuckRoll})
		if err != nil {
			return nil, err
		}
		return levelUpResult{SkillPoints: points, Perks: rules.PerksPerLevel}, nil
	case "spend":
		var allGiven bool = input.Skill != nil && input.Tagged != nil && input.Points != nil
		if err := requireField(input.Operation, "skill, tagged and points", allGiven); err != nil {
			return nil, err
		}
		var cost int64 = rules.SkillCost.Normal
		if *input.Tagged {
			cost = rules.SkillCost.Tag
		}
		var ranks int64 = min(*input.Points/cost, max(rules.SkillCap-*input.Skill, 0))
		return spendResult{Ranks: ranks, Spent: ranks * cost, Left: *input.Points - ranks*cost}, nil
	}
	return nil, fmt.Errorf("progression op %q is not defined", input.Operation)
}
