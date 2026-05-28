package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/alpardfm/cost-efficient-go/types"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestProperty_AdapterDelegationFidelity(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generate random Go source code snippets and verify adapter calls Detect()
	// exactly once per non-nil node
	sources := []string{
		`package main`,
		`package main; func f() {}`,
		`package main; func f() { x := 1; _ = x }`,
		`package main; import "fmt"; func main() { fmt.Println("hi") }`,
		`package main; type S struct { A int; B string }`,
		`package main; func f() { for i := 0; i < 10; i++ {} }`,
	}

	properties.Property("adapter calls Detect once per non-nil node", prop.ForAll(
		func(idx int) bool {
			src := sources[idx%len(sources)]

			// Count nodes independently
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", src, 0)
			if err != nil {
				return true // skip unparseable
			}

			expectedCount := 0
			ast.Inspect(file, func(n ast.Node) bool {
				if n != nil {
					expectedCount++
				}
				return true
			})

			// Run adapter with counting mock
			mock := &countingDetector{}
			adapter := NewFileDetector(mock)

			fset2 := token.NewFileSet()
			file2, _ := parser.ParseFile(fset2, "test.go", src, 0)
			adapter.DetectFile(fset2, file2)

			return mock.callCount == expectedCount
		},
		gen.IntRange(0, 599),
	))

	properties.TestingRun(t)
}

type countingDetector struct {
	callCount int
}

func (d *countingDetector) Detect(ctx types.ASTContext) []types.Finding {
	d.callCount++
	return nil
}

func (d *countingDetector) Rule() types.Rule {
	return types.Rule{ID: "TEST-COUNT"}
}
