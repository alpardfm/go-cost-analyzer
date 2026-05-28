package main

import (
	"go/parser"
	"go/token"
	"testing"
)

var validPatterns = map[string]bool{
	"CEG-016": true,
	"CEG-017": true,
	"CEG-018": true,
}

func TestParseSuppression_LineLevelSameLine(t *testing.T) {
	src := `package main

func main() {
	x := 1 //noinspect:CEG-016
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	result := ParseSuppressions(file, fset, validPatterns)

	// The comment is on line 4 (same line as x := 1)
	if !result.IsSuppressed(4, "CEG-016") {
		t.Error("expected line 4 to be suppressed for CEG-016")
	}
}

func TestParseSuppression_LineLevelLineBefore(t *testing.T) {
	src := `package main

func main() {
	//noinspect:CEG-016
	x := 1
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	result := ParseSuppressions(file, fset, validPatterns)

	// Comment is on line 4, code is on line 5
	// IsSuppressed checks line-1, so line 5 should be suppressed
	if !result.IsSuppressed(5, "CEG-016") {
		t.Error("expected line 5 to be suppressed for CEG-016 (comment on line before)")
	}
}

func TestParseSuppression_FileLevel(t *testing.T) {
	src := `package main //noinspect:all

func main() {
	x := 1
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	result := ParseSuppressions(file, fset, validPatterns)

	if !result.FileSkipped {
		t.Error("expected FileSkipped to be true when //noinspect:all is on package declaration line")
	}
}

func TestParseSuppression_InvalidPattern(t *testing.T) {
	src := `package main

func main() {
	x := 1 //noinspect:nonexistent
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	result := ParseSuppressions(file, fset, validPatterns)

	if len(result.InvalidPatterns) == 0 {
		t.Fatal("expected InvalidPatterns to contain 'nonexistent'")
	}

	found := false
	for _, p := range result.InvalidPatterns {
		if p == "nonexistent" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'nonexistent' in InvalidPatterns, got %v", result.InvalidPatterns)
	}
}

func TestParseSuppression_MultiplePatterns(t *testing.T) {
	src := `package main

func main() {
	x := 1 //noinspect:CEG-016
	//noinspect:CEG-017
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	result := ParseSuppressions(file, fset, validPatterns)

	if !result.IsSuppressed(4, "CEG-016") {
		t.Error("expected line 4 to be suppressed for CEG-016")
	}
	if !result.IsSuppressed(5, "CEG-017") {
		t.Error("expected line 5 to be suppressed for CEG-017")
	}
}

func TestParseSuppression_DoesNotAffectOtherPatterns(t *testing.T) {
	src := `package main

func main() {
	a := 1
	b := 2 //noinspect:CEG-016
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	result := ParseSuppressions(file, fset, validPatterns)

	// Line 5 is suppressed for CEG-016
	if !result.IsSuppressed(5, "CEG-016") {
		t.Error("expected line 5 to be suppressed for CEG-016")
	}
	// Line 5 should NOT be suppressed for CEG-017
	if result.IsSuppressed(5, "CEG-017") {
		t.Error("expected line 5 to NOT be suppressed for CEG-017")
	}
}

func TestParseSuppression_WithSpace(t *testing.T) {
	src := `package main

func main() {
	x := 1 // noinspect:CEG-016
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	result := ParseSuppressions(file, fset, validPatterns)

	if !result.IsSuppressed(4, "CEG-016") {
		t.Error("expected line 4 to be suppressed for CEG-016 with space after //")
	}
}

func TestIsSuppressed_FileSkipped(t *testing.T) {
	result := &SuppressionResult{
		FileSkipped:     true,
		SuppressedLines: make(map[int][]string),
	}

	// When FileSkipped is true, any line/pattern should be suppressed
	if !result.IsSuppressed(1, "CEG-016") {
		t.Error("expected IsSuppressed to return true for any line when FileSkipped is true")
	}
	if !result.IsSuppressed(100, "CEG-017") {
		t.Error("expected IsSuppressed to return true for any line/pattern when FileSkipped is true")
	}
	if !result.IsSuppressed(50, "anything") {
		t.Error("expected IsSuppressed to return true for any pattern when FileSkipped is true")
	}
}

func TestParseSuppression_NilFile(t *testing.T) {
	fset := token.NewFileSet()

	// Should not panic and return empty result
	result := ParseSuppressions(nil, fset, validPatterns)

	if result.FileSkipped {
		t.Error("expected FileSkipped to be false for nil file")
	}
	if len(result.SuppressedLines) != 0 {
		t.Error("expected SuppressedLines to be empty for nil file")
	}
	if len(result.InvalidPatterns) != 0 {
		t.Error("expected InvalidPatterns to be empty for nil file")
	}
}
