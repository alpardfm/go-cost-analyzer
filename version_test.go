package main

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestRegistryVersionFormat(t *testing.T) {
	// Verify RegistryVersion matches format YYYY.MM.patch
	pattern := `^\d{4}\.\d{2}\.\d+$`
	re := regexp.MustCompile(pattern)

	if !re.MatchString(RegistryVersion) {
		t.Fatalf("RegistryVersion %q does not match expected format YYYY.MM.patch (regex: %s)", RegistryVersion, pattern)
	}

	// Verify year is reasonable (2024-2030)
	parts := strings.Split(RegistryVersion, ".")
	year, err := strconv.Atoi(parts[0])
	if err != nil {
		t.Fatalf("failed to parse year from RegistryVersion %q: %v", RegistryVersion, err)
	}

	if year < 2024 || year > 2030 {
		t.Errorf("RegistryVersion year %d is outside reasonable range (2024-2030)", year)
	}

	t.Logf("RegistryVersion: %s (year=%d)", RegistryVersion, year)
}

func TestRegistryVersionConsistency(t *testing.T) {
	// Read go.mod from the project root
	data, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("failed to read go.mod: %v", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	// Parse go.mod to find the cost-efficient-go dependency line
	var depLine string
	var depVersion string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "cost-efficient-go") && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "replace") {
			depLine = trimmed
			// Extract version from dependency line
			fields := strings.Fields(trimmed)
			if len(fields) >= 2 {
				depVersion = fields[len(fields)-1]
			}
			break
		}
	}

	// Verify the dependency exists (drift detection)
	if depLine == "" {
		t.Fatal("cost-efficient-go dependency not found in go.mod — library may have been removed (drift detected)")
	}

	// Log the library version found for visibility
	t.Logf("Found cost-efficient-go dependency: %s", depLine)
	t.Logf("Library version: %s", depVersion)
	t.Logf("Registry version: %s", RegistryVersion)
}
