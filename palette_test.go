package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPaletteExportOpensFormatSubmenu(t *testing.T) {
	m := freshModel(t)
	m = send(t, m, tea.KeyMsg{Type: tea.KeyCtrlP})
	if m.modal != modalPalette {
		t.Fatal("expected the command palette to open")
	}
	m = send(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Export")})
	m = send(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.palette.stage != paletteExport {
		t.Fatalf("expected the export format sub-page, got %v", m.palette.stage)
	}
	view := m.palette.View()
	for _, want := range []string{"CSV", "XLSX", "JSON"} {
		if !strings.Contains(view, want) {
			t.Fatalf("export page missing %q, got:\n%s", want, view)
		}
	}
}

func testPaletteExportFormat(t *testing.T, query, wantExt string) {
	t.Helper()
	home := withFakeHome(t)
	m := freshModel(t)
	m = send(t, m, tea.KeyMsg{Type: tea.KeyCtrlP})
	m = send(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Export")})
	m = send(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	m = send(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(query)})
	m = send(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.modal != modalNone {
		t.Fatal("expected export to execute immediately and close the palette")
	}
	if m.noticeMessage == "" {
		t.Fatal("expected an export confirmation notice")
	}
	if !strings.Contains(m.noticeMessage, "Downloads") {
		t.Fatalf("expected the confirmation to mention ~/Downloads, got %q", m.noticeMessage)
	}

	dir := filepath.Join(home, "Downloads")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	found := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "export-") && filepath.Ext(e.Name()) == wantExt {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a %s export file in ~/Downloads, got entries: %v", wantExt, entries)
	}
}

func TestPaletteExportCSV(t *testing.T)  { testPaletteExportFormat(t, "CSV", ".csv") }
func TestPaletteExportXLSX(t *testing.T) { testPaletteExportFormat(t, "XLSX", ".xlsx") }
func TestPaletteExportJSON(t *testing.T) { testPaletteExportFormat(t, "JSON", ".json") }

func TestPaletteThemeSelectionRecolors(t *testing.T) {
	m := freshModel(t)
	originalAccent := activeTheme.Accent

	m = send(t, m, tea.KeyMsg{Type: tea.KeyCtrlP})
	m = send(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Theme")})
	m = send(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.palette.stage != paletteTheme {
		t.Fatalf("expected the theme sub-page, got %v", m.palette.stage)
	}

	m = send(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("dracula")})
	m = send(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.modal != modalNone {
		t.Fatal("expected the palette to close after selecting a theme")
	}
	if activeTheme.Accent == originalAccent {
		t.Fatal("expected activeTheme to actually change color")
	}

	setActiveTheme("ansi-dark") // reset for other tests sharing package-level theme state
}
