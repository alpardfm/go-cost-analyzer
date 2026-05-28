package main

// AnalysisConfig holds all configuration for a single analysis run.
type AnalysisConfig struct {
	ScanConfig
	RootPath     string
	Format       string   // Output format: "text" or "json"
	Verbose      bool     // Print per-file progress to stderr
	Patterns     []string // Filter: only these patterns (empty = all)
	Threshold    int      // Minimum overall score (0 = disabled)
	Exclude      []string // Additional directories to exclude
	IncludeTests bool     // Whether to include *_test.go files
	Scale        string   // Impact projection scale (e.g., "1M", "10M", "100M")
	OutputFile   string   // Output file path (empty = stdout)
}
