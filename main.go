package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
)

const usage = `go-cost-analyzer - Analyze Go projects for cost-efficiency anti-patterns

Usage:
  go-cost-analyzer [flags] <path-to-go-project>

Flags:
  -o, --output         Output file path (default: stdout)
  -f, --format         Output format: "text" or "json" (default: "text")
  -v, --verbose        Enable verbose mode (default: false)
  -p, --patterns       Comma-separated pattern filter (default: all)
  -t, --threshold      Minimum score 0-100 (default: 0)
      --exclude        Comma-separated additional dirs to exclude
      --include-tests  Include *_test.go files (default: false)
      --scale          Impact scale: "1M", "10M", "100M" (default: "1M")
      --diff           Only analyze files changed vs this branch (e.g., main)

Examples:
  go-cost-analyzer ./my-project
  go-cost-analyzer -f json -o report.json ./my-project
  go-cost-analyzer --threshold 80 --patterns slice-performance,sync-pool ./my-project
  go-cost-analyzer --diff main ./my-project
`

func main() {
	// If no arguments at all, show help and exit 0
	if len(os.Args) == 1 {
		fmt.Print(usage)
		os.Exit(0)
	}

	// Define flags
	var (
		output       string
		format       string
		verbose      bool
		patterns     string
		threshold    int
		exclude      string
		includeTests bool
		scale        string
		diffBase     string
	)

	flag.StringVar(&output, "output", "", "Output file path")
	flag.StringVar(&output, "o", "", "Output file path (shorthand)")
	flag.StringVar(&format, "format", "text", "Output format: text or json")
	flag.StringVar(&format, "f", "text", "Output format (shorthand)")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose mode")
	flag.BoolVar(&verbose, "v", false, "Enable verbose mode (shorthand)")
	flag.StringVar(&patterns, "patterns", "", "Comma-separated pattern filter")
	flag.StringVar(&patterns, "p", "", "Comma-separated pattern filter (shorthand)")
	flag.IntVar(&threshold, "threshold", 0, "Minimum score 0-100")
	flag.IntVar(&threshold, "t", 0, "Minimum score 0-100 (shorthand)")
	flag.StringVar(&exclude, "exclude", "", "Comma-separated additional dirs to exclude")
	flag.BoolVar(&includeTests, "include-tests", false, "Include *_test.go files")
	flag.StringVar(&scale, "scale", "1M", `Impact scale: "1M", "10M", "100M"`)
	flag.StringVar(&diffBase, "diff", "", "Only analyze files changed vs this branch (e.g., main)")

	flag.Usage = func() {
		fmt.Print(usage)
	}

	flag.Parse()

	// Positional argument: path to Go project
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: path to Go project is required\n")
		fmt.Fprintf(os.Stderr, "Run 'go-cost-analyzer' without arguments for usage information\n")
		os.Exit(1)
	}
	rootPath := args[0]

	// Validate --format
	if format != "text" && format != "json" {
		fmt.Fprintf(os.Stderr, "Error: --format must be \"text\" or \"json\", got %q\n", format)
		os.Exit(1)
	}

	// Validate --threshold
	if threshold < 0 || threshold > 100 {
		fmt.Fprintf(os.Stderr, "Error: --threshold must be between 0 and 100, got %d\n", threshold)
		os.Exit(1)
	}

	// Validate --scale
	validScales := map[string]bool{"1M": true, "10M": true, "100M": true}
	if !validScales[scale] {
		fmt.Fprintf(os.Stderr, "Error: --scale must be \"1M\", \"10M\", or \"100M\", got %q\n", scale)
		os.Exit(1)
	}

	// Parse comma-separated values
	var patternList []string
	if patterns != "" {
		for _, p := range strings.Split(patterns, ",") {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				patternList = append(patternList, trimmed)
			}
		}
	}

	var excludeList []string
	if exclude != "" {
		for _, e := range strings.Split(exclude, ",") {
			trimmed := strings.TrimSpace(e)
			if trimmed != "" {
				excludeList = append(excludeList, trimmed)
			}
		}
	}

	// Build AnalysisConfig
	config := AnalysisConfig{
		ScanConfig: ScanConfig{
			RootPath:     rootPath,
			ExcludeDirs:  excludeList,
			IncludeTests: includeTests,
		},
		RootPath:     rootPath,
		Format:       format,
		Verbose:      verbose,
		Patterns:     patternList,
		Threshold:    threshold,
		Exclude:      excludeList,
		IncludeTests: includeTests,
		Scale:        scale,
		OutputFile:   output,
		DiffBase:     diffBase,
	}

	// Load project config file
	projectConfig, err := LoadProjectConfig(rootPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	// Merge project config (CLI flags take precedence)
	MergeConfig(&config, projectConfig,
		exclude != "",    // cliExcludeSet
		patterns != "",   // cliPatternsSet
		threshold != 0,   // cliThresholdSet
		format != "text", // cliFormatSet (non-default means explicitly set)
		scale != "1M",    // cliScaleSet
		includeTests,     // cliIncludeTestsSet
	)

	// Create and run orchestrator
	orch := NewOrchestrator(config)
	report, err := orch.Run(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	// Select reporter
	var reporter Reporter
	switch config.Format {
	case "json":
		reporter = &JSONReporter{}
	default:
		reporter = &TextReporter{}
	}

	// Render report
	reportOutput, err := reporter.Render(report)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering report: %v\n", err)
		os.Exit(2)
	}

	// Write output
	if config.OutputFile != "" {
		if err := os.WriteFile(config.OutputFile, reportOutput, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(2)
		}
	} else {
		fmt.Print(string(reportOutput))
	}

	// Check threshold
	if config.Threshold > 0 && report.OverallScore >= 0 && report.OverallScore < config.Threshold {
		os.Exit(1)
	}
	// N/A score always passes threshold (exit 0)
}
