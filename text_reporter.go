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

	// Sort by severity (Critical first)
	sort.Slice(allFindings, func(i, j int) bool {
		return allFindings[i].finding.Severity > allFindings[j].finding.Severity
	})

	sb.WriteString(fmt.Sprintf("%s── Findings ──%s\n", colorBold, colorReset))
	sb.WriteString("\n")

	for _, item := range allFindings {
		f := item.finding
		r := item.rule

		sevColor := severityColor(f.Severity)
		sb.WriteString(fmt.Sprintf("  %s[%s]%s %s %s\n",
			sevColor, severityString(f.Severity), colorReset,
			r.ID, r.Name))
		sb.WriteString(fmt.Sprintf("    %sFile:%s %s:%d\n", colorDim, colorReset, f.FilePath, f.Line))

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
