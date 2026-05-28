package main

import "fmt"

// PatternImpact holds heuristic impact data for a single pattern.
// All numeric fields have explicit units in their names.
type PatternImpact struct {
	PatternID               string
	BytesSavedPerOccurrence int64   // bytes saved per single fix
	AllocsReducedPercent    float64 // 0.0-1.0 ratio (e.g., 0.91 = 91% reduction)
	Description             string  // Human-readable impact description
	Source                  string  // Audit trail: "README benchmark, Go X.Y, hardware, date"
}

// ImpactRegistry maps pattern IDs to their impact data.
// All values derived from pattern README benchmarks.
var ImpactRegistry = map[string]PatternImpact{
	"CEG-001": {PatternID: "CEG-001", BytesSavedPerOccurrence: 4096, AllocsReducedPercent: 0.99, Description: "Batch processing reduces round-trips by 99%", Source: "README benchmark, Go 1.24, MacBook Air M2, 2025-01"},
	"CEG-002": {PatternID: "CEG-002", BytesSavedPerOccurrence: 8192, AllocsReducedPercent: 0.99, Description: "Cache hit 21,872x faster than DB query", Source: "README benchmark, Go 1.24, MacBook Air M2, 2025-01"},
	"CEG-003": {PatternID: "CEG-003", BytesSavedPerOccurrence: 256, AllocsReducedPercent: 0.30, Description: "Buffered channels 3-4x faster than unbuffered", Source: "README benchmark, Go 1.24, MacBook Air M2, 2025-01"},
	"CEG-004": {PatternID: "CEG-004", BytesSavedPerOccurrence: 32768, AllocsReducedPercent: 0.975, Description: "Connection pooling 2.7x faster, 40x less memory per request", Source: "README benchmark, Go 1.24, MacBook Air M2, 2025-01"},
	"CEG-005": {PatternID: "CEG-005", BytesSavedPerOccurrence: 2048, AllocsReducedPercent: 0.15, Description: "Context cancellation saves 15% CPU at 20% cancel rate", Source: "README benchmark, Go 1.24, MacBook Air M2, 2025-01"},
	"CEG-006": {PatternID: "CEG-006", BytesSavedPerOccurrence: 512, AllocsReducedPercent: 0.90, Description: "Structured logging 10x+ faster, zero allocations", Source: "README benchmark, Go 1.24, MacBook Air M2, 2025-01"},
	"CEG-007": {PatternID: "CEG-007", BytesSavedPerOccurrence: 64, AllocsReducedPercent: 0.95, Description: "Sentinel errors eliminate 5M allocs/day", Source: "README benchmark, Go 1.24, MacBook Air M2, 2025-01"},
	"CEG-008": {PatternID: "CEG-008", BytesSavedPerOccurrence: 4096, AllocsReducedPercent: 0.99, Description: "Prevent 172-691 MB/day memory waste from leaked goroutines", Source: "README benchmark, Go 1.24, MacBook Air M2, 2025-01"},
	"CEG-009": {PatternID: "CEG-009", BytesSavedPerOccurrence: 1024, AllocsReducedPercent: 0.60, Description: "HTTP client reuse 2.6x faster with body drain", Source: "README benchmark, Go 1.24, MacBook Air M2, 2025-01"},
	"CEG-010": {PatternID: "CEG-010", BytesSavedPerOccurrence: 8, AllocsReducedPercent: 0.03, Description: "Interface dispatch ~1-3ns/call overhead", Source: "README benchmark, Go 1.24, MacBook Air M2, 2025-01"},
	"CEG-011": {PatternID: "CEG-011", BytesSavedPerOccurrence: 2048, AllocsReducedPercent: 0.77, Description: "Streaming JSON 2x faster, 77% less bandwidth", Source: "README benchmark, Go 1.24, MacBook Air M2, 2025-01"},
	"CEG-012": {PatternID: "CEG-012", BytesSavedPerOccurrence: 512, AllocsReducedPercent: 0.40, Description: "Map pre-allocation reduces hidden memory overhead", Source: "README benchmark, Go 1.24, MacBook Air M2, 2025-01"},
	"CEG-013": {PatternID: "CEG-013", BytesSavedPerOccurrence: 128, AllocsReducedPercent: 0.20, Description: "Correct profiling techniques for accurate measurement", Source: "README benchmark, Go 1.24, MacBook Air M2, 2025-01"},
	"CEG-014": {PatternID: "CEG-014", BytesSavedPerOccurrence: 16384, AllocsReducedPercent: 0.98, Description: "Query optimization 4.4x faster SELECT, 50x with batch", Source: "README benchmark, Go 1.24, MacBook Air M2, 2025-01"},
	"CEG-015": {PatternID: "CEG-015", BytesSavedPerOccurrence: 4096, AllocsReducedPercent: 0.80, Description: "Redis pipeline 50-100x faster, 80% latency reduction", Source: "README benchmark, Go 1.24, MacBook Air M2, 2025-01"},
	"CEG-016": {PatternID: "CEG-016", BytesSavedPerOccurrence: 1024, AllocsReducedPercent: 0.91, Description: "Slice pre-allocation 4x faster, 91% fewer allocations", Source: "README benchmark, Go 1.24, MacBook Air M2, 2025-01"},
	"CEG-017": {PatternID: "CEG-017", BytesSavedPerOccurrence: 2048, AllocsReducedPercent: 0.95, Description: "strings.Builder 5-20x faster than + at 100+ concats", Source: "README benchmark, Go 1.24, MacBook Air M2, 2025-01"},
	"CEG-018": {PatternID: "CEG-018", BytesSavedPerOccurrence: 8, AllocsReducedPercent: 0.25, Description: "Struct alignment saves 25% memory via field reordering", Source: "README benchmark, Go 1.24, MacBook Air M2, 2025-01"},
	"CEG-019": {PatternID: "CEG-019", BytesSavedPerOccurrence: 4096, AllocsReducedPercent: 0.99, Description: "sync.Pool 50%+ GC reduction, 99% fewer allocations", Source: "README benchmark, Go 1.24, MacBook Air M2, 2025-01"},
	"CEG-020": {PatternID: "CEG-020", BytesSavedPerOccurrence: 8192, AllocsReducedPercent: 0.999, Description: "Worker pool 99.9% less goroutine memory", Source: "README benchmark, Go 1.24, MacBook Air M2, 2025-01"},
}

// ParseScale converts a scale string to a numeric multiplier.
// Supported values: "1M", "10M", "100M". Default is 1,000,000.
func ParseScale(scale string) int64 {
	switch scale {
	case "1M":
		return 1_000_000
	case "10M":
		return 10_000_000
	case "100M":
		return 100_000_000
	default:
		return 1_000_000
	}
}

// CalculateImpact computes the impact level and estimated savings for a pattern.
// It looks up the pattern in ImpactRegistry, calculates total bytes saved,
// and returns a human-readable level and savings string.
func CalculateImpact(patternID string, occurrences int, scale int64) (level string, savings string) {
	impact, ok := ImpactRegistry[patternID]
	if !ok {
		return "Low", "unknown"
	}

	totalBytes := impact.BytesSavedPerOccurrence * int64(occurrences) * scale / 1_000_000

	switch {
	case totalBytes > 10*1024*1024:
		level = "High"
	case totalBytes > 1*1024*1024:
		level = "Medium"
	default:
		level = "Low"
	}

	scaleStr := "1M"
	switch scale {
	case 10_000_000:
		scaleStr = "10M"
	case 100_000_000:
		scaleStr = "100M"
	}

	savings = fmt.Sprintf("~%s at %s scale", FormatBytes(totalBytes), scaleStr)
	return level, savings
}

// FormatBytes converts a byte count to a human-readable string.
func FormatBytes(bytes int64) string {
	switch {
	case bytes < 1024:
		return fmt.Sprintf("%d B", bytes)
	case bytes < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	case bytes < 1024*1024*1024:
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	default:
		return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024))
	}
}
