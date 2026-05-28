package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/alpardfm/cost-efficient-go/types"
)

// newTestReport creates a minimal Report for testing purposes.
func newTestReport() *Report {
	return &Report{
		ProjectPath:     "/tmp/test-project",
		RegistryVersion: "v1.0.0",
		Timestamp:       "2024-01-15T10:30:00Z",
		OverallScore:    75,
		Patterns: []PatternResult{
			{
				Rule: types.Rule{
					ID:             "CEG-001",
					Name:           "Slice Pre-allocation",
					Severity:       types.Major,
					Category:       types.Memory,
					ReferenceLinks: []string{"https://example.com/slice"},
				},
				Score:            80,
				TotalOccurrences: 10,
				OptimalCount:     8,
				SuboptimalCount:  2,
				ImpactLevel:      "Medium",
				Findings: []types.Finding{
					{
						RuleID:       "CEG-001",
						FilePath:     "main.go",
						Line:         42,
						Explanation:  "Slice grows without pre-allocation",
						SuggestedFix: "Use make([]T, 0, expectedCap)",
						Severity:     types.Major,
						Category:     types.Memory,
					},
				},
			},
			{
				Rule: types.Rule{
					ID:       "CEG-002",
					Name:     "Connection Pooling",
					Severity: types.Minor,
					Category: types.IO,
				},
				Score:            90,
				TotalOccurrences: 5,
				OptimalCount:     4,
				SuboptimalCount:  1,
				ImpactLevel:      "Low",
			},
		},
		UnparsableFiles: []string{"broken.go"},
		Disclaimer:      "This is a static analysis tool. Results may vary.",
	}
}

func TestJSONReporter_Schema(t *testing.T) {
	report := newTestReport()
	reporter := &JSONReporter{}

	output, err := reporter.Render(report)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	requiredFields := []string{
		"project_path",
		"registry_version",
		"timestamp",
		"overall_score",
		"patterns",
		"unparsable_files",
		"disclaimer",
	}

	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("Required field %q missing from JSON output", field)
		}
	}
}

func TestJSONReporter_RoundTrip(t *testing.T) {
	report := newTestReport()
	reporter := &JSONReporter{}

	output, err := reporter.Render(report)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	var jr jsonReport
	if err := json.Unmarshal(output, &jr); err != nil {
		t.Fatalf("Failed to unmarshal into jsonReport: %v", err)
	}

	if jr.ProjectPath != report.ProjectPath {
		t.Errorf("ProjectPath: got %q, want %q", jr.ProjectPath, report.ProjectPath)
	}
	if jr.RegistryVersion != report.RegistryVersion {
		t.Errorf("RegistryVersion: got %q, want %q", jr.RegistryVersion, report.RegistryVersion)
	}
	if jr.Timestamp != report.Timestamp {
		t.Errorf("Timestamp: got %q, want %q", jr.Timestamp, report.Timestamp)
	}
	if jr.Disclaimer != report.Disclaimer {
		t.Errorf("Disclaimer: got %q, want %q", jr.Disclaimer, report.Disclaimer)
	}

	// OverallScore is an interface{}, check as float64 (JSON number)
	scoreFloat, ok := jr.OverallScore.(float64)
	if !ok {
		t.Fatalf("OverallScore: expected float64, got %T", jr.OverallScore)
	}
	if int(scoreFloat) != report.OverallScore {
		t.Errorf("OverallScore: got %v, want %d", scoreFloat, report.OverallScore)
	}

	if len(jr.Patterns) != len(report.Patterns) {
		t.Fatalf("Patterns count: got %d, want %d", len(jr.Patterns), len(report.Patterns))
	}

	// Verify first pattern values
	if jr.Patterns[0].ID != report.Patterns[0].Rule.ID {
		t.Errorf("Pattern[0].ID: got %q, want %q", jr.Patterns[0].ID, report.Patterns[0].Rule.ID)
	}
	if jr.Patterns[0].Name != report.Patterns[0].Rule.Name {
		t.Errorf("Pattern[0].Name: got %q, want %q", jr.Patterns[0].Name, report.Patterns[0].Rule.Name)
	}

	if len(jr.UnparsableFiles) != len(report.UnparsableFiles) {
		t.Errorf("UnparsableFiles count: got %d, want %d", len(jr.UnparsableFiles), len(report.UnparsableFiles))
	}
}

func TestJSONReporter_NAScore(t *testing.T) {
	report := newTestReport()
	report.OverallScore = -1

	reporter := &JSONReporter{}
	output, err := reporter.Render(report)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify the raw JSON contains "overall_score": "N/A"
	if !strings.Contains(string(output), `"overall_score": "N/A"`) {
		t.Errorf("Expected JSON to contain '\"overall_score\": \"N/A\"', got:\n%s", string(output))
	}
}

func TestJSONReporter_NAPattern(t *testing.T) {
	report := &Report{
		ProjectPath:     "/tmp/test",
		RegistryVersion: "v1.0.0",
		Timestamp:       "2024-01-15T10:30:00Z",
		OverallScore:    50,
		Patterns: []PatternResult{
			{
				Rule: types.Rule{
					ID:       "CEG-003",
					Name:     "Goroutine Leak",
					Severity: types.Critical,
					Category: types.Concurrency,
				},
				NotApplicable: true,
				Score:         0,
			},
		},
		UnparsableFiles: []string{},
		Disclaimer:      "Disclaimer text",
	}

	reporter := &JSONReporter{}
	output, err := reporter.Render(report)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify the pattern has "score": "N/A"
	if !strings.Contains(string(output), `"score": "N/A"`) {
		t.Errorf("Expected JSON to contain '\"score\": \"N/A\"' for NotApplicable pattern, got:\n%s", string(output))
	}
}

func TestTextReporter_ContainsScore(t *testing.T) {
	report := newTestReport()
	reporter := &TextReporter{}

	output, err := reporter.Render(report)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	text := stripColors(string(output))
	if !strings.Contains(text, "75/100") {
		t.Errorf("Expected text output to contain score '75/100', got:\n%s", text)
	}
}

func TestTextReporter_FindingsSortedBySeverity(t *testing.T) {
	report := &Report{
		ProjectPath:     "/tmp/test",
		RegistryVersion: "v1.0.0",
		Timestamp:       "2024-01-15T10:30:00Z",
		OverallScore:    60,
		Patterns: []PatternResult{
			{
				Rule: types.Rule{
					ID:       "CEG-010",
					Name:     "Minor Pattern",
					Severity: types.Minor,
					Category: types.Memory,
				},
				Score:            70,
				TotalOccurrences: 5,
				OptimalCount:     3,
				SuboptimalCount:  2,
				ImpactLevel:      "Low",
				Findings: []types.Finding{
					{
						RuleID:   "CEG-010",
						FilePath: "minor.go",
						Line:     10,
						Severity: types.Minor,
					},
				},
			},
			{
				Rule: types.Rule{
					ID:       "CEG-005",
					Name:     "Critical Pattern",
					Severity: types.Critical,
					Category: types.Concurrency,
				},
				Score:            30,
				TotalOccurrences: 10,
				OptimalCount:     3,
				SuboptimalCount:  7,
				ImpactLevel:      "High",
				Findings: []types.Finding{
					{
						RuleID:   "CEG-005",
						FilePath: "critical.go",
						Line:     20,
						Severity: types.Critical,
					},
				},
			},
		},
		UnparsableFiles: []string{},
		Disclaimer:      "Disclaimer",
	}

	reporter := &TextReporter{}
	output, err := reporter.Render(report)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	text := stripColors(string(output))

	criticalIdx := strings.Index(text, "[Critical]")
	minorIdx := strings.Index(text, "[Minor]")

	if criticalIdx == -1 {
		t.Fatal("Expected output to contain '[Critical]'")
	}
	if minorIdx == -1 {
		t.Fatal("Expected output to contain '[Minor]'")
	}
	if criticalIdx >= minorIdx {
		t.Errorf("Expected Critical to appear before Minor in output (Critical at %d, Minor at %d)", criticalIdx, minorIdx)
	}
}

func TestTextReporter_NAScoreRendered(t *testing.T) {
	report := newTestReport()
	report.OverallScore = -1

	reporter := &TextReporter{}
	output, err := reporter.Render(report)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	text := stripColors(string(output))
	if !strings.Contains(text, "N/A") {
		t.Errorf("Expected text output to contain 'N/A' for score -1, got:\n%s", text)
	}
}

func TestTextReporter_StripColors(t *testing.T) {
	input := "\033[31mHello\033[0m \033[1m\033[32mWorld\033[0m"
	result := stripColors(input)

	if strings.Contains(result, "\033") {
		t.Errorf("stripColors did not remove all ANSI codes, got: %q", result)
	}

	expected := "Hello World"
	if result != expected {
		t.Errorf("stripColors: got %q, want %q", result, expected)
	}
}
