// Package commandsummons works out what each of a commander's summons gets (docs/core/commanding.md,
// GD_S06_03): an order, the command action, gives the summons it reaches AP for their next turn,
// and a held command spell gives every summon a bonus to attack while the commander holds it.
// The kernel has no party: the caller names the summons and passes each its modifiers.
package commandsummons

import (
	"fmt"

	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
)

const (
	heldEffect   = "core:effect/command_summons"
	moraleOrigin = "core:origin/morale"
)

// Commander is the one who commands.
type Commander struct {
	Name          string              `json:"name"`
	Perk          string              `json:"perk"`          // a feature with a command object
	Modifiers     []modifier.Modifier `json:"modifiers"`     // those on command change an order's cost
	Concentration *string             `json:"concentration"` // the held spell key, or nil
}

// Order is an order given this turn.
type Order struct {
	Targets []string `json:"targets"`
}

// Input is the `command_summons` command's input.
type Input struct {
	Commander Commander `json:"commander"`
	Summons   []string  `json:"summons"`
	Order     *Order    `json:"order"`
}

// Result is the order's cost and, per summon, the modifiers it gets.
type Result struct {
	ActionPoints int64                          `json:"ap"`
	Modifiers    map[string][]modifier.Modifier `json:"modifiers"`
}

// Run is the `command_summons` command.
func Run(ruleset *data.Ruleset, input Input) (Result, error) {
	var commander *Commander = &input.Commander
	var result Result = Result{Modifiers: map[string][]modifier.Modifier{}}
	for _, name := range input.Summons {
		result.Modifiers[name] = []modifier.Modifier{}
	}
	if input.Order != nil {
		perk, ok := ruleset.Features.Get(commander.Perk)
		if !ok || perk.Command == nil {
			return Result{}, fmt.Errorf("an order needs a perk with a command object")
		}
		for _, name := range input.Order.Targets {
			if _, active := result.Modifiers[name]; !active {
				return Result{}, fmt.Errorf("%s is not an active summon", name)
			}
		}
		if perk.Command.Reach == "one" && len(input.Order.Targets) > 1 {
			return Result{}, fmt.Errorf("an order with %s reaches 1 summon, not %d", perk.Key, len(input.Order.Targets))
		}
		action, err := ruleset.Actions.Actions.Require("action", "command")
		if err != nil {
			return Result{}, err
		}
		cost, err := modifier.Value(action.Cost, modifier.Target(commander.Modifiers, "command"), ruleset.Origins, 0, nil)
		if err != nil {
			return Result{}, err
		}
		result.ActionPoints = max(cost, 0)
		for _, name := range input.Order.Targets {
			result.Modifiers[name] = append(result.Modifiers[name], modifier.Modifier{
				Target: "ap", Operation: modifier.Add, Value: float64(perk.Command.AP),
				Source: "command:" + perk.Key + ":" + commander.Name, Origin: moraleOrigin,
				Duration: "turns", Turns: perk.Command.Turns,
			})
		}
	}
	if commander.Concentration != nil {
		if spell, ok := ruleset.Spells.Get(*commander.Concentration); ok {
			for _, applied := range spell.Effects {
				if applied.Effect != heldEffect {
					continue
				}
				for _, name := range input.Summons {
					result.Modifiers[name] = append(result.Modifiers[name], modifier.Modifier{
						Target: "attack_roll", Operation: modifier.Add, Value: applied.Magnitude,
						Source: "command:" + spell.Key + ":" + commander.Name, Origin: moraleOrigin,
						Duration: "while_applied",
					})
				}
			}
		}
	}
	return result, nil
}
