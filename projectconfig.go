package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectConfig represents the .cost-analyzer.json configuration file.
type ProjectConfig struct {
	// Exclude additional directories from scanning
	Exclude []string `json:"exclude,omitempty"`

	// Include test files in analysis
	IncludeTests *bool `json:"include_tests,omitempty"`

	// Patterns to disable (won't be analyzed)
	DisablePatterns []string `json:"disable_patterns,omitempty"`

	// Only analyze these patterns (if set, overrides disable)
	EnablePatterns []string `json:"enable_patterns,omitempty"`

	// Minimum score threshold
	Threshold *int `json:"threshold,omitempty"`

	// Output format: "text" or "json"
	Format string `json:"format,omitempty"`

	// Impact projection scale
	Scale string `json:"scale,omitempty"`

	// Project-level suppressions by path glob
	Suppress []SuppressionRule `json:"suppress,omitempty"`
}

// SuppressionRule defines a project-level suppression for specific paths.
type SuppressionRule struct {
	Pattern string   `json:"pattern"`         // Pattern ID (e.g., "CEG-006")
	Paths   []string `json:"paths,omitempty"` // Glob patterns (e.g., "internal/legacy/**")
}

const configFileName = ".cost-analyzer.json"

// LoadProjectConfig reads and parses the config file from the project root.
// Returns nil config (not error) if file doesn't exist.
func LoadProjectConfig(projectRoot string) (*ProjectConfig, error) {
	configPath := filepath.Join(projectRoot, configFileName)

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No config file — that's fine
		}
		return nil, fmt.Errorf("Error reading %s: %v", configFileName, err)
	}

	var config ProjectConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("Error parsing %s: %v", configFileName, err)
	}

	return &config, nil
}

// MergeConfig merges ProjectConfig into AnalysisConfig.
// CLI flags take precedence over config file values.
// "cliSet" parameters indicate which flags were explicitly set by the user.
func MergeConfig(analysis *AnalysisConfig, project *ProjectConfig, cliExcludeSet bool, cliPatternsSet bool, cliThresholdSet bool, cliFormatSet bool, cliScaleSet bool, cliIncludeTestsSet bool) {
	if project == nil {
		return
	}

	// Exclude: append config excludes to CLI excludes
	if len(project.Exclude) > 0 {
		analysis.Exclude = append(analysis.Exclude, project.Exclude...)
		analysis.ScanConfig.ExcludeDirs = append(analysis.ScanConfig.ExcludeDirs, project.Exclude...)
	}

	// IncludeTests: config file only if CLI didn't set it
	if !cliIncludeTestsSet && project.IncludeTests != nil {
		analysis.IncludeTests = *project.IncludeTests
		analysis.ScanConfig.IncludeTests = *project.IncludeTests
	}

	// Patterns: handle disable/enable from config
	if !cliPatternsSet {
		if len(project.EnablePatterns) > 0 {
			analysis.Patterns = project.EnablePatterns
		} else if len(project.DisablePatterns) > 0 {
			// Store disabled patterns — orchestrator will filter them out
			analysis.DisablePatterns = project.DisablePatterns
		}
	}

	// Threshold: config only if CLI didn't set it
	if !cliThresholdSet && project.Threshold != nil {
		analysis.Threshold = *project.Threshold
	}

	// Format: config only if CLI didn't set it
	if !cliFormatSet && project.Format != "" {
		analysis.Format = project.Format
	}

	// Scale: config only if CLI didn't set it
	if !cliScaleSet && project.Scale != "" {
		analysis.Scale = project.Scale
	}

	// Suppress: always merge from config (no CLI equivalent)
	if len(project.Suppress) > 0 {
		analysis.SuppressRules = append(analysis.SuppressRules, project.Suppress...)
	}
}

// IsPathSuppressed checks if a given file path is suppressed for a specific pattern
// based on the project-level suppression rules.
// Supports glob patterns: "**" matches any number of path segments,
// "*" matches within a single segment.
func IsPathSuppressed(filePath string, patternID string, rules []SuppressionRule, projectRoot string) bool {
	for _, rule := range rules {
		if rule.Pattern != patternID {
			continue
		}
		for _, glob := range rule.Paths {
			if matchPath(filePath, glob, projectRoot) {
				return true
			}
		}
	}
	return false
}

// matchPath checks if filePath matches a glob pattern relative to projectRoot.
// Supports:
//   - "**" to match any number of directories
//   - "*" to match any characters within a single path segment
//   - Direct prefix matching for directory patterns
func matchPath(filePath string, pattern string, projectRoot string) bool {
	// Make filePath relative to project root for matching
	relPath := filePath
	if projectRoot != "" {
		if rel, err := filepath.Rel(projectRoot, filePath); err == nil {
			relPath = rel
		}
	}

	// Normalize separators
	relPath = filepath.ToSlash(relPath)
	pattern = filepath.ToSlash(pattern)

	// Handle "dir/**" pattern — match anything under that directory
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		return strings.HasPrefix(relPath, prefix+"/") || relPath == prefix
	}

	// Handle "dir/*" pattern — match files directly in that directory
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		if !strings.HasPrefix(relPath, prefix+"/") {
			return false
		}
		// Check it's a direct child (no more slashes after prefix/)
		remainder := relPath[len(prefix)+1:]
		return !strings.Contains(remainder, "/")
	}

	// Handle plain directory name — match anything under it
	if !strings.Contains(pattern, "*") && !strings.Contains(pattern, ".") {
		return strings.HasPrefix(relPath, pattern+"/") || relPath == pattern
	}

	// Fallback: use filepath.Match for simple glob patterns
	matched, _ := filepath.Match(pattern, relPath)
	return matched
}
