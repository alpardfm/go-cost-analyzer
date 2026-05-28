package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ScanConfig controls file discovery behavior.
type ScanConfig struct {
	RootPath     string   // Project root (must contain go.mod)
	ExcludeDirs  []string // Additional directories to exclude
	IncludeTests bool     // Whether to include *_test.go files
}

// Scanner discovers Go source files in a project.
type Scanner struct {
	config ScanConfig
}

// DefaultExcludeDirs are always excluded from scanning.
var DefaultExcludeDirs = []string{"vendor", ".git", "testdata"}

// NewScanner creates a new Scanner with the given configuration.
func NewScanner(config ScanConfig) *Scanner {
	return &Scanner{config: config}
}

// ValidateProject checks that the root path exists and contains go.mod.
func (s *Scanner) ValidateProject() error {
	// Check if path exists
	info, err := os.Stat(s.config.RootPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Error: path '%s' does not exist", s.config.RootPath)
		}
		return fmt.Errorf("Error: cannot access path '%s': %v", s.config.RootPath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("Error: path '%s' is not a directory", s.config.RootPath)
	}

	// Check if go.mod exists
	goModPath := filepath.Join(s.config.RootPath, "go.mod")
	if _, err := os.Stat(goModPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Error: no go.mod found in '%s'. Not a valid Go project", s.config.RootPath)
		}
		return fmt.Errorf("Error: cannot access go.mod in '%s': %v", s.config.RootPath, err)
	}

	return nil
}

// Scan returns all .go file paths in the project, respecting exclusion rules.
func (s *Scanner) Scan() ([]string, error) {
	var files []string

	// Build a set of excluded directory names for fast lookup
	excludeSet := make(map[string]bool)
	for _, d := range DefaultExcludeDirs {
		excludeSet[d] = true
	}
	for _, d := range s.config.ExcludeDirs {
		excludeSet[strings.TrimSpace(d)] = true
	}

	err := filepath.WalkDir(s.config.RootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Handle permission errors gracefully — skip the file/dir
			return nil
		}

		// If it's a directory, check if it should be excluded
		if d.IsDir() {
			dirName := d.Name()
			if excludeSet[dirName] {
				return filepath.SkipDir
			}
			return nil
		}

		// Only collect .go files
		if !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}

		// Skip test files unless IncludeTests is set
		if !s.config.IncludeTests && strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}

		files = append(files, path)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("Error: failed to scan directory '%s': %v", s.config.RootPath, err)
	}

	sort.Strings(files)
	return files, nil
}
