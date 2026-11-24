// Command benchdiff compares two `go test -bench` outputs (QA_S24_02): each benchmark's ns/op
// before and after, the change in percent, and a failure when one is slower than the threshold.
//
//	go run ./tools/benchdiff -threshold 5 before.txt after.txt
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
)

// benchmarkLine is a result line: the name with its -N suffix, the runs, and the ns/op.
var benchmarkLine = regexp.MustCompile(`^(Benchmark\S+?)(?:-\d+)?\s+\d+\s+([0-9.]+) ns/op`)

// Parse is each benchmark's ns/op; a benchmark run more than once keeps its fastest.
func Parse(reader io.Reader) map[string]float64 {
	var results map[string]float64 = map[string]float64{}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		match := benchmarkLine.FindStringSubmatch(scanner.Text())
		if match == nil {
			continue
		}
		value, err := strconv.ParseFloat(match[2], 64)
		if err != nil {
			continue
		}
		if known, ok := results[match[1]]; !ok || value < known {
			results[match[1]] = value
		}
	}
	return results
}

// Diff is a table line per benchmark in both runs, and the benchmarks slower than threshold
// percent.
func Diff(before, after map[string]float64, threshold float64) ([]string, []string) {
	var names []string
	for name := range before {
		if _, ok := after[name]; ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	var lines, slower []string
	for _, name := range names {
		change := (after[name] - before[name]) / before[name] * 100
		lines = append(lines, fmt.Sprintf("%-40s %14.0f %14.0f %+7.1f%%", name, before[name], after[name], change))
		if change > threshold {
			slower = append(slower, name)
		}
	}
	return lines, slower
}

func main() {
	threshold := flag.Float64("threshold", 5, "the slowdown in percent that fails")
	flag.Parse()
	if flag.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "usage: benchdiff [-threshold 5] before.txt after.txt")
		os.Exit(2)
	}
	var runs [2]map[string]float64
	for index := range runs {
		file, err := os.Open(flag.Arg(index))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		runs[index] = Parse(file)
		file.Close()
	}
	lines, slower := Diff(runs[0], runs[1], *threshold)
	fmt.Printf("%-40s %14s %14s %8s\n", "benchmark", "before ns/op", "after ns/op", "change")
	for _, line := range lines {
		fmt.Println(line)
	}
	if len(slower) > 0 {
		fmt.Printf("slower than %.0f%%: %v\n", *threshold, slower)
		os.Exit(1)
	}
}
