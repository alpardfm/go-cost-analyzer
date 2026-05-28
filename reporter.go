package main

import (
	"encoding/json"

	"github.com/alpardfm/cost-efficient-go/types"
)

// Reporter formats and outputs the analysis report.
type Reporter interface {
	Render(report *Report) ([]byte, error)
}

// --- JSON Report Structs ---

type jsonReport struct {
	ProjectPath     string        `json:"project_path"`
	RegistryVersion string        `json:"registry_version"`
	Timestamp       string        `json:"timestamp"`
	OverallScore    interface{}   `json:"overall_score"`
	Patterns        []jsonPattern `json:"patterns"`
	UnparsableFiles []string      `json:"unparsable_files"`
	Disclaimer      string        `json:"disclaimer"`
}

type jsonPattern struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	Category         string        `json:"category"`
	Severity         string        `json:"severity"`
	Score            interface{}   `json:"score"`
	TotalOccurrences int           `json:"total_occurrences"`
	OptimalCount     int           `json:"optimal_count"`
	SuboptimalCount  int           `json:"suboptimal_count"`
	SuppressedCount  int           `json:"suppressed_count"`
	NotApplicable    bool          `json:"not_applicable"`
	ImpactLevel      string        `json:"impact_level"`
	EstimatedSavings string        `json:"estimated_savings"`
	Findings         []jsonFinding `json:"findings"`
}

type jsonFinding struct {
	File         string `json:"file"`
	Line         int    `json:"line"`
	Explanation  string `json:"explanation"`
	SuggestedFix string `json:"suggested_fix"`
	CodeContext  string `json:"code_context"`
	Reference    string `json:"reference"`
	Confidence   string `json:"confidence"`
}

// --- JSONReporter ---

// JSONReporter renders the analysis report as formatted JSON.
type JSONReporter struct{}

// Render converts a Report into indented JSON bytes.
func (r *JSONReporter) Render(report *Report) ([]byte, error) {
	jr := jsonReport{
		ProjectPath:     report.ProjectPath,
		RegistryVersion: report.RegistryVersion,
		Timestamp:       report.Timestamp,
		UnparsableFiles: report.UnparsableFiles,
		Disclaimer:      report.Disclaimer,
	}

	// OverallScore: -1 means N/A
	if report.OverallScore == -1 {
		jr.OverallScore = "N/A"
	} else {
		jr.OverallScore = report.OverallScore
	}

	// Ensure non-nil slices for clean JSON output
	if jr.UnparsableFiles == nil {
		jr.UnparsableFiles = []string{}
	}

	jr.Patterns = make([]jsonPattern, 0, len(report.Patterns))
	for _, p := range report.Patterns {
		jp := jsonPattern{
			ID:               p.Rule.ID,
			Name:             p.Rule.Name,
			Category:         categoryString(p.Rule.Category),
			Severity:         severityString(p.Rule.Severity),
			TotalOccurrences: p.TotalOccurrences,
			OptimalCount:     p.OptimalCount,
			SuboptimalCount:  p.SuboptimalCount,
			SuppressedCount:  p.Suppressed,
			NotApplicable:    p.NotApplicable,
			ImpactLevel:      p.ImpactLevel,
			EstimatedSavings: p.EstimatedSavings,
		}

		// Score: NotApplicable patterns output "N/A"
		if p.NotApplicable {
			jp.Score = "N/A"
		} else {
			jp.Score = p.Score
		}

		// Convert findings
		jp.Findings = make([]jsonFinding, 0, len(p.Findings))
		for _, f := range p.Findings {
			jf := jsonFinding{
				File:         f.FilePath,
				Line:         f.Line,
				Explanation:  f.Explanation,
				SuggestedFix: f.SuggestedFix,
				CodeContext:  f.CodeContext,
				Confidence:   confidenceString(f.Confidence),
			}
			// Add reference link from Rule if available
			if len(p.Rule.ReferenceLinks) > 0 {
				jf.Reference = p.Rule.ReferenceLinks[0]
			}
			jp.Findings = append(jp.Findings, jf)
		}

		jr.Patterns = append(jr.Patterns, jp)
	}

	return json.MarshalIndent(jr, "", "  ")
}

// --- Helper Functions ---

// severityString converts a types.Severity to its string representation.
func severityString(s types.Severity) string {
	switch s {
	case types.Critical:
		return "Critical"
	case types.Major:
		return "Major"
	case types.Minor:
		return "Minor"
	default:
		return "Unknown"
	}
}

// categoryString converts a types.Category to its string representation.
func categoryString(c types.Category) string {
	switch c {
	case types.Memory:
		return "Memory"
	case types.Concurrency:
		return "Concurrency"
	case types.IO:
		return "IO"
	case types.ErrorHandling:
		return "ErrorHandling"
	default:
		return "Unknown"
	}
}

// confidenceString converts a types.Confidence to its string representation.
func confidenceString(c types.Confidence) string {
	switch c {
	case types.ConfidenceHigh:
		return "high"
	case types.ConfidenceMedium:
		return "medium"
	default:
		return "low"
	}
}
