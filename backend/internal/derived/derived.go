// Package derived computes the stats an actor's attributes and level set: hit points, mana,
// stamina, action points, carry weight, initiative (docs/core/derived.md), and the action pool and
// action costs (docs/core/actions.md). Every coefficient is a row in the data.
package derived

import (
	"encoding/json"
	"fmt"
	"maps"
	"math"
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
		// The derived ap is the action pool, kept between its floor and cap (GD_S19_07).
		if stat.Key == apTarget {
			modifiedValue = min(max(modifiedValue, ruleset.Actions.Pool.Floor), ruleset.Actions.Pool.Cap)
		}
		values[index] = modifiedValue
	}
	return values, nil
}

// apTarget is the modifier target and derived stat of the action pool.
const apTarget = "ap"

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
	// Speed is the race's speed; with it the result says how many squares the moves cover (GD_S19_02).
	Speed *int64 `json:"speed"`
	// Carried is the AP carried into this turn; HeldSpent is held AP spent before the next turn,
	// on overwatch (docs/core/actions.md#held-ap, GD_S19_05; Trigger Watch, GD_S17_06).
	Carried   int64 `json:"carried"`
	HeldSpent int64 `json:"held_spent"`
}

// ActionsResult is the pool and the cost of the named actions, in AP and in real-time stamina.
type ActionsResult struct {
	Pool    int64 `json:"pool"`
	Cost    int64 `json:"cost"`
	Stamina int64 `json:"stamina"` // Cost times stamina_per_ap (docs/core/actions.md#one-rule-set-two-clocks)
	// Squares the actions cover, given a speed (docs/core/actions.md#how-far-a-move-goes).
	Squares *int64 `json:"squares,omitempty"`
	// Held AP (docs/core/actions.md#held-ap): the pool with carried AP, what the turn leaves held,
	// the evasion it guards with, what carries, and the next turn's pool.
	Available int64 `json:"available"`
	Held      int64 `json:"held"`
	Guard     int64 `json:"guard"`
	Carry     int64 `json:"carry"`
	NextPool  int64 `json:"next_pool"`
}

// RunActions is the `actions` command: the pool is base + attribute * per_point / divisor.
func RunActions(ruleset *data.Ruleset, input ActionsInput) (ActionsResult, error) {
	var pool *data.ActionPool = &ruleset.Actions.Pool
	if !input.Actor.Attributes.Has(pool.Ability) {
		return ActionsResult{}, &data.MissingInputError{Subject: "actor", Needs: "the " + pool.Ability.String() + " attribute"}
	}
	var attributeValue int64 = input.Actor.Attributes.Get(pool.Ability)
	// Modifiers on ap change the pool, kept between its floor and cap (GD_S19_07).
	poolValue, err := modifier.Value(pool.Base+formula.FloorDivide(attributeValue*pool.PerPoint, pool.PerPointDivisor), modifier.Target(input.Modifiers, apTarget), ruleset.Origins, 0, nil)
	if err != nil {
		return ActionsResult{}, err
	}
	var result ActionsResult = ActionsResult{Pool: min(max(poolValue, pool.Floor), pool.Cap)}
	var squares int64 = 0
	for _, key := range input.Actions {
		action, err := ruleset.Actions.Actions.Require("action", key)
		if err != nil {
			return ActionsResult{}, err
		}
		// A move's squares come from speed alone; modifiers on it change its AP (GD_S19_02).
		if input.Speed != nil && action.SpeedDivisor > 0 {
			squares += formula.FloorDivide(*input.Speed, action.SpeedDivisor)
		}
		cost, err := modifier.Value(action.Cost, modifier.Target(input.Modifiers, key), ruleset.Origins, 0, nil)
		if err != nil {
			return ActionsResult{}, err
		}
		result.Cost += max(cost, 0)
	}
	result.Stamina = result.Cost * ruleset.StaminaPerAP
	if input.Speed != nil {
		result.Squares = &squares
	}
	if input.Carried > pool.Carry {
		return ActionsResult{}, fmt.Errorf("carried %d is more than pool.carry %d", input.Carried, pool.Carry)
	}
	result.Available = min(result.Pool+input.Carried, pool.Cap)
	result.Held = max(result.Available-result.Cost, 0)
	if input.HeldSpent > result.Held {
		return ActionsResult{}, fmt.Errorf("held_spent %d is more than the %d AP held", input.HeldSpent, result.Held)
	}
	var left int64 = result.Held - input.HeldSpent
	result.Guard = left * pool.GuardPerAP
	result.Carry = min(left, pool.Carry)
	result.NextPool = min(result.Pool+result.Carry, pool.Cap)
	return result, nil
}

// MonsterRow is what the sheet reads of one composed monster row.
type MonsterRow struct {
	Attributes actor.Attributes `json:"attributes"`
	Skills     map[string]int64 `json:"skills"`
	Level      *int64           `json:"level"`
	// Role scales the computed HP by progression's monster_hp_multiplier; empty is standard (DE_S02_03).
	Role string `json:"role"`
	// Perks are the row's perks at their rank (M3-15); their modifiers reach the sheet.
	Perks []struct {
		Perk string `json:"perk"`
		Rank *int64 `json:"rank"`
	} `json:"perks"`
	// Maneuvers are the ones the row lists (M3-24); its perks may unlock more (GD_S05_04).
	Maneuvers []string `json:"maneuvers"`
}

// maneuversAndRiders are the row's maneuvers, listed or unlocked by its perks, sorted, and
// the maneuvers its held rider perks put on its plain attacks, sorted. Two held perks that
// exclude each other are an error.
func maneuversAndRiders(ruleset *data.Ruleset, row *MonsterRow) ([]string, []string, error) {
	var known map[string]bool = map[string]bool{}
	for _, ref := range row.Maneuvers {
		known[ref] = true
	}
	var held []*data.Feature
	var riders []string
	for _, item := range row.Perks {
		feature, err := ruleset.Features.Require("perk", item.Perk[strings.LastIndex(item.Perk, "/")+1:])
		if err != nil {
			return nil, nil, err
		}
		for _, ref := range feature.Unlocks {
			if strings.HasPrefix(ref, "core:maneuver/") {
				known[ref] = true
			}
		}
		if feature.Rider != "" {
			riders = append(riders, feature.Rider)
		}
		held = append(held, feature)
	}
	for index, first := range held {
		for _, second := range held[index+1:] {
			if data.Exclude(first, second) {
				return nil, nil, fmt.Errorf("perks %s and %s exclude each other", first.Key, second.Key)
			}
		}
	}
	var maneuvers []string = make([]string, 0, len(known))
	for ref := range known {
		maneuvers = append(maneuvers, ref)
	}
	slices.Sort(maneuvers)
	slices.Sort(riders)
	return maneuvers, riders, nil
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

// checkMonsterLimits refuses the two limits a custom monster keeps besides its references: a level
// over max_level and a written skill over skill_cap (docs/core/actor-sheet.md#a-custom-monster).
// Point-buy, the attribute cap, and perk prerequisites and ranks do not apply to a monster.
func checkMonsterLimits(ruleset *data.Ruleset, row *MonsterRow) error {
	var progression = &ruleset.Progression
	if row.Level != nil && *row.Level > progression.MaximumLevel {
		return fmt.Errorf("level %d is over the maximum level %d", *row.Level, progression.MaximumLevel)
	}
	for _, key := range slices.Sorted(maps.Keys(row.Skills)) {
		if row.Skills[key] > progression.SkillCap {
			return fmt.Errorf("skill %q %d is over the skill cap %d", key, row.Skills[key], progression.SkillCap)
		}
	}
	return nil
}

// MonsterSheetInput is the `monster_sheet` command's input.
type MonsterSheetInput struct {
	Monster    MonsterRow `json:"monster"`
	ReadSkills []string   `json:"read_skills"`
}

// RunMonsterSheet is the `monster_sheet` command (docs/core/actor-sheet.md): a monster's stats
// and skills by the rules a player uses; its HP is the player formula before perks times its
// role's monster_hp_multiplier, rounded down and at least 1, then its perks' hp modifiers (DE_S02_03).
func RunMonsterSheet(ruleset *data.Ruleset, input MonsterSheetInput) (map[string]any, error) {
	var row *MonsterRow = &input.Monster
	if err := checkMonsterLimits(ruleset, row); err != nil {
		return nil, err
	}
	perks, err := perkModifiers(ruleset, row)
	if err != nil {
		return nil, err
	}
	var beforeHP []modifier.Modifier
	for _, current := range perks {
		if current.Target != "hp" {
			beforeHP = append(beforeHP, current)
		}
	}
	values, err := Compute(ruleset, actor.Actor{Attributes: row.Attributes, Level: row.Level}, beforeHP)
	if err != nil {
		return nil, err
	}
	var result map[string]any = make(map[string]any, len(values)+1)
	for index, stat := range ruleset.Derived.Rows() {
		result[stat.Key] = values[index]
	}
	var role string = row.Role
	if role == "" {
		role = "standard"
	}
	share, ok := ruleset.Progression.MonsterHPMultiplier[role]
	if !ok {
		return nil, fmt.Errorf("monster role %q has no monster_hp_multiplier", role)
	}
	base, _ := result["hp"].(int64)
	// The tolerance keeps a share stored just under its decimal (0.7 is 0.6999...) from losing a
	// point: 90 x 0.7 is 63, not 62. Python's scaled_hp does the same.
	var scaled int64 = max(1, int64(math.Floor(float64(base)*share+1e-9)))
	hp, err := modifier.Value(scaled, modifier.Target(perks, "hp"), ruleset.Origins, 0, nil)
	if err != nil {
		return nil, err
	}
	result["hp"] = hp
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
	maneuvers, riders, err := maneuversAndRiders(ruleset, row)
	if err != nil {
		return nil, err
	}
	result["maneuvers"] = maneuvers
	if len(riders) > 0 {
		result["riders"] = riders
	}
	return result, nil
}
