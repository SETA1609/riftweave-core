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
	opLit opcode = iota
	opAttr
	opVar
	opAdd
	opSub
	opMul
	opDiv
)

const stackSize = 32

type instr struct {
	op  opcode
	arg int64 // literal, ability, variable slot, or argument count
}

// Program is a compiled formula. Its zero value evaluates to an error.
type Program struct {
	code []instr
	vars []string
}

// Var is a named input of a formula, such as `level` or `base_cp`.
type Var struct {
	Name string
	Val  int64
}

type node struct {
	Lit  *int64  `json:"lit"`
	Attr *string `json:"attr"`
	Var  *string `json:"var"`
	Op   string  `json:"op"`
	Args []node  `json:"args"`
}

// UnmarshalJSON compiles the tree it reads, so a data struct can hold a Program directly.
func (p *Program) UnmarshalJSON(b []byte) error {
	var root node
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&root); err != nil {
		return err
	}
	*p = Program{}
	depth, err := p.emit(root)
	if err != nil {
		return err
	}
	if depth > stackSize {
		return fmt.Errorf("formula needs a stack of %d, the limit is %d", depth, stackSize)
	}
	return nil
}

// emit appends the node's instructions and returns the stack depth they need.
func (p *Program) emit(n node) (int, error) {
	switch {
	case n.Lit != nil:
		p.code = append(p.code, instr{opLit, *n.Lit})
		return 1, nil
	case n.Attr != nil:
		ab, ok := actor.ByKey(*n.Attr)
		if !ok {
			return 0, fmt.Errorf("%q is not an attribute", *n.Attr)
		}
		p.code = append(p.code, instr{opAttr, int64(ab)})
		return 1, nil
	case n.Var != nil:
		slot := -1
		for i, name := range p.vars {
			if name == *n.Var {
				slot = i
			}
		}
		if slot < 0 {
			slot = len(p.vars)
			p.vars = append(p.vars, *n.Var)
		}
		p.code = append(p.code, instr{opVar, int64(slot)})
		return 1, nil
	}
	var op opcode
	switch n.Op {
	case "add":
		op = opAdd
	case "sub":
		op = opSub
	case "mul":
		op = opMul
	case "div":
		op = opDiv
		if len(n.Args) != 2 {
			return 0, errors.New("div takes two arguments")
		}
	default:
		return 0, fmt.Errorf("formula op %q has no implementation", n.Op)
	}
	if op == opSub && len(n.Args) == 0 {
		return 0, errors.New("sub takes at least one argument")
	}
	depth := 0
	for i, arg := range n.Args {
		d, err := p.emit(arg)
		if err != nil {
			return 0, err
		}
		depth = max(depth, i+d)
	}
	p.code = append(p.code, instr{op, int64(len(n.Args))})
	return max(depth, 1), nil
}

// FloorDiv is integer division that rounds toward negative infinity. b must not be 0.
func FloorDiv(a, b int64) int64 {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

// Eval runs the program. attrs may be nil when the formula reads no attribute.
func (p *Program) Eval(attrs *actor.Attrs, vars ...Var) (int64, error) {
	if len(p.code) == 0 {
		return 0, errors.New("empty formula")
	}
	var stack [stackSize]int64
	sp := 0
	for _, in := range p.code {
		switch in.op {
		case opLit:
			stack[sp] = in.arg
			sp++
		case opAttr:
			ab := actor.Ability(in.arg)
			if attrs == nil || !attrs.Has(ab) {
				return 0, fmt.Errorf("formula reads attribute %q, which the actor does not have", ab)
			}
			stack[sp] = attrs.Get(ab)
			sp++
		case opVar:
			name := p.vars[in.arg]
			found := false
			for _, v := range vars {
				if v.Name == name {
					stack[sp], found = v.Val, true
					break
				}
			}
			if !found {
				return 0, fmt.Errorf("formula reads variable %q, which was not given", name)
			}
			sp++
		default:
			n := int(in.arg)
			args := stack[sp-n : sp]
			var out int64
			switch in.op {
			case opAdd:
				for _, a := range args {
					out += a
				}
			case opSub:
				out = args[0]
				for _, a := range args[1:] {
					out -= a
				}
			case opMul:
				out = 1
				for _, a := range args {
					out *= a
				}
			case opDiv:
				if args[1] == 0 {
					return 0, errors.New("formula divides by zero")
				}
				out = FloorDiv(args[0], args[1])
			}
			sp -= n
			stack[sp] = out
			sp++
		}
	}
	return stack[0], nil
}
