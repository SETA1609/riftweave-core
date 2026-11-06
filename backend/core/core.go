// Package core is the one door to the ruleset. An adapter (agent CLI, human CLI, HTTP API)
// opens a Core, then calls Run with a command name and JSON, and gets JSON back. It parses,
// calls, and formats; every rule is in a package under internal/.
//
// Each command is a typed function `func(*data.Ruleset, In) (Out, error)` in its own package,
// bound here to JSON in a single table. Adding a command is one package and one line in
// `commands`.
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
// It matches ErrRule.
type RuleError struct{ Msg string }

func (e *RuleError) Error() string        { return e.Msg }
func (e *RuleError) Is(target error) bool { return target == ErrRule }

type handler func(rs *data.Ruleset, input []byte) (any, error)

// bind decodes the input strictly into In (an unknown field is an error, as in the contract)
// and runs the typed command.
func bind[In, Out any](run func(*data.Ruleset, In) (Out, error)) handler {
	return func(rs *data.Ruleset, input []byte) (any, error) {
		var in In
		dec := json.NewDecoder(bytes.NewReader(input))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&in); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInput, err)
		}
		if dec.More() {
			return nil, fmt.Errorf("%w: more than one JSON value", ErrInput)
		}
		out, err := run(rs, in)
		if err != nil {
			return nil, &RuleError{err.Error()}
		}
		return out, nil
	}
}

// commands is the registry, keyed by the command and vector function name.
var commands = map[string]handler{
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

// New loads the ruleset from the data directory held by fsys.
func New(fsys fs.FS) (*Core, error) {
	rs, err := data.Load(fsys)
	if err != nil {
		return nil, err
	}
	return &Core{rs}, nil
}

// Open loads the ruleset from a directory such as ruleset/data.
func Open(dir string) (*Core, error) { return New(os.DirFS(dir)) }

// Commands lists the command names, sorted.
func Commands() []string {
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// Run runs a command on JSON input and returns its JSON result. A rule error comes back
// both as `{"error": "..."}` bytes, so any call can be saved as a vector, and as an
// error wrapping ErrRule; bad input wraps ErrInput; a missing command, ErrUnknownCommand.
func (c *Core) Run(name string, input []byte) ([]byte, error) {
	h, ok := commands[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownCommand, name)
	}
	out, err := h(c.rules, input)
	if err != nil {
		if !errors.Is(err, ErrRule) {
			return nil, err
		}
		var rule *RuleError
		errors.As(err, &rule)
		body, _ := json.Marshal(map[string]string{"error": rule.Msg})
		return body, err
	}
	return json.Marshal(out)
}

type formulaInput struct {
	Formula string           `json:"formula"`
	Expr    *formula.Program `json:"expr"`
	Actor   actor.Actor      `json:"actor"`
}

func runFormula(rs *data.Ruleset, in formulaInput) (map[string]int64, error) {
	program := in.Expr
	if program == nil {
		var ok bool
		if program, ok = rs.Formulas[in.Formula]; !ok {
			return nil, fmt.Errorf("formula %q is not defined", in.Formula)
		}
	}
	a := in.Actor.Attributes
	value, err := program.Eval(&a)
	return map[string]int64{"value": value}, err
}

type modifiersInput struct {
	Value          int64               `json:"value"`
	Target         *string             `json:"target"`
	Modifiers      []modifier.Modifier `json:"modifiers"`
	ElapsedTurns   int64               `json:"elapsed_turns"`
	RemovedSources []string            `json:"removed_sources"`
	Packet         *struct {
		Amount int64  `json:"amount"`
		Tag    string `json:"tag"`
	} `json:"packet"`
}

type modifiersResult struct {
	Value  int64    `json:"value"`
	Flags  []string `json:"flags"`
	Packet *int64   `json:"packet,omitempty"`
}

func runModifiers(rs *data.Ruleset, in modifiersInput) (modifiersResult, error) {
	own := in.Modifiers
	if in.Target != nil {
		own = modifier.Target(in.Modifiers, *in.Target)
	}
	value, err := modifier.Value(in.Value, own, rs.Origins, in.ElapsedTurns, in.RemovedSources)
	if err != nil {
		return modifiersResult{}, err
	}
	res := modifiersResult{Value: value, Flags: modifier.Flags(in.Modifiers, in.ElapsedTurns, in.RemovedSources)}
	if in.Packet != nil {
		packet, err := modifier.Packet(in.Packet.Amount, in.Packet.Tag, in.Modifiers, rs.Origins, in.ElapsedTurns, in.RemovedSources)
		if err != nil {
			return modifiersResult{}, err
		}
		res.Packet = &packet
	}
	return res, nil
}
