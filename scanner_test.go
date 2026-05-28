package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

// --- Scanner Tests ---

func TestScanner_Scan_RecursiveDiscovery(t *testing.T) {
	// Create a temp dir with nested .go files
	tmpDir := t.TempDir()

	// Create nested structure: root/a.go, root/sub/b.go, root/sub/deep/c.go
	dirs := []string{
		tmpDir,
		filepath.Join(tmpDir, "sub"),
		filepath.Join(tmpDir, "sub", "deep"),
	}
	for _, d := range dirs[1:] {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("failed to create dir %s: %v", d, err)
		}
	}

	goFiles := []string{
		filepath.Join(tmpDir, "a.go"),
		filepath.Join(tmpDir, "sub", "b.go"),
		filepath.Join(tmpDir, "sub", "deep", "c.go"),
	}
	for _, f := range goFiles {
		if err := os.WriteFile(f, []byte("package main\n"), 0o644); err != nil {
			t.Fatalf("failed to write file %s: %v", f, err)
		}
	}

	scanner := NewScanner(ScanConfig{RootPath: tmpDir})
	found, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	if len(found) != 3 {
		t.Fatalf("expected 3 files, got %d: %v", len(found), found)
	}

	// Verify all expected files are present
	foundSet := make(map[string]bool)
	for _, f := range found {
		foundSet[f] = true
	}
	for _, expected := range goFiles {
		if !foundSet[expected] {
			t.Errorf("expected file %s not found in results", expected)
		}
	}
}

func TestScanner_Scan_DefaultExclusion(t *testing.T) {
	tmpDir := t.TempDir()

	// Create default excluded dirs: vendor/, .git/, testdata/
	excludedDirs := []string{
		filepath.Join(tmpDir, "vendor"),
		filepath.Join(tmpDir, ".git"),
		filepath.Join(tmpDir, "testdata"),
	}
	for _, d := range excludedDirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("failed to create dir %s: %v", d, err)
		}
		// Put a .go file in each excluded dir
		goFile := filepath.Join(d, "excluded.go")
		if err := os.WriteFile(goFile, []byte("package excluded\n"), 0o644); err != nil {
			t.Fatalf("failed to write file %s: %v", goFile, err)
		}
	}

	// Create a non-excluded .go file at root
	rootFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(rootFile, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	scanner := NewScanner(ScanConfig{RootPath: tmpDir})
	found, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	// Should only find the root file, not files in excluded dirs
	if len(found) != 1 {
		t.Fatalf("expected 1 file, got %d: %v", len(found), found)
	}
	if found[0] != rootFile {
		t.Errorf("expected %s, got %s", rootFile, found[0])
	}
}

func TestScanner_Scan_CustomExclusion(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a "custom" directory with a .go file
	customDir := filepath.Join(tmpDir, "custom")
	if err := os.MkdirAll(customDir, 0o755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	customFile := filepath.Join(customDir, "custom.go")
	if err := os.WriteFile(customFile, []byte("package custom\n"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Create a non-excluded .go file at root
	rootFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(rootFile, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	scanner := NewScanner(ScanConfig{
		RootPath:    tmpDir,
		ExcludeDirs: []string{"custom"},
	})
	found, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	if len(found) != 1 {
		t.Fatalf("expected 1 file, got %d: %v", len(found), found)
	}
	if found[0] != rootFile {
		t.Errorf("expected %s, got %s", rootFile, found[0])
	}
}

func TestScanner_Scan_TestFileFiltering(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a regular file and a test file
	regularFile := filepath.Join(tmpDir, "foo.go")
	testFile := filepath.Join(tmpDir, "foo_test.go")
	if err := os.WriteFile(regularFile, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.WriteFile(testFile, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Without IncludeTests — test file should be skipped
	scanner := NewScanner(ScanConfig{RootPath: tmpDir, IncludeTests: false})
	found, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}
	if len(found) != 1 {
		t.Fatalf("expected 1 file (no tests), got %d: %v", len(found), found)
	}
	if found[0] != regularFile {
		t.Errorf("expected %s, got %s", regularFile, found[0])
	}

	// With IncludeTests — test file should be included
	scanner = NewScanner(ScanConfig{RootPath: tmpDir, IncludeTests: true})
	found, err = scanner.Scan()
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}
	if len(found) != 2 {
		t.Fatalf("expected 2 files (with tests), got %d: %v", len(found), found)
	}
}

func TestScanner_ValidateProject_Valid(t *testing.T) {
	// Use testdata/clean-project/ which has go.mod
	scanner := NewScanner(ScanConfig{RootPath: "testdata/clean-project"})
	err := scanner.ValidateProject()
	if err != nil {
		t.Fatalf("expected no error for valid project, got: %v", err)
	}
}

func TestScanner_ValidateProject_NoGoMod(t *testing.T) {
	// Use testdata/no-gomod/ which has no go.mod
	scanner := NewScanner(ScanConfig{RootPath: "testdata/no-gomod"})
	err := scanner.ValidateProject()
	if err == nil {
		t.Fatal("expected error for project without go.mod, got nil")
	}
	if !contains(err.Error(), "no go.mod found") {
		t.Errorf("expected error to contain 'no go.mod found', got: %v", err)
	}
}

func TestScanner_ValidateProject_NonExistent(t *testing.T) {
	scanner := NewScanner(ScanConfig{RootPath: "/nonexistent/path"})
	err := scanner.ValidateProject()
	if err == nil {
		t.Fatal("expected error for non-existent path, got nil")
	}
	if !contains(err.Error(), "does not exist") {
		t.Errorf("expected error to contain 'does not exist', got: %v", err)
	}
}

// --- IsGeneratedFile Tests ---

func TestIsGeneratedFile_ValidHeader(t *testing.T) {
	src := `// Code generated by protoc-gen-go. DO NOT EDIT.
package main
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "gen.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	if !IsGeneratedFile(file) {
		t.Error("expected IsGeneratedFile to return true for valid generated header")
	}
}

func TestIsGeneratedFile_InvalidHeader(t *testing.T) {
	src := `// This file was generated by our tool
package main
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "gen.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	if IsGeneratedFile(file) {
		t.Error("expected IsGeneratedFile to return false for non-standard generated comment")
	}
}

func TestIsGeneratedFile_SecondCommentGroup(t *testing.T) {
	src := `// Package main provides the entry point.
package main

// Code generated by some tool. DO NOT EDIT.
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "gen.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	if IsGeneratedFile(file) {
		t.Error("expected IsGeneratedFile to return false when 'Code generated' is in second comment group")
	}
}

func TestIsGeneratedFile_EmptyFile(t *testing.T) {
	src := `package main
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "empty.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	if IsGeneratedFile(file) {
		t.Error("expected IsGeneratedFile to return false for file with no comments")
	}
}

func TestIsGeneratedFile_NilFile(t *testing.T) {
	if IsGeneratedFile(nil) {
		t.Error("expected IsGeneratedFile to return false for nil file")
	}
}

// --- Helper ---

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
