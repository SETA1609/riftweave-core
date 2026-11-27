// Package encounter runs a fight offline (docs/backend/sessions.md#encounters, BE_S08_07): the
// initiative order, whose turn it is, and the round. It rolls no dice: the roll-off that breaks a
// tie (docs/core/derived.md#initiative-ties) is an input, kept in the step, so a replay is the same.
package encounter

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/SETA1609/riftweave-core/backend/internal/data"
)

// Operation is what an encounter call does.
type Operation string

// The operations.
const (
	OperationStart   Operation = "start"
	OperationTurnEnd Operation = "turn_end"
)

// Start is the actions command's start for a turn.
type Start string

// The starts a turn may read.
const (
	StartCombat   Start = "combat"
	StartSurprise Start = "surprise"
)

// Combatant is one actor in the fight.
type Combatant struct {
	ID         string  `json:"id"`
	Initiative int64   `json:"initiative"`
	Agility    int64   `json:"agi"`
	Luck       int64   `json:"lck"`
	Rolls      []int64 `json:"rolls"`
	Surprise   bool    `json:"surprise"`
}

// State is the fight: round 0 is the surprise round, in which only Surprise act.
type State struct {
	Round    int64    `json:"round"`
	Order    []string `json:"order"`
	Turn     int64    `json:"turn"`
	Acted    []string `json:"acted"`
	Surprise []string `json:"surprise"`
}

// Input is the `encounter` command's input.
type Input struct {
	Operation  Operation   `json:"operation"`
	Combatants []Combatant `json:"combatants"`
	Encounter  *State      `json:"encounter"`
}

// Current is whose turn it is, and the start it reads.
type Current struct {
	ID    string `json:"id"`
	Start Start  `json:"start,omitempty"`
}

// Result is the fight, whose turn it is, and whose comes next.
type Result struct {
	Encounter State   `json:"encounter"`
	Current   Current `json:"current"`
	Next      string  `json:"next"`
}

// Run is the `encounter` command.
func Run(_ *data.Ruleset, input Input) (Result, error) {
	switch input.Operation {
	case OperationStart:
		return start(input.Combatants)
	case OperationTurnEnd:
		if input.Encounter == nil {
			return Result{}, errors.New("turn_end needs the encounter")
		}
		return view(advance(*input.Encounter)), nil
	}
	return Result{}, fmt.Errorf("encounter operation %q is not defined", input.Operation)
}

// compare orders two combatants: highest initiative, then AGI, then LCK, then the roll-off lowest first.
func compare(first, second Combatant) int {
	return cmp.Or(
		cmp.Compare(second.Initiative, first.Initiative),
		cmp.Compare(second.Agility, first.Agility),
		cmp.Compare(second.Luck, first.Luck),
		slices.Compare(first.Rolls, second.Rolls),
	)
}

func start(combatants []Combatant) (Result, error) {
	var seen map[string]bool = map[string]bool{}
	for _, combatant := range combatants {
		if seen[combatant.ID] {
			return Result{}, fmt.Errorf("%s is in the fight twice", combatant.ID)
		}
		seen[combatant.ID] = true
	}
	var ranked []Combatant = slices.Clone(combatants)
	slices.SortStableFunc(ranked, compare)
	for index := 1; index < len(ranked); index++ {
		if compare(ranked[index-1], ranked[index]) == 0 {
			return Result{}, fmt.Errorf("%s and %s tie: roll off again", ranked[index-1].ID, ranked[index].ID)
		}
	}
	var state State = State{Round: 1, Order: make([]string, len(ranked)), Acted: []string{}, Surprise: []string{}}
	for index, combatant := range ranked {
		state.Order[index] = combatant.ID
		if combatant.Surprise {
			if len(state.Surprise) == 0 {
				state.Round, state.Turn = 0, int64(index)
			}
			state.Surprise = append(state.Surprise, combatant.ID)
		}
	}
	return view(state), nil
}

// advance is the state after the current turn ends: the next actor in the round (in the surprise
// round, the next ambusher), or the next round from the top.
func advance(state State) State {
	var current string = state.Order[state.Turn]
	var next State = state
	next.Acted = slices.Clone(state.Acted)
	if !slices.Contains(next.Acted, current) {
		next.Acted = append(next.Acted, current)
	}
	for index := state.Turn + 1; index < int64(len(state.Order)); index++ {
		if state.Round > 0 || slices.Contains(state.Surprise, state.Order[index]) {
			next.Turn = index
			return next
		}
	}
	next.Round, next.Turn = state.Round+1, 0
	return next
}

// view is the state with whose turn it is and whose comes next.
func view(state State) Result {
	var id string = state.Order[state.Turn]
	var current Current = Current{ID: id}
	if state.Round == 0 {
		current.Start = StartSurprise
	} else if !slices.Contains(state.Acted, id) {
		current.Start = StartCombat
	}
	var after State = advance(state)
	return Result{Encounter: state, Current: current, Next: after.Order[after.Turn]}
}
