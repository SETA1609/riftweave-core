package main

import (
	"strings"
	"testing"
)

// branches is a function of complexity 1 + n: one if per branch.
func branches(name string, n int) string {
	var body strings.Builder
	for index := 0; index < n; index++ {
		body.WriteString("\tif x == 0 {\n\t\tx++\n\t}\n")
	}
	return "func " + name + "(x int) int {\n" + body.String() + "\treturn x\n}\n"
}

const nested = `func deep(x int) int {
	if x > 0 {
		for x > 1 {
			if x > 2 {
				for x > 3 {
					if x > 4 {
						x--
					}
				}
			}
		}
	}
	return x
}
`

func measure(t *testing.T, source string) []Function {
	t.Helper()
	functions, err := Measure("sample.go", []byte("package sample\n\n"+source))
	if err != nil {
		t.Fatal(err)
	}
	return functions
}

func TestComplexityCountsBranchesAndBooleans(t *testing.T) {
	functions := measure(t, `func f(x int, ok bool) int {
	if x > 0 && ok {
		return 1
	}
	for index := 0; index < x; index++ {
		switch {
		case x == 1:
			return 2
		case x == 2 || ok:
			return 3
		}
	}
	return 0
}
`)
	// 1, the if, &&, the for, two cases, and ||.
	if len(functions) != 1 || functions[0].Complexity != 7 || functions[0].Depth != 2 {
		t.Fatalf("want complexity 7 and depth 2, got %+v", functions)
	}
}

func TestTheGateFailsAFunctionOverTheLimits(t *testing.T) {
	functions := measure(t, branches("fine", 39)+branches("over", 40)+nested)
	problems := Check(functions, Allowed{})
	if len(problems) != 2 || !strings.Contains(problems[0], "over") || !strings.Contains(problems[1], "deep: nesting depth 5 is over 4") {
		t.Fatalf("want over and deep, got %v", problems)
	}
	if strings.Contains(strings.Join(problems, "\n"), "fine") {
		t.Errorf("complexity 40 is at the limit, not over it: %v", problems)
	}
}

func TestTheAllowListOnlyShrinks(t *testing.T) {
	functions := measure(t, branches("big", 50)+branches("mended", 3))
	allowed := Allowed{"sample.go big": {Complexity: 51, Depth: 1}, "sample.go mended": {Complexity: 45, Depth: 1}}
	problems := Check(functions, allowed)
	if len(problems) != 1 || !strings.Contains(problems[0], "mended") || !strings.Contains(problems[0], "remove it from the allow list") {
		t.Fatalf("a function back under the limits must leave the list: %v", problems)
	}
	allowed["sample.go big"] = Entry{Complexity: 49, Depth: 1}
	delete(allowed, "sample.go mended")
	problems = Check(functions, allowed)
	if len(problems) != 1 || !strings.Contains(problems[0], "big: complexity 51 is over its allowed 49") {
		t.Fatalf("an allowed function may not grow: %v", problems)
	}
}

func TestTheAllowListParses(t *testing.T) {
	allowed, err := ParseAllowed("# comment\ninternal/attack/attack.go Run 133 5\n\ncore/author.go refIndex 12 5\n")
	if err != nil || len(allowed) != 2 || allowed["internal/attack/attack.go Run"] != (Entry{133, 5}) {
		t.Fatalf("got %v, %v", allowed, err)
	}
	if _, err := ParseAllowed("internal/attack/attack.go Run many 5\n"); err == nil {
		t.Error("a bad number is an error")
	}
}
