package main

import (
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// chartItem is one row in a bar chart: a label and a signed value.
type chartItem struct {
	Label  string
	Amount float64
}

// renderBarChart mirrors charts.py's render_bar_chart: a horizontal bar per
// item with a blank line of spacing between them, one color for every bar
// unless negativeColor is given (used for a signed series, e.g. income vs
// spend) - never a per-category rainbow, and never a per-bar value, just
// the proportion. If showAxis (and fmt) are given, a trailing "0 .. max"
// scale line is appended under the bars.
func renderBarChart(items []chartItem, color lipgloss.Color, negativeColor lipgloss.Color, width int, showAxis bool, fmt func(float64) string) string {
	if len(items) == 0 {
		return styleMuted.Italic(true).Render("No data")
	}
	if width <= 0 {
		width = 16
	}

	maxAbs := 0.0
	labelWidth := 0
	for _, it := range items {
		if abs := absFloat(it.Amount); abs > maxAbs {
			maxAbs = abs
		}
		if l := len([]rune(it.Label)); l > labelWidth {
			labelWidth = l
		}
	}
	if maxAbs == 0 {
		maxAbs = 1
	}

	lines := make([]string, len(items))
	for i, it := range items {
		barColor := color
		if it.Amount < 0 && negativeColor != "" {
			barColor = negativeColor
		}
		filled := roundBar(width, absFloat(it.Amount), maxAbs)
		bar := barStyle(barColor).Render(strings.Repeat("█", filled)) +
			mutedBarStyle().Render(strings.Repeat("░", width-filled))
		label := styleBold.Render(padRight(it.Label, labelWidth))
		lines[i] = label + " " + bar
	}

	result := strings.Join(lines, "\n\n")

	if showAxis && fmt != nil {
		left := "0"
		right := fmt(maxAbs)
		gap := width - len(left) - len(right)
		if gap < 1 {
			gap = 1
		}
		axis := strings.Repeat(" ", labelWidth+1) + left + strings.Repeat(" ", gap) + right
		result += "\n" + mutedBarStyle().Render(axis)
	}

	return result
}

// renderBudgetGauge mirrors charts.py's render_budget_gauge: a
// threshold-colored progress bar (good below 70%, warning below 100%,
// critical at/above), the one chart that does show numbers.
func renderBudgetGauge(spent, budget float64, fmt func(float64) string, good, warning, critical lipgloss.Color, width int) string {
	if budget <= 0 {
		return styleMuted.Italic(true).Render("No budget set")
	}
	if width <= 0 {
		width = 14
	}

	ratio := spent / budget
	filled := roundBar(width, minFloat(ratio, 1.0), 1.0)
	if filled > width {
		filled = width
	}
	color := good
	switch {
	case ratio >= 1.0:
		color = critical
	case ratio >= 0.7:
		color = warning
	}

	var b strings.Builder
	b.WriteString(fmt(spent) + " / " + fmt(budget) + "\n")
	b.WriteString(barStyle(color).Render(strings.Repeat("█", filled)))
	b.WriteString(mutedBarStyle().Render(strings.Repeat("░", width-filled)))
	b.WriteString(barStyle(color).Render(" " + formatPercent(ratio*100)))
	if ratio > 1.0 {
		b.WriteString("\n" + barStyle(critical).Bold(true).Render("over budget"))
	}
	return b.String()
}

// roundBar mirrors Python's round(width * amount / maxAbs) - round-half-
// to-even, same as Python's built-in round() - so bar-fill widths land on
// exactly the same pixel counts as the original for tied .5 cases.
func roundBar(width int, amount, maxAbs float64) int {
	return roundHalfToEven(float64(width) * amount / maxAbs)
}

func roundHalfToEven(x float64) int {
	floor := math.Floor(x)
	diff := x - floor
	switch {
	case diff < 0.5:
		return int(floor)
	case diff > 0.5:
		return int(floor) + 1
	default:
		if int64(floor)%2 == 0 {
			return int(floor)
		}
		return int(floor) + 1
	}
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func padRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

// formatPercent mirrors Python's f"{x:,.0f}" for the gauge's "NN%" label -
// a whole-number percentage, comma-grouped (moot below 1000% in practice,
// kept for fidelity).
func formatPercent(v float64) string {
	return groupThousands(strconv.Itoa(roundHalfToEven(v))) + "%"
}
