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

// renderFieldRow/renderButtonRow match splitsy's (pear's Go sibling)
// add/edit form chrome exactly: a fixed-width label (bold+accent with a
// "▸" prefix when focused, muted otherwise) directly followed by the
// input's own value - no per-field box - and a row of plain-colored-text
// buttons with no visible button chrome at all.
func renderFieldRow(label string, input textinput.Model, focused bool) string {
	labelStyle := styleFieldLabel
	prefix := "  "
	if focused {
		labelStyle = styleFieldFocused
		prefix = styleAccent.Render("▸") + " "
	}
	return prefix + labelStyle.Render(label) + input.View()
}

func renderButtonRow(labels []string, styles []lipgloss.Style) string {
	parts := make([]string, len(labels))
	for i, l := range labels {
		parts[i] = styles[i].Render(l)
	}
	return strings.Join(parts, "    ")
}

// transactionFormModel mirrors TransactionForm: a modal for adding a new
// transaction or editing an existing one, four fields (date, category,
// amount, note), Tab cycling between them, Enter from any field
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
		ti.Prompt = ""
		ti.Placeholder = placeholder
		ti.Width = 30
		ti.SetValue(value)
		return ti
	}

	f := transactionFormModel{editing: editing}
	if editing != nil {
		f.date = mk("YYYY-MM-DD", editing.Date)
		f.category = mk("e.g. Food", editing.Category)
		f.amount = mk("negative = income", formatFixed2(editing.Amount))
		f.note = mk("optional", editing.Note)
	} else {
		f.date = mk("YYYY-MM-DD", defaultDate)
		f.category = mk("e.g. Food", "")
		f.amount = mk("negative = income", "")
		f.note = mk("optional", "")
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

func (f transactionFormModel) title() string {
	if f.editing != nil {
		return "Edit transaction"
	}
	return "Add transaction"
}

func (f transactionFormModel) View() string {
	var b strings.Builder
	b.WriteString(styleModalTitle.Render(f.title()))
	b.WriteString("\n")
	b.WriteString(renderFieldRow("Date", f.date, f.focus == 0))
	b.WriteString("\n")
	b.WriteString(renderFieldRow("Category", f.category, f.focus == 1))
	b.WriteString("\n")
	b.WriteString(renderFieldRow("Amount", f.amount, f.focus == 2))
	b.WriteString("\n")
	b.WriteString(renderFieldRow("Note", f.note, f.focus == 3))
	b.WriteString("\n\n")
	if f.errorMsg != "" {
		b.WriteString(styleError.Render(f.errorMsg))
		b.WriteString("\n\n")
	}
	b.WriteString(renderButtonRow([]string{"Cancel", "Save"}, []lipgloss.Style{styleButtonCancel, styleButtonSave}))
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
	b.WriteString(styleModalTitle.Render("Confirm delete"))
	b.WriteString("\n")
	b.WriteString(c.message)
	b.WriteString("\n\n")
	b.WriteString(renderButtonRow([]string{"Cancel (esc)", "Delete (enter)"}, []lipgloss.Style{styleButtonCancel, styleButtonDelete}))
	return styleModalBoxDanger.Render(b.String())
}
