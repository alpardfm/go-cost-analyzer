package main

import (
	"go/ast"
	"go/token"
	"strings"
)

// SuppressionResult holds suppression state for a file.
type SuppressionResult struct {
	FileSkipped     bool             // true if //noinspect:all on package decl
	SuppressedLines map[int][]string // line -> pattern names suppressed
	InvalidPatterns []string         // unrecognized pattern names in comments
}

// ParseSuppressions extracts all noinspect comments from a file.
// It iterates over all comments, looking for //noinspect:<pattern> directives.
// If //noinspect:all appears on the package declaration line, the entire file is skipped.
// Invalid pattern names (not in validPatterns and not "all") are tracked for warnings.
func ParseSuppressions(file *ast.File, fset *token.FileSet, validPatterns map[string]bool) SuppressionResult {
	result := SuppressionResult{
		SuppressedLines: make(map[int][]string),
	}

	if file == nil || fset == nil {
		return result
	}

	packageLine := fset.Position(file.Package).Line

	for _, cg := range file.Comments {
		for _, comment := range cg.List {
			text := comment.Text

			// Strip the leading "//" from the comment
			if !strings.HasPrefix(text, "//") {
				continue
			}
			body := text[2:]

			// Handle both "//noinspect:X" (no space) and "// noinspect:X" (with space)
			body = strings.TrimSpace(body)

			if !strings.HasPrefix(body, "noinspect:") {
				continue
			}

			// Extract pattern name after "noinspect:"
			pattern := strings.TrimPrefix(body, "noinspect:")
			pattern = strings.TrimSpace(pattern)

			if pattern == "" {
				continue
			}

			commentLine := fset.Position(comment.Pos()).Line

			// Check for file-level suppression: //noinspect:all on package declaration line
			if pattern == "all" && commentLine == packageLine {
				result.FileSkipped = true
				continue
			}

			// Validate pattern name
			if pattern != "all" && !validPatterns[pattern] {
				result.InvalidPatterns = append(result.InvalidPatterns, pattern)
			}

			// Record line-level suppression
			result.SuppressedLines[commentLine] = append(result.SuppressedLines[commentLine], pattern)
		}
	}

	return result
}

// IsSuppressed checks if a finding at a given line is suppressed for a pattern.
// Returns true if:
// - The entire file is skipped (//noinspect:all on package decl)
// - The pattern is suppressed on the same line
// - The pattern is suppressed on the line before (comment above the code)
func (s *SuppressionResult) IsSuppressed(line int, patternName string) bool {
	if s.FileSkipped {
		return true
	}

	// Check same line
	if patterns, ok := s.SuppressedLines[line]; ok {
		for _, p := range patterns {
			if p == patternName || p == "all" {
				return true
			}
		}
	}

	// Check line before (comment on line above)
	if patterns, ok := s.SuppressedLines[line-1]; ok {
		for _, p := range patterns {
			if p == patternName || p == "all" {
				return true
			}
		}
	}

	return false
}
