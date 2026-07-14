package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type stage string

const (
	stageYear  stage = "year"
	stageMonth stage = "month"
	stageTable stage = "table"
)

var stageOrder = []stage{stageYear, stageMonth, stageTable}

func stageIndex(s stage) int {
	for i, v := range stageOrder {
		if v == s {
			return i
		}
	}
	return 0
}

var monthNames = []string{
	"January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December",
}

type modalKind int

const (
	modalNone modalKind = iota
	modalTransactionForm
	modalConfirmDelete
	modalPalette
)

const epsilon = 0.005

type model struct {
	transactions []Transaction
	settings     Settings

	stage         stage
	selectedYear  string
	selectedMonth string

	width, height           int
	mainWidth, sidebarWidth int
	bodyHeight              int

	yearFilter  textinput.Model
	yearItems   []periodItem
	yearCursor  int
	monthFilter textinput.Model
	monthItems  []periodItem
	monthCursor int

	// periodRows is the current period's transactions, most recent first -
	// tableCursor indexes into it (selectedTransactionID/
	// getSelectedTransaction resolve a position back to the real
	// Transaction, mirroring what the row's DataTable *key* (t.id) does
	// in the Python original). Hand-rolled instead of bubbles/table: that
	// library's cell truncation isn't ANSI-aware, so a colored amount
	// cell's escape codes get counted as visible width and the cell gets
	// truncated to a few characters - see renderTableStage/ledgerRow.
	periodRows  []Transaction
	tableCursor int

	budgetInput   textinput.Model
	budgetFocused bool // table stage's only other focusable widget besides the ledger table

	modal         modalKind
	form          transactionFormModel
	confirm       confirmDeleteModel
	confirmTarget string
	palette       paletteModel

	noticeMessage string
	errorMessage  string
}

// NewModel mirrors FinanceApp.__init__ + on_mount: stage always starts at
// "table" (the selected year/month, the last theme, and the last currency
// all persist across launches), the ledger table gets its
// Date/Category/Amount/Balance/Note columns, and the budget input is
// pre-filled from settings if a budget was already set.
func NewModel(transactions []Transaction, settings Settings) model {
	if settings.Theme != "" {
		setActiveTheme(settings.Theme)
	}

	m := model{
		transactions: transactions,
		settings:     settings,
		stage:        stageTable,
	}
	m.selectedYear, m.selectedMonth = m.initialPeriod()

	m.yearFilter = textinput.New()
	m.yearFilter.Placeholder = "Type to filter years..."
	m.monthFilter = textinput.New()
	m.monthFilter.Placeholder = "Type to filter months..."

	m.budgetInput = textinput.New()
	m.budgetInput.Placeholder = "e.g. 2000"
	if settings.MonthlyBudget > 0 {
		m.budgetInput.SetValue(formatFixed2(settings.MonthlyBudget))
	}

	m.refreshTable(settings.LastTransactionID)
	return m
}

// initialPeriod mirrors _initial_period: last_year/last_month from
// settings if both are present, else the most recent month any
// transaction falls in, else the real current year/month.
func (m model) initialPeriod() (string, string) {
	if m.settings.LastYear != "" && m.settings.LastMonth != "" {
		return m.settings.LastYear, m.settings.LastMonth
	}
	months := monthsPresent(m.transactions)
	if len(months) > 0 {
		last := months[len(months)-1]
		return last[:4], last[5:7]
	}
	now := time.Now()
	return fmt.Sprintf("%04d", now.Year()), fmt.Sprintf("%02d", int(now.Month()))
}

func (m model) selectedPeriod() string {
	return m.selectedYear + "-" + m.selectedMonth
}

func (m model) Init() tea.Cmd {
	return nil
}

// -- update ------------------------------------------------------------

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.applySizes()
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	if m.modal != modalNone {
		return m.handleModalKey(msg)
	}

	m.errorMessage = ""
	m.noticeMessage = ""

	switch msg.String() {
	case "ctrl+p":
		m.openPalette()
		return m, nil
	case "alt+right":
		m.nextStage()
		return m, nil
	case "alt+left":
		m.prevStage()
		return m, nil
	}

	switch m.stage {
	case stageYear:
		return m.handleYearKey(msg)
	case stageMonth:
		return m.handleMonthKey(msg)
	default:
		return m.handleTableKey(msg)
	}
}

func (m model) handleModalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.modal {
	case modalTransactionForm:
		result, closed, cmd := m.form.handleKey(msg)
		if closed {
			m.modal = modalNone
			if result != nil {
				m.handleFormResult(*result)
			}
		}
		return m, cmd
	case modalConfirmDelete:
		switch msg.String() {
		case "enter":
			m.handleDeleteConfirm(true)
			m.modal = modalNone
		case "esc":
			m.modal = modalNone
		}
		return m, nil
	case modalPalette:
		cmd := m.handlePaletteKey(msg)
		return m, cmd
	}
	return m, nil
}

func (m *model) closeModal() {
	m.modal = modalNone
}

func (m model) handleYearKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up":
		if m.yearCursor > 0 {
			m.yearCursor--
		}
		return m, nil
	case "down":
		if m.yearCursor < len(m.yearItems)-1 {
			m.yearCursor++
		}
		return m, nil
	case "enter":
		if m.yearCursor >= 0 && m.yearCursor < len(m.yearItems) {
			m.selectYear(m.yearItems[m.yearCursor].value)
		}
		return m, nil
	case "q":
		// Swallowed - the filter Input has focus in this stage, so typing
		// "q" types a "q", it doesn't quit (mirrors the Python original:
		// app-level bindings only fire when the focused widget doesn't
		// consume the key itself, and Input consumes every printable one).
	}
	var cmd tea.Cmd
	m.yearFilter, cmd = m.yearFilter.Update(msg)
	m.populateYearList()
	return m, cmd
}

func (m model) handleMonthKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up":
		if m.monthCursor > 0 {
			m.monthCursor--
		}
		return m, nil
	case "down":
		if m.monthCursor < len(m.monthItems)-1 {
			m.monthCursor++
		}
		return m, nil
	case "enter":
		if m.monthCursor >= 0 && m.monthCursor < len(m.monthItems) {
			m.selectMonth(m.monthItems[m.monthCursor].value)
		}
		return m, nil
	case "q":
		// see handleYearKey
	}
	var cmd tea.Cmd
	m.monthFilter, cmd = m.monthFilter.Update(msg)
	m.populateMonthList()
	return m, cmd
}

func (m model) handleTableKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab", "shift+tab":
		m.budgetFocused = !m.budgetFocused
		if m.budgetFocused {
			m.budgetInput.Focus()
		} else {
			m.budgetInput.Blur()
		}
		return m, nil
	case "q":
		if !m.budgetFocused {
			return m, tea.Quit
		}
	case "a":
		if !m.budgetFocused {
			m.startAdd()
			return m, nil
		}
	case "e":
		if !m.budgetFocused {
			m.startEdit()
			return m, nil
		}
	case "d":
		if !m.budgetFocused {
			m.startDelete()
			return m, nil
		}
	case "enter":
		if m.budgetFocused {
			m.submitBudget()
			return m, nil
		}
	}

	if m.budgetFocused {
		var cmd tea.Cmd
		m.budgetInput, cmd = m.budgetInput.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "up":
		if m.tableCursor > 0 {
			m.tableCursor--
		}
	case "down":
		if m.tableCursor < len(m.periodRows)-1 {
			m.tableCursor++
		}
	}
	m.settings.LastTransactionID = m.selectedTransactionID()
	_ = SaveSettings(m.settings)
	return m, nil
}

// submitBudget mirrors on_input_submitted's "budget-input" branch: a
// non-numeric or negative value is silently ignored (no save, no notice) -
// Textual's Number(minimum=0) validator rejects it before the handler's
// own try/except would even run.
func (m *model) submitBudget() {
	value, err := strconv.ParseFloat(strings.TrimSpace(m.budgetInput.Value()), 64)
	if err != nil || value < 0 {
		return
	}
	m.settings.MonthlyBudget = value
	_ = SaveSettings(m.settings)
	m.noticeMessage = "Monthly budget set to " + fmtAmount(value, m.settings.Currency)
}

// -- navigation ----------------------------------------------------------

func (m *model) nextStage() {
	idx := stageIndex(m.stage)
	if idx < len(stageOrder)-1 {
		m.setStage(stageOrder[idx+1])
	}
}

func (m *model) prevStage() {
	idx := stageIndex(m.stage)
	if idx > 0 {
		m.setStage(stageOrder[idx-1])
	}
}

func (m *model) setStage(s stage) {
	m.stage = s
	switch s {
	case stageYear:
		m.yearFilter.SetValue("")
		m.populateYearList()
		m.yearFilter.Focus()
	case stageMonth:
		m.monthFilter.SetValue("")
		m.populateMonthList()
		m.monthFilter.Focus()
	default:
		m.budgetFocused = false
		m.budgetInput.Blur()
	}
}

func (m *model) selectYear(year string) {
	m.selectedYear = year
	m.refreshAll("")
	m.setStage(stageMonth)
}

func (m *model) selectMonth(month string) {
	m.selectedMonth = month
	m.settings.LastYear = m.selectedYear
	m.settings.LastMonth = m.selectedMonth
	_ = SaveSettings(m.settings)
	m.refreshAll("")
	m.setStage(stageTable)
}

// -- year / month list population ----------------------------------------

func (m *model) populateYearList() {
	query := strings.TrimSpace(m.yearFilter.Value())
	years := map[string]bool{fmt.Sprintf("%04d", time.Now().Year()): true}
	for _, t := range m.transactions {
		years[dateYear(t.Date)] = true
	}
	sorted := make([]string, 0, len(years))
	for y := range years {
		if query == "" || strings.Contains(y, query) {
			sorted = append(sorted, y)
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(sorted)))

	items := make([]periodItem, len(sorted))
	for i, y := range sorted {
		spent := totalSpent(filterByYear(m.transactions, y))
		items[i] = periodItem{value: y, label: fmt.Sprintf("%s    %s spent", y, fmtAmount(spent, m.settings.Currency))}
	}
	m.yearItems = items
	m.yearCursor = highlightPeriod(items, m.selectedYear)
}

func (m *model) populateMonthList() {
	query := strings.ToLower(strings.TrimSpace(m.monthFilter.Value()))
	var items []periodItem
	for i := 1; i <= 12; i++ {
		mm := fmt.Sprintf("%02d", i)
		name := monthNames[i-1]
		if query != "" && !strings.Contains(mm, query) && !strings.Contains(strings.ToLower(name), query) {
			continue
		}
		period := m.selectedYear + "-" + mm
		spent := totalSpent(filterByMonth(m.transactions, period))
		items = append(items, periodItem{value: mm, label: fmt.Sprintf("%s %s %s spent", mm, padRight(name, 10), fmtAmount(spent, m.settings.Currency))})
	}
	m.monthItems = items
	m.monthCursor = highlightPeriod(items, m.selectedMonth)
}

// highlightPeriod mirrors _highlight_option: land the cursor back on
// whichever item's value matches want, or the first item if it isn't
// present in the (possibly filtered) list.
func highlightPeriod(items []periodItem, want string) int {
	for i, it := range items {
		if it.value == want {
			return i
		}
	}
	return 0
}

func filterByYear(transactions []Transaction, year string) []Transaction {
	var out []Transaction
	for _, t := range transactions {
		if dateYear(t.Date) == year {
			out = append(out, t)
		}
	}
	return out
}

func filterByMonth(transactions []Transaction, period string) []Transaction {
	var out []Transaction
	for _, t := range transactions {
		if dateMonth(t.Date) == period {
			out = append(out, t)
		}
	}
	return out
}

// -- table refresh / selection --------------------------------------------

// refreshTable mirrors refresh_table: rows scoped to the selected period,
// most recent first (a stable descending sort, so same-date transactions
// keep their original relative order - matching Python's sorted(...,
// reverse=True)) - the cursor re-lands on selectID (or whatever was
// already selected if selectID is ""). Cell coloring happens at render
// time (renderTableStage in view.go), not here - see periodRows/
// tableCursor's doc comment for why this table isn't bubbles/table.
func (m *model) refreshTable(selectID string) {
	target := selectID
	if target == "" {
		target = m.selectedTransactionID()
	}

	rows := filterByMonth(m.transactions, m.selectedPeriod())
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Date > rows[j].Date })
	m.periodRows = rows

	m.tableCursor = 0
	if target != "" {
		for i, t := range rows {
			if t.ID == target {
				m.tableCursor = i
				break
			}
		}
	}
	m.settings.LastTransactionID = m.selectedTransactionID()
	_ = SaveSettings(m.settings)
}

func (m *model) refreshAll(selectID string) {
	m.refreshTable(selectID)
}

func (m model) selectedTransactionID() string {
	if m.tableCursor < 0 || m.tableCursor >= len(m.periodRows) {
		return ""
	}
	return m.periodRows[m.tableCursor].ID
}

func (m model) getSelectedTransaction() *Transaction {
	id := m.selectedTransactionID()
	if id == "" {
		return nil
	}
	for i := range m.transactions {
		if m.transactions[i].ID == id {
			return &m.transactions[i]
		}
	}
	return nil
}

// -- transaction mutation --------------------------------------------------

func (m *model) startAdd() {
	defaultDate := fmt.Sprintf("%s-%s-01", m.selectedYear, m.selectedMonth)
	m.form = newTransactionForm(nil, defaultDate)
	m.modal = modalTransactionForm
}

func (m *model) startEdit() {
	current := m.getSelectedTransaction()
	if current == nil {
		return
	}
	cp := *current
	m.form = newTransactionForm(&cp, "")
	m.modal = modalTransactionForm
}

func (m *model) startDelete() {
	current := m.getSelectedTransaction()
	if current == nil {
		return
	}
	m.confirmTarget = current.ID
	message := fmt.Sprintf("Delete %s  %s  %s?", current.Date, current.Category, formatComma2(current.Amount))
	m.confirm = newConfirmDelete(message)
	m.modal = modalConfirmDelete
}

func (m *model) handleFormResult(result Transaction) {
	if m.form.editing != nil {
		for i := range m.transactions {
			if m.transactions[i].ID == result.ID {
				m.transactions[i] = result
				break
			}
		}
	} else {
		m.transactions = append(m.transactions, result)
	}
	_ = SaveTransactions(m.transactions)
	m.followPeriod(result.Date)
	m.refreshAll(result.ID)
}

func (m *model) handleDeleteConfirm(confirmed bool) {
	if !confirmed {
		return
	}
	kept := make([]Transaction, 0, len(m.transactions))
	for _, t := range m.transactions {
		if t.ID != m.confirmTarget {
			kept = append(kept, t)
		}
	}
	m.transactions = kept
	_ = SaveTransactions(m.transactions)
	m.refreshAll("")
}

// followPeriod mirrors _follow_period: adding/editing a transaction dated
// outside the currently-viewed month re-points the view at that
// transaction's month.
func (m *model) followPeriod(date string) {
	if dateMonth(date) == m.selectedPeriod() {
		return
	}
	m.selectedYear, m.selectedMonth = dateYear(date), date[5:7]
	m.settings.LastYear = m.selectedYear
	m.settings.LastMonth = m.selectedMonth
	_ = SaveSettings(m.settings)
}
