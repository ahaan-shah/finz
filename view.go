package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// applySizes derives every widget's width/height from the last known
// terminal size - mirrors the CSS layout: Header/Footer are 1 row each,
// #ledger is 3fr wide, #sidebar is 1fr wide with a 38-column floor.
func (m *model) applySizes() {
	m.bodyHeight = m.height - 2
	if m.bodyHeight < 1 {
		m.bodyHeight = 1
	}

	sidebarWidth := m.width / 4
	if sidebarWidth < 38 {
		sidebarWidth = 38
	}
	if sidebarWidth > m.width-20 {
		sidebarWidth = m.width - 20
	}
	if sidebarWidth < 10 {
		sidebarWidth = 10
	}
	m.sidebarWidth = sidebarWidth
	m.mainWidth = m.width - sidebarWidth
	if m.mainWidth < 10 {
		m.mainWidth = 10
	}

	m.yearFilter.Width = m.mainWidth - 6
	m.monthFilter.Width = m.mainWidth - 6

}

// -- header / footer -------------------------------------------------------

// renderHeader mirrors Textual's Header widget layout exactly: an 8-column
// docked-left icon block (1 pad + icon + 6 pad, per HeaderIcon's own
// `padding: 0 1; width: 8`), a 10-column docked-right (empty) clock space,
// and the title/subtitle centered in whatever's left between them.
func (m model) renderHeader() string {
	const iconWidth = 8
	const clockWidth = 10
	middleWidth := m.width - iconWidth - clockWidth
	if middleWidth < 0 {
		middleWidth = 0
	}

	title := "Tally"
	subtitle := " — Finance Tracker"
	plainWidth := lipgloss.Width(title) + lipgloss.Width(subtitle)
	pad := middleWidth - plainWidth
	if pad < 0 {
		pad = 0
	}
	left := pad / 2
	right := pad - left

	icon := styleHeader.Render(padRight(" ⭘", iconWidth))
	middle := styleHeader.Render(strings.Repeat(" ", left)+title) + styleHeaderSubtitle.Render(subtitle) + styleHeader.Render(strings.Repeat(" ", right))
	clock := styleHeader.Render(strings.Repeat(" ", clockWidth))

	line := icon + middle + clock
	return repaintWith(line, panelRepaint)
}

// renderFooter mirrors Footer: key hints on the left (a/e/d/q hidden
// whenever a text-entry widget has effective focus, since typing there
// would just type the letter instead of firing the binding - matches the
// Python original's observed behavior exactly, see app.go's key
// dispatch), the command-palette hint permanently docked right behind a
// dim vertical separator.
func (m model) renderFooter() string {
	var groups [][2]string
	if m.stage == stageTable && !m.budgetFocused {
		groups = append(groups, [2]string{"a", "Add"}, [2]string{"e", "Edit"}, [2]string{"d", "Delete"})
	}
	groups = append(groups, [2]string{"alt+→", "Next"}, [2]string{"alt+←", "Prev"})
	if m.stage == stageTable && !m.budgetFocused {
		groups = append(groups, [2]string{"q", "Quit"})
	}

	var left strings.Builder
	for i, g := range groups {
		if i > 0 {
			left.WriteString(styleFooterDescription.Render(" "))
		}
		left.WriteString(styleFooterKey.Render(g[0]) + styleFooterDescription.Render(" "+g[1]+" "))
	}

	right := styleFooterSeparator.Render("▏") + styleFooterKey.Render("^p") + styleFooterDescription.Render(" palette")

	leftStr := left.String()
	gap := m.width - lipgloss.Width(leftStr) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	line := leftStr + styleFooter.Render(strings.Repeat(" ", gap)) + right
	return repaintWith(line, footerRepaint)
}

// -- year / month stage ----------------------------------------------------

// renderYearMonthStage builds the filter Input box (height 3: border,
// content, border) directly above the option-list box (fills the rest of
// bodyHeight), both using Textual's own "tall" border glyphs.
func (m model) renderYearMonthStage(filterView string, items []periodItem, cursor int) string {
	innerWidth := m.mainWidth - 4
	if innerWidth < 4 {
		innerWidth = 4
	}
	filterBox := renderTallBox(filterView, innerWidth, 2)

	listHeight := m.bodyHeight - 3
	if listHeight < 1 {
		listHeight = 1
	}
	listView := renderPeriodList(items, cursor, innerWidth-2, listHeight)
	listBox := renderTallBox(listView, innerWidth, 1)

	return lipgloss.JoinVertical(lipgloss.Left, filterBox, listBox)
}

// renderTallBox manually draws Textual's "tall" border (see tallBorder in
// style.go) around already-rendered content, rather than handing that
// content to another lipgloss Style.Render() call - wrapping pre-rendered
// ANSI text (the filter Input's own cursor styling, or a period list's
// cursor-row highlight) in a bordered Style().Render() call re-processes
// it as one opaque run of text and corrupts the escape sequence stream
// (the same class of bug renderBudgetBox's doc comment describes).
func renderTallBox(content string, innerWidth, padding int) string {
	top := styleAccent.Render(tallBorder.TopLeft + strings.Repeat(tallBorder.Top, innerWidth) + tallBorder.TopRight)
	bottom := styleAccent.Render(tallBorder.BottomLeft + strings.Repeat(tallBorder.Bottom, innerWidth) + tallBorder.BottomRight)
	left := styleAccent.Render(tallBorder.Left)
	right := styleAccent.Render(tallBorder.Right)
	pad := strings.Repeat(" ", padding)

	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines)+2)
	out = append(out, top)
	for _, l := range lines {
		trailing := innerWidth - 2*padding - lipgloss.Width(l)
		if trailing < 0 {
			trailing = 0
		}
		out = append(out, left+pad+l+strings.Repeat(" ", trailing)+pad+right)
	}
	out = append(out, bottom)
	return repaintWith(strings.Join(out, "\n"), surfaceRepaint)
}

// renderPeriodList mirrors an OptionList's body: one line per item,
// full-row highlighted (styleTableCursor, matching Textual's own
// $block-cursor-background highlight color for a focused OptionList) when
// it's the cursor's row, scrolled so the cursor always stays in view.
func renderPeriodList(items []periodItem, cursor, width, height int) string {
	blank := strings.Repeat(" ", width)
	if len(items) == 0 {
		lines := make([]string, height)
		for i := range lines {
			lines[i] = blank
		}
		return strings.Join(lines, "\n")
	}

	start := 0
	if len(items) > height {
		start = cursor - height/2
		if start < 0 {
			start = 0
		}
		if start > len(items)-height {
			start = len(items) - height
		}
	}
	end := start + height
	if end > len(items) {
		end = len(items)
	}

	lines := make([]string, 0, height)
	for i := start; i < end; i++ {
		line := padRight(items[i].label, width)
		if i == cursor {
			line = styleTableCursor.Render(line)
		}
		lines = append(lines, line)
	}
	for len(lines) < height {
		lines = append(lines, blank)
	}
	return strings.Join(lines, "\n")
}

// -- table stage -------------------------------------------------------

const ledgerDateWidth = 10 // "YYYY-MM-DD" is always exactly 10 characters

// renderTableStage mirrors the DataTable ledger: Date/Category/Amount/
// Balance/Note columns (auto-sized to their content, like Rich's own
// DataTable, not fixed pixel widths), amount/balance colored from the
// theme's error/success roles, cursor row highlighted with
// styleTableCursor. Column widths are measured from each row's *plain*
// text before any styling is applied, and only ever used for measurement
// - see periodRows/tableCursor's doc comment in app.go for why this isn't
// bubbles/table.
func (m model) renderTableStage() string {
	balances := runningBalances(m.transactions)

	type rowText struct {
		date, category, amountPlain, balancePlain, note string
		amountIsIncome, balanceNegative                 bool
	}
	texts := make([]rowText, len(m.periodRows))
	categoryWidth := lipgloss.Width("Category")
	amountWidth := lipgloss.Width("Amount")
	balanceWidth := lipgloss.Width("Balance")
	for i, t := range m.periodRows {
		var amountPlain string
		if t.Amount < 0 {
			amountPlain = "+" + fmtAmount(-t.Amount, m.settings.Currency)
		} else {
			amountPlain = "-" + fmtAmount(t.Amount, m.settings.Currency)
		}
		balance := balances[t.ID]
		texts[i] = rowText{
			date: t.Date, category: t.Category,
			amountPlain: amountPlain, balancePlain: fmtAmount(balance, m.settings.Currency),
			note: t.Note, amountIsIncome: t.Amount < 0, balanceNegative: balance < -epsilon,
		}
		if w := lipgloss.Width(t.Category); w > categoryWidth {
			categoryWidth = w
		}
		if w := lipgloss.Width(texts[i].amountPlain); w > amountWidth {
			amountWidth = w
		}
		if w := lipgloss.Width(texts[i].balancePlain); w > balanceWidth {
			balanceWidth = w
		}
	}

	const leadingSpace = 1 // Textual's DataTable has a small implicit left inset
	const gaps = 4*2 + leadingSpace
	noteWidth := m.mainWidth - ledgerDateWidth - categoryWidth - amountWidth - balanceWidth - gaps
	if noteWidth < lipgloss.Width("Note") {
		noteWidth = lipgloss.Width("Note")
	}

	header := " " + styleBold.Render(padRight("Date", ledgerDateWidth)) + "  " +
		styleBold.Render(padRight("Category", categoryWidth)) + "  " +
		styleBold.Render(padRight("Amount", amountWidth)) + "  " +
		styleBold.Render(padRight("Balance", balanceWidth)) + "  " +
		styleBold.Render(padRight("Note", noteWidth))

	lines := []string{header}
	for i, rt := range texts {
		amountStyle := styleError
		if rt.amountIsIncome {
			amountStyle = styleSuccess
		}
		balanceStyle := lipgloss.NewStyle()
		if rt.balanceNegative {
			balanceStyle = styleError.Bold(true)
		}

		line := " " + padRight(rt.date, ledgerDateWidth) + "  " +
			padRight(rt.category, categoryWidth) + "  " +
			amountStyle.Render(padRight(rt.amountPlain, amountWidth)) + "  " +
			balanceStyle.Render(padRight(rt.balancePlain, balanceWidth)) + "  " +
			padRight(truncateWidth(rt.note, noteWidth), noteWidth)

		if i == m.tableCursor {
			line = styleTableCursor.Render(padRight(line, m.mainWidth))
		}
		lines = append(lines, line)
	}
	for len(lines) < m.bodyHeight {
		lines = append(lines, "")
	}
	return repaintLedgerView(strings.Join(lines, "\n"))
}

// truncateWidth trims s to at most width columns (ansi-aware, though note
// text is always plain here), appending "…" when it had to cut.
func truncateWidth(s string, width int) string {
	if lipgloss.Width(s) <= width {
		return s
	}
	r := []rune(s)
	if width <= 1 {
		return string(r[:width])
	}
	return string(r[:width-1]) + "…"
}

// -- sidebar -----------------------------------------------------------

// renderSidebar mirrors the Python original's entirely stage-dependent
// right pane: empty while picking a year, only the spend-by-month chart
// while picking a month, and the full budget/summary/chart dashboard once
// a specific month's ledger is open. Built by hand (border character +
// padding concatenated onto each line) rather than handing the already-
// rendered, ANSI-heavy body to styleSidebar.Render() - a color-bearing
// Style's Render() re-processes its *entire* input as one opaque run of
// text, and doing that to text that already contains embedded escape
// sequences from earlier Render() calls corrupts the stream (this is what
// made every non-ansi-* theme render as a blank pane before this fix -
// ansi-dark/light happened to work because their styles carry no color
// properties at all, so the risky re-processing was accidentally a no-op).
func (m model) renderSidebar() string {
	var body string
	switch m.stage {
	case stageYear:
		body = ""
	case stageMonth:
		body = m.renderChart()
	default:
		body = m.renderBudgetBox() + "\n\n" + m.renderSummary() + "\n\n" + m.renderChart()
	}

	innerHeight := m.bodyHeight - 2 // Padding(1, 2)'s top/bottom inset
	if innerHeight < 1 {
		innerHeight = 1
	}
	lines := strings.Split(body, "\n")
	for len(lines) < innerHeight {
		lines = append(lines, "")
	}
	if len(lines) > innerHeight {
		lines = lines[:innerHeight]
	}

	border := styleAccent.Render("│")
	out := make([]string, 0, innerHeight+2)
	out = append(out, "")
	for _, l := range lines {
		out = append(out, border+"  "+l)
	}
	out = append(out, "")
	return repaintWith(strings.Join(out, "\n"), canvasRepaint)
}

// renderBudgetBox mirrors the "Budget" box: round $accent border, a
// compact budget Input, and the color-thresholded gauge from charts.go.
// Every border glyph is its own small Render() call over plain text - the
// input/gauge content (already fully rendered, with its own embedded
// color codes) is only ever concatenated next to those spans, never
// handed back into another Style.Render() call, which is what corrupts
// the escape sequence stream (lipgloss's per-line wrap treats an already-
// ANSI multi-line block as one opaque run of text to re-wrap, and nesting
// several of those is what produced the garbled output this comment used
// to describe finding).
func (m model) renderBudgetBox() string {
	innerWidth := 28
	title := " Budget "
	dashes := innerWidth - lipgloss.Width(title)
	if dashes < 0 {
		dashes = 0
	}
	top := styleAccent.Render("╭─" + title + strings.Repeat("─", dashes) + "╮")

	inputWidth := innerWidth - 4
	if inputWidth < 4 {
		inputWidth = 4
	}
	input := renderTallBox(m.budgetInput.View(), inputWidth, 2)

	good, warning, critical := colorSuccess, colorWarning, colorError
	spent := totalSpent(filterByMonth(m.transactions, m.selectedPeriod()))
	gauge := renderBudgetGauge(spent, m.settings.MonthlyBudget, func(v float64) string { return fmtAmount(v, m.settings.Currency) }, good, warning, critical, 14)

	lines := []string{top}
	for _, l := range strings.Split(input, "\n") {
		lines = append(lines, styleAccent.Render("│ ")+l+styleAccent.Render(" │"))
	}
	for _, l := range strings.Split(gauge, "\n") {
		pad := innerWidth - 2 - lipgloss.Width(l)
		if pad < 0 {
			pad = 0
		}
		lines = append(lines, styleAccent.Render("│ ")+l+strings.Repeat(" ", pad)+styleAccent.Render(" │"))
	}
	lines = append(lines, styleAccent.Render("╰"+strings.Repeat("─", innerWidth)+"╯"))

	return strings.Join(lines, "\n")
}

// renderSummary mirrors refresh_summary: Spent/Income/Net for the viewed
// month, then a by-category expense breakdown (never the raw signed
// totals - see analytics.go), amounts right-aligned to the widest one.
func (m model) renderSummary() string {
	period := m.selectedPeriod()
	rows := filterByMonth(m.transactions, period)

	var b strings.Builder
	b.WriteString(styleBold.Render("Spent:") + "  " + fmtAmount(totalSpent(rows), m.settings.Currency) + "\n")
	b.WriteString(styleBold.Render("Income:") + " " + fmtAmount(totalIncome(rows), m.settings.Currency) + "\n")
	b.WriteString(styleBold.Render("Net:") + "    " + fmtAmount(netBalance(rows), m.settings.Currency) + "\n\n")

	b.WriteString(styleBold.Render("By category") + "\n")
	byCategory := expenseTotalsByCategoryForMonth(m.transactions, period)
	delete(byCategory, "") // categories are never blank in practice; guards the loop below
	type entry struct {
		label  string
		amount float64
	}
	var entries []entry
	for c, a := range byCategory {
		if strings.ToLower(strings.TrimSpace(c)) == "income" {
			continue
		}
		entries = append(entries, entry{c, a})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].amount > entries[j].amount })

	if len(entries) == 0 {
		b.WriteString("  (none)")
		return b.String()
	}

	labelWidth, amountWidth := 0, 0
	amountStrs := make([]string, len(entries))
	for i, e := range entries {
		if l := lipgloss.Width(e.label); l > labelWidth {
			labelWidth = l
		}
		amountStrs[i] = fmtAmount(e.amount, m.settings.Currency)
		if l := lipgloss.Width(amountStrs[i]); l > amountWidth {
			amountWidth = l
		}
	}
	lines := make([]string, len(entries))
	for i, e := range entries {
		lines[i] = "  " + padRight(e.label, labelWidth) + "  " + padLeftTo(amountStrs[i], amountWidth)
	}
	b.WriteString(strings.Join(lines, "\n"))
	return b.String()
}

// renderChart mirrors refresh_chart: a category breakdown for the viewed
// month while stage == "table" (top 8 plus an "Other" bucket for the
// rest), or a spend-per-month overview for the viewed year while stage ==
// "month".
func (m model) renderChart() string {
	if m.stage == stageTable {
		period := m.selectedPeriod()
		byCategory := expenseTotalsByCategoryForMonth(m.transactions, period)
		type entry struct {
			label  string
			amount float64
		}
		var ranked []entry
		for c, a := range byCategory {
			if strings.ToLower(strings.TrimSpace(c)) == "income" {
				continue
			}
			ranked = append(ranked, entry{c, a})
		}
		sort.Slice(ranked, func(i, j int) bool { return ranked[i].amount > ranked[j].amount })

		items := make([]chartItem, 0, 9)
		for i, e := range ranked {
			if i >= 8 {
				break
			}
			items = append(items, chartItem{Label: e.label, Amount: e.amount})
		}
		other := 0.0
		for i := 8; i < len(ranked); i++ {
			other += ranked[i].amount
		}
		if other > 0 {
			items = append(items, chartItem{Label: "Other", Amount: other})
		}

		header := styleBold.Render(fmt.Sprintf("Spending by category -- %s", period))
		body := renderBarChart(items, colorAccent, "", 16, false, nil)
		return header + "\n\n" + body
	}

	items := make([]chartItem, 12)
	for i := 1; i <= 12; i++ {
		period := fmt.Sprintf("%s-%02d", m.selectedYear, i)
		items[i-1] = chartItem{Label: monthNames[i-1], Amount: totalSpent(filterByMonth(m.transactions, period))}
	}
	header := styleBold.Render("Spent by month -- " + m.selectedYear)
	body := renderBarChart(items, colorAccent, "", 16, true, func(v float64) string { return fmtAmount(v, m.settings.Currency) })
	return header + "\n\n" + body
}

func padLeftTo(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return strings.Repeat(" ", width-w) + s
}

// -- top-level view ----------------------------------------------------

func (m model) View() string {
	if m.width == 0 {
		return "tally: loading...\n"
	}

	if m.modal != modalNone {
		return placeOnCanvas(m.width, m.height, repaintWith(m.renderModal(), surfaceRepaint))
	}

	header := m.renderHeader()

	var main string
	switch m.stage {
	case stageYear:
		main = m.renderYearMonthStage(m.yearFilter.View(), m.yearItems, m.yearCursor)
	case stageMonth:
		main = m.renderYearMonthStage(m.monthFilter.View(), m.monthItems, m.monthCursor)
	default:
		main = m.renderTableStage()
	}
	sidebar := m.renderSidebar()
	// main/sidebar are already fully rendered ANSI (table cells, budget
	// box, etc.) - JoinHorizontal only splits on "\n" and pads with plain
	// spaces/blank lines, unlike wrapping them in one more
	// lipgloss.Style{Width,Height}.Render() call, which re-processes the
	// whole block as opaque text and corrupts already-embedded escape
	// sequences (this is what caused the table/header to render as
	// entirely blank before this fix).
	body := lipgloss.JoinHorizontal(lipgloss.Top, main, sidebar)

	var footer string
	if m.errorMessage != "" {
		footer = repaintWith(styleError.Render(" "+m.errorMessage), footerRepaint)
	} else if m.noticeMessage != "" {
		footer = repaintWith(styleAccent.Render(" "+m.noticeMessage), footerRepaint)
	} else {
		footer = m.renderFooter()
	}

	full := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
	return repaintCanvas(fillCanvas(full, m.width, m.height))
}

// fillCanvas pads s out to exactly width x height with the active theme's
// background/foreground, using lipgloss.Place (ansi-aware, only ever
// *appends* whitespace - never re-processes s itself) rather than
// styleAppCanvas.Width(w).Height(h).Render(s), which - like renderSidebar
// used to - would hand the fully-composed, ANSI-heavy frame back into a
// color-bearing Style's Render() and corrupt it.
func fillCanvas(s string, width, height int) string {
	if appBackground == "" {
		return lipgloss.PlaceVertical(height, lipgloss.Top, lipgloss.PlaceHorizontal(width, lipgloss.Left, s))
	}
	opts := []lipgloss.WhitespaceOption{
		lipgloss.WithWhitespaceBackground(appBackground),
		lipgloss.WithWhitespaceForeground(appForeground),
	}
	return lipgloss.PlaceVertical(height, lipgloss.Top, lipgloss.PlaceHorizontal(width, lipgloss.Left, s, opts...), opts...)
}

// placeOnCanvas centers content over a full-screen fill of the active
// theme's background/foreground (falling back to the terminal's own
// colors for the ansi-* themes, which leave appBackground/appForeground
// empty on purpose).
func placeOnCanvas(width, height int, content string) string {
	if appBackground == "" {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content,
		lipgloss.WithWhitespaceBackground(appBackground),
		lipgloss.WithWhitespaceForeground(appForeground))
}

func (m model) renderModal() string {
	switch m.modal {
	case modalTransactionForm:
		return m.form.View()
	case modalConfirmDelete:
		return m.confirm.View()
	case modalPalette:
		return m.palette.View()
	}
	return ""
}
