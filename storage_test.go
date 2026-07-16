package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadTransactionsMissingFileReturnsEmpty(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	got, err := LoadTransactions()
	if err != nil {
		t.Fatalf("LoadTransactions: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected no transactions, got %d", len(got))
	}
}

func TestSaveLoadTransactionsRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	want := []Transaction{NewTransaction("2026-07-01", "Food", 42.5, "Lunch")}
	if err := SaveTransactions(want); err != nil {
		t.Fatalf("SaveTransactions: %v", err)
	}
	got, err := LoadTransactions()
	if err != nil {
		t.Fatalf("LoadTransactions: %v", err)
	}
	if len(got) != 1 || got[0] != want[0] {
		t.Fatalf("round-trip mismatch: got %+v, want %+v", got, want)
	}
}

// TestLoadTransactionsBackfillsMissingID mirrors load_transactions's
// backfill for legacy rows that predate the id field.
func TestLoadTransactionsBackfillsMissingID(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path, err := transactionsPath()
	if err != nil {
		t.Fatalf("transactionsPath: %v", err)
	}
	raw := `[{"date":"2026-07-01","category":"Food","amount":10,"note":""}]`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := LoadTransactions()
	if err != nil {
		t.Fatalf("LoadTransactions: %v", err)
	}
	if len(got) != 1 || got[0].ID == "" {
		t.Fatalf("expected a backfilled id, got %+v", got)
	}

	// The backfill should have been persisted too.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(data), got[0].ID) {
		t.Fatal("expected the backfilled id to be saved back to disk")
	}
}

func TestLoadSettingsDefaultsMissingFields(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path, err := settingsPath()
	if err != nil {
		t.Fatalf("settingsPath: %v", err)
	}
	if err := os.WriteFile(path, []byte(`{"currency":"EUR"}`), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings: %v", err)
	}
	if got.Currency != "EUR" {
		t.Fatalf("Currency = %q, want EUR", got.Currency)
	}
	if got.MonthlyBudget != 0.0 {
		t.Fatalf("MonthlyBudget = %v, want the zero-value default", got.MonthlyBudget)
	}
}

// withFakeHome points $HOME (and, on Windows, the equivalent
// os.UserHomeDir() actually reads) at a scratch dir for the duration of
// the test, so export tests never touch the real user's ~/Downloads.
func withFakeHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	return home
}

func testExportFormat(t *testing.T, export func([]Transaction) (string, error), wantExt string) {
	t.Helper()
	home := withFakeHome(t)
	transactions := []Transaction{
		NewTransaction("2026-07-01", "Food", 100, "Groceries"),
		NewTransaction("2026-07-02", "Transport", 50, "Bus"),
	}
	path, err := export(transactions)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if filepath.Ext(path) != wantExt {
		t.Fatalf("expected a %s file, got %s", wantExt, path)
	}
	wantDir := filepath.Join(home, "Downloads")
	if filepath.Dir(path) != wantDir {
		t.Fatalf("expected the export in %s, got %s", wantDir, path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected the export file to exist: %v", err)
	}
}

func TestExportCSVWritesRunningBalance(t *testing.T) {
	home := withFakeHome(t)
	transactions := []Transaction{
		NewTransaction("2026-07-01", "Food", 100, "Groceries"),
		NewTransaction("2026-07-02", "Transport", 50, "Bus"),
	}
	path, err := ExportCSV(transactions)
	if err != nil {
		t.Fatalf("ExportCSV: %v", err)
	}
	if filepath.Dir(path) != filepath.Join(home, "Downloads") {
		t.Fatalf("expected the export in ~/Downloads, got %s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)
	for _, want := range []string{"date,category,amount,note,balance", "-100.00", "-150.00"} {
		if !strings.Contains(content, want) {
			t.Fatalf("export missing %q, got:\n%s", want, content)
		}
	}
}

func TestExportXLSXWritesToDownloads(t *testing.T) {
	testExportFormat(t, ExportXLSX, ".xlsx")
}

func TestExportJSONWritesToDownloads(t *testing.T) {
	testExportFormat(t, ExportJSON, ".json")
}

func TestSettingsJSONOmitsUnsetOptionalFields(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := SaveSettings(DefaultSettings()); err != nil {
		t.Fatalf("SaveSettings: %v", err)
	}
	path, err := settingsPath()
	if err != nil {
		t.Fatalf("settingsPath: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	for _, key := range []string{"last_year", "last_month", "last_transaction_id", "theme"} {
		if _, present := raw[key]; present {
			t.Fatalf("expected %q to be omitted from a fresh DefaultSettings(), got: %v", key, raw)
		}
	}
}

// TestThemeChoicePersistsAcrossLaunches guards against pear reverting to
// always reopening on ansi-dark - a picked theme must round-trip through
// settings.json and get applied by the next NewModel call, unlike the
// Python original (which intentionally forgets it every launch).
func TestThemeChoicePersistsAcrossLaunches(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	defer setActiveTheme("ansi-dark") // reset for other tests sharing package-level theme state

	settings := DefaultSettings()
	settings.Theme = "dracula"
	if err := SaveSettings(settings); err != nil {
		t.Fatalf("SaveSettings: %v", err)
	}

	loaded, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings: %v", err)
	}
	if loaded.Theme != "dracula" {
		t.Fatalf("expected loaded settings to keep the saved theme, got %q", loaded.Theme)
	}

	setActiveTheme("ansi-dark") // simulate the package-level default before a fresh launch
	NewModel(nil, loaded)
	if activeTheme.Name != "dracula" {
		t.Fatalf("expected NewModel to apply the saved theme, got %q active", activeTheme.Name)
	}
}
