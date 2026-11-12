// Package consume uses items and lets time pass (docs/core/modifier.md#effects-over-time).
// A pool restore is instant or ticks each second; any other effect is held for its duration;
// a `replace` origin removes every running effect of that origin first.
package consume

import (
	"slices"
	"strings"

	"github.com/SETA1609/riftweave-core/backend/internal/data"
)

// Pools are the resources an item can restore.
type Pools struct {
	HitPoints  int64 `json:"hp"`
	ManaPoints int64 `json:"mp"`
	Stamina    int64 `json:"stamina"`
}

// Event is one step: use an item, or advance by seconds or by turns.
type Event struct {
	Item     string `json:"item"`
	InCombat bool   `json:"in_combat"`
	Advance  int64  `json:"advance"`
	Turns    int64  `json:"turns"`
}

// Current is the pools before the events and the stored toxicity (0 when not named).
type Current struct {
	Pools
	Toxicity int64 `json:"toxicity"`
}

// Maximum is each pool's limit and the toxicity limit (nil when not named: no overdose).
type Maximum struct {
	Pools
	Toxicity *int64 `json:"toxicity"`
}

// Input is the `consume` command's input.
type Input struct {
	Current Current `json:"current"`
	Maximum Maximum `json:"max"`
	Events  []Event `json:"events"`
}

// overdosed is the condition held while toxicity is above the limit.
const overdosed = "core:condition/overdosed"

// Tick is a running pool effect.
type Tick struct {
	Source      string `json:"source"`
	Origin      string `json:"origin"`
	Pool        string `json:"pool"`
	Magnitude   int64  `json:"magnitude"`
	SecondsLeft int64  `json:"seconds_left"`
}

// Held is a running stat effect.
type Held struct {
	Source      string `json:"source"`
	Origin      string `json:"origin"`
	Target      string `json:"target"`
	Operation   string `json:"op"`
	Value       int64  `json:"value"`
	SecondsLeft int64  `json:"seconds_left"`
}

// Result is the pools after the events and what is still running.
type Result struct {
	Pools
	Ticks      []Tick   `json:"ticks"`
	Conditions []string `json:"conditions"` // overdosed while toxicity is above the limit
	Toxicity   int64    `json:"toxicity"`
	Modifiers  []Held   `json:"modifiers"`
	Refused    []string `json:"refused"`
}

var poolOfParameter = map[string]string{"health": "hp", "mana": "mp", "stamina": "stamina"}

// state is the running result and the limits.
type state struct {
	result  Result
	maximum Pools
}

func pointer(pools *Pools, name string) *int64 {
	switch name {
	case "hp":
		return &pools.HitPoints
	case "mp":
		return &pools.ManaPoints
	}
	return &pools.Stamina
}

func (running *state) add(pool string, amount int64) {
	var value *int64 = pointer(&running.result.Pools, pool)
	*value = min(*value+amount, *pointer(&running.maximum, pool))
}

func (running *state) advance(ruleset *data.Ruleset, seconds int64) {
	running.result.Toxicity = max(0, running.result.Toxicity-ruleset.ToxicityDecayPerSecond*seconds)
	for index := range running.result.Ticks {
		var tick *Tick = &running.result.Ticks[index]
		running.add(tick.Pool, tick.Magnitude*min(seconds, tick.SecondsLeft))
		tick.SecondsLeft -= seconds
	}
	for index := range running.result.Modifiers {
		running.result.Modifiers[index].SecondsLeft -= seconds
	}
	running.result.Ticks = slices.DeleteFunc(running.result.Ticks, func(tick Tick) bool { return tick.SecondsLeft <= 0 })
	running.result.Modifiers = slices.DeleteFunc(running.result.Modifiers, func(held Held) bool { return held.SecondsLeft <= 0 })
}

func originOf(ruleset *data.Ruleset, item *data.Item, effect *data.ConsumableEffect) string {
	if effect.Origin != "" {
		return effect.Origin
	}
	if item.Type == "food" {
		return ruleset.OriginFor["food"]
	}
	return ruleset.OriginFor["consumable"]
}

func (running *state) use(ruleset *data.Ruleset, event Event) error {
	item, known := ruleset.Items["core:consumable/"+event.Item]
	if !known || item.Consumable == nil {
		return &data.NotDefinedError{Kind: "consumable", Key: event.Item}
	}
	var food bool = item.Type == "food"
	var delivery string = item.Consumable.Delivery
	if (delivery != "drink" && delivery != "eat") || (food && event.InCombat) {
		running.result.Refused = append(running.result.Refused, event.Item)
		return nil
	}
	running.result.Toxicity += item.Consumable.Toxicity
	var source string = "item:" + event.Item
	if food {
		source = "food:" + event.Item
	}
	var gone map[string]bool = map[string]bool{}
	for index := range item.Consumable.Effects {
		var origin string = originOf(ruleset, &item, &item.Consumable.Effects[index])
		if ruleset.OriginReplaces[origin] {
			gone[origin] = true
		}
	}
	running.result.Ticks = slices.DeleteFunc(running.result.Ticks, func(tick Tick) bool { return gone[tick.Origin] || tick.Source == source })
	running.result.Modifiers = slices.DeleteFunc(running.result.Modifiers, func(held Held) bool { return gone[held.Origin] || held.Source == source })
	for index := range item.Consumable.Effects {
		var effect *data.ConsumableEffect = &item.Consumable.Effects[index]
		var key string = effect.Effect[strings.LastIndex(effect.Effect, "/")+1:]
		var duration int64 = max(effect.Duration, 1)
		var origin string = originOf(ruleset, &item, effect)
		var magnitude int64 = 1
		if effect.Magnitude != nil {
			magnitude = *effect.Magnitude
		}
		if key == "restore_resource" {
			pool, ok := poolOfParameter[effect.Parameter]
			if !ok {
				return &data.NotDefinedError{Kind: "resource", Key: effect.Parameter}
			}
			if duration == 1 {
				running.add(pool, magnitude)
			} else {
				running.result.Ticks = append(running.result.Ticks, Tick{Source: source, Origin: origin, Pool: pool, Magnitude: magnitude, SecondsLeft: duration})
			}
			continue
		}
		if duration > 1 {
			var operation string = "add"
			if key == "resist" {
				operation = "resist"
			} else if effect.Magnitude == nil {
				operation = "grant"
			}
			var target string = effect.Parameter
			if target == "" {
				target = key
			}
			running.result.Modifiers = append(running.result.Modifiers, Held{Source: source, Origin: origin, Target: target, Operation: operation, Value: magnitude, SecondsLeft: duration})
		}
	}
	return nil
}

// Run is the `consume` command.
func Run(ruleset *data.Ruleset, input Input) (Result, error) {
	var running state = state{
		result: Result{
			Pools: input.Current.Pools, Toxicity: input.Current.Toxicity,
			Ticks: []Tick{}, Modifiers: []Held{}, Refused: []string{}, Conditions: []string{},
		},
		maximum: input.Maximum.Pools,
	}
	for _, event := range input.Events {
		switch {
		case event.Item != "":
			if err := running.use(ruleset, event); err != nil {
				return Result{}, err
			}
		case event.Turns > 0:
			running.advance(ruleset, event.Turns*ruleset.SecondsPerTurn)
		default:
			running.advance(ruleset, event.Advance)
		}
	}
	if input.Maximum.Toxicity != nil && running.result.Toxicity > *input.Maximum.Toxicity {
		running.result.Conditions = append(running.result.Conditions, overdosed)
	}
	return running.result, nil
}
