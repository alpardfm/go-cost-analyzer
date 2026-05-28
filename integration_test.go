package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestIntegration_CleanProject(t *testing.T) {
	cfg := AnalysisConfig{
		RootPath: "testdata/clean-project",
	}
	orch := NewOrchestrator(cfg)
	report, err := orch.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.OverallScore == -1 {
		t.Errorf("expected OverallScore >= 0, got -1 (N/A)")
	}
}

func TestIntegration_MessyProject(t *testing.T) {
	cfg := AnalysisConfig{
		RootPath: "testdata/messy-project",
	}
	orch := NewOrchestrator(cfg)
	report, err := orch.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	totalFindings := 0
	for _, p := range report.Patterns {
		totalFindings += len(p.Findings)
	}
	if totalFindings == 0 {
		t.Errorf("expected findings > 0 for messy-project, got 0")
	}
}

func TestIntegration_EdgeProject(t *testing.T) {
	cfg := AnalysisConfig{
		RootPath: "testdata/edge-project",
	}
	orch := NewOrchestrator(cfg)
	report, err := orch.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestIntegration_GeneratedSkipped(t *testing.T) {
	cfg := AnalysisConfig{
		RootPath: "testdata/generated",
	}
	orch := NewOrchestrator(cfg)
	report, err := orch.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, p := range report.Patterns {
		for _, f := range p.Findings {
			if strings.Contains(f.FilePath, "generated.pb.go") {
				t.Errorf("finding references generated.pb.go: %s at %s:%d", f.RuleID, f.FilePath, f.Line)
			}
		}
	}
}

func TestIntegration_NoGoMod(t *testing.T) {
	cfg := AnalysisConfig{
		RootPath: "testdata/no-gomod",
	}
	orch := NewOrchestrator(cfg)
	_, err := orch.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for project without go.mod, got nil")
	}
	if !strings.Contains(err.Error(), "no go.mod") {
		t.Errorf("expected error containing 'no go.mod', got: %v", err)
	}
}

func TestIntegration_NAProject(t *testing.T) {
	cfg := AnalysisConfig{
		RootPath: "testdata/na-project",
	}
	orch := NewOrchestrator(cfg)
	report, err := orch.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestIntegration_JSONOutput(t *testing.T) {
	cfg := AnalysisConfig{
		RootPath: "testdata/messy-project",
	}
	orch := NewOrchestrator(cfg)
	report, err := orch.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reporter := &JSONReporter{}
	output, err := reporter.Render(report)
	if err != nil {
		t.Fatalf("JSONReporter.Render failed: %v", err)
	}
	if !json.Valid(output) {
		t.Errorf("JSONReporter output is not valid JSON:\n%s", string(output))
	}
}

func TestIntegration_ThresholdPass(t *testing.T) {
	cfg := AnalysisConfig{
		RootPath:  "testdata/clean-project",
		Threshold: 0,
	}
	orch := NewOrchestrator(cfg)
	report, err := orch.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Threshold=0 means no minimum required; any score passes.
	// OverallScore should be >= threshold (0), so it would exit 0.
	if report.OverallScore < cfg.Threshold {
		t.Errorf("expected OverallScore >= %d (threshold), got %d", cfg.Threshold, report.OverallScore)
	}
}

func TestIntegration_PatternFilter(t *testing.T) {
	cfg := AnalysisConfig{
		RootPath: "testdata/messy-project",
		Patterns: []string{"CEG-016"},
	}
	orch := NewOrchestrator(cfg)
	report, err := orch.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, p := range report.Patterns {
		if p.Rule.ID != "CEG-016" {
			t.Errorf("expected only CEG-016 findings, but found pattern %s", p.Rule.ID)
		}
		for _, f := range p.Findings {
			if f.RuleID != "CEG-016" {
				t.Errorf("expected only CEG-016 findings, got finding with RuleID=%s", f.RuleID)
			}
		}
	}

	// Verify we actually got CEG-016 findings
	found := false
	for _, p := range report.Patterns {
		if p.Rule.ID == "CEG-016" && len(p.Findings) > 0 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected at least one CEG-016 finding in messy-project")
	}
}
