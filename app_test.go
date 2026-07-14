package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func send(t *testing.T, m model, msg tea.Msg) model {
	t.Helper()
	updated, _ := m.Update(msg)
	mm, ok := updated.(model)
	if !ok {
		t.Fatalf("Update did not return a model, got %T", updated)
	}
	return mm
}

func freshModel(t *testing.T) model {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	transactions := seedIfEmpty(nil)
	m := NewModel(transactions, DefaultSettings())
	return send(t, m, tea.WindowSizeMsg{Width: 140, Height: 36})
}

func TestInitialStageIsTable(t *testing.T) {
	m := freshModel(t)
	if m.stage != stageTable {
		t.Fatalf("expected initial stage to be table (Python hard-codes self.stage = \"table\"), got %v", m.stage)
	}
	if len(m.periodRows) == 0 {
		t.Fatal("expected the seeded July transactions to populate the table")
	}
}

func TestNextPrevStageCycles(t *testing.T) {
	m := freshModel(t)
	if m.stage != stageTable {
		t.Fatalf("expected to start at table stage, got %v", m.stage)
	}
	m.prevStage()
	if m.stage != stageMonth {
		t.Fatalf("prevStage from table should step to month, got %v", m.stage)
	}
	m.prevStage()
	if m.stage != stageYear {
		t.Fatalf("prevStage from month should step to year, got %v", m.stage)
	}
	m.prevStage() // already at the first stage - no-op
	if m.stage != stageYear {
		t.Fatalf("prevStage at the year boundary should stay put, got %v", m.stage)
	}
	m.nextStage()
	m.nextStage()
	if m.stage != stageTable {
		t.Fatalf("expected two nextStage calls to reach table, got %v", m.stage)
	}
	m.nextStage() // already at the last stage - no-op
	if m.stage != stageTable {
		t.Fatalf("nextStage at the table boundary should stay put, got %v", m.stage)
	}
}

func TestStageIndexHelpers(t *testing.T) {
	if stageIndex(stageYear) != 0 || stageIndex(stageMonth) != 1 || stageIndex(stageTable) != 2 {
		t.Fatal("stageOrder must be year, month, table in that order")
	}
}

func TestAddTransactionPersistsAndFollowsPeriod(t *testing.T) {
	m := freshModel(t)
	before := len(m.transactions)

	m.startAdd()
	if m.modal != modalTransactionForm {
		t.Fatal("expected the add form to open")
	}
	m.form.date.SetValue("2026-09-15")
	m.form.category.SetValue("Health")
	m.form.amount.SetValue("99.99")
	m.form.note.SetValue("Checkup")

	result, closed, _ := m.form.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if !closed || result == nil {
		t.Fatal("expected a valid submission to close the form with a result")
	}
	m.modal = modalNone
	m.handleFormResult(*result)

	if len(m.transactions) != before+1 {
		t.Fatalf("expected %d transactions, got %d", before+1, len(m.transactions))
	}
	if m.selectedYear != "2026" || m.selectedMonth != "09" {
		t.Fatalf("expected the view to follow the new transaction's period, got %s-%s", m.selectedYear, m.selectedMonth)
	}
}

func TestDeleteTransactionRemovesIt(t *testing.T) {
	m := freshModel(t)
	before := len(m.transactions)
	target := m.getSelectedTransaction()
	if target == nil {
		t.Fatal("expected a selected transaction in the seeded ledger")
	}
	m.confirmTarget = target.ID
	m.handleDeleteConfirm(true)

	if len(m.transactions) != before-1 {
		t.Fatalf("expected %d transactions after delete, got %d", before-1, len(m.transactions))
	}
	for _, tx := range m.transactions {
		if tx.ID == target.ID {
			t.Fatal("deleted transaction is still present")
		}
	}
}

func TestFormValidationRejectsBadDate(t *testing.T) {
	f := newTransactionForm(nil, "2026-07-01")
	f.category.SetValue("Food")
	f.amount.SetValue("10")
	f.date.SetValue("not-a-date")

	_, err := f.attemptSave()
	if err == "" {
		t.Fatal("expected a validation error for a malformed date")
	}
}

func TestSelectYearAdvancesToMonthStage(t *testing.T) {
	m := freshModel(t)
	m.setStage(stageYear)
	m.selectYear("2026")
	if m.stage != stageMonth {
		t.Fatalf("expected selectYear to advance to month stage, got %v", m.stage)
	}
	if m.selectedYear != "2026" {
		t.Fatalf("selectedYear = %q, want 2026", m.selectedYear)
	}
}
