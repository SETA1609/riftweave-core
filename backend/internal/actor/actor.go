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
	Str Ability = iota
	Per
	End
	Int
	Wil
	Agi
	Cha
	Lck
	Count
)

var keys = [Count]string{"str", "per", "end", "int", "wil", "agi", "cha", "lck"}

// ByKey returns the ability with the given key ("str", "lck", ...).
func ByKey(key string) (Ability, bool) {
	for i, k := range keys {
		if k == key {
			return Ability(i), true
		}
	}
	return 0, false
}

func (a Ability) String() string { return keys[a] }

// Attrs are an actor's attribute values.
type Attrs struct {
	v   [Count]int64
	set uint8
}

// Get returns the attribute, or 0 when it was not given.
func (a Attrs) Get(ab Ability) int64 { return a.v[ab] }

// Has reports whether the attribute was given.
func (a Attrs) Has(ab Ability) bool { return a.set&(1<<ab) != 0 }

// With returns a copy with one attribute set.
func (a Attrs) With(ab Ability, value int64) Attrs {
	a.v[ab] = value
	a.set |= 1 << ab
	return a
}

// UnmarshalJSON reads an object keyed by ability and rejects any other key.
func (a *Attrs) UnmarshalJSON(b []byte) error {
	var m map[string]int64
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	*a = Attrs{}
	for key, value := range m {
		ab, ok := ByKey(key)
		if !ok {
			return fmt.Errorf("%q is not an attribute", key)
		}
		*a = a.With(ab, value)
	}
	return nil
}

// Actor is the shape of the `actor` input (backend/contract/common.json).
type Actor struct {
	Attributes Attrs            `json:"attributes"`
	Skills     map[string]int64 `json:"skills"`
	Level      *int64           `json:"level"`
}

// LevelOr returns the level, or def when none was given.
func (a Actor) LevelOr(def int64) int64 {
	if a.Level == nil {
		return def
	}
	return *a.Level
}
