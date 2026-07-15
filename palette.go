package main

import (
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// The command palette (ctrl+p) mirrors Textual's own built-in command
// palette, which tui.py adds "Currency" and "Export CSV" entries to
// alongside Textual's built-in "Theme" - here it's all one hand-rolled
// palette, since there's no framework providing it for free. Rendering is
// deliberately styled to match splitsy (finz's Go sibling) one for one -
// a single-line "▸ label  hint" entry per row - rather than Textual's own
// two-line title+description palette look. "Keybindings" is this port's
// own addition (Textual's palette also lists framework commands like
// Keys/Maximize/Screenshot/Quit that don't have a meaningful terminal-app
// equivalent here, so they're swapped for a reference page instead), and
// "Export" is a sub-page offering CSV/XLSX/JSON, matching splitsy's own
// Export sub-page exactly.

type paletteStage string

const (
	paletteRoot     paletteStage = "root"
	paletteCurrency paletteStage = "currency"
	paletteTheme    paletteStage = "theme"
	paletteExport   paletteStage = "export"
	paletteKeybinds paletteStage = "keybinds"
)

type paletteOption struct {
	id    string
	title string
	desc  string
}

type paletteModel struct {
	stage  paletteStage
	filter textinput.Model
	cursor int
}

func newPalette() paletteModel {
	ti := textinput.New()
	ti.Prompt = "/ "
	ti.Placeholder = "Type a command..."
	ti.Focus()
	return paletteModel{stage: paletteRoot, filter: ti}
}

func rootPaletteOptions() []paletteOption {
	return []paletteOption{
		{id: "currency", title: "Currency", desc: "Choose the currency amounts are displayed in"},
		{id: "theme", title: "Theme", desc: "Change the current theme"},
		{id: "export", title: "Export", desc: "Write all activity to a file"},
		{id: "keybinds", title: "Keybindings", desc: "Show all keyboard shortcuts"},
	}
}

// exportPaletteOptions lists the file formats the Export sub-page offers -
// every one of them lands in ~/Downloads (see downloadsDir in storage.go),
// same as any other download a browser hands you.
func exportPaletteOptions() []paletteOption {
	return []paletteOption{
		{id: "csv", title: "CSV", desc: "Opens in any spreadsheet app"},
		{id: "xlsx", title: "XLSX", desc: "Excel workbook"},
		{id: "json", title: "JSON", desc: "For scripts and other tools"},
	}
}

func currencyPaletteOptions() []paletteOption {
	opts := make([]paletteOption, len(currencyOrder))
	for i, code := range currencyOrder {
		opts[i] = paletteOption{id: code, title: code + " (" + currencySymbols[code] + ")", desc: "Display amounts in " + code}
	}
	return opts
}

func themePaletteOptions() []paletteOption {
	opts := make([]paletteOption, len(themes))
	for i, t := range themes {
		desc := ""
		if t.Name == activeTheme.Name {
			desc = "current"
		}
		opts[i] = paletteOption{id: t.Name, title: t.Name, desc: desc}
	}
	return opts
}

// filterOptions applies the palette's live substring filter against the
// title, approximating Textual's fuzzy command-palette search closely
// enough for these small, fixed option lists.
func filterOptions(options []paletteOption, query string) []paletteOption {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return options
	}
	var out []paletteOption
	for _, o := range options {
		if strings.Contains(strings.ToLower(o.title), query) {
			out = append(out, o)
		}
	}
	return out
}

func (m *model) currentPaletteOptions() []paletteOption {
	switch m.palette.stage {
	case paletteCurrency:
		return filterOptions(currencyPaletteOptions(), m.palette.filter.Value())
	case paletteTheme:
		return filterOptions(themePaletteOptions(), m.palette.filter.Value())
	case paletteExport:
		return filterOptions(exportPaletteOptions(), m.palette.filter.Value())
	case paletteKeybinds:
		return nil
	default:
		return filterOptions(rootPaletteOptions(), m.palette.filter.Value())
	}
}

func (m *model) openPalette() {
	m.palette = newPalette()
	m.modal = modalPalette
}

func (m *model) handlePaletteKey(msg tea.KeyMsg) tea.Cmd {
	p := &m.palette
	opts := m.currentPaletteOptions()

	switch msg.String() {
	case "esc":
		if p.stage == paletteRoot {
			m.closeModal()
		} else {
			p.stage = paletteRoot
			p.filter.SetValue("")
			p.cursor = 0
		}
		return nil
	case "enter":
		m.submitPaletteSelection(opts)
		return nil
	case "up":
		if p.cursor > 0 {
			p.cursor--
		}
		return nil
	case "down":
		if p.cursor < len(opts)-1 {
			p.cursor++
		}
		return nil
	}

	if p.stage == paletteKeybinds {
		return nil
	}
	var cmd tea.Cmd
	p.filter, cmd = p.filter.Update(msg)
	p.cursor = 0
	return cmd
}

func (m *model) submitPaletteSelection(opts []paletteOption) {
	p := &m.palette
	if p.stage != paletteKeybinds && p.cursor >= len(opts) {
		return
	}

	switch p.stage {
	case paletteRoot:
		switch opts[p.cursor].id {
		case "currency":
			p.stage, p.cursor = paletteCurrency, 0
			p.filter.SetValue("")
		case "theme":
			p.stage, p.cursor = paletteTheme, 0
			p.filter.SetValue("")
		case "export":
			p.stage, p.cursor = paletteExport, 0
			p.filter.SetValue("")
		case "keybinds":
			p.stage = paletteKeybinds
		}
	case paletteExport:
		switch opts[p.cursor].id {
		case "csv":
			m.exportFile(ExportCSV)
		case "xlsx":
			m.exportFile(ExportXLSX)
		case "json":
			m.exportFile(ExportJSON)
		}
		m.closeModal()
	case paletteCurrency:
		code := opts[p.cursor].id
		m.settings.Currency = code
		_ = SaveSettings(m.settings)
		m.refreshAll("")
		m.noticeMessage = "Currency set to " + code
		m.closeModal()
	case paletteTheme:
		name := opts[p.cursor].id
		if setActiveTheme(name) {
			m.settings.Theme = name
			_ = SaveSettings(m.settings)
			m.noticeMessage = "Theme changed"
		}
		m.closeModal()
	case paletteKeybinds:
		m.closeModal()
	}
}

// exportFile runs any of the ExportXxx([]Transaction) (string, error)
// functions in storage.go and turns the result into the same notice/
// error banner treatment, regardless of which format was picked on the
// Export sub-page - mirrors splitsy's exportFile exactly.
func (m *model) exportFile(export func([]Transaction) (string, error)) {
	path, err := export(m.transactions)
	if err != nil {
		m.errorMessage = "Export failed: " + err.Error()
		return
	}
	m.noticeMessage = "Exported to ~/Downloads/" + filepath.Base(path)
}

// -- rendering --------------------------------------------------------------

func (p paletteModel) View() string {
	switch p.stage {
	case paletteCurrency:
		return renderPaletteList("Currency", p.filter, filterOptions(currencyPaletteOptions(), p.filter.Value()), p.cursor)
	case paletteTheme:
		return renderPaletteList("Theme", p.filter, filterOptions(themePaletteOptions(), p.filter.Value()), p.cursor)
	case paletteExport:
		return renderPaletteList("Export", p.filter, filterOptions(exportPaletteOptions(), p.filter.Value()), p.cursor)
	case paletteKeybinds:
		return renderKeybindsPage()
	default:
		return renderPaletteList("Command Palette", p.filter, filterOptions(rootPaletteOptions(), p.filter.Value()), p.cursor)
	}
}

// renderPaletteList matches splitsy's renderPaletteList exactly: a bold
// title, a rounded-border filter box, then one "▸ label  hint" line per
// option (bold + accent arrow when selected, plain otherwise) - no boxed
// rows, no two-line title/description layout.
func renderPaletteList(title string, filter textinput.Model, opts []paletteOption, cursor int) string {
	var b strings.Builder
	b.WriteString(styleModalTitle.Render(title))
	b.WriteString("\n")
	b.WriteString(styleFilterBox.Width(36).Render(filter.View()))
	b.WriteString("\n\n")

	if len(opts) == 0 {
		b.WriteString(styleMuted.Render("No matches"))
		b.WriteString("\n")
	}
	for i, o := range opts {
		prefix := "  "
		nameStyle := lipgloss.NewStyle()
		if i == cursor {
			prefix = styleAccent.Render("▸") + " "
			nameStyle = styleBold
		}
		line := prefix + nameStyle.Render(o.title)
		if o.desc != "" {
			line += "  " + styleMuted.Render(o.desc)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(styleMuted.Render("esc: back   enter: select"))
	return styleModalBox.Render(b.String())
}

func renderKeybindsPage() string {
	var b strings.Builder
	b.WriteString(styleModalTitle.Render("Keybindings"))
	b.WriteString("\n")

	section := func(name string, rows [][2]string) {
		b.WriteString(styleBold.Render(name))
		b.WriteString("\n")
		for _, r := range rows {
			b.WriteString(styleAccent.Render(padRight(r[0], 14)) + r[1])
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	section("Global", [][2]string{
		{"ctrl+p", "command palette"},
		{"q / ctrl+c", "quit"},
	})
	section("Navigation", [][2]string{
		{"alt+→", "next stage (year → month → table)"},
		{"alt+←", "prev stage"},
		{"↑ / ↓", "move highlight (year/month) or table row"},
		{"tab", "switch focus to/from the budget input (table stage)"},
	})
	section("Ledger (table stage)", [][2]string{
		{"a", "add transaction"},
		{"e", "edit transaction"},
		{"d", "delete transaction"},
	})

	b.WriteString(styleMuted.Render("esc/enter: back"))
	return styleModalBox.Render(b.String())
}
