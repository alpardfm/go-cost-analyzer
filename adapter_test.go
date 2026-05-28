package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/alpardfm/cost-efficient-go/types"
)

// mockDetector is a configurable test double for types.Detector.
type mockDetector struct {
	callCount int
	findings  []types.Finding
	rule      types.Rule
}

func (m *mockDetector) Detect(ctx types.ASTContext) []types.Finding {
	m.callCount++
	return m.findings
}

func (m *mockDetector) Rule() types.Rule {
	return m.rule
}

func TestFileDetectorAdapter_NilFile(t *testing.T) {
	mock := &mockDetector{}
	adapter := NewFileDetector(mock)

	fset := token.NewFileSet()
	result := adapter.DetectFile(fset, nil)

	if result != nil {
		t.Fatalf("expected nil for nil *ast.File, got %v", result)
	}
}

func TestFileDetectorAdapter_EmptyFile(t *testing.T) {
	mock := &mockDetector{}
	adapter := NewFileDetector(mock)

	src := `package main`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "empty.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	result := adapter.DetectFile(fset, file)

	if len(result) != 0 {
		t.Fatalf("expected empty findings for empty file, got %d findings", len(result))
	}
}

func TestFileDetectorAdapter_DelegationCount(t *testing.T) {
	mock := &mockDetector{}
	adapter := NewFileDetector(mock)

	src := `package main

func hello() {
	x := 1
	_ = x
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "count.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	// Count non-nil nodes visited by ast.Inspect independently.
	expectedCount := 0
	ast.Inspect(file, func(n ast.Node) bool {
		if n != nil {
			expectedCount++
		}
		return true
	})

	result := adapter.DetectFile(fset, file)

	if mock.callCount != expectedCount {
		t.Fatalf("expected Detect() to be called %d times, got %d", expectedCount, mock.callCount)
	}

	// With no findings configured, result should be empty.
	if len(result) != 0 {
		t.Fatalf("expected empty findings, got %d", len(result))
	}
}

func TestFileDetectorAdapter_MultipleMatches(t *testing.T) {
	finding := types.Finding{
		RuleID:      "TEST-001",
		Explanation: "test finding",
		Severity:    types.Minor,
		Category:    types.Memory,
	}

	// mockIdentDetector returns a finding only for *ast.Ident nodes.
	identDetector := &mockIdentDetector{
		finding: finding,
	}
	adapter := NewFileDetector(identDetector)

	src := `package main

func foo() {
	bar := baz
	_ = bar
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "multi.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	// Count *ast.Ident nodes independently.
	identCount := 0
	ast.Inspect(file, func(n ast.Node) bool {
		if _, ok := n.(*ast.Ident); ok {
			identCount++
		}
		return true
	})

	if identCount == 0 {
		t.Fatal("test source should contain at least one *ast.Ident node")
	}

	result := adapter.DetectFile(fset, file)

	if len(result) != identCount {
		t.Fatalf("expected %d findings (one per *ast.Ident), got %d", identCount, len(result))
	}
}

func TestFileDetectorAdapter_NoRelevantNodes(t *testing.T) {
	// Detector that always returns empty slice.
	mock := &mockDetector{
		findings: []types.Finding{},
	}
	adapter := NewFileDetector(mock)

	src := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "norelevant.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	result := adapter.DetectFile(fset, file)

	if len(result) != 0 {
		t.Fatalf("expected empty findings when detector returns nothing, got %d", len(result))
	}

	// Verify the detector was still called (nodes were visited).
	if mock.callCount == 0 {
		t.Fatal("expected Detect() to be called at least once")
	}
}

func TestFileDetectorAdapter_Rule(t *testing.T) {
	expectedRule := types.Rule{
		ID:          "CEG-TEST",
		Name:        "Test Rule",
		Description: "A test rule for unit testing",
		Severity:    types.Major,
		Category:    types.Concurrency,
		Suggestion:  "Fix the thing",
	}

	mock := &mockDetector{
		rule: expectedRule,
	}
	adapter := NewFileDetector(mock)

	got := adapter.Rule()

	if got.ID != expectedRule.ID {
		t.Errorf("Rule().ID = %q, want %q", got.ID, expectedRule.ID)
	}
	if got.Name != expectedRule.Name {
		t.Errorf("Rule().Name = %q, want %q", got.Name, expectedRule.Name)
	}
	if got.Description != expectedRule.Description {
		t.Errorf("Rule().Description = %q, want %q", got.Description, expectedRule.Description)
	}
	if got.Severity != expectedRule.Severity {
		t.Errorf("Rule().Severity = %v, want %v", got.Severity, expectedRule.Severity)
	}
	if got.Category != expectedRule.Category {
		t.Errorf("Rule().Category = %v, want %v", got.Category, expectedRule.Category)
	}
	if got.Suggestion != expectedRule.Suggestion {
		t.Errorf("Rule().Suggestion = %q, want %q", got.Suggestion, expectedRule.Suggestion)
	}
}

// mockIdentDetector returns a finding only for *ast.Ident nodes.
type mockIdentDetector struct {
	finding types.Finding
}

func (m *mockIdentDetector) Detect(ctx types.ASTContext) []types.Finding {
	if _, ok := ctx.Node.(*ast.Ident); ok {
		return []types.Finding{m.finding}
	}
	return nil
}

func (m *mockIdentDetector) Rule() types.Rule {
	return types.Rule{ID: "TEST-IDENT"}
}
