package main

import (
	"os"
	"path/filepath"
	"strings"
)

// PatternDependencies maps pattern IDs to the Go module paths they require.
// If a pattern has dependencies listed here, it will only run if at least one
// of those modules is found in the target project's go.mod.
var PatternDependencies = map[string][]string{
	"CEG-004": {"database/sql", "github.com/jackc/pgx", "github.com/lib/pq", "gorm.io/gorm", "github.com/jmoiron/sqlx"},
	"CEG-015": {"github.com/redis/go-redis", "github.com/gomodule/redigo", "github.com/go-redis/redis"},
	"CEG-014": {"database/sql", "github.com/jackc/pgx", "github.com/lib/pq", "gorm.io/gorm", "github.com/jmoiron/sqlx"},
}

// ParseGoModDependencies reads go.mod from the project root and returns
// a set of all dependency module paths found (both direct and indirect).
// Also includes standard library packages that are imported (detected via import paths without dots).
func ParseGoModDependencies(projectRoot string) map[string]bool {
	deps := make(map[string]bool)

	goModPath := filepath.Join(projectRoot, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return deps
	}

	lines := strings.Split(string(data), "\n")
	inRequire := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "require (" {
			inRequire = true
			continue
		}
		if trimmed == ")" {
			inRequire = false
			continue
		}

		// Single-line require
		if strings.HasPrefix(trimmed, "require ") && !strings.Contains(trimmed, "(") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				deps[parts[1]] = true
			}
			continue
		}

		// Multi-line require block
		if inRequire && trimmed != "" && !strings.HasPrefix(trimmed, "//") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 1 {
				deps[parts[0]] = true
			}
		}
	}

	return deps
}

// ShouldSkipPattern checks if a pattern should be skipped because
// its required dependencies are not present in the target project.
// Returns true if the pattern has dependency requirements AND none are satisfied.
// Returns false (don't skip) if the pattern has no dependency requirements.
func ShouldSkipPattern(patternID string, projectDeps map[string]bool) bool {
	requiredDeps, hasRequirements := PatternDependencies[patternID]
	if !hasRequirements {
		return false // No requirements — always run
	}

	// Check if any required dependency is present
	for _, dep := range requiredDeps {
		if projectDeps[dep] {
			return false // Found a matching dependency — run the pattern
		}
		// Also check prefix match (e.g., "github.com/jackc/pgx" matches "github.com/jackc/pgx/v5")
		for projectDep := range projectDeps {
			if strings.HasPrefix(projectDep, dep) || strings.HasPrefix(dep, projectDep) {
				return false
			}
		}
	}

	return true // No matching dependencies — skip
}
