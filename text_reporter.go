package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/alpardfm/cost-efficient-go/types"
)

// ANSI color codes for terminal output.
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

// TextReporter renders a Report as human-readable colored terminal output.
type TextReporter struct{}

// Render formats the report as colored text output.
func (t *TextReporter) Render(report *Report) ([]byte, error) {
	var sb strings.Builder

	t.renderHeader(&sb, report)
	t.renderOverallScore(&sb, report)
	t.renderSummary(&sb, report)
	t.renderPatternTable(&sb, report)
	t.renderFindings(&sb, report)
	t.renderDisclaimer(&sb, report)

	output := sb.String()

	if !isTerminal() {
		output = stripColors(output)
	}

	return []byte(output), nil
}

// renderHeader writes the header box with project metadata.
func (t *TextReporter) renderHeader(sb *strings.Builder, report *Report) {
	border := strings.Repeat("═", 60)
	sb.WriteString(fmt.Sprintf("%s╔%s╗%s\n", colorCyan, border, colorReset))
	sb.WriteString(fmt.Sprintf("%s║%s  Go Cost-Efficiency Analyzer Report%s%s║%s\n",
		colorCyan, colorBold, strings.Repeat(" ", 22), colorCyan, colorReset))
	sb.WriteString(fmt.Sprintf("%s╠%s╣%s\n", colorCyan, border, colorReset))
	sb.WriteString(fmt.Sprintf("%s║%s  Project:  %-48s%s║%s\n", colorCyan, colorReset, report.ProjectPath, colorCyan, colorReset))
	sb.WriteString(fmt.Sprintf("%s║%s  Registry: %-48s%s║%s\n", colorCyan, colorReset, report.RegistryVersion, colorCyan, colorReset))
	sb.WriteString(fmt.Sprintf("%s║%s  Date:     %-48s%s║%s\n", colorCyan, colorReset, report.Timestamp, colorCyan, colorReset))
	sb.WriteString(fmt.Sprintf("%s╚%s╝%s\n", colorCyan, border, colorReset))
	sb.WriteString("\n")
}

// renderOverallScore writes the overall score with color coding.
func (t *TextReporter) renderOverallScore(sb *strings.Builder, report *Report) {
	sb.WriteString(fmt.Sprintf("%s── Overall Score ──%s\n", colorBold, colorReset))

	if report.OverallScore < 0 {
		sb.WriteString(fmt.Sprintf("  Score: %s%sN/A%s (no applicable patterns)\n", colorDim, colorBold, colorReset))
	} else {
		color := scoreColor(report.OverallScore)
		sb.WriteString(fmt.Sprintf("  Score: %s%s%d/100%s\n", color, colorBold, report.OverallScore, colorReset))
	}
	sb.WriteString("\n")
}

// renderSummary writes a summary section with key stats.
func (t *TextReporter) renderSummary(sb *strings.Builder, report *Report) {
	// Count stats
	totalFindings := 0
	patternsWithFindings := 0
	for _, p := range report.Patterns {
		if len(p.Findings) > 0 {
			patternsWithFindings++
			totalFindings += len(p.Findings)
		}
	}

	sb.WriteString(fmt.Sprintf("%s── Summary ──%s\n", colorBold, colorReset))
	sb.WriteString(fmt.Sprintf("  Patterns checked: %d | With findings: %d\n", len(report.Patterns), patternsWithFindings))
	sb.WriteString(fmt.Sprintf("  Total findings: %d\n", totalFindings))

	if totalFindings > 0 {
		sb.WriteString(fmt.Sprintf("  %sTop issues:%s\n", colorBold, colorReset))

		// Sort patterns by finding count descending, show top 3
		type patternCount struct {
			id    string
			name  string
			sev   string
			count int
		}
		var counts []patternCount
		for _, p := range report.Patterns {
			if len(p.Findings) > 0 {
				counts = append(counts, patternCount{
					id:    p.Rule.ID,
					name:  p.Rule.Name,
					sev:   severityString(p.Rule.Severity),
					count: len(p.Findings),
				})
			}
		}
		sort.Slice(counts, func(i, j int) bool {
			return counts[i].count > counts[j].count
		})

		limit := 3
		if len(counts) < limit {
			limit = len(counts)
		}
		for i := 0; i < limit; i++ {
			c := counts[i]
			sb.WriteString(fmt.Sprintf("    %d. %s (%s) %s — %d findings\n", i+1, c.id, c.sev, truncate(c.name, 30), c.count))
		}
	}

	sb.WriteString("\n")
}

// renderPatternTable writes the pattern score table sorted by severity.
func (t *TextReporter) renderPatternTable(sb *strings.Builder, report *Report) {
	sb.WriteString(fmt.Sprintf("%s── Pattern Scores ──%s\n", colorBold, colorReset))
	sb.WriteString("\n")

	// Header row
	sb.WriteString(fmt.Sprintf("  %-8s %-28s %-14s %-10s %-6s %-10s\n",
		"ID", "Pattern", "Category", "Severity", "Score", "Impact"))
	sb.WriteString(fmt.Sprintf("  %s\n", strings.Repeat("─", 80)))

	// Sort patterns by severity (Critical first, then Major, then Minor)
	sorted := make([]PatternResult, len(report.Patterns))
	copy(sorted, report.Patterns)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Rule.Severity > sorted[j].Rule.Severity
	})

	for _, p := range sorted {
		if p.NotApplicable {
			// Show N/A patterns grayed out
			sb.WriteString(fmt.Sprintf("  %s%-8s %-28s %-14s %-10s %-6s %-10s%s\n",
				colorDim,
				p.Rule.ID,
				truncate(p.Rule.Name, 28),
				categoryString(p.Rule.Category),
				severityString(p.Rule.Severity),
				"N/A",
				"—",
				colorReset))
			continue
		}

		color := scoreColor(p.Score)
		sb.WriteString(fmt.Sprintf("  %s%-8s %-28s %-14s %-10s %-6s %-10s%s\n",
			color,
			p.Rule.ID,
			truncate(p.Rule.Name, 28),
			categoryString(p.Rule.Category),
			severityString(p.Rule.Severity),
			fmt.Sprintf("%d", p.Score),
			p.ImpactLevel,
			colorReset))
	}
	sb.WriteString("\n")
}

// renderFindings writes detailed findings sorted by severity.
// When a pattern has more than 5 findings, groups them by file for readability.
func (t *TextReporter) renderFindings(sb *strings.Builder, report *Report) {
	// Collect all findings across patterns
	type findingWithRule struct {
		finding types.Finding
		rule    types.Rule
	}

	var allFindings []findingWithRule
	for _, p := range report.Patterns {
		for _, f := range p.Findings {
			allFindings = append(allFindings, findingWithRule{finding: f, rule: p.Rule})
		}
	}

	if len(allFindings) == 0 {
		return
	}

	sb.WriteString(fmt.Sprintf("%s── Findings ──%s\n", colorBold, colorReset))
	sb.WriteString("\n")

	// Group findings by pattern
	patternFindings := make(map[string][]findingWithRule)
	var patternOrder []string
	for _, item := range allFindings {
		id := item.rule.ID
		if _, exists := patternFindings[id]; !exists {
			patternOrder = append(patternOrder, id)
		}
		patternFindings[id] = append(patternFindings[id], item)
	}

	// Sort pattern order by severity (Critical first)
	sort.Slice(patternOrder, func(i, j int) bool {
		fi := patternFindings[patternOrder[i]][0]
		fj := patternFindings[patternOrder[j]][0]
		return fi.finding.Severity > fj.finding.Severity
	})

	const groupThreshold = 5 // Group by file when more than this many findings

	for _, patternID := range patternOrder {
		findings := patternFindings[patternID]
		rule := findings[0].rule
		sevColor := severityColor(rule.Severity)

		if len(findings) > groupThreshold {
			// Grouped mode: show summary per file
			sb.WriteString(fmt.Sprintf("  %s[%s]%s %s %s — %d findings\n",
				sevColor, severityString(rule.Severity), colorReset,
				rule.ID, rule.Name, len(findings)))

			// Group by file
			fileFindings := make(map[string][]findingWithRule)
			var fileOrder []string
			for _, f := range findings {
				fp := f.finding.FilePath
				if _, exists := fileFindings[fp]; !exists {
					fileOrder = append(fileOrder, fp)
				}
				fileFindings[fp] = append(fileFindings[fp], f)
			}

			// Sort files by finding count descending
			sort.Slice(fileOrder, func(i, j int) bool {
				return len(fileFindings[fileOrder[i]]) > len(fileFindings[fileOrder[j]])
			})

			for _, fp := range fileOrder {
				ffs := fileFindings[fp]
				// Collect line numbers
				var lines []string
				for _, ff := range ffs {
					lines = append(lines, fmt.Sprintf("%d", ff.finding.Line))
				}
				lineStr := strings.Join(lines, ", ")
				if len(lines) > 5 {
					lineStr = strings.Join(lines[:5], ", ") + fmt.Sprintf(" (+%d more)", len(lines)-5)
				}
				sb.WriteString(fmt.Sprintf("    %s%s%s: %d occurrences (lines: %s)\n",
					colorDim, fp, colorReset, len(ffs), lineStr))
			}

			// Show suggestion once
			if findings[0].finding.SuggestedFix != "" {
				sb.WriteString(fmt.Sprintf("    %sFix:%s %s\n", colorDim, colorReset, findings[0].finding.SuggestedFix))
			}
			if len(rule.ReferenceLinks) > 0 {
				sb.WriteString(fmt.Sprintf("    %sRef:%s %s\n", colorDim, colorReset, rule.ReferenceLinks[0]))
			}
			sb.WriteString("\n")
		} else {
			// Detailed mode: show each finding individually
			for _, item := range findings {
				f := item.finding
				r := item.rule

				sb.WriteString(fmt.Sprintf("  %s[%s]%s %s %s\n",
					sevColor, severityString(f.Severity), colorReset,
					r.ID, r.Name))
				sb.WriteString(fmt.Sprintf("    %sFile:%s %s:%d\n", colorDim, colorReset, f.FilePath, f.Line))
				sb.WriteString(fmt.Sprintf("    %sConfidence:%s %s\n", colorDim, colorReset, confidenceString(f.Confidence)))

				if f.Explanation != "" {
					sb.WriteString(fmt.Sprintf("    %sExplanation:%s %s\n", colorDim, colorReset, f.Explanation))
				}
				if f.SuggestedFix != "" {
					sb.WriteString(fmt.Sprintf("    %sSuggested fix:%s %s\n", colorDim, colorReset, f.SuggestedFix))
				}
				if len(r.ReferenceLinks) > 0 {
					sb.WriteString(fmt.Sprintf("    %sReference:%s %s\n", colorDim, colorReset, r.ReferenceLinks[0]))
				}
				sb.WriteString("\n")
			}
		}
	}
}

// renderDisclaimer writes the disclaimer at the bottom.
func (t *TextReporter) renderDisclaimer(sb *strings.Builder, report *Report) {
	if report.Disclaimer == "" {
		return
	}
	sb.WriteString(fmt.Sprintf("%s%s%s\n", colorDim, strings.Repeat("─", 60), colorReset))
	sb.WriteString(fmt.Sprintf("%s%s%s\n", colorDim, report.Disclaimer, colorReset))
}

// isTerminal checks if stdout is connected to a terminal.
func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// stripColors removes all ANSI escape sequences from the output.
func stripColors(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\033' {
			// Skip until 'm' (end of ANSI escape sequence)
			for i < len(s) && s[i] != 'm' {
				i++
			}
			if i < len(s) {
				i++ // skip the 'm'
			}
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String()
}

// scoreColor returns the ANSI color code based on score value.
func scoreColor(score int) string {
	switch {
	case score >= 80:
		return colorGreen
	case score >= 50:
		return colorYellow
	default:
		return colorRed
	}
}

// severityColor returns the ANSI color code for a severity level.
func severityColor(sev types.Severity) string {
	switch sev {
	case types.Critical:
		return colorRed
	case types.Major:
		return colorYellow
	case types.Minor:
		return colorCyan
	default:
		return colorReset
	}
}

// truncate shortens a string to maxLen, adding "…" if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "…"
	}
	return s[:maxLen-1] + "…"
}
