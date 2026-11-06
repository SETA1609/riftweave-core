// Package formula compiles a formula tree (docs/core/formula.md) into a flat program and
// evaluates it with a fixed stack.
//
// The tree in the data is for people; the program is for the machine: one slice of
// instructions in postfix order, no pointers, no allocation while evaluating. Integer math
// only, and `div` is floor division.
package formula

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/SETA1609/riftweave-core/backend/internal/actor"
)

type opcode uint8

const (
	opcodeLiteral opcode = iota
	opcodeAttribute
	opcodeVariable
	opcodeAdd
	opcodeSubtract
	opcodeMultiply
	opcodeDivide
)

const stackSize = 32

type instruction struct {
	operation opcode
	argument  int64 // literal, ability, variable slot, or argument count
}

// Program is a compiled formula. Its zero value evaluates to an error.
type Program struct {
	instructions  []instruction
	variableNames []string
}

// Variable is a named input of a formula, such as `level` or `base_cp`.
type Variable struct {
	Name  string
	Value int64
}

type node struct {
	Literal   *int64  `json:"lit"`
	Attribute *string `json:"attr"`
	Variable  *string `json:"var"`
	Operation string  `json:"op"`
	Arguments []node  `json:"args"`
}

// UnmarshalJSON compiles the tree it reads, so a data struct can hold a Program directly.
func (program *Program) UnmarshalJSON(encoded []byte) error {
	var root node
	var decoder *json.Decoder = json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&root); err != nil {
		return err
	}
	*program = Program{}
	depth, err := program.emit(root)
	if err != nil {
		return err
	}
	if depth > stackSize {
		return fmt.Errorf("formula needs a stack of %d, the limit is %d", depth, stackSize)
	}
	return nil
}

// emit appends the node's instructions and returns the stack depth they need.
func (program *Program) emit(tree node) (int, error) {
	switch {
	case tree.Literal != nil:
		program.instructions = append(program.instructions, instruction{opcodeLiteral, *tree.Literal})
		return 1, nil
	case tree.Attribute != nil:
		ability, ok := actor.ByKey(*tree.Attribute)
		if !ok {
			return 0, fmt.Errorf("%q is not an attribute", *tree.Attribute)
		}
		program.instructions = append(program.instructions, instruction{opcodeAttribute, int64(ability)})
		return 1, nil
	case tree.Variable != nil:
		var slot int = -1
		for index, name := range program.variableNames {
			if name == *tree.Variable {
				slot = index
			}
		}
		if slot < 0 {
			slot = len(program.variableNames)
			program.variableNames = append(program.variableNames, *tree.Variable)
		}
		program.instructions = append(program.instructions, instruction{opcodeVariable, int64(slot)})
		return 1, nil
	}
	var operation opcode
	switch tree.Operation {
	case "add":
		operation = opcodeAdd
	case "sub":
		operation = opcodeSubtract
	case "mul":
		operation = opcodeMultiply
	case "div":
		operation = opcodeDivide
		if len(tree.Arguments) != 2 {
			return 0, errors.New("div takes two arguments")
		}
	default:
		return 0, fmt.Errorf("formula op %q has no implementation", tree.Operation)
	}
	if operation == opcodeSubtract && len(tree.Arguments) == 0 {
		return 0, errors.New("sub takes at least one argument")
	}
	var depth int = 0
	for index, argument := range tree.Arguments {
		argumentDepth, err := program.emit(argument)
		if err != nil {
			return 0, err
		}
		depth = max(depth, index+argumentDepth)
	}
	program.instructions = append(program.instructions, instruction{operation, int64(len(tree.Arguments))})
	return max(depth, 1), nil
}

// FloorDivide is integer division that rounds toward negative infinity. divisor must not be 0.
func FloorDivide(dividend, divisor int64) int64 {
	var quotient int64 = dividend / divisor
	if dividend%divisor != 0 && (dividend < 0) != (divisor < 0) {
		quotient--
	}
	return quotient
}

// Evaluate runs the program. attributes may be nil when the formula reads no attribute.
func (program *Program) Evaluate(attributes *actor.Attributes, variables ...Variable) (int64, error) {
	if len(program.instructions) == 0 {
		return 0, errors.New("empty formula")
	}
	var stack [stackSize]int64
	var stackPointer int = 0
	for _, step := range program.instructions {
		switch step.operation {
		case opcodeLiteral:
			stack[stackPointer] = step.argument
			stackPointer++
		case opcodeAttribute:
			var ability actor.Ability = actor.Ability(step.argument)
			if attributes == nil || !attributes.Has(ability) {
				return 0, fmt.Errorf("formula reads attribute %q, which the actor does not have", ability)
			}
			stack[stackPointer] = attributes.Get(ability)
			stackPointer++
		case opcodeVariable:
			var name string = program.variableNames[step.argument]
			var found bool = false
			for _, variable := range variables {
				if variable.Name == name {
					stack[stackPointer], found = variable.Value, true
					break
				}
			}
			if !found {
				return 0, fmt.Errorf("formula reads variable %q, which was not given", name)
			}
			stackPointer++
		default:
			var argumentCount int = int(step.argument)
			var arguments []int64 = stack[stackPointer-argumentCount : stackPointer]
			var result int64
			switch step.operation {
			case opcodeAdd:
				for _, argument := range arguments {
					result += argument
				}
			case opcodeSubtract:
				result = arguments[0]
				for _, argument := range arguments[1:] {
					result -= argument
				}
			case opcodeMultiply:
				result = 1
				for _, argument := range arguments {
					result *= argument
				}
			case opcodeDivide:
				if arguments[1] == 0 {
					return 0, errors.New("formula divides by zero")
				}
				result = FloorDivide(arguments[0], arguments[1])
			}
			stackPointer -= argumentCount
			stack[stackPointer] = result
			stackPointer++
		}
	}
	return stack[0], nil
}
