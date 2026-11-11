package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sort"

	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/stored"
)

// SessionStep is one command in a session's log, with its input as it was given.
type SessionStep struct {
	Command string          `json:"command"`
	Input   json.RawMessage `json:"input"`
	As      string          `json:"as,omitempty"`
}

// SessionRuleset is what the run started on.
type SessionRuleset struct {
	Modules []string `json:"modules"`
}

// Session is one run (docs/core/state.md): a starting actor, a log, and the current snapshot.
type Session struct {
	Ruleset SessionRuleset `json:"ruleset"`
	Start   stored.Actor   `json:"start"`
	Log     []SessionStep  `json:"log"`
	Actor   stored.Actor   `json:"actor"`
}

type sessionInput struct {
	Op      string        `json:"op"`
	Start   *stored.Actor `json:"start"`
	Session *Session      `json:"session"`
	Step    *SessionStep  `json:"step"`
}

type sessionResult struct {
	Session Session         `json:"session"`
	Result  json.RawMessage `json:"result,omitempty"`
	Matches *bool           `json:"matches,omitempty"`
}

func init() { commands["session"] = bind(runSession) }

// runSession is the `session` command (docs/backend/sessions.md). It holds no rule.
func runSession(ruleset *data.Ruleset, input sessionInput) (sessionResult, error) {
	switch input.Op {
	case "start":
		if input.Start == nil {
			return sessionResult{}, errors.New("session start needs a start actor")
		}
		var modules []string = slices.Clone(ruleset.Modules)
		if modules == nil {
			modules = []string{}
		}
		sort.Strings(modules)
		actor, err := copyActor(*input.Start)
		if err != nil {
			return sessionResult{}, err
		}
		return sessionResult{Session: Session{Ruleset: SessionRuleset{Modules: modules}, Start: *input.Start, Log: []SessionStep{}, Actor: actor}}, nil
	case "step":
		if input.Session == nil || input.Step == nil {
			return sessionResult{}, errors.New("session step needs a session and a step")
		}
		var session Session = *input.Session
		actor, output, err := applyStep(ruleset, session.Actor, *input.Step)
		if err != nil {
			return sessionResult{}, err
		}
		session.Actor = actor
		session.Log = append(slices.Clone(session.Log), *input.Step)
		return sessionResult{Session: session, Result: output}, nil
	case "replay":
		if input.Session == nil {
			return sessionResult{}, errors.New("session replay needs a session")
		}
		var session Session = *input.Session
		actor, err := copyActor(session.Start)
		if err != nil {
			return sessionResult{}, err
		}
		for _, step := range session.Log {
			if actor, _, err = applyStep(ruleset, actor, step); err != nil {
				return sessionResult{}, err
			}
		}
		saved, rebuilt := encodedValue(session.Actor), encodedValue(actor)
		var matches bool = reflect.DeepEqual(saved, rebuilt)
		session.Actor = actor
		return sessionResult{Session: session, Matches: &matches}, nil
	}
	return sessionResult{}, &data.NotDefinedError{Kind: "session op", Key: input.Op}
}

func encodedValue(value any) any {
	body, _ := json.Marshal(value)
	var decoded any
	_ = json.Unmarshal(body, &decoded)
	return decoded
}

func copyActor(actor stored.Actor) (stored.Actor, error) {
	body, err := json.Marshal(actor)
	if err != nil {
		return stored.Actor{}, err
	}
	var copied stored.Actor
	err = json.Unmarshal(body, &copied)
	return copied, err
}

// section returns the object at key, creating it when missing.
func section(object map[string]any, key string) map[string]any {
	if inner, ok := object[key].(map[string]any); ok {
		return inner
	}
	inner := map[string]any{}
	object[key] = inner
	return inner
}

func setDefault(object map[string]any, key string, value any) {
	if _, present := object[key]; !present {
		object[key] = value
	}
}

func derivedStats(ruleset *data.Ruleset, actor *stored.Actor) (map[string]any, error) {
	body, err := json.Marshal(map[string]any{"actor": characterOf(actor)})
	if err != nil {
		return nil, err
	}
	output, err := commands["derived"](ruleset, body)
	if err != nil {
		return nil, err
	}
	stats, _ := encodedValue(output).(map[string]any)
	return stats, nil
}

// characterOf is the part of a stored actor the `actor` input of a command reads.
func characterOf(actor *stored.Actor) map[string]any {
	return map[string]any{"attributes": actor.Attributes, "level": actor.Level, "growth": actor.Growth, "skills": actor.Skills}
}

// fillInput adds what the actor knows to the fields the step did not name.
func fillInput(ruleset *data.Ruleset, actor *stored.Actor, step *SessionStep, input map[string]any) error {
	switch step.Command {
	case "derived":
		setDefault(input, "actor", characterOf(actor))
	case "combat":
		if step.As == "defender" {
			defender := section(input, "defender")
			setDefault(defender, "hp", actor.Resources.HitPoints)
			setDefault(defender, "attributes", actor.Attributes)
		} else {
			setDefault(section(input, "attacker"), "attributes", actor.Attributes)
		}
	case "rest":
		setDefault(input, "actor", characterOf(actor))
		setDefault(input, "current", map[string]any{
			"hp": actor.Resources.HitPoints, "mp": actor.Resources.ManaPoints,
			"stamina": actor.Resources.Stamina, "ap": actor.Resources.ActionPoints,
		})
		if _, present := section(input, "max")["stamina"]; !present {
			stats, err := derivedStats(ruleset, actor)
			if err != nil {
				return err
			}
			section(input, "max")["stamina"] = stats["sp"]
		}
	case "casting":
		caster := section(input, "caster")
		setDefault(caster, "skills", actor.Skills)
		setDefault(caster, "attributes", actor.Attributes)
		setDefault(caster, "mana", actor.Resources.ManaPoints)
	case "progression":
		if input["op"] == "level_up" {
			setDefault(input, "actor", map[string]any{"attributes": actor.Attributes})
		}
	case "consume":
		stats, err := derivedStats(ruleset, actor)
		if err != nil {
			return err
		}
		setDefault(input, "current", map[string]any{"hp": actor.Resources.HitPoints, "mp": actor.Resources.ManaPoints, "stamina": actor.Resources.Stamina})
		setDefault(input, "max", map[string]any{"hp": stats["hp"], "mp": stats["mp"], "stamina": stats["sp"]})
	case "conditions":
		if _, present := input["conditions"]; !present {
			held := map[string]int64{}
			for _, condition := range actor.Conditions {
				held[condition.Key] = condition.Stacks
			}
			input["conditions"] = held
		}
	}
	return nil
}

func number(value any) int64 {
	amount, _ := value.(float64)
	return int64(amount)
}

// writeResult changes the actor by what the command returned.
func writeResult(actor *stored.Actor, step *SessionStep, input, result map[string]any) {
	switch step.Command {
	case "combat":
		if step.As == "defender" {
			actor.Resources.HitPoints = number(result["hp"])
		}
	case "rest":
		actor.Resources.HitPoints, actor.Resources.ManaPoints = number(result["hp"]), number(result["mp"])
		actor.Resources.Stamina, actor.Resources.ActionPoints = number(result["stamina"]), number(result["ap"])
		kept, _ := result["conditions"].([]any)
		named, _ := input["conditions"].([]any)
		var cleared []string
		for _, entry := range named {
			key, _ := entry.(map[string]any)["key"].(string)
			if !slices.Contains(kept, any(key)) {
				cleared = append(cleared, key)
			}
		}
		actor.Conditions = slices.DeleteFunc(actor.Conditions, func(condition stored.Condition) bool {
			return slices.Contains(cleared, condition.Key)
		})
		for _, key := range cleared {
			actor.Flags = slices.DeleteFunc(actor.Flags, func(flag string) bool { return flag == key })
		}
	case "casting":
		actor.Resources.ManaPoints = number(result["mana"])
	case "progression":
		switch input["op"] {
		case "level":
			actor.XP, actor.Level = number(input["xp"]), number(result["level"])
		case "level_up":
			actor.Growth.HitPoints += number(input["hp_roll"])
			actor.Growth.ManaPoints += number(input["mp_roll"])
			actor.Growth.Stamina += number(input["sp_roll"])
		}
	case "consume":
		actor.Resources.HitPoints, actor.Resources.ManaPoints = number(result["hp"]), number(result["mp"])
		actor.Resources.Stamina = number(result["stamina"])
	case "conditions":
		held, _ := result["conditions"].(map[string]any)
		keys := make([]string, 0, len(held))
		for key := range held {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		actor.Conditions = []stored.Condition{}
		for _, key := range keys {
			actor.Conditions = append(actor.Conditions, stored.Condition{Key: key, Stacks: number(held[key])})
		}
		flags, _ := result["flags"].([]any)
		actor.Flags = []string{}
		for _, flag := range flags {
			actor.Flags = append(actor.Flags, flag.(string))
		}
	}
}

// applyStep runs one step on a copy of the actor, so a step that fails changes nothing.
func applyStep(ruleset *data.Ruleset, current stored.Actor, step SessionStep) (stored.Actor, json.RawMessage, error) {
	handler, known := commands[step.Command]
	if !known || step.Command == "session" {
		return stored.Actor{}, nil, fmt.Errorf("session step: unknown command %q", step.Command)
	}
	actor, err := copyActor(current)
	if err != nil {
		return stored.Actor{}, nil, err
	}
	var input map[string]any
	if err := json.Unmarshal(step.Input, &input); err != nil {
		return stored.Actor{}, nil, err
	}
	if err := fillInput(ruleset, &actor, &step, input); err != nil {
		return stored.Actor{}, nil, err
	}
	body, err := json.Marshal(input)
	if err != nil {
		return stored.Actor{}, nil, err
	}
	output, err := handler(ruleset, body)
	if err != nil {
		return stored.Actor{}, nil, err
	}
	encoded, err := json.Marshal(output)
	if err != nil {
		return stored.Actor{}, nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(encoded, &result); err != nil {
		return stored.Actor{}, nil, err
	}
	writeResult(&actor, &step, input, result)
	return actor, encoded, nil
}
