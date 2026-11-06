// Package modifier is the kernel's one stacking function (docs/core/modifier.md).
//
// The order is fixed: drop expired and removed modifiers; same origin, source, target, and
// op keeps the strongest; a `highest` origin keeps its strongest bonus and penalty per
// target and op; then add, mul (floored), set. grant makes a flag and resist reduces a
// packet. Nothing here reads the ruleset: the caller passes the origin policies, so the
// package is a pure function over a modifier list.
package modifier

import (
	"fmt"
	"math"
	"slices"
)

// Op is what a modifier does to its target.
type Op uint8

const (
	Add Op = iota
	Mul
	Set
	Grant
	Resist
)

var opNames = [...]string{"add", "mul", "set", "grant", "resist"}

// UnmarshalText reads an op name.
func (o *Op) UnmarshalText(b []byte) error {
	for i, name := range opNames {
		if name == string(b) {
			*o = Op(i)
			return nil
		}
	}
	return fmt.Errorf("modifier op %q is not defined", b)
}

func (o Op) String() string { return opNames[o] }

// Modifier is one opcode on one target (ruleset/schemas/schema.json, "modifier").
type Modifier struct {
	Target   string  `json:"target"`
	Op       Op      `json:"op"`
	Value    float64 `json:"value"`
	Source   string  `json:"source"`
	Origin   string  `json:"origin,omitempty"`
	Duration string  `json:"duration,omitempty"`
	Turns    int64   `json:"turns,omitempty"`
	Stack    string  `json:"stack,omitempty"`
}

// Highest maps an origin ref to whether its policy is `highest`. An origin that is not in
// the map has no policy, which is an error.
type Highest map[string]bool

// Live drops the modifiers of removed sources and the `turns` modifiers that have elapsed.
func Live(mods []Modifier, elapsedTurns int64, removed []string) []Modifier {
	live := make([]Modifier, 0, len(mods))
	for _, m := range mods {
		if slices.Contains(removed, m.Source) {
			continue
		}
		if m.Duration == "turns" && elapsedTurns >= m.Turns {
			continue
		}
		live = append(live, m)
	}
	return live
}

type groupKey struct {
	origin, source, target string
	op                     Op
}

// keepStrongest groups the modifiers at idx by key and keeps the strongest of each group,
// returning the survivors' indexes in list order.
func keepStrongest(mods []Modifier, idx []int, key func(*Modifier) groupKey) []int {
	groups := map[groupKey][]int{}
	var order []groupKey
	for _, i := range idx {
		k := key(&mods[i])
		if _, seen := groups[k]; !seen {
			order = append(order, k)
		}
		groups[k] = append(groups[k], i)
	}
	var kept []int
	for _, k := range order {
		kept = append(kept, strongest(mods, groups[k])...)
	}
	slices.Sort(kept)
	return kept
}

// strongest is the strongest bonus and the strongest penalty of one group (one op). `set`
// is not ranked: the latest wins. The neutral value is 1 for mul and 0 for the others.
func strongest(mods []Modifier, group []int) []int {
	if mods[group[0]].Op == Set {
		return group[len(group)-1:]
	}
	neutral := 0.0
	if mods[group[0]].Op == Mul {
		neutral = 1
	}
	best, worst := -1, -1
	for _, i := range group {
		v := mods[i].Value
		if v > neutral && (best < 0 || v > mods[best].Value) {
			best = i
		}
		if v < neutral && (worst < 0 || v < mods[worst].Value) {
			worst = i
		}
	}
	var kept []int
	if best >= 0 {
		kept = append(kept, best)
	}
	if worst >= 0 {
		kept = append(kept, worst)
	}
	return kept
}

// Dedupe is steps 2 and 3: the same origin, source, target, and op keeps the strongest;
// then a `highest` origin keeps its strongest bonus and penalty per target and op.
func Dedupe(mods []Modifier, policy Highest) ([]Modifier, error) {
	all := make([]int, len(mods))
	for i := range all {
		all[i] = i
	}
	kept := keepStrongest(mods, all, func(m *Modifier) groupKey { return groupKey{m.Origin, m.Source, m.Target, m.Op} })
	var rest, highest []int
	for _, i := range kept {
		if origin := mods[i].Origin; origin != "" {
			isHighest, ok := policy[origin]
			if !ok {
				return nil, fmt.Errorf("origin %q has no stacking policy", origin)
			}
			if isHighest {
				highest = append(highest, i)
				continue
			}
		}
		rest = append(rest, i)
	}
	highest = keepStrongest(mods, highest, func(m *Modifier) groupKey { return groupKey{m.Origin, "", m.Target, m.Op} })
	order := slices.Concat(rest, highest)
	slices.Sort(order)
	out := make([]Modifier, len(order))
	for n, i := range order {
		out[n] = mods[i]
	}
	return out, nil
}

// Value is steps 1 to 6 for one numeric target: expire, dedupe, then add, mul (floored),
// set. The caller passes the modifiers of one target; grant and resist carry no number.
func Value(base int64, mods []Modifier, policy Highest, elapsedTurns int64, removed []string) (int64, error) {
	deduped, err := Dedupe(Live(mods, elapsedTurns, removed), policy)
	if err != nil {
		return 0, err
	}
	value, factor, set, hasSet := float64(base), 1.0, 0.0, false
	for _, m := range deduped {
		switch m.Op {
		case Add:
			value += m.Value
		case Mul:
			factor *= m.Value
		case Set:
			set, hasSet = m.Value, true
		}
	}
	if factor != 1 {
		value = math.Floor(value * factor)
	}
	if hasSet {
		value = set
	}
	return int64(math.Floor(value)), nil
}

// Flags are the targets that at least one live grant names, sorted.
func Flags(mods []Modifier, elapsedTurns int64, removed []string) []string {
	flags := []string{}
	for _, m := range Live(mods, elapsedTurns, removed) {
		if m.Op == Grant && !slices.Contains(flags, m.Target) {
			flags = append(flags, m.Target)
		}
	}
	slices.Sort(flags)
	return flags
}

// Packet is a packet's amount after every live resist on its tag, never below zero.
func Packet(amount int64, tag string, mods []Modifier, policy Highest, elapsedTurns int64, removed []string) (int64, error) {
	deduped, err := Dedupe(Live(mods, elapsedTurns, removed), policy)
	if err != nil {
		return 0, err
	}
	total := 0.0
	for _, m := range deduped {
		if m.Op == Resist && m.Target == tag {
			total += m.Value
		}
	}
	return max(int64(math.Floor(float64(amount)-total)), 0), nil
}

// Target filters the modifiers that name one target.
func Target(mods []Modifier, target string) []Modifier {
	var out []Modifier
	for _, m := range mods {
		if m.Target == target {
			out = append(out, m)
		}
	}
	return out
}
