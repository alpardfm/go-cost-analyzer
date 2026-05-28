package main

import (
	"testing"

	"github.com/alpardfm/cost-efficient-go/types"
)

func TestCalculatePatternScore_Normal(t *testing.T) {
	result := &PatternResult{
		OptimalCount:    7,
		SuboptimalCount: 3,
	}
	CalculatePatternScore(result)
	if result.Score != 70 {
		t.Errorf("expected Score=70, got %d", result.Score)
	}
	if result.NotApplicable {
		t.Error("expected NotApplicable=false")
	}
}

func TestCalculatePatternScore_AllOptimal(t *testing.T) {
	result := &PatternResult{
		OptimalCount:    10,
		SuboptimalCount: 0,
	}
	CalculatePatternScore(result)
	if result.Score != 100 {
		t.Errorf("expected Score=100, got %d", result.Score)
	}
	if result.NotApplicable {
		t.Error("expected NotApplicable=false")
	}
}

func TestCalculatePatternScore_AllSuboptimal(t *testing.T) {
	result := &PatternResult{
		OptimalCount:    0,
		SuboptimalCount: 5,
	}
	CalculatePatternScore(result)
	if result.Score != 0 {
		t.Errorf("expected Score=0, got %d", result.Score)
	}
	if result.NotApplicable {
		t.Error("expected NotApplicable=false")
	}
}

func TestCalculatePatternScore_NotApplicable(t *testing.T) {
	result := &PatternResult{
		OptimalCount:    0,
		SuboptimalCount: 0,
	}
	CalculatePatternScore(result)
	if !result.NotApplicable {
		t.Error("expected NotApplicable=true")
	}
	if result.Score != 0 {
		t.Errorf("expected Score=0, got %d", result.Score)
	}
}

func TestCalculateOverallScore_Mixed(t *testing.T) {
	patterns := []PatternResult{
		{
			Rule:  types.Rule{Severity: types.Critical},
			Score: 80,
		},
		{
			Rule:  types.Rule{Severity: types.Major},
			Score: 60,
		},
		{
			Rule:  types.Rule{Severity: types.Minor},
			Score: 40,
		},
	}
	// Minor score 40 is floored to 50 (ScoreFloor for Minor = 50)
	// Weighted: (80*3 + 60*2 + 50*1) / (3+2+1) = (240+120+50)/6 = 410/6 = 68
	score := CalculateOverallScore(patterns)
	expected := 68
	if score != expected {
		t.Errorf("expected overall score=%d, got %d", expected, score)
	}
}

func TestCalculateOverallScore_AllNA(t *testing.T) {
	patterns := []PatternResult{
		{Rule: types.Rule{Severity: types.Critical}, NotApplicable: true, Score: 0},
		{Rule: types.Rule{Severity: types.Major}, NotApplicable: true, Score: 0},
	}
	score := CalculateOverallScore(patterns)
	if score != -1 {
		t.Errorf("expected -1 for all N/A patterns, got %d", score)
	}
}

func TestCalculateOverallScore_SinglePattern(t *testing.T) {
	patterns := []PatternResult{
		{
			Rule:  types.Rule{Severity: types.Critical},
			Score: 80,
		},
	}
	score := CalculateOverallScore(patterns)
	if score != 80 {
		t.Errorf("expected overall score=80, got %d", score)
	}
}

func TestCalculateImpact_High(t *testing.T) {
	// CEG-001: BytesSavedPerOccurrence=4096
	// totalBytes = 4096 * 100 * 10_000_000 / 1_000_000 = 4,096,000 bytes
	// Wait, let's recalculate: 4096 * 100 * 10_000_000 / 1_000_000 = 4,096,000,000 / 1,000,000 = 4096
	// Need larger numbers. Use scale=100M, occurrences=100
	// totalBytes = 4096 * 100 * 100_000_000 / 1_000_000 = 40,960,000 > 10*1024*1024 = 10,485,760 → High
	level, savings := CalculateImpact("CEG-001", 100, ParseScale("100M"))
	if level != "High" {
		t.Errorf("expected level=High, got %s", level)
	}
	if savings == "" {
		t.Error("expected non-empty savings string")
	}
}

func TestCalculateImpact_Medium(t *testing.T) {
	// CEG-003: BytesSavedPerOccurrence=256
	// totalBytes = 256 * 50 * 100_000_000 / 1_000_000 = 1,280,000
	// 1,280,000 > 1*1024*1024 (1,048,576) → Medium
	level, savings := CalculateImpact("CEG-003", 50, ParseScale("100M"))
	if level != "Medium" {
		t.Errorf("expected level=Medium, got %s", level)
	}
	if savings == "" {
		t.Error("expected non-empty savings string")
	}
}

func TestCalculateImpact_Low(t *testing.T) {
	// CEG-010: BytesSavedPerOccurrence=8
	// totalBytes = 8 * 5 * 1_000_000 / 1_000_000 = 40
	// 40 < 1*1024*1024 → Low
	level, savings := CalculateImpact("CEG-010", 5, ParseScale("1M"))
	if level != "Low" {
		t.Errorf("expected level=Low, got %s", level)
	}
	if savings == "" {
		t.Error("expected non-empty savings string")
	}
}

func TestCalculateImpact_UnknownPattern(t *testing.T) {
	level, savings := CalculateImpact("CEG-999", 10, ParseScale("1M"))
	if level != "Low" {
		t.Errorf("expected level=Low for unknown pattern, got %s", level)
	}
	if savings != "unknown" {
		t.Errorf("expected savings=unknown, got %s", savings)
	}
}

func TestParseScale(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1M", 1_000_000},
		{"10M", 10_000_000},
		{"100M", 100_000_000},
		{"invalid", 1_000_000},
	}
	for _, tc := range tests {
		got := ParseScale(tc.input)
		if got != tc.expected {
			t.Errorf("ParseScale(%q) = %d, want %d", tc.input, got, tc.expected)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{500, "500 B"},
		{2048, "2.0 KB"},
		{5242880, "5.0 MB"},
		{2147483648, "2.0 GB"},
	}
	for _, tc := range tests {
		got := FormatBytes(tc.input)
		if got != tc.expected {
			t.Errorf("FormatBytes(%d) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}
