package main

import (
	"context"
	"sort"
	"testing"

	"github.com/alpardfm/cost-efficient-go/types"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestProperty_ConcurrentSafety(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("concurrent analysis produces same findings regardless of run", prop.ForAll(
		func(dummy int) bool {
			cfg := AnalysisConfig{
				RootPath: "testdata/messy-project",
			}

			// Run analysis twice
			orch1 := NewOrchestrator(cfg)
			report1, err1 := orch1.Run(context.Background())
			if err1 != nil {
				return false
			}

			orch2 := NewOrchestrator(cfg)
			report2, err2 := orch2.Run(context.Background())
			if err2 != nil {
				return false
			}

			// Compare findings (set equality)
			findings1 := collectFindings(report1)
			findings2 := collectFindings(report2)

			if len(findings1) != len(findings2) {
				return false
			}

			sort.Strings(findings1)
			sort.Strings(findings2)

			for i := range findings1 {
				if findings1[i] != findings2[i] {
					return false
				}
			}

			return report1.OverallScore == report2.OverallScore
		},
		gen.IntRange(0, 99),
	))

	properties.TestingRun(t)
}

func collectFindings(report *Report) []string {
	var result []string
	for _, p := range report.Patterns {
		for _, f := range p.Findings {
			key := f.RuleID + ":" + f.FilePath + ":" + string(rune(f.Line))
			result = append(result, key)
		}
	}
	return result
}

// Ensure types.Finding is used (avoid unused import)
var _ types.Finding
