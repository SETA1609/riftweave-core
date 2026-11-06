// Package actor holds the inputs every command shares: the eight attributes and an actor.
//
// Attributes are a fixed array indexed by Ability, not a map, so a formula reads one with
// no hashing and no allocation. A bit mask remembers which were given, because a missing
// attribute reads as 0 in most commands but is an error in a formula.
package actor

import (
	"encoding/json"
	"fmt"
)

// Ability is one of the eight core attributes (ruleset/schemas/schema.json, "ability").
type Ability uint8

const (
	Strength Ability = iota
	Perception
	Endurance
	Intelligence
	Willpower
	Agility
	Charisma
	Luck
	AbilityCount
)

var abilityKeys = [AbilityCount]string{"str", "per", "end", "int", "wil", "agi", "cha", "lck"}

// ByKey returns the ability with the given key ("str", "lck", ...).
func ByKey(key string) (Ability, bool) {
	for index, candidateKey := range abilityKeys {
		if candidateKey == key {
			return Ability(index), true
		}
	}
	return 0, false
}

func (ability Ability) String() string { return abilityKeys[ability] }

// Attributes are an actor's attribute values.
type Attributes struct {
	values    [AbilityCount]int64
	givenMask uint8
}

// Get returns the attribute, or 0 when it was not given.
func (attributes Attributes) Get(ability Ability) int64 { return attributes.values[ability] }

// Has reports whether the attribute was given.
func (attributes Attributes) Has(ability Ability) bool {
	return attributes.givenMask&(1<<ability) != 0
}

// With returns a copy with one attribute set.
func (attributes Attributes) With(ability Ability, value int64) Attributes {
	attributes.values[ability] = value
	attributes.givenMask |= 1 << ability
	return attributes
}

// UnmarshalJSON reads an object keyed by ability and rejects any other key.
func (attributes *Attributes) UnmarshalJSON(encoded []byte) error {
	var valuesByKey map[string]int64
	if err := json.Unmarshal(encoded, &valuesByKey); err != nil {
		return err
	}
	*attributes = Attributes{}
	for key, value := range valuesByKey {
		ability, ok := ByKey(key)
		if !ok {
			return fmt.Errorf("%q is not an attribute", key)
		}
		*attributes = attributes.With(ability, value)
	}
	return nil
}

// Actor is the shape of the `actor` input (backend/contract/common.json).
type Actor struct {
	Attributes Attributes       `json:"attributes"`
	Skills     map[string]int64 `json:"skills"`
	Level      *int64           `json:"level"`
}

// LevelOrDefault returns the level, or defaultLevel when none was given.
func (actor Actor) LevelOrDefault(defaultLevel int64) int64 {
	if actor.Level == nil {
		return defaultLevel
	}
	return *actor.Level
}
