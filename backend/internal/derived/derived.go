// Package derived computes the stats an actor's attributes and level set: hit points, mana,
// action points, carry weight, initiative (docs/core/derived.md), and the action pool and
// action costs (docs/core/actions.md). Every coefficient is a row in the data.
package derived

import (
	"fmt"

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

// Compute returns one value per row of rs.Derived, in row order: the formula, then the
// modifiers that name the stat. A missing attribute reads as 0 and a missing level as 1.
func Compute(rs *data.Ruleset, a actor.Actor, mods []modifier.Modifier) ([]int64, error) {
	level := a.LevelOr(1)
	rows := rs.Derived.Rows()
	out := make([]int64, len(rows))
	for i := range rows {
		stat := &rows[i]
		attr := a.Attributes.Get(stat.AttrID)
		value := stat.Base +
			formula.FloorDiv(attr*stat.PerPoint, stat.PerPointDivisor) +
			level*(attr*stat.PerLevelAttr+stat.PerLevelFlat)
		v, err := modifier.Value(value, modifier.Target(mods, stat.Key), rs.Origins, 0, nil)
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}

// Run is the `derived` command: one field per derived stat.
func Run(rs *data.Ruleset, in Input) (map[string]int64, error) {
	values, err := Compute(rs, in.Actor, in.Modifiers)
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(values))
	for i, row := range rs.Derived.Rows() {
		out[row.Key] = values[i]
	}
	return out, nil
}

// ActionsInput is the `actions` command's input.
type ActionsInput struct {
	Actor   actor.Actor `json:"actor"`
	Actions []string    `json:"actions"`
}

// ActionsResult is the pool and the cost of the named actions.
type ActionsResult struct {
	Pool int64 `json:"pool"`
	Cost int64 `json:"cost"`
}

// RunActions is the `actions` command: the pool is base + attr * per_point / divisor.
func RunActions(rs *data.Ruleset, in ActionsInput) (ActionsResult, error) {
	pool := &rs.Actions.Pool
	if !in.Actor.Attributes.Has(pool.AttrID) {
		return ActionsResult{}, fmt.Errorf("actor needs the %s attribute", pool.AttrID)
	}
	res := ActionsResult{Pool: pool.Base + formula.FloorDiv(in.Actor.Attributes.Get(pool.AttrID)*pool.PerPoint, pool.PerPointDivisor)}
	for _, key := range in.Actions {
		action, ok := rs.Actions.Actions.Get(key)
		if !ok {
			return ActionsResult{}, fmt.Errorf("action %q is not defined", key)
		}
		res.Cost += action.Cost
	}
	return res, nil
}
