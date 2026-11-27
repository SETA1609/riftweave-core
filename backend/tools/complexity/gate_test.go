package main

import (
	"os"
	"testing"
)

// The gate on the backend itself (QA_S24_02): every non-test function outside the allow list is
// within the limits, and every allowed one is still over them and has not grown.
func TestTheBackendPassesTheGate(t *testing.T) {
	content, err := os.ReadFile("allow.txt")
	if err != nil {
		t.Fatal(err)
	}
	allowed, err := ParseAllowed(string(content))
	if err != nil {
		t.Fatal(err)
	}
	functions, err := MeasureTree("../..")
	if err != nil {
		t.Fatal(err)
	}
	if len(functions) < 300 {
		t.Fatalf("measured only %d functions", len(functions))
	}
	for _, problem := range Check(functions, allowed) {
		t.Error(problem)
	}
}
