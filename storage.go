package main

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/xuri/excelize/v2"
)

// Settings mirrors storage.py's settings dict. LastYear/LastMonth/
// LastTransactionID are the "reopen exactly where I left off" state tui.py
// writes on every navigation/selection change - omitempty so a fresh
// settings.json (matching Python's DEFAULT_SETTINGS) doesn't grow keys
// that were never set.
type Settings struct {
	Currency          string  `json:"currency"`
	MonthlyBudget     float64 `json:"monthly_budget"`
	LastTransactionID string  `json:"last_transaction_id,omitempty"`
	LastYear          string  `json:"last_year,omitempty"`
	LastMonth         string  `json:"last_month,omitempty"`
	Theme             string  `json:"theme,omitempty"`
}

// DefaultSettings mirrors storage.py's DEFAULT_SETTINGS.
func DefaultSettings() Settings {
	return Settings{Currency: "USD", MonthlyBudget: 0.0}
}

// dataDir returns ~/.config/pear (or the platform equivalent), creating it
// if necessary. The Python original keeps transactions.json/settings.json
// next to the source files; a Go binary has no such fixed "next to the
// script" location once installed, so this uses the same XDG convention as
// splitsy (pear's Go sibling) instead - an internal storage-location
// choice, not a user-visible behavior difference.
func dataDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "pear")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// downloadsDir returns ~/Downloads, creating it if necessary. Every export
// format lands here rather than in dataDir() - dataDir is where pear's
// own state lives, but an export is something the user asked for and
// wants to actually find, and every OS already puts a "Downloads" folder
// in front of the user for exactly that. Matches splitsy (pear's Go
// sibling) exactly.
func downloadsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, "Downloads")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func transactionsPath() (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "transactions.json"), nil
}

func settingsPath() (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.json"), nil
}

// LoadTransactions reads transactions.json. A missing file is not an error:
// it returns an empty slice so the caller can decide whether to seed
// sample data, mirroring load_transactions's `if not DATA_FILE.exists()`.
func LoadTransactions() ([]Transaction, error) {
	path, err := transactionsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var transactions []Transaction
	if err := json.Unmarshal(data, &transactions); err != nil {
		return nil, err
	}
	// Backfill an id for any legacy rows that predate that field, same as
	// load_transactions's `if any("id" not in item for item in raw)`.
	needsSave := false
	for i := range transactions {
		if transactions[i].ID == "" {
			transactions[i].ID = newID()
			needsSave = true
		}
	}
	if needsSave {
		if err := SaveTransactions(transactions); err != nil {
			return nil, err
		}
	}
	return transactions, nil
}

func SaveTransactions(transactions []Transaction) error {
	path, err := transactionsPath()
	if err != nil {
		return err
	}
	if transactions == nil {
		transactions = []Transaction{}
	}
	data, err := json.MarshalIndent(transactions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// LoadSettings reads settings.json, falling back to DefaultSettings for any
// field missing from the file - mirroring `{**DEFAULT_SETTINGS, **raw}`.
func LoadSettings() (Settings, error) {
	path, err := settingsPath()
	if err != nil {
		return DefaultSettings(), err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return DefaultSettings(), nil
	}
	if err != nil {
		return DefaultSettings(), err
	}
	s := DefaultSettings()
	if err := json.Unmarshal(data, &s); err != nil {
		return DefaultSettings(), err
	}
	return s, nil
}

func SaveSettings(s Settings) error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// exportHeader/exportRecord/exportRows are the shared row model behind
// every export format, so CSV/XLSX/JSON can never drift apart on what a
// row of activity actually contains - mirrors splitsy's exportHeader/
// exportRecord/exportRows exactly, adapted to pear's Transaction shape.
var exportHeader = []string{"date", "category", "amount", "note", "balance"}

type exportRecord struct {
	Date     string
	Category string
	Amount   string
	Note     string
	Balance  string
}

func exportRows(transactions []Transaction) []exportRecord {
	balances := runningBalances(transactions)
	sorted := append([]Transaction{}, transactions...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Date < sorted[j].Date })

	rows := make([]exportRecord, len(sorted))
	for i, t := range sorted {
		rows[i] = exportRecord{
			Date: t.Date, Category: t.Category, Amount: formatFixed2(t.Amount),
			Note: t.Note, Balance: formatFixed2(balances[t.ID]),
		}
	}
	return rows
}

// exportPath builds a dated path in ~/Downloads for the given extension,
// used by every ExportXxx function below so they only ever disagree on
// format, never on where the file ends up.
func exportPath(ext string) (string, error) {
	dir, err := downloadsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "export-"+time.Now().Format("2006-01-02")+"."+ext), nil
}

// ExportCSV mirrors export_csv: a dated ledger export (date, category,
// amount, note, balance) for every transaction, sorted chronologically,
// written to ~/Downloads. Returns the path written.
func ExportCSV(transactions []Transaction) (string, error) {
	path, err := exportPath("csv")
	if err != nil {
		return "", err
	}

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write(exportHeader); err != nil {
		return "", err
	}
	for _, r := range exportRows(transactions) {
		row := []string{r.Date, r.Category, r.Amount, r.Note, r.Balance}
		if err := w.Write(row); err != nil {
			return "", err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}
	return path, nil
}

// ExportXLSX writes the same activity as ExportCSV to a dated .xlsx
// workbook in ~/Downloads - a single "Activity" sheet, header row bolded.
// Matches splitsy's ExportXLSX exactly.
func ExportXLSX(transactions []Transaction) (string, error) {
	path, err := exportPath("xlsx")
	if err != nil {
		return "", err
	}

	f := excelize.NewFile()
	defer f.Close()
	const sheet = "Activity"
	f.SetSheetName(f.GetSheetName(0), sheet)

	boldStyle, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	if err != nil {
		return "", err
	}
	for i, h := range exportHeader {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, boldStyle)
	}
	for i, r := range exportRows(transactions) {
		row := i + 2
		values := []string{r.Date, r.Category, r.Amount, r.Note, r.Balance}
		for col, v := range values {
			cell, _ := excelize.CoordinatesToCellName(col+1, row)
			f.SetCellValue(sheet, cell, v)
		}
	}
	for i := range exportHeader {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, 18)
	}

	if err := f.SaveAs(path); err != nil {
		return "", err
	}
	return path, nil
}

// ExportJSON writes the same activity as ExportCSV/ExportXLSX to a dated
// .json file in ~/Downloads - useful for piping into another tool or
// script rather than opening in a spreadsheet. Matches splitsy's
// ExportJSON exactly.
func ExportJSON(transactions []Transaction) (string, error) {
	path, err := exportPath("json")
	if err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(exportRows(transactions), "", "  ")
	if err != nil {
		return "", err
	}
	return path, os.WriteFile(path, data, 0o644)
}
