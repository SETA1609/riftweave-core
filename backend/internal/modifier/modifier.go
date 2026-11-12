// Package modifier is the kernel's one stacking function (docs/core/modifier.md).
//
// The order is fixed: drop expired and removed modifiers; same origin, source, target, and
// op keeps the strongest; a `highest` origin keeps its strongest bonus and penalty per
// target and op; then add, mul (floored), set. grant makes a flag and resist reduces a
// packet. Nothing here reads the ruleset: the caller passes the origin policies, so the
// package is a pure function over a modifier list.
package modifier

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
)

// Operation is what a modifier does to its target.
type Operation uint8

const (
	Add Operation = iota
	Multiply
	Set
	Grant
	Resist
)

var operationNames = [...]string{"add", "mul", "set", "grant", "resist"}

// shortOperationNames are the names a vector still uses until M5-46's vector part renames
// them; the data spells them out.
var shortOperationNames = map[string]Operation{"multiply": Multiply}

// UnmarshalText reads an operation name.
func (operation *Operation) UnmarshalText(encoded []byte) error {
	for index, name := range operationNames {
		if name == string(encoded) {
			*operation = Operation(index)
			return nil
		}
	}
	if spelled, ok := shortOperationNames[string(encoded)]; ok {
		*operation = spelled
		return nil
	}
	return fmt.Errorf("modifier operation %q is not defined", encoded)
}

// MarshalText writes the operation name, so a stored modifier round-trips.
func (operation Operation) MarshalText() ([]byte, error) { return []byte(operation.String()), nil }

func (operation Operation) String() string { return operationNames[operation] }

// Modifier is one opcode on one target (ruleset/schemas/schema.json, "modifier").
type Modifier struct {
	Target    string    `json:"target"`
	Operation Operation `json:"op"`
	Value     float64   `json:"value"`
	Source    string    `json:"source"`
	Origin    string    `json:"origin,omitempty"`
	Duration  string    `json:"duration,omitempty"`
	Turns     int64     `json:"turns,omitempty"`
	Stack     string    `json:"stack,omitempty"`
}

// UnmarshalJSON reads `operation`, as the data writes it, or the short `op` a vector still
// uses until M5-46's vector part renames it.
func (modifier *Modifier) UnmarshalJSON(encoded []byte) error {
	type plain Modifier
	var read struct {
		plain
		SpelledOperation *Operation `json:"operation"`
	}
	if err := json.Unmarshal(encoded, &read); err != nil {
		return err
	}
	*modifier = Modifier(read.plain)
	if read.SpelledOperation != nil {
		modifier.Operation = *read.SpelledOperation
	}
	return nil
}

// OriginPolicies maps an origin ref to whether its policy is `highest`. An origin that is
// not in the map has no policy, which is an error.
type OriginPolicies map[string]bool

// Live drops the modifiers of removed sources and the `turns` modifiers that have elapsed.
func Live(modifiers []Modifier, elapsedTurns int64, removedSources []string) []Modifier {
	var live []Modifier = make([]Modifier, 0, len(modifiers))
	for _, candidate := range modifiers {
		if slices.Contains(removedSources, candidate.Source) {
			continue
		}
		if candidate.Duration == "turns" && elapsedTurns >= candidate.Turns {
			continue
		}
		live = append(live, candidate)
	}
	return live
}

type groupKey struct {
	origin, source, target string
	operation              Operation
}

// keepStrongest groups the modifiers at indexes by key and keeps the strongest of each
// group, returning the survivors' indexes in list order.
func keepStrongest(modifiers []Modifier, indexes []int, keyOf func(*Modifier) groupKey) []int {
	var groups map[groupKey][]int = map[groupKey][]int{}
	var order []groupKey
	for _, index := range indexes {
		var key groupKey = keyOf(&modifiers[index])
		if _, seen := groups[key]; !seen {
			order = append(order, key)
		}
		groups[key] = append(groups[key], index)
	}
	var kept []int
	for _, key := range order {
		kept = append(kept, strongest(modifiers, groups[key])...)
	}
	slices.Sort(kept)
	return kept
}

// strongest is the strongest bonus and the strongest penalty of one group (one operation).
// `set` is not ranked: the latest wins. The neutral value is 1 for mul and 0 for the others.
func strongest(modifiers []Modifier, group []int) []int {
	if modifiers[group[0]].Operation == Set {
		return group[len(group)-1:]
	}
	var neutral float64 = 0.0
	if modifiers[group[0]].Operation == Multiply {
		neutral = 1
	}
	var bestIndex, worstIndex int = -1, -1
	for _, index := range group {
		var value float64 = modifiers[index].Value
		if value > neutral && (bestIndex < 0 || value > modifiers[bestIndex].Value) {
			bestIndex = index
		}
		if value < neutral && (worstIndex < 0 || value < modifiers[worstIndex].Value) {
			worstIndex = index
		}
	}
	var kept []int
	if bestIndex >= 0 {
		kept = append(kept, bestIndex)
	}
	if worstIndex >= 0 {
		kept = append(kept, worstIndex)
	}
	return kept
}

// Deduplicate is steps 2 and 3: the same origin, source, target, and operation keeps the
// strongest; then a `highest` origin keeps its strongest bonus and penalty per target and
// operation.
func Deduplicate(modifiers []Modifier, policies OriginPolicies) ([]Modifier, error) {
	var allIndexes []int = make([]int, len(modifiers))
	for index := range allIndexes {
		allIndexes[index] = index
	}
	var kept []int = keepStrongest(modifiers, allIndexes, func(candidate *Modifier) groupKey {
		return groupKey{candidate.Origin, candidate.Source, candidate.Target, candidate.Operation}
	})
	var otherIndexes, highestIndexes []int
	for _, index := range kept {
		if origin := modifiers[index].Origin; origin != "" {
			isHighest, ok := policies[origin]
			if !ok {
				return nil, fmt.Errorf("origin %q has no stacking policy", origin)
			}
			if isHighest {
				highestIndexes = append(highestIndexes, index)
				continue
			}
		}
		otherIndexes = append(otherIndexes, index)
	}
	highestIndexes = keepStrongest(modifiers, highestIndexes, func(candidate *Modifier) groupKey {
		return groupKey{candidate.Origin, "", candidate.Target, candidate.Operation}
	})
	var order []int = slices.Concat(otherIndexes, highestIndexes)
	slices.Sort(order)
	var deduplicated []Modifier = make([]Modifier, len(order))
	for position, index := range order {
		deduplicated[position] = modifiers[index]
	}
	return deduplicated, nil
}

// Value is steps 1 to 6 for one numeric target: expire, deduplicate, then add, mul
// (floored), set. The caller passes the modifiers of one target; grant and resist carry no
// number.
func Value(base int64, modifiers []Modifier, policies OriginPolicies, elapsedTurns int64, removedSources []string) (int64, error) {
	deduplicated, err := Deduplicate(Live(modifiers, elapsedTurns, removedSources), policies)
	if err != nil {
		return 0, err
	}
	var value float64 = float64(base)
	var factor float64 = 1.0
	var setValue float64 = 0.0
	var hasSet bool = false
	for _, current := range deduplicated {
		switch current.Operation {
		case Add:
			value += current.Value
		case Multiply:
			factor *= current.Value
		case Set:
			setValue, hasSet = current.Value, true
		}
	}
	if factor != 1 {
		value = math.Floor(value * factor)
	}
	if hasSet {
		value = setValue
	}
	return int64(math.Floor(value)), nil
}

// Flags are the targets that at least one live grant names, sorted.
func Flags(modifiers []Modifier, elapsedTurns int64, removedSources []string) []string {
	var flags []string = []string{}
	for _, current := range Live(modifiers, elapsedTurns, removedSources) {
		if current.Operation == Grant && !slices.Contains(flags, current.Target) {
			flags = append(flags, current.Target)
		}
	}
	slices.Sort(flags)
	return flags
}

// Packet is a packet's amount after every live resist on its tag, never below zero.
func Packet(amount int64, tag string, modifiers []Modifier, policies OriginPolicies, elapsedTurns int64, removedSources []string) (int64, error) {
	deduplicated, err := Deduplicate(Live(modifiers, elapsedTurns, removedSources), policies)
	if err != nil {
		return 0, err
	}
	var totalResist float64 = 0.0
	for _, current := range deduplicated {
		if current.Operation == Resist && current.Target == tag {
			totalResist += current.Value
		}
	}
	return max(int64(math.Floor(float64(amount)-totalResist)), 0), nil
}

// Target filters the modifiers that name one target.
func Target(modifiers []Modifier, target string) []Modifier {
	var matching []Modifier
	for _, current := range modifiers {
		if current.Target == target {
			matching = append(matching, current)
		}
	}
	return matching
}
