package main

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var dateRE = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// transactionFormModel mirrors TransactionForm: a modal for adding a new
// transaction or editing an existing one, four boxed inputs (date,
// category, amount, note), Tab cycling between them, Enter from any field
// attempting a save (on_input_submitted fires for every Input in the
// screen, not just the last one), Esc cancelling.
type transactionFormModel struct {
	editing *Transaction // nil when adding

	date     textinput.Model
	category textinput.Model
	amount   textinput.Model
	note     textinput.Model
	focus    int // 0-3, indexes {date, category, amount, note}

	errorMsg string
}

func newTransactionForm(editing *Transaction, defaultDate string) transactionFormModel {
	mk := func(placeholder, value string) textinput.Model {
		ti := textinput.New()
		ti.Placeholder = placeholder
		ti.SetValue(value)
		return ti
	}

	f := transactionFormModel{editing: editing}
	if editing != nil {
		f.date = mk("Date (YYYY-MM-DD)", editing.Date)
		f.category = mk("Category", editing.Category)
		f.amount = mk("Amount (negative = income)", formatFixed2(editing.Amount))
		f.note = mk("Note (optional)", editing.Note)
	} else {
		f.date = mk("Date (YYYY-MM-DD)", defaultDate)
		f.category = mk("Category", "")
		f.amount = mk("Amount (negative = income)", "")
		f.note = mk("Note (optional)", "")
	}
	f.date.Focus()
	return f
}

func (f *transactionFormModel) fields() []*textinput.Model {
	return []*textinput.Model{&f.date, &f.category, &f.amount, &f.note}
}

func (f *transactionFormModel) focusField(i int) {
	fields := f.fields()
	for idx, field := range fields {
		if idx == i {
			field.Focus()
		} else {
			field.Blur()
		}
	}
	f.focus = i
}

// handleKey processes a key while the form is open. Returns (result,
// handled) - result is non-nil only once the user has successfully saved
// (mirrors TransactionForm.dismiss(result)); handled tells the caller
// whether to close the modal (save succeeded, or cancel).
func (f *transactionFormModel) handleKey(msg tea.KeyMsg) (result *Transaction, closed bool, cmd tea.Cmd) {
	switch msg.String() {
	case "esc":
		return nil, true, nil
	case "tab", "down":
		f.focusField((f.focus + 1) % 4)
		return nil, false, nil
	case "shift+tab", "up":
		f.focusField((f.focus + 3) % 4)
		return nil, false, nil
	case "enter":
		t, err := f.attemptSave()
		if err != "" {
			f.errorMsg = err
			return nil, false, nil
		}
		return &t, true, nil
	}

	fields := f.fields()
	var c tea.Cmd
	*fields[f.focus], c = fields[f.focus].Update(msg)
	return nil, false, c
}

// attemptSave mirrors TransactionForm.attempt_save: validates every field,
// returning the first error message (matching Python's return-on-first-
// failure order: date, category, amount) or a built Transaction.
func (f *transactionFormModel) attemptSave() (Transaction, string) {
	date := strings.TrimSpace(f.date.Value())
	category := strings.TrimSpace(f.category.Value())
	amountStr := strings.TrimSpace(f.amount.Value())
	note := strings.TrimSpace(f.note.Value())

	if !dateRE.MatchString(date) {
		return Transaction{}, "Date must be in YYYY-MM-DD format"
	}
	if category == "" {
		return Transaction{}, "Category is required"
	}
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return Transaction{}, "Amount must be a number"
	}

	if f.editing != nil {
		return Transaction{Date: date, Category: category, Amount: amount, Note: note, ID: f.editing.ID}, ""
	}
	return NewTransaction(date, category, amount, note), ""
}

func (f transactionFormModel) View() string {
	title := "Add transaction"
	if f.editing != nil {
		title = "Edit transaction"
	}

	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n\n")
	for _, field := range f.fields() {
		b.WriteString(renderTallBox(field.View(), 46, 2))
		b.WriteString("\n\n")
	}
	if f.errorMsg != "" {
		b.WriteString(styleError.Render(f.errorMsg))
	}
	b.WriteString("\n")
	b.WriteString(centerLine(styleButtonCancel.Render("Cancel")+"   "+styleButtonSave.Render("Save"), 50))

	return styleModalBox.Render(b.String())
}

// confirmDeleteModel mirrors ConfirmDelete: a plain confirmation dialog,
// Enter deletes (this port's keyboard-first equivalent of clicking the
// "Delete" text button), Esc cancels.
type confirmDeleteModel struct {
	message string
}

func newConfirmDelete(message string) confirmDeleteModel {
	return confirmDeleteModel{message: message}
}

func (c confirmDeleteModel) View() string {
	var b strings.Builder
	b.WriteString(c.message)
	b.WriteString("\n\n")
	b.WriteString(centerLine(styleButtonCancel.Render("Cancel")+"   "+styleButtonDelete.Render("Delete"), 46))
	return styleModalBoxDanger.Width(46).Render(b.String())
}

func centerLine(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	left := (width - w) / 2
	return strings.Repeat(" ", left) + s
}
