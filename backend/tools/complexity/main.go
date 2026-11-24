// Command complexity is the backend's complexity gate (QA_S24_02, docs/backend/README.md#checks): a
// non-test function may not be over a cyclomatic complexity of 40 or a nesting depth of 4, except
// the ones in allow.txt, which may not grow and leave the list when they are back under.
//
//	go run ./tools/complexity            check the backend against tools/complexity/allow.txt
//	go run ./tools/complexity -list      print every function over the limits, as allow.txt lines
package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// The limits; a function at a limit passes.
const (
	maxComplexity = 40
	maxDepth      = 4
)

// Function is one measured function.
type Function struct {
	File       string
	Name       string
	Line       int
	Complexity int
	Depth      int
}

// key names a function in the allow list.
func (function Function) key() string { return function.File + " " + function.Name }

// Entry is an allowed function's complexity and depth.
type Entry struct{ Complexity, Depth int }

// Allowed is the allow list, by file and function.
type Allowed map[string]Entry

// Measure is every function in one Go source file.
func Measure(file string, source []byte) ([]Function, error) {
	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, file, source, 0)
	if err != nil {
		return nil, err
	}
	var functions []Function
	for _, declaration := range parsed.Decls {
		declared, ok := declaration.(*ast.FuncDecl)
		if !ok || declared.Body == nil {
			continue
		}
		complexity, depth := score(declared.Body)
		functions = append(functions, Function{File: file, Name: name(declared), Line: fileSet.Position(declared.Pos()).Line, Complexity: complexity, Depth: depth})
	}
	return functions, nil
}

// MeasureTree is every non-test function under root, with paths relative to it.
func MeasureTree(root string) ([]Function, error) {
	var functions []Function
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		measured, err := Measure(filepath.ToSlash(relative), source)
		functions = append(functions, measured...)
		return err
	})
	return functions, err
}

// name is a function's name, with its receiver's type for a method.
func name(declared *ast.FuncDecl) string {
	if declared.Recv == nil || len(declared.Recv.List) == 0 {
		return declared.Name.Name
	}
	var receiver ast.Expr = declared.Recv.List[0].Type
	if star, ok := receiver.(*ast.StarExpr); ok {
		receiver = star.X
	}
	if index, ok := receiver.(*ast.IndexExpr); ok {
		receiver = index.X
	}
	if ident, ok := receiver.(*ast.Ident); ok {
		return ident.Name + "." + declared.Name.Name
	}
	return declared.Name.Name
}

// score is a body's cyclomatic complexity (1, plus each if, loop, case, and && or ||) and its
// deepest nesting of if, for, range, switch, and select. A function literal counts toward the
// function that holds it.
func score(body *ast.BlockStmt) (int, int) {
	var complexity, deepest int = 1, 0
	var walk func(node ast.Node, depth int)
	walk = func(node ast.Node, depth int) {
		ast.Inspect(node, func(child ast.Node) bool {
			if child == nil || child == node {
				return true
			}
			switch typed := child.(type) {
			case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt, *ast.SelectStmt:
				switch typed.(type) {
				case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt:
					complexity++
				}
				deepest = max(deepest, depth+1)
				walk(typed, depth+1)
				return false
			case *ast.CaseClause, *ast.CommClause:
				complexity++
			case *ast.BinaryExpr:
				if typed.Op == token.LAND || typed.Op == token.LOR {
					complexity++
				}
			}
			return true
		})
	}
	walk(body, 0)
	return complexity, deepest
}

// Check is the gate's problems, sorted by file and line: a function over the limits that is not
// allowed, an allowed one that grew, and an allowed one back under the limits.
func Check(functions []Function, allowed Allowed) []string {
	sort.SliceStable(functions, func(i, j int) bool {
		if functions[i].File != functions[j].File {
			return functions[i].File < functions[j].File
		}
		return functions[i].Line < functions[j].Line
	})
	var problems []string
	for _, function := range functions {
		var where string = fmt.Sprintf("%s:%d %s", function.File, function.Line, function.Name)
		entry, listed := allowed[function.key()]
		over := function.Complexity > maxComplexity || function.Depth > maxDepth
		switch {
		case listed && !over:
			problems = append(problems, fmt.Sprintf("%s: complexity %d and depth %d are within the limits; remove it from the allow list", where, function.Complexity, function.Depth))
		case listed && function.Complexity > entry.Complexity:
			problems = append(problems, fmt.Sprintf("%s: complexity %d is over its allowed %d", where, function.Complexity, entry.Complexity))
		case listed && function.Depth > entry.Depth:
			problems = append(problems, fmt.Sprintf("%s: nesting depth %d is over its allowed %d", where, function.Depth, entry.Depth))
		case !listed && function.Complexity > maxComplexity:
			problems = append(problems, fmt.Sprintf("%s: complexity %d is over %d", where, function.Complexity, maxComplexity))
		case !listed && function.Depth > maxDepth:
			problems = append(problems, fmt.Sprintf("%s: nesting depth %d is over %d", where, function.Depth, maxDepth))
		}
	}
	return problems
}

// ParseAllowed reads allow.txt: "file function complexity depth" a line, # for a comment.
func ParseAllowed(content string) (Allowed, error) {
	var allowed Allowed = Allowed{}
	scanner := bufio.NewScanner(strings.NewReader(content))
	for line := 1; scanner.Scan(); line++ {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		if len(fields) != 4 {
			return nil, fmt.Errorf("allow list line %d: want file, function, complexity, and depth", line)
		}
		complexity, errComplexity := strconv.Atoi(fields[2])
		depth, errDepth := strconv.Atoi(fields[3])
		if errComplexity != nil || errDepth != nil {
			return nil, fmt.Errorf("allow list line %d: complexity and depth are numbers", line)
		}
		allowed[fields[0]+" "+fields[1]] = Entry{Complexity: complexity, Depth: depth}
	}
	return allowed, scanner.Err()
}

func main() {
	list := flag.Bool("list", false, "print every function over the limits as allow.txt lines")
	root := flag.String("root", ".", "the backend directory")
	allowFile := flag.String("allow", "tools/complexity/allow.txt", "the allow list")
	flag.Parse()
	functions, err := MeasureTree(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if *list {
		for _, function := range functions {
			if function.Complexity > maxComplexity || function.Depth > maxDepth {
				fmt.Printf("%s %s %d %d\n", function.File, function.Name, function.Complexity, function.Depth)
			}
		}
		return
	}
	content, err := os.ReadFile(*allowFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	allowed, err := ParseAllowed(string(content))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	problems := Check(functions, allowed)
	for _, problem := range problems {
		fmt.Println(problem)
	}
	if len(problems) > 0 {
		os.Exit(1)
	}
}
