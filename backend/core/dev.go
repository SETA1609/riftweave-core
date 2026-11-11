package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/SETA1609/riftweave-core/backend/internal/actor"
	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
	"github.com/SETA1609/riftweave-core/backend/internal/stored"
)

// errDevOff is every dev operation, step, or replay without --dev (docs/backend/dev-mode.md).
var errDevOff = errors.New("dev mode is off; run with --dev")

// EnableDev turns on dev mode for this core, as the --dev flag does.
func (core *Core) EnableDev() { core.rules.DevMode = true }

// devOperation is the input of one dev operation, without the session.
type devOperation struct {
	Set       *int64 `json:"set"`
	Add       *int64 `json:"add"`
	Stacks    *int64 `json:"stacks"`
	Magnitude *int64 `json:"magnitude"`
	Remove    *int64 `json:"remove"`
	Quantity  *int64 `json:"quantity"`
	Operation string `json:"operation"`
	Pool      string `json:"pool"`
	Key       string `json:"key"`
	Effect    string `json:"effect"`
	Parameter string `json:"parameter"`
	Origin    string `json:"origin"`
	Base      string `json:"base"`
	Material  string `json:"material"`
	Turns     int64  `json:"turns"`
	Clear     bool   `json:"clear"`
}

type devInput struct {
	devOperation
	Session *Session `json:"session"`
}

func init() { commands["dev"] = bind(runDev) }

// runDev is the `dev` command: one operation on the session's actor, logged, and the session
// marked dev.
func runDev(ruleset *data.Ruleset, input devInput) (sessionResult, error) {
	if !ruleset.DevMode {
		return sessionResult{}, errDevOff
	}
	if input.Session == nil {
		return sessionResult{}, &data.MissingInputError{Subject: "dev", Needs: "session"}
	}
	encoded, err := json.Marshal(input.devOperation)
	if err != nil {
		return sessionResult{}, err
	}
	var step SessionStep = SessionStep{Command: "dev", Input: compactOperation(encoded)}
	actor, _, err := applyStep(ruleset, input.Session.Actor, step)
	if err != nil {
		return sessionResult{}, err
	}
	var session Session = *input.Session
	session.Actor, session.Dev = actor, true
	session.Log = append(slices.Clone(session.Log), step)
	return sessionResult{Session: session}, nil
}

// compactOperation drops the fields an operation left at their zero value, so the log holds
// what was given.
func compactOperation(encoded []byte) json.RawMessage {
	var fields map[string]any
	_ = json.Unmarshal(encoded, &fields)
	for key, value := range fields {
		if value == nil || value == "" || value == false || (key == "turns" && value == 0.0) {
			delete(fields, key)
		}
	}
	compacted, _ := json.Marshal(fields)
	return compacted
}

// setOrAdd is the new value of current: set, or current plus add; exactly one is given.
func setOrAdd(operation *devOperation, current int64) (int64, error) {
	switch {
	case operation.Set != nil && operation.Add != nil:
		return 0, fmt.Errorf("dev %s takes set or add, not both", operation.Operation)
	case operation.Set != nil:
		return *operation.Set, nil
	case operation.Add != nil:
		return current + *operation.Add, nil
	}
	return 0, &data.MissingInputError{Subject: "dev " + operation.Operation, Needs: "set or add"}
}

// applyDev changes the actor by one operation. It is a session step, so a replay repeats it.
func applyDev(ruleset *data.Ruleset, target *stored.Actor, encoded json.RawMessage) error {
	if !ruleset.DevMode {
		return errDevOff
	}
	var operation devOperation
	if err := json.Unmarshal(encoded, &operation); err != nil {
		return err
	}
	switch operation.Operation {
	case "resource":
		return devResource(target, &operation)
	case "xp":
		value, err := setOrAdd(&operation, target.XP)
		target.XP = max(0, value)
		return err
	case "level":
		return devLevel(ruleset, target, &operation)
	case "attribute":
		if _, known := actor.ByKey(operation.Key); !known {
			return &data.NotDefinedError{Kind: "attribute", Key: operation.Key}
		}
		value, err := setOrAdd(&operation, target.Attributes[operation.Key])
		if err == nil {
			target.Attributes[operation.Key] = max(1, value)
		}
		return err
	case "skill":
		if _, known := ruleset.Skills.Get(operation.Key); !known {
			return &data.NotDefinedError{Kind: "skill", Key: operation.Key}
		}
		value, err := setOrAdd(&operation, target.Skills[operation.Key])
		if err == nil {
			if target.Skills == nil {
				target.Skills = map[string]int64{}
			}
			target.Skills[operation.Key] = min(100, max(0, value))
		}
		return err
	case "condition":
		return devCondition(ruleset, target, &operation)
	case "effect":
		return devEffect(ruleset, target, &operation)
	case "item":
		return devItem(ruleset, target, &operation)
	case "turns":
		return devTurns(target, &operation)
	}
	return &data.NotDefinedError{Kind: "dev operation", Key: operation.Operation}
}

func devResource(target *stored.Actor, operation *devOperation) error {
	var pool *int64
	switch operation.Pool {
	case "hp":
		pool = &target.Resources.HitPoints
	case "mp":
		pool = &target.Resources.ManaPoints
	case "sp":
		pool = &target.Resources.Stamina
	case "ap":
		pool = &target.Resources.ActionPoints
	default:
		return &data.NotDefinedError{Kind: "pool", Key: operation.Pool}
	}
	value, err := setOrAdd(operation, *pool)
	if err == nil {
		*pool = max(0, value)
	}
	return err
}

// devLevel sets the level and the XP total of that level, from the progression command.
func devLevel(ruleset *data.Ruleset, target *stored.Actor, operation *devOperation) error {
	if operation.Set == nil || operation.Add != nil {
		return &data.MissingInputError{Subject: "dev " + operation.Operation, Needs: "set"}
	}
	body, _ := json.Marshal(map[string]any{"op": "xp", "level": *operation.Set})
	output, err := commands["progression"](ruleset, body)
	if err != nil {
		return err
	}
	result, _ := encodedValue(output).(map[string]any)
	target.Level, target.XP = *operation.Set, number(result["xp_total"])
	return nil
}

func devCondition(ruleset *data.Ruleset, target *stored.Actor, operation *devOperation) error {
	if _, known := ruleset.Conditions.Table.Get(operation.Key); !known {
		return &data.NotDefinedError{Kind: "condition", Key: operation.Key}
	}
	if operation.Stacks == nil {
		return &data.MissingInputError{Subject: "dev " + operation.Operation, Needs: "stacks"}
	}
	target.Conditions = slices.DeleteFunc(slices.Clone(target.Conditions), func(condition stored.Condition) bool {
		return condition.Key == operation.Key
	})
	if *operation.Stacks > 0 {
		target.Conditions = append(target.Conditions, stored.Condition{Key: operation.Key, Stacks: *operation.Stacks})
	}
	return nil
}

// devEffect holds an effect as one modifier with source dev:<effect key>, or clears it.
func devEffect(ruleset *data.Ruleset, target *stored.Actor, operation *devOperation) error {
	if !ruleset.EffectRefs[operation.Effect] {
		return &data.NotDefinedError{Kind: "effect", Key: operation.Effect}
	}
	var key string = operation.Effect[strings.LastIndex(operation.Effect, "/")+1:]
	var source string = "dev:" + key
	target.Modifiers = slices.DeleteFunc(slices.Clone(target.Modifiers), func(held stored.ActiveModifier) bool {
		return held.Modifier.Source == source
	})
	if operation.Clear {
		return nil
	}
	if _, known := ruleset.Origins[operation.Origin]; !known {
		return &data.NotDefinedError{Kind: "origin", Key: operation.Origin}
	}
	if operation.Magnitude == nil {
		return &data.MissingInputError{Subject: "dev " + operation.Operation, Needs: "magnitude"}
	}
	var held modifier.Modifier = modifier.Modifier{
		Target: cmpOr(operation.Parameter, key), Operation: modifier.Add, Value: float64(*operation.Magnitude),
		Source: source, Origin: operation.Origin, Duration: "while_applied",
	}
	if operation.Turns > 0 {
		held.Duration, held.Turns = "turns", operation.Turns
	}
	target.Modifiers = append(target.Modifiers, stored.ActiveModifier{Modifier: held})
	return nil
}

func cmpOr(first, second string) string {
	if first != "" {
		return first
	}
	return second
}

func devItem(ruleset *data.Ruleset, target *stored.Actor, operation *devOperation) error {
	if operation.Remove != nil {
		var before int = len(target.Inventory)
		target.Inventory = slices.DeleteFunc(slices.Clone(target.Inventory), func(item stored.ItemInstance) bool {
			return item.ID == *operation.Remove
		})
		if len(target.Inventory) == before {
			return &data.NotDefinedError{Kind: "inventory id", Key: fmt.Sprint(*operation.Remove)}
		}
		for slot, id := range target.Equipped {
			if id == *operation.Remove {
				delete(target.Equipped, slot)
			}
		}
		return nil
	}
	if _, known := ruleset.Items[operation.Base]; !known {
		return &data.NotDefinedError{Kind: "item", Key: operation.Base}
	}
	var id int64 = 1
	for _, item := range target.Inventory {
		id = max(id, item.ID+1)
	}
	var quantity int64 = 1
	if operation.Quantity != nil {
		quantity = *operation.Quantity
	}
	target.Inventory = append(target.Inventory, stored.ItemInstance{ID: id, Base: operation.Base, Material: operation.Material, Mods: []json.RawMessage{}, Quantity: quantity})
	return nil
}

// devTurns advances every modifier; a turns modifier that has run its turns ends.
func devTurns(target *stored.Actor, operation *devOperation) error {
	if operation.Turns <= 0 {
		return &data.MissingInputError{Subject: "dev " + operation.Operation, Needs: "turns"}
	}
	var kept []stored.ActiveModifier = []stored.ActiveModifier{}
	for _, held := range target.Modifiers {
		held.TurnsElapsed += operation.Turns
		if held.Modifier.Duration == "turns" && held.TurnsElapsed >= held.Modifier.Turns {
			continue
		}
		kept = append(kept, held)
	}
	target.Modifiers = kept
	return nil
}
