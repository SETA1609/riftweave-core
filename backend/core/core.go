// Package core is the one door to the ruleset. An adapter (agent CLI, human CLI, HTTP API)
// opens a Core, then calls Run with a command name and JSON, and gets JSON back. It parses,
// calls, and formats; every rule is in a package under internal/.
//
// Each command is a typed function `func(*data.Ruleset, Input) (Output, error)` in its own
// package, bound here to JSON in a single table. Adding a command is one package and one
// line in `commands`.
package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"

	"github.com/SETA1609/riftweave-core/backend/internal/actor"
	"github.com/SETA1609/riftweave-core/backend/internal/combat"
	"github.com/SETA1609/riftweave-core/backend/internal/condition"
	"github.com/SETA1609/riftweave-core/backend/internal/data"
	"github.com/SETA1609/riftweave-core/backend/internal/derived"
	"github.com/SETA1609/riftweave-core/backend/internal/economy"
	"github.com/SETA1609/riftweave-core/backend/internal/formula"
	"github.com/SETA1609/riftweave-core/backend/internal/loot"
	"github.com/SETA1609/riftweave-core/backend/internal/modifier"
	"github.com/SETA1609/riftweave-core/backend/internal/progression"
	"github.com/SETA1609/riftweave-core/backend/internal/resolution"
	"github.com/SETA1609/riftweave-core/backend/internal/rest"
	"github.com/SETA1609/riftweave-core/backend/internal/wuxing"
)

// Errors an adapter maps to an exit code or a status.
var (
	ErrUnknownCommand = errors.New("unknown command")
	ErrInput          = errors.New("bad input")
	ErrRule           = errors.New("rule error")
)

// RuleError is a command that ran and refused: the wrong number of rolls, an unknown key.
// It matches ErrRule, and errors.As reaches its Cause, so an adapter can tell a
// NotDefinedError (an unknown key, 404) from a MissingInputError (400) from the rest.
type RuleError struct {
	Cause   error
	Message string
}

func (ruleError *RuleError) Error() string        { return ruleError.Message }
func (ruleError *RuleError) Is(target error) bool { return target == ErrRule }
func (ruleError *RuleError) Unwrap() error        { return ruleError.Cause }

// The typed causes a RuleError can carry.
type (
	NotDefinedError   = data.NotDefinedError
	MissingInputError = data.MissingInputError
)

type commandHandler func(ruleset *data.Ruleset, encodedInput []byte) (any, error)

// bind decodes the input strictly into Input (an unknown field is an error, as in the
// contract) and runs the typed command.
func bind[Input, Output any](run func(*data.Ruleset, Input) (Output, error)) commandHandler {
	return func(ruleset *data.Ruleset, encodedInput []byte) (any, error) {
		var input Input
		var decoder *json.Decoder = json.NewDecoder(bytes.NewReader(encodedInput))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInput, err)
		}
		if decoder.More() {
			return nil, fmt.Errorf("%w: more than one JSON value", ErrInput)
		}
		output, err := run(ruleset, input)
		if err != nil {
			return nil, &RuleError{Cause: err, Message: err.Error()}
		}
		return output, nil
	}
}

// commands is the registry, keyed by the command and vector function name.
var commands map[string]commandHandler = map[string]commandHandler{
	"resolution":  bind(resolution.Run),
	"derived":     bind(derived.Run),
	"actions":     bind(derived.RunActions),
	"formula":     bind(runFormula),
	"modifiers":   bind(runModifiers),
	"conditions":  bind(condition.Run),
	"wuxing":      bind(wuxing.Run),
	"economy":     bind(economy.Run),
	"rest":        bind(rest.Run),
	"combat":      bind(combat.Run),
	"loot":        bind(loot.Run),
	"progression": bind(progression.Run),
}

// Core is an opened ruleset.
type Core struct{ rules *data.Ruleset }

// New loads the ruleset from the data directory held by fileSystem.
func New(fileSystem fs.FS) (*Core, error) {
	ruleset, err := data.Load(fileSystem)
	if err != nil {
		return nil, err
	}
	return &Core{ruleset}, nil
}

// Open loads the ruleset from a directory such as ruleset/data.
func Open(directory string) (*Core, error) { return New(os.DirFS(directory)) }

// Commands lists the command names, sorted.
func Commands() []string {
	var names []string = make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// Run runs a command on JSON input and returns its JSON result. A rule error comes back
// both as `{"error": "..."}` bytes, so any call can be saved as a vector, and as an
// error wrapping ErrRule; bad input wraps ErrInput; a missing command, ErrUnknownCommand.
func (core *Core) Run(name string, input []byte) ([]byte, error) {
	handler, ok := commands[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownCommand, name)
	}
	output, err := handler(core.rules, input)
	if err != nil {
		if !errors.Is(err, ErrRule) {
			return nil, err
		}
		var ruleError *RuleError
		errors.As(err, &ruleError)
		body, _ := json.Marshal(map[string]string{"error": ruleError.Message})
		return body, err
	}
	return json.Marshal(output)
}

type formulaInput struct {
	Expression *formula.Program `json:"expr"`
	Formula    string           `json:"formula"`
	Actor      actor.Actor      `json:"actor"`
}

func runFormula(ruleset *data.Ruleset, input formulaInput) (map[string]int64, error) {
	var program *formula.Program = input.Expression
	if program == nil {
		var ok bool
		if program, ok = ruleset.Formulas[input.Formula]; !ok {
			return nil, &data.NotDefinedError{Kind: "formula", Key: input.Formula}
		}
	}
	var attributes actor.Attributes = input.Actor.Attributes
	value, err := program.Evaluate(&attributes)
	return map[string]int64{"value": value}, err
}

type damagePacket struct {
	Tag    string `json:"tag"`
	Amount int64  `json:"amount"`
}

type modifiersInput struct {
	Target         *string             `json:"target"`
	Packet         *damagePacket       `json:"packet"`
	Modifiers      []modifier.Modifier `json:"modifiers"`
	RemovedSources []string            `json:"removed_sources"`
	Value          int64               `json:"value"`
	ElapsedTurns   int64               `json:"elapsed_turns"`
}

type modifiersResult struct {
	Value  int64    `json:"value"`
	Flags  []string `json:"flags"`
	Packet *int64   `json:"packet,omitempty"`
}

func runModifiers(ruleset *data.Ruleset, input modifiersInput) (modifiersResult, error) {
	var targetModifiers []modifier.Modifier = input.Modifiers
	if input.Target != nil {
		targetModifiers = modifier.Target(input.Modifiers, *input.Target)
	}
	value, err := modifier.Value(input.Value, targetModifiers, ruleset.Origins, input.ElapsedTurns, input.RemovedSources)
	if err != nil {
		return modifiersResult{}, err
	}
	var result modifiersResult = modifiersResult{
		Value: value,
		Flags: modifier.Flags(input.Modifiers, input.ElapsedTurns, input.RemovedSources),
	}
	if input.Packet != nil {
		packet, err := modifier.Packet(input.Packet.Amount, input.Packet.Tag, input.Modifiers, ruleset.Origins, input.ElapsedTurns, input.RemovedSources)
		if err != nil {
			return modifiersResult{}, err
		}
		result.Packet = &packet
	}
	return result, nil
}
