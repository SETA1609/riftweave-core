package main

import (
	"strings"
	"testing"
)

const before = `goos: linux
BenchmarkCommand/combat-8         	   20000	     50000 ns/op	    4096 B/op	      40 allocs/op
BenchmarkCommand/casting-8        	   10000	    100000 ns/op
BenchmarkLoad-8                   	     100	  12000000 ns/op
PASS
`

const after = `BenchmarkCommand/combat-8         	   25000	     40000 ns/op	    2048 B/op	      20 allocs/op
BenchmarkCommand/casting-8        	   10000	    110000 ns/op
BenchmarkLoad-8                   	     100	  12000000 ns/op
`

func TestParseReadsNanosecondsPerOperation(t *testing.T) {
	results := Parse(strings.NewReader(before))
	if len(results) != 3 || results["BenchmarkCommand/combat"] != 50000 || results["BenchmarkLoad"] != 12000000 {
		t.Fatalf("got %v", results)
	}
}

func TestDiffShowsTheChangeAndFailsASlowdownOverTheThreshold(t *testing.T) {
	lines, slower := Diff(Parse(strings.NewReader(before)), Parse(strings.NewReader(after)), 5)
	var table string = strings.Join(lines, "\n")
	for _, want := range []string{"BenchmarkCommand/combat", "-20.0%", "BenchmarkCommand/casting", "+10.0%", "BenchmarkLoad", "+0.0%"} {
		if !strings.Contains(table, want) {
			t.Errorf("missing %q in\n%s", want, table)
		}
	}
	if len(slower) != 1 || slower[0] != "BenchmarkCommand/casting" {
		t.Errorf("casting is 10%% slower, over 5%%: %v", slower)
	}
}
