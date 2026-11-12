// Package progression is XP, level, the skill points of a level-up, spending them on a skill
// or a skill group, and reading a skill off a sheet (docs/core/progression.md). The curve, the
// caps, and the costs are in the data.
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
	HPRoll     *int64       `json:"hp_roll"`
	MPRoll     *int64       `json:"mp_roll"`
	SPRoll     *int64       `json:"sp_roll"`
	Skill      *int64       `json:"skill"`
	Tagged     *bool        `json:"tagged"`
	Points     *int64       `json:"points"`
	Group      *string      `json:"group"`
	Read       *string      `json:"read"`
	// Groups is a grouped sheet: the ranks of each tracked skill group. Nil is a full sheet.
	Groups map[string]int64 `json:"groups"`
}

type readResult struct {
	Value    int64  `json:"value"`
	ReadFrom string `json:"read_from"`
}

type experienceResult struct {
	Total int64  `json:"xp_total"`
	Step  *int64 `json:"xp_step"`
}

type levelResult struct {
	Level int64 `json:"level"`
}

type levelUpResult struct {
	Gain        map[string]int64 `json:"gain"`
	SkillPoints int64            `json:"skill_points"`
	Perks       int64            `json:"perks"`
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
	return &data.MissingInputError{Subject: "progression " + operation, Needs: field}
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
		if err := requireField(input.Operation, "actor and the four rolls", input.Actor != nil && input.LuckRoll != nil && input.HPRoll != nil && input.MPRoll != nil && input.SPRoll != nil); err != nil {
			return nil, err
		}
		var attributes actor.Attributes = input.Actor.Attributes
		if !attributes.Has(rules.LuckMaximumAbility) {
			return nil, &data.MissingInputError{Subject: "progression level_up", Needs: "the " + rules.LuckMaximumAbility.String() + " attribute"}
		}
		if *input.LuckRoll < rules.LuckRoll.Minimum || *input.LuckRoll > attributes.Get(rules.LuckMaximumAbility) {
			return nil, errors.New("luck_roll out of range")
		}
		highest, err := rules.PoolRollMaximum.Evaluate(&attributes)
		if err != nil {
			return nil, err
		}
		var gain map[string]int64 = make(map[string]int64, 3)
		for _, pool := range []struct {
			roll *int64
			key  string
		}{{input.HPRoll, "hp"}, {input.MPRoll, "mp"}, {input.SPRoll, "sp"}} {
			if *pool.roll < 0 || *pool.roll > highest {
				return nil, errors.New(pool.key + "_roll out of range")
			}
			stat, err := ruleset.Derived.Require("derived stat", pool.key)
			if err != nil {
				return nil, err
			}
			gain[pool.key] = stat.PerLevelFlat + *pool.roll
		}
		points, err := rules.SkillPoints.Evaluate(&attributes, formula.Variable{Name: "luck_roll", Value: *input.LuckRoll})
		if err != nil {
			return nil, err
		}
		return levelUpResult{SkillPoints: points, Perks: rules.PerksPerLevel, Gain: gain}, nil
	case "spend":
		var allGiven bool = input.Skill != nil && input.Tagged != nil && input.Points != nil
		if err := requireField(input.Operation, "skill, tagged and points", allGiven); err != nil {
			return nil, err
		}
		var cost int64 = rules.SkillCost.Normal
		if *input.Tagged {
			cost = rules.SkillCost.Tag
		}
		if input.Group != nil {
			group, err := ruleset.SkillGroups.Require("skill group", *input.Group)
			if err != nil {
				return nil, err
			}
			if !group.IsTracked {
				return nil, fmt.Errorf("skill group %q is not tracked", group.Key)
			}
			var members int64 = int64(len(group.SkillKeys()))
			var step int64 = rules.SkillGroupCost.MembersPerStep
			cost *= (members + step - 1) / step
		}
		var ranks int64 = min(*input.Points/cost, max(rules.SkillCap-*input.Skill, 0))
		return spendResult{Ranks: ranks, Spent: ranks * cost, Left: *input.Points - ranks*cost}, nil
	case "read_skill":
		if err := requireField(input.Operation, "actor and read", input.Actor != nil && input.Read != nil); err != nil {
			return nil, err
		}
		return readSkill(ruleset, input.Actor, *input.Read, input.Groups)
	}
	return nil, &data.NotDefinedError{Kind: "progression op", Key: input.Operation}
}

// readSkill reads one skill off a sheet (docs/core/progression.md#two-sheets-one-rule-set). A
// full sheet reads the written value, else the skill's start; a grouped sheet reads the start
// plus the member's group ranks, or the start alone for an untracked group.
func readSkill(ruleset *data.Ruleset, sheet *actor.Actor, key string, groups map[string]int64) (readResult, error) {
	skill, err := ruleset.Skills.Require("skill", key)
	if err != nil {
		return readResult{}, err
	}
	start, err := ruleset.Progression.SkillStart.Evaluate(&sheet.Attributes, formula.Variable{Name: "attribute", Value: sheet.Attributes.Get(skill.Ability)})
	if err != nil {
		return readResult{}, err
	}
	if groups == nil {
		if written, ok := sheet.Skills[key]; ok {
			return readResult{Value: written, ReadFrom: "sheet"}, nil
		}
		return readResult{Value: start, ReadFrom: "start"}, nil
	}
	group, ok := ruleset.GroupOfSkill[key]
	if !ok {
		return readResult{}, &data.NotDefinedError{Kind: "skill group of skill", Key: key}
	}
	if !group.IsTracked {
		return readResult{Value: start, ReadFrom: "start"}, nil
	}
	return readResult{Value: start + groups[group.Key], ReadFrom: "group"}, nil
}
