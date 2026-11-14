// Package derived computes the stats an actor's attributes and level set: hit points, mana,
// stamina, action points, carry weight, initiative (docs/core/derived.md), and the action pool and
// action costs (docs/core/actions.md). Every coefficient is a row in the data.
package derived

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/SETA1609/riftweave-core/backend/internal/actor"
	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/formula"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
)

// Input is the `derived` command's input.
type Input struct {
	Actor     actor.Actor         `json:"actor"`
	Modifiers []modifier.Modifier `json:"modifiers"`
}

// Compute returns one value per row of ruleset.Derived, in row order: the formula, then the
// modifiers that name the stat. A missing attribute reads as 0 and a missing level as 1.
func Compute(ruleset *data.Ruleset, character actor.Actor, modifiers []modifier.Modifier) ([]int64, error) {
	var level int64 = character.LevelOrDefault(1)
	var rows []data.DerivedStat = ruleset.Derived.Rows()
	var values []int64 = make([]int64, len(rows))
	for index := range rows {
		var stat *data.DerivedStat = &rows[index]
		var attributeValue int64 = character.Attributes.Get(stat.Ability)
		var unmodifiedValue int64 = stat.Base +
			formula.FloorDivide(attributeValue*stat.PerPoint, stat.PerPointDivisor) +
			level*(attributeValue*stat.PerLevelAttribute+stat.PerLevelFlat) +
			character.Growth[stat.Key]
		modifiedValue, err := modifier.Value(unmodifiedValue, modifier.Target(modifiers, stat.Key), ruleset.Origins, 0, nil)
		if err != nil {
			return nil, err
		}
		values[index] = modifiedValue
	}
	return values, nil
}

// Run is the `derived` command: one field per derived stat.
func Run(ruleset *data.Ruleset, input Input) (map[string]int64, error) {
	values, err := Compute(ruleset, input.Actor, input.Modifiers)
	if err != nil {
		return nil, err
	}
	var valuesByStat map[string]int64 = make(map[string]int64, len(values))
	for index, row := range ruleset.Derived.Rows() {
		valuesByStat[row.Key] = values[index]
	}
	return valuesByStat, nil
}

// ActionsInput is the `actions` command's input.
type ActionsInput struct {
	Actor     actor.Actor         `json:"actor"`
	Actions   []string            `json:"actions"`
	Modifiers []modifier.Modifier `json:"modifiers"` // adds on an action's key change its cost
}

// ActionsResult is the pool and the cost of the named actions, in AP and in real-time stamina.
type ActionsResult struct {
	Pool    int64 `json:"pool"`
	Cost    int64 `json:"cost"`
	Stamina int64 `json:"stamina"` // Cost times stamina_per_ap (docs/core/actions.md#one-rule-set-two-clocks)
}

// RunActions is the `actions` command: the pool is base + attribute * per_point / divisor.
func RunActions(ruleset *data.Ruleset, input ActionsInput) (ActionsResult, error) {
	var pool *data.ActionPool = &ruleset.Actions.Pool
	if !input.Actor.Attributes.Has(pool.Ability) {
		return ActionsResult{}, &data.MissingInputError{Subject: "actor", Needs: "the " + pool.Ability.String() + " attribute"}
	}
	var attributeValue int64 = input.Actor.Attributes.Get(pool.Ability)
	var result ActionsResult = ActionsResult{Pool: pool.Base + formula.FloorDivide(attributeValue*pool.PerPoint, pool.PerPointDivisor)}
	for _, key := range input.Actions {
		action, err := ruleset.Actions.Actions.Require("action", key)
		if err != nil {
			return ActionsResult{}, err
		}
		cost, err := modifier.Value(action.Cost, modifier.Target(input.Modifiers, key), ruleset.Origins, 0, nil)
		if err != nil {
			return ActionsResult{}, err
		}
		result.Cost += max(cost, 0)
	}
	result.Stamina = result.Cost * ruleset.StaminaPerAP
	return result, nil
}

// MonsterRow is what the sheet reads of one composed monster row.
type MonsterRow struct {
	Attributes actor.Attributes `json:"attributes"`
	Skills     map[string]int64 `json:"skills"`
	Level      *int64           `json:"level"`
	HitPoints  *int64           `json:"hitPoints"`
	// Perks are the row's perks at their rank (M3-15); their modifiers reach the sheet.
	Perks []struct {
		Perk string `json:"perk"`
		Rank *int64 `json:"rank"`
	} `json:"perks"`
}

// perkModifiers are the row's perks emitted at their rank, in list order.
func perkModifiers(ruleset *data.Ruleset, row *MonsterRow) ([]modifier.Modifier, error) {
	var modifiers []modifier.Modifier
	for _, held := range row.Perks {
		feature, err := ruleset.Features.Require("perk", held.Perk[strings.LastIndex(held.Perk, "/")+1:])
		if err != nil {
			return nil, err
		}
		var rank int64 = 1
		if held.Rank != nil {
			rank = *held.Rank
		}
		encoded, err := json.Marshal(feature.EmitAt(rank))
		if err != nil {
			return nil, err
		}
		var emitted []modifier.Modifier
		if err := json.Unmarshal(encoded, &emitted); err != nil {
			return nil, err
		}
		modifiers = append(modifiers, emitted...)
	}
	return modifiers, nil
}

// MonsterSheetInput is the `monster_sheet` command's input.
type MonsterSheetInput struct {
	Monster    MonsterRow `json:"monster"`
	ReadSkills []string   `json:"read_skills"`
}

// RunMonsterSheet is the `monster_sheet` command (docs/core/actor-sheet.md): a monster's stats
// and skills by the rules a player uses, with authored hit points replacing the computed value.
func RunMonsterSheet(ruleset *data.Ruleset, input MonsterSheetInput) (map[string]any, error) {
	var row *MonsterRow = &input.Monster
	perks, err := perkModifiers(ruleset, row)
	if err != nil {
		return nil, err
	}
	values, err := Compute(ruleset, actor.Actor{Attributes: row.Attributes, Level: row.Level}, perks)
	if err != nil {
		return nil, err
	}
	var result map[string]any = make(map[string]any, len(values)+1)
	for index, stat := range ruleset.Derived.Rows() {
		result[stat.Key] = values[index]
	}
	if row.HitPoints != nil {
		hp, err := modifier.Value(*row.HitPoints, modifier.Target(perks, "hp"), ruleset.Origins, 0, nil)
		if err != nil {
			return nil, err
		}
		result["hp"] = hp
	}
	var skills map[string]int64 = make(map[string]int64, len(input.ReadSkills))
	for _, key := range input.ReadSkills {
		if written, ok := row.Skills[key]; ok {
			value, err := modifier.Value(written, modifier.Target(perks, key), ruleset.Origins, 0, nil)
			if err != nil {
				return nil, err
			}
			skills[key] = value
			continue
		}
		skill, err := ruleset.Skills.Require("skill", key)
		if err != nil {
			return nil, err
		}
		start, err := ruleset.Progression.SkillStart.Evaluate(&row.Attributes, formula.Variable{Name: "attribute", Value: row.Attributes.Get(skill.Ability)})
		if err != nil {
			return nil, err
		}
		var growth = &ruleset.Progression.MonsterSkillGrowth
		if slices.Contains(growth.Categories, skill.Category) {
			var level int64 = 1
			if row.Level != nil {
				level = *row.Level
			}
			start = min(ruleset.Progression.SkillCap, start+growth.PerLevel*(level-1))
		}
		value, err := modifier.Value(start, modifier.Target(perks, key), ruleset.Origins, 0, nil)
		if err != nil {
			return nil, err
		}
		skills[key] = value
	}
	result["skills"] = skills
	return result, nil
}
