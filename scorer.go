package main

import "github.com/alpardfm/cost-efficient-go/types"

// PatternResult holds the analysis result for a single pattern.
type PatternResult struct {
	Rule             types.Rule
	Score            int             // 0-100
	TotalOccurrences int             // Total relevant occurrences found
	OptimalCount     int             // Occurrences following best practice
	SuboptimalCount  int             // Occurrences not following best practice
	Findings         []types.Finding // Suboptimal findings with location
	Suppressed       int             // Count of suppressed occurrences
	NotApplicable    bool            // true if no relevant occurrences
	ImpactLevel      string          // "High", "Medium", "Low"
	EstimatedSavings string          // Human-readable savings estimate
}

// Report is the complete analysis output.
type Report struct {
	ProjectPath     string
	RegistryVersion string
	Timestamp       string // RFC3339
	OverallScore    int    // 0-100 weighted average
	Patterns        []PatternResult
	UnparsableFiles []string // Files with syntax errors
	Disclaimer      string
}

// SeverityWeight maps Severity to scoring weight.
var SeverityWeight = map[types.Severity]int{
	types.Critical: 3,
	types.Major:    2,
	types.Minor:    1,
}

// CalculatePatternScore computes the score for a single pattern result.
// Score = (OptimalCount / TotalOccurrences) * 100.
// If TotalOccurrences == 0, the pattern is marked NotApplicable with Score = 0.
func CalculatePatternScore(result *PatternResult) {
	result.TotalOccurrences = result.OptimalCount + result.SuboptimalCount
	if result.TotalOccurrences == 0 {
		result.NotApplicable = true
		result.Score = 0
		return
	}
	result.Score = (result.OptimalCount * 100) / result.TotalOccurrences
}

// CalculateOverallScore computes a weighted average score across all applicable patterns.
// Patterns where NotApplicable == true are excluded.
// Returns -1 if no applicable patterns remain (represents N/A).
// Weight is determined by SeverityWeight (Critical=3, Major=2, Minor=1).
func CalculateOverallScore(patterns []PatternResult) int {
	var weightedSum int
	var totalWeight int

	for _, p := range patterns {
		if p.NotApplicable {
			continue
		}
		weight := SeverityWeight[p.Rule.Severity]
		weightedSum += p.Score * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return -1
	}

	return weightedSum / totalWeight
}
