// Package stored is the actor a session saves (docs/core/state.md#the-stored-actor): the
// primary facts only. Maximum pools and derived stats are computed from it.
package stored

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"

	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/derived"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
)

// Pools is the growth from level-up rolls.
type Pools struct {
	HitPoints  int64 `json:"hp"`
	ManaPoints int64 `json:"mp"`
	Stamina    int64 `json:"sp"`
}

// Resources are the current pools.
type Resources struct {
	HitPoints    int64 `json:"hp"`
	ManaPoints   int64 `json:"mp"`
	Stamina      int64 `json:"sp"`
	ActionPoints int64 `json:"ap"`
}

// Condition is an active condition.
type Condition struct {
	Key    string `json:"key"`
	Stacks int64  `json:"stacks"`
}

// ActiveModifier is a modifier and how long it has run.
type ActiveModifier struct {
	Modifier     modifier.Modifier `json:"modifier"`
	TurnsElapsed int64             `json:"turns_elapsed"`
}

// ItemInstance is one item the actor holds, pointing at a row.
type ItemInstance struct {
	ID           int64             `json:"id"`
	Base         string            `json:"base"`
	Material     string            `json:"material,omitempty"`
	Size         string            `json:"size,omitempty"`
	AddedEffects []json.RawMessage `json:"added_effects"`
	Quantity     int64             `json:"quantity"`
}

// Actor is the stored actor.
type Actor struct {
	Attributes map[string]int64 `json:"attributes"`
	Level      int64            `json:"level"`
	XP         int64            `json:"xp"`
	Skills     map[string]int64 `json:"skills"`
	Tags       []string         `json:"tags"`
	Growth     Pools            `json:"growth"`
	Resources  Resources        `json:"resources"`
	Conditions []Condition      `json:"conditions"`
	Modifiers  []ActiveModifier `json:"modifiers"`
	Flags      []string         `json:"flags"`
	// Perks, known spells and effects, and coins (copper) are left out when empty.
	Perks        map[string]int64 `json:"perks,omitempty"`
	Spells       []string         `json:"spells,omitempty"`
	KnownEffects []string         `json:"known_effects,omitempty"`
	Coins        int64            `json:"coins,omitempty"`
	Inventory    []ItemInstance   `json:"inventory"`
	Equipped     map[string]int64 `json:"equipped"`
	// Templates are the character templates held, and how each came (docs/core/templates.md);
	// left out when none.
	Templates []HeldTemplate `json:"templates,omitempty"`
}

// HeldTemplate is a held character template: its ref, and how it came (granted or outside_rules).
type HeldTemplate struct {
	Template string `json:"template"`
	How      string `json:"how"`
}

// Input is the `stored_actor` command's input: an actor to check, or a new one to make.
type Input struct {
	Actor *Actor `json:"actor"`
	New   *New   `json:"new"`
}

// New is what a new stored actor starts from (docs/core/state.md#the-stored-actor, BE_S23_03).
type New struct {
	Attributes map[string]int64 `json:"attributes"`
	Level      int64            `json:"level"`
	Tags       []string         `json:"tags"`
}

// Result is the actor, unchanged.
type Result struct {
	Actor Actor `json:"actor"`
}

// Run is the `stored_actor` command: the actor comes back unchanged, if it is consistent; with
// new, a new one is made.
func Run(ruleset *data.Ruleset, input Input) (Result, error) {
	switch {
	case input.New != nil && input.Actor != nil:
		return Result{}, errors.New("stored_actor takes actor or new, not both")
	case input.New != nil:
		made, err := newActor(ruleset, *input.New)
		return Result{Actor: made}, err
	case input.Actor == nil:
		return Result{}, &data.MissingInputError{Subject: "stored_actor", Needs: "actor or new"}
	}
	return check(*input.Actor)
}

// newActor is a new stored actor: that level (default 1), xp 0, nothing held, and every pool full
// at the maximum the derived stats give.
func newActor(ruleset *data.Ruleset, start New) (Actor, error) {
	var level int64 = max(start.Level, 1)
	encoded, err := json.Marshal(map[string]any{"actor": map[string]any{"attributes": start.Attributes, "level": level}})
	if err != nil {
		return Actor{}, err
	}
	var input derived.Input
	if err := json.Unmarshal(encoded, &input); err != nil {
		return Actor{}, err
	}
	maxima, err := derived.Run(ruleset, input)
	if err != nil {
		return Actor{}, err
	}
	var tags []string = append([]string{}, start.Tags...)
	return Actor{
		Attributes: start.Attributes, Level: level, Skills: map[string]int64{}, Tags: tags,
		Resources:  Resources{HitPoints: maxima["hp"], ManaPoints: maxima["mp"], Stamina: maxima["sp"], ActionPoints: maxima["ap"]},
		Conditions: []Condition{}, Modifiers: []ActiveModifier{}, Flags: []string{}, Inventory: []ItemInstance{}, Equipped: map[string]int64{},
	}, nil
}

// check is the actor unchanged, if its inventory ids are unique and equipped names only them.
func check(actor Actor) (Result, error) {
	var seen []int64
	for _, item := range actor.Inventory {
		if slices.Contains(seen, item.ID) {
			return Result{}, fmt.Errorf("inventory: id %d is used twice", item.ID)
		}
		seen = append(seen, item.ID)
	}
	var slots []string
	for slot := range actor.Equipped {
		slots = append(slots, slot)
	}
	sort.Strings(slots)
	for _, slot := range slots {
		if !slices.Contains(seen, actor.Equipped[slot]) {
			return Result{}, fmt.Errorf("equipped %s: no inventory item %d", slot, actor.Equipped[slot])
		}
	}
	return Result{Actor: actor}, nil
}
