package main

import (
	"testing"

	"github.com/alpardfm/cost-efficient-go/types"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestProperty_ScoreDeterminism(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	severities := []types.Severity{types.Minor, types.Major, types.Critical}

	properties.Property("scoring produces identical results on repeated runs", prop.ForAll(
		func(optimal, suboptimal int, sevIdx int) bool {
			if optimal < 0 {
				optimal = 0
			}
			if suboptimal < 0 {
				suboptimal = 0
			}

			sev := severities[sevIdx%len(severities)]

			patterns := []PatternResult{
				{
					Rule:            types.Rule{Severity: sev},
					OptimalCount:    optimal,
					SuboptimalCount: suboptimal,
				},
			}

			// Run scoring twice
			p1 := make([]PatternResult, len(patterns))
			copy(p1, patterns)
			CalculatePatternScore(&p1[0])
			score1 := CalculateOverallScore(p1)

			p2 := make([]PatternResult, len(patterns))
			copy(p2, patterns)
			CalculatePatternScore(&p2[0])
			score2 := CalculateOverallScore(p2)

			return score1 == score2 && p1[0].Score == p2[0].Score
		},
		gen.IntRange(0, 100),
		gen.IntRange(0, 100),
		gen.IntRange(0, 2),
	))

	properties.TestingRun(t)
}
