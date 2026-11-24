// Package overwatch holds a shot or a zone under fire until the shooter's next turn: Pinning
// Fire and Trigger Watch (docs/core/actions.md#ranged-maneuvers, GD_S17_06). It says which
// events draw a strike and what is spent; the caller resolves each strike with combat.
package overwatch

import (
	"fmt"
	"slices"

	"github.com/SETA1609/riftweave-core/backend/internal/data"
)

// EventKind is what a foe did.
type EventKind string

// The kinds of event the table reports.
const (
	EventMove   EventKind = "move"
	EventAttack EventKind = "attack"
)

// handGun is the loadout kind of a pistol or a hand crossbow: two of them hold a zone though they load.
const handGun = "hand_gun"

// Shooter is who holds the shot.
type Shooter struct {
	// Perks are the held perks at their ranks; nil skips the unlock check.
	Perks []data.HeldRef `json:"perks"`
	// Maneuvers are a monster sheet's maneuvers, made with no perk.
	Maneuvers        []string `json:"maneuvers"`
	WeaponProperties []string `json:"weapon_properties"`
}

// Event is one thing the table saw until the shooter's next turn.
type Event struct {
	Foe     int64     `json:"foe"`
	Kind    EventKind `json:"kind"`
	InZone  bool      `json:"in_zone"`
	InRange bool      `json:"in_range"`
}

// Input is the `overwatch` command's input.
type Input struct {
	Maneuver string        `json:"maneuver"`
	Shooter  Shooter       `json:"shooter"`
	Loadout  *data.Loadout `json:"loadout"`
	// AP is the AP left (Pinning Fire) or held (Trigger Watch).
	AP int64 `json:"ap"`
	// Rounds is the rounds carried; nil, rounds are not counted.
	Rounds *int64  `json:"rounds"`
	Events []Event `json:"events"`
}

// Strike is an event that draws a strike, or a foe's attack from the zone, at its attack add.
type Strike struct {
	Event  int64 `json:"event"`
	Foe    int64 `json:"foe"`
	Attack int64 `json:"attack"`
}

// Result is what the maneuver spent and the strikes it drew.
type Result struct {
	AP          int64    `json:"ap"`
	APReturned  int64    `json:"ap_returned"`
	Rounds      int64    `json:"rounds"`
	RoundsLeft  *int64   `json:"rounds_left,omitempty"`
	Shots       []Strike `json:"shots"`
	ZoneAttacks []Strike `json:"zone_attacks"`
}

// Run is the `overwatch` command.
func Run(ruleset *data.Ruleset, input Input) (Result, error) {
	maneuver, err := ruleset.Maneuvers.Require("maneuver", input.Maneuver)
	if err != nil {
		return Result{}, err
	}
	if err := ruleset.CheckManeuver(maneuver, data.ManeuverOverwatch, input.Shooter.Perks, input.Shooter.Maneuvers); err != nil {
		return Result{}, err
	}
	var attackAP int64
	attack, err := ruleset.Actions.Actions.Require("action", "attack")
	if err != nil {
		return Result{}, err
	}
	attackAP = attack.Cost
	var twoHandGuns bool
	if input.Loadout != nil {
		sequence := ruleset.AttackSequences.Covering(input.Loadout.Main, input.Loadout.Off)
		if sequence == nil {
			return Result{}, &data.NoSequenceError{Loadout: *input.Loadout}
		}
		attackAP = sequence.AP
		twoHandGuns = input.Loadout.Main == handGun && input.Loadout.Off == handGun
	}
	for _, refused := range maneuver.WithoutWeaponProperties {
		if slices.Contains(input.Shooter.WeaponProperties, refused) && !twoHandGuns {
			return Result{}, fmt.Errorf("%s is refused with a %s weapon", maneuver.Key, refused)
		}
	}
	var cost int64 = max(attackAP+maneuver.AP, 1)
	var rounds int64 = maneuver.RoundsSpent()
	if input.Rounds != nil && *input.Rounds < rounds {
		return Result{}, fmt.Errorf("%s needs %d rounds", maneuver.Key, rounds)
	}
	var result Result = Result{Shots: []Strike{}, ZoneAttacks: []Strike{}}
	if maneuver.Hold {
		if input.AP < cost {
			return Result{}, fmt.Errorf("%s needs %d held AP", maneuver.Key, cost)
		}
		result.APReturned = cost
		for index, event := range input.Events {
			if event.InRange {
				result.Shots = append(result.Shots, Strike{Event: int64(index), Foe: event.Foe, Attack: maneuver.Attack})
				result.AP, result.APReturned, result.Rounds = cost, 0, rounds
				break
			}
		}
	} else {
		if input.AP < cost {
			return Result{}, fmt.Errorf("%s needs %d AP", maneuver.Key, cost)
		}
		result.AP, result.Rounds = input.AP, rounds
		var struck map[int64]bool = map[int64]bool{}
		for index, event := range input.Events {
			if !event.InZone {
				continue
			}
			switch event.Kind {
			case EventAttack:
				if maneuver.Zone != nil {
					result.ZoneAttacks = append(result.ZoneAttacks, Strike{Event: int64(index), Foe: event.Foe, Attack: maneuver.Zone.Attack})
				}
			case EventMove:
				if !struck[event.Foe] && int64(len(result.Shots)) < rounds {
					struck[event.Foe] = true
					result.Shots = append(result.Shots, Strike{Event: int64(index), Foe: event.Foe, Attack: maneuver.Attack})
				}
			}
		}
	}
	if input.Rounds != nil {
		var left int64 = *input.Rounds - result.Rounds
		result.RoundsLeft = &left
	}
	return result, nil
}
