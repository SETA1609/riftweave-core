// Package condition applies effect and remove events to an actor's active conditions, then
// reads the values and flags the conditions' modifiers give (docs/core/conditions.md).
package condition

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
)

// activeCondition is one condition on an actor: its row and its stacks.
type activeCondition struct {
	row    *data.Condition
	key    string
	stacks int64
}

// Stacks is an object of condition key to stacks. It keeps the order of the keys, because
// the order of an actor's conditions is the order of their modifiers, and `set` is won by
// the last.
type Stacks struct {
	keys   []string
	counts []int64
}

// UnmarshalJSON reads the object in key order.
func (stacks *Stacks) UnmarshalJSON(encoded []byte) error {
	var decoder *json.Decoder = json.NewDecoder(strings.NewReader(string(encoded)))
	if token, err := decoder.Token(); err != nil || token != json.Delim('{') {
		return errors.New("conditions must be an object")
	}
	*stacks = Stacks{}
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		var count int64
		if err := decoder.Decode(&count); err != nil {
			return err
		}
		stacks.keys = append(stacks.keys, token.(string))
		stacks.counts = append(stacks.counts, count)
	}
	return nil
}

// Event is an effect landing on the actor, or a condition removed by its ref.
type Event struct {
	Effect string `json:"effect"`
	Remove string `json:"remove"`
}

// Input is the `conditions` command's input.
type Input struct {
	Base       map[string]int64    `json:"base"`
	Modifiers  []modifier.Modifier `json:"modifiers"`
	Conditions Stacks              `json:"conditions"`
	Events     []Event             `json:"events"`
}

// Result is the conditions left, the flags granted, and one value per base target.
type Result struct {
	Conditions map[string]int64 `json:"conditions"`
	Flags      []string         `json:"flags"`
	Values     map[string]int64 `json:"values"`
}

type activeConditions []activeCondition

func (conditions activeConditions) find(key string) int {
	for index := range conditions {
		if conditions[index].key == key {
			return index
		}
	}
	return -1
}

func (conditions *activeConditions) remove(key string) {
	if index := conditions.find(key); index >= 0 {
		*conditions = append((*conditions)[:index], (*conditions)[index+1:]...)
	}
}

// Run is the `conditions` command. An effect first removes the conditions it is
// `removedBy`, then applies the ones it is `appliedBy`, each adding a stack up to its cap.
func Run(ruleset *data.Ruleset, input Input) (Result, error) {
	var table *data.Conditions = &ruleset.Conditions
	var active activeConditions
	for index, key := range input.Conditions.keys {
		row, err := table.Table.Require("condition", key)
		if err != nil {
			return Result{}, err
		}
		active = append(active, activeCondition{row: row, key: key, stacks: input.Conditions.counts[index]})
	}
	for _, event := range input.Events {
		switch {
		case event.Remove != "":
			active.remove(event.Remove[strings.LastIndex(event.Remove, "/")+1:])
		case event.Effect != "":
			for _, rowIndex := range table.RemovedBy[event.Effect] {
				active.remove(table.Table.At(rowIndex).Key)
			}
			for _, rowIndex := range table.AppliedBy[event.Effect] {
				var row *data.Condition = table.Table.At(rowIndex)
				if position := active.find(row.Key); position >= 0 {
					active[position].stacks = min(active[position].stacks+1, row.MaximumStacks)
				} else {
					active = append(active, activeCondition{row: row, key: row.Key, stacks: 1})
				}
			}
		default:
			return Result{}, &data.MissingInputError{Subject: "an event", Needs: "`effect` or `remove`"}
		}
	}
	var modifiers []modifier.Modifier = append([]modifier.Modifier(nil), input.Modifiers...)
	var result Result = Result{Conditions: map[string]int64{}, Values: map[string]int64{}}
	for _, condition := range active {
		modifiers = append(modifiers, condition.row.Modifiers...)
		result.Conditions[condition.key] = condition.stacks
	}
	for target, base := range input.Base {
		value, err := modifier.Value(base, modifier.Target(modifiers, target), ruleset.Origins, 0, nil)
		if err != nil {
			return Result{}, err
		}
		result.Values[target] = value
	}
	result.Flags = modifier.Flags(modifiers, 0, nil)
	return result, nil
}
