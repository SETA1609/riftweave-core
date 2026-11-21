// Package stance runs a fighting stance over a stretch of turns
// (docs/core/combat.md#fighting-stances): take or change one, end it, or change the loadout,
// which ends a stance it no longer fits. At most one is active. Taking or changing one is
// free at the start of the actor's own turn and costs mid_turn_ap later. The stance's
// modifiers are what the caller passes to the combat and attack_sequence commands.
package stance

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
)

// Event is one step: Take names a stance (MidTurn when not at the start of the turn), End
// ends the active one, and Loadout is what the hands now hold.
type Event struct {
	Take    *string       `json:"take"`
	MidTurn bool          `json:"mid_turn"`
	End     bool          `json:"end"`
	Loadout *data.Loadout `json:"loadout"`
	// By is the action that paid for a loadout change (draw, sheathe, drop, or swap, GD_S19_03).
	By *string `json:"by"`
}

// Input is the `stance` command's input.
type Input struct {
	// Perks are the actor's perks as core:feature refs; their unlocks name the stances.
	Perks   []string     `json:"perks"`
	Loadout data.Loadout `json:"loadout"`
	// Stance is the one active before the first event, if any.
	Stance *string `json:"stance"`
	Events []Event `json:"events"`
}

// Step is one event's outcome: the stance after it, the AP it cost, and the stance it ended.
type Step struct {
	AP     int64   `json:"ap"`
	Ended  *string `json:"ended"`
	Stance *string `json:"stance"`
}

// Result is the stance at the end, the AP spent, its modifiers, and each step.
type Result struct {
	Stance    *string             `json:"stance"`
	AP        int64               `json:"ap"`
	Modifiers []modifier.Modifier `json:"modifiers"`
	Steps     []Step              `json:"steps"`
}

// Run is the `stance` command.
func Run(ruleset *data.Ruleset, input Input) (Result, error) {
	var rows *data.Stances = &ruleset.Stances
	var unlocked []string
	var perkModifiers []modifier.Modifier
	for _, ref := range input.Perks {
		key, isCore := strings.CutPrefix(ref, "core:feature/")
		var feature *data.Feature
		if isCore {
			feature, _ = ruleset.Features.Get(key)
		}
		if feature == nil {
			return Result{}, &data.NotDefinedError{Kind: "feature", Key: ref}
		}
		unlocked = append(unlocked, feature.Unlocks...)
		encoded, err := json.Marshal(feature.EmitAt(1))
		if err != nil {
			return Result{}, err
		}
		var emitted []modifier.Modifier
		if err := json.Unmarshal(encoded, &emitted); err != nil {
			return Result{}, err
		}
		perkModifiers = append(perkModifiers, emitted...)
	}
	var find func(string) *data.Stance = func(key string) *data.Stance {
		for index := range rows.Rows {
			if rows.Rows[index].Key == key {
				return &rows.Rows[index]
			}
		}
		return nil
	}
	var check func(string, data.Loadout) error = func(key string, loadout data.Loadout) error {
		var row *data.Stance = find(key)
		if row == nil {
			return &data.NotDefinedError{Kind: "stance", Key: key}
		}
		if !slices.Contains(unlocked, "core:stance/"+key) {
			return &data.MissingInputError{Subject: fmt.Sprintf("stance %q", key), Needs: "a perk that unlocks it"}
		}
		return row.Misfit(loadout)
	}
	var loadout data.Loadout = input.Loadout
	var held *string = input.Stance
	if held != nil {
		if err := check(*held, loadout); err != nil {
			return Result{}, err
		}
	}
	var result Result = Result{Steps: []Step{}, Modifiers: []modifier.Modifier{}}
	for _, event := range input.Events {
		var step Step
		var set int
		for _, given := range []bool{event.Take != nil, event.End, event.Loadout != nil} {
			if given {
				set++
			}
		}
		if set > 1 {
			return Result{}, errors.New("a stance event takes one of take, end, or loadout")
		}
		switch {
		case event.Take != nil:
			if err := check(*event.Take, loadout); err != nil {
				return Result{}, err
			}
			if held == nil || *held != *event.Take {
				step.AP = rows.Rules.TurnStartAP
				if event.MidTurn {
					step.AP = rows.Rules.MidTurnAP
				}
				step.Ended, held = held, event.Take
			}
		case event.End:
			step.Ended, held = held, nil
		case event.Loadout != nil:
			if event.By != nil {
				action, err := ruleset.Actions.Actions.Require("action", *event.By)
				if err != nil {
					return Result{}, err
				}
				if !action.ChangesLoadout {
					return Result{}, fmt.Errorf("action %q does not change the loadout", *event.By)
				}
				cost, err := modifier.Value(action.Cost, modifier.Target(perkModifiers, *event.By), ruleset.Origins, 0, nil)
				if err != nil {
					return Result{}, err
				}
				step.AP = max(cost, 0)
			}
			loadout = *event.Loadout
			if held != nil && find(*held).Misfit(loadout) != nil {
				step.Ended, held = held, nil
			}
		default:
			return Result{}, errors.New("a stance event needs take, end, or loadout")
		}
		result.AP += step.AP
		step.Stance = held
		result.Steps = append(result.Steps, step)
	}
	result.Stance = held
	if held != nil {
		result.Modifiers = slices.Clone(find(*held).Modifiers)
	}
	return result, nil
}
