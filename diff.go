package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetChangedFiles returns the list of .go files that changed compared to the base branch.
// Uses `git diff --name-only --diff-filter=ACMR <base>...HEAD` to get changed files.
// Returns absolute paths. Returns error if git command fails.
func GetChangedFiles(projectRoot string, baseBranch string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=ACMR", baseBranch+"...HEAD")
	cmd.Dir = projectRoot

	output, err := cmd.Output()
	if err != nil {
		// Try without ...HEAD (for uncommitted changes)
		cmd2 := exec.Command("git", "diff", "--name-only", "--diff-filter=ACMR", baseBranch)
		cmd2.Dir = projectRoot
		output, err = cmd2.Output()
		if err != nil {
			return nil, fmt.Errorf("git diff failed: %v", err)
		}
	}

	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line == "" {
			continue
		}
		if strings.HasSuffix(line, ".go") {
			absPath := filepath.Join(projectRoot, line)
			files = append(files, absPath)
		}
	}

	return files, nil
}
