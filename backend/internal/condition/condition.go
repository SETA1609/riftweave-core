// Package condition applies effect and remove events to an actor's active conditions, then
// reads the values and flags the conditions' modifiers give (docs/core/conditions.md).
package condition

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
)

// active is one condition on an actor: its row and its stacks.
type active struct {
	key    string
	row    *data.Condition
	stacks int64
}

// Stacks is an object of condition key to stacks. It keeps the order of the keys, because
// the order of an actor's conditions is the order of their modifiers, and `set` is won by
// the last.
type Stacks struct {
	keys []string
	vals []int64
}

// UnmarshalJSON reads the object in key order.
func (s *Stacks) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(strings.NewReader(string(b)))
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		return errors.New("conditions must be an object")
	}
	*s = Stacks{}
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		var n int64
		if err := dec.Decode(&n); err != nil {
			return err
		}
		s.keys = append(s.keys, tok.(string))
		s.vals = append(s.vals, n)
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

type state []active

func (s state) find(key string) int {
	for i := range s {
		if s[i].key == key {
			return i
		}
	}
	return -1
}

func (s *state) remove(key string) {
	if i := s.find(key); i >= 0 {
		*s = append((*s)[:i], (*s)[i+1:]...)
	}
}

// Run is the `conditions` command. An effect first removes the conditions it is
// `removedBy`, then applies the ones it is `appliedBy`, each adding a stack up to its cap.
func Run(rs *data.Ruleset, in Input) (Result, error) {
	table := &rs.Conditions
	var conds state
	for i, key := range in.Conditions.keys {
		row, ok := table.Table.Get(key)
		if !ok {
			return Result{}, fmt.Errorf("condition %q is not defined", key)
		}
		conds = append(conds, active{key, row, in.Conditions.vals[i]})
	}
	for _, event := range in.Events {
		switch {
		case event.Remove != "":
			conds.remove(event.Remove[strings.LastIndex(event.Remove, "/")+1:])
		case event.Effect != "":
			for _, i := range table.RemovedBy[event.Effect] {
				conds.remove(table.Table.At(i).Key)
			}
			for _, i := range table.AppliedBy[event.Effect] {
				row := table.Table.At(i)
				if n := conds.find(row.Key); n >= 0 {
					conds[n].stacks = min(conds[n].stacks+1, row.MaxStacks)
				} else {
					conds = append(conds, active{row.Key, row, 1})
				}
			}
		default:
			return Result{}, errors.New("an event needs `effect` or `remove`")
		}
	}
	mods := append([]modifier.Modifier(nil), in.Modifiers...)
	res := Result{Conditions: map[string]int64{}, Values: map[string]int64{}}
	for _, c := range conds {
		mods = append(mods, c.row.Modifiers...)
		res.Conditions[c.key] = c.stacks
	}
	for target, base := range in.Base {
		v, err := modifier.Value(base, modifier.Target(mods, target), rs.Origins, 0, nil)
		if err != nil {
			return Result{}, err
		}
		res.Values[target] = v
	}
	res.Flags = modifier.Flags(mods, 0, nil)
	return res, nil
}
