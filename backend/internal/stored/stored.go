// Package stored is the actor a session saves (docs/core/state.md#the-stored-actor): the
// primary facts only. Maximum pools and derived stats are computed from it.
package stored

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"

	"github.com/SETA1609/riftweave-core/backend/internal/data"
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
	ID       int64             `json:"id"`
	Base     string            `json:"base"`
	Material string            `json:"material,omitempty"`
	Size     string            `json:"size,omitempty"`
	Mods     []json.RawMessage `json:"mods"`
	Quantity int64             `json:"quantity"`
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
}

// Input is the `stored_actor` command's input.
type Input struct {
	Actor Actor `json:"actor"`
}

// Result is the actor, unchanged.
type Result struct {
	Actor Actor `json:"actor"`
}

// Run is the `stored_actor` command: the actor comes back unchanged, if it is consistent.
func Run(_ *data.Ruleset, input Input) (Result, error) {
	var seen []int64
	for _, item := range input.Actor.Inventory {
		if slices.Contains(seen, item.ID) {
			return Result{}, fmt.Errorf("inventory: id %d is used twice", item.ID)
		}
		seen = append(seen, item.ID)
	}
	var slots []string
	for slot := range input.Actor.Equipped {
		slots = append(slots, slot)
	}
	sort.Strings(slots)
	for _, slot := range slots {
		if !slices.Contains(seen, input.Actor.Equipped[slot]) {
			return Result{}, fmt.Errorf("equipped %s: no inventory item %d", slot, input.Actor.Equipped[slot])
		}
	}
	return Result{Actor: input.Actor}, nil
}
