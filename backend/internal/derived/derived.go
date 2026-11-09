// Package derived computes the stats an actor's attributes and level set: hit points, mana,
// stamina, action points, carry weight, initiative (docs/core/derived.md), and the action pool and
// action costs (docs/core/actions.md). Every coefficient is a row in the data.
package derived

import (
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

// ActionsResult is the pool and the cost of the named actions.
type ActionsResult struct {
	Pool int64 `json:"pool"`
	Cost int64 `json:"cost"`
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
	return result, nil
}

// MonsterRow is what the sheet reads of one composed monster row.
type MonsterRow struct {
	Attributes actor.Attributes `json:"attributes"`
	Skills     map[string]int64 `json:"skills"`
	Level      *int64           `json:"level"`
	HitPoints  *int64           `json:"hitPoints"`
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
	values, err := Compute(ruleset, actor.Actor{Attributes: row.Attributes, Level: row.Level}, nil)
	if err != nil {
		return nil, err
	}
	var result map[string]any = make(map[string]any, len(values)+1)
	for index, stat := range ruleset.Derived.Rows() {
		result[stat.Key] = values[index]
	}
	if row.HitPoints != nil {
		result["hp"] = *row.HitPoints
	}
	var skills map[string]int64 = make(map[string]int64, len(input.ReadSkills))
	for _, key := range input.ReadSkills {
		if written, ok := row.Skills[key]; ok {
			skills[key] = written
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
		skills[key] = start
	}
	result["skills"] = skills
	return result, nil
}
