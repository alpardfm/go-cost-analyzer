package main

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/alpardfm/cost-efficient-go/registry"
	"github.com/alpardfm/cost-efficient-go/types"
)

// Disclaimer is the standard disclaimer appended to all reports.
const Disclaimer = "Estimasi berdasarkan analisis statis dan benchmark heuristic. " +
	"Actual savings bergantung pada runtime behavior dan deployment configuration."

// fileResult holds per-file analysis results collected from workers.
type fileResult struct {
	filePath string
	findings []types.Finding
	parseErr bool
	skipped  bool // generated or suppressed
}

// Orchestrator coordinates the full analysis pipeline.
type Orchestrator struct {
	config    AnalysisConfig
	detectors []FileDetector
}

// NewOrchestrator creates an Orchestrator with detectors from the registry.
// If config.Patterns is non-empty, only detectors whose Rule().ID or Rule().Name
// matches one of the patterns are included.
func NewOrchestrator(config AnalysisConfig) *Orchestrator {
	allDetectors := registry.AllDetectors()

	var filtered []types.Detector
	if len(config.Patterns) > 0 {
		patternSet := make(map[string]bool, len(config.Patterns))
		for _, p := range config.Patterns {
			patternSet[p] = true
		}
		for _, d := range allDetectors {
			rule := d.Rule()
			if patternSet[rule.ID] || patternSet[rule.Name] {
				filtered = append(filtered, d)
			}
		}
	} else {
		filtered = allDetectors
	}

	detectors := make([]FileDetector, 0, len(filtered))
	for _, d := range filtered {
		detectors = append(detectors, NewFileDetector(d))
	}

	return &Orchestrator{
		config:    config,
		detectors: detectors,
	}
}

// Run executes the full analysis pipeline. Context is used for cancellation.
func (o *Orchestrator) Run(ctx context.Context) (*Report, error) {
	// 1. Validate project
	scanner := NewScanner(ScanConfig{
		RootPath:     o.config.RootPath,
		ExcludeDirs:  o.config.Exclude,
		IncludeTests: o.config.IncludeTests,
	})

	if err := scanner.ValidateProject(); err != nil {
		return nil, err
	}

	// 2. Scan files
	files, err := scanner.Scan()
	if err != nil {
		return nil, err
	}

	if o.config.Verbose {
		fmt.Fprintf(os.Stderr, "[verbose] Found %d Go files to analyze\n", len(files))
	}

	// 3. Process files concurrently using bounded channel-based worker pool
	numWorkers := runtime.NumCPU()
	fileCh := make(chan string, numWorkers*2)
	resultsCh := make(chan fileResult, numWorkers*2)

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			o.worker(ctx, fileCh, resultsCh)
		}()
	}

	// Feed files to workers
	go func() {
		defer close(fileCh)
		for _, f := range files {
			select {
			case fileCh <- f:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// 4. Collect results
	var (
		allResults      []fileResult
		unparsableFiles []string
	)

	for {
		select {
		case result, ok := <-resultsCh:
			if !ok {
				goto doneCollecting
			}
			allResults = append(allResults, result)
			if result.parseErr {
				unparsableFiles = append(unparsableFiles, result.filePath)
			}
			if o.config.Verbose && !result.skipped && !result.parseErr {
				fmt.Fprintf(os.Stderr, "[verbose] Analyzed: %s (%d findings)\n", result.filePath, len(result.findings))
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
doneCollecting:

	// 5. Calculate scores
	totalFilesScanned := 0
	for _, r := range allResults {
		if !r.parseErr && !r.skipped {
			totalFilesScanned++
		}
	}

	// Build pattern results: for each detector, aggregate findings
	patternResults := make([]PatternResult, 0, len(o.detectors))
	for _, det := range o.detectors {
		rule := det.Rule()

		var findings []types.Finding
		filesWithFindings := make(map[string]bool)

		for _, r := range allResults {
			if r.parseErr || r.skipped {
				continue
			}
			for _, f := range r.findings {
				if f.RuleID == rule.ID {
					findings = append(findings, f)
					filesWithFindings[r.filePath] = true
				}
			}
		}

		pr := PatternResult{
			Rule:            rule,
			Findings:        findings,
			SuboptimalCount: len(findings),
			OptimalCount:    totalFilesScanned - len(filesWithFindings),
		}

		CalculatePatternScore(&pr)

		// 6. Calculate impact for patterns with findings
		if len(findings) > 0 {
			scale := ParseScale(o.config.Scale)
			pr.ImpactLevel, pr.EstimatedSavings = CalculateImpact(rule.ID, len(findings), scale)
		}

		patternResults = append(patternResults, pr)
	}

	// 7. Build Report
	overallScore := CalculateOverallScore(patternResults)

	report := &Report{
		ProjectPath:     o.config.RootPath,
		RegistryVersion: RegistryVersion,
		Timestamp:       time.Now().Format(time.RFC3339),
		OverallScore:    overallScore,
		Patterns:        patternResults,
		UnparsableFiles: unparsableFiles,
		Disclaimer:      Disclaimer,
	}

	return report, nil
}

// worker processes files from fileCh and sends results to resultsCh.
func (o *Orchestrator) worker(ctx context.Context, fileCh <-chan string, resultsCh chan<- fileResult) {
	// Build valid patterns set for suppression checking
	validPatterns := make(map[string]bool, len(o.detectors))
	for _, det := range o.detectors {
		rule := det.Rule()
		validPatterns[rule.ID] = true
		validPatterns[rule.Name] = true
	}

	for path := range fileCh {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return
		default:
		}

		result := o.processFile(path, validPatterns)
		resultsCh <- result
	}
}

// processFile handles parsing and detection for a single file.
func (o *Orchestrator) processFile(path string, validPatterns map[string]bool) fileResult {
	result := fileResult{filePath: path}

	// a. Parse file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		result.parseErr = true
		return result
	}

	// b. Check if generated file → skip
	if IsGeneratedFile(file) {
		result.skipped = true
		return result
	}

	// c. Check file-level suppression
	suppression := ParseSuppressions(file, fset, validPatterns)
	if suppression.FileSkipped {
		result.skipped = true
		return result
	}

	// d. Run all FileDetector adapters on the file
	var allFindings []types.Finding
	for _, det := range o.detectors {
		findings := det.DetectFile(fset, file)
		allFindings = append(allFindings, findings...)
	}

	// e. Filter findings through line-level suppression
	var filtered []types.Finding
	for _, f := range allFindings {
		// Look up the pattern name for suppression check
		patternName := f.RuleID
		// Also check by detector name
		if !suppression.IsSuppressed(f.Line, patternName) {
			filtered = append(filtered, f)
		}
	}

	result.findings = filtered
	return result
}
