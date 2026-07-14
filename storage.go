package main

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
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
}

// DefaultSettings mirrors storage.py's DEFAULT_SETTINGS.
func DefaultSettings() Settings {
	return Settings{Currency: "USD", MonthlyBudget: 0.0}
}

// dataDir returns ~/.config/tally (or the platform equivalent), creating it
// if necessary. The Python original keeps transactions.json/settings.json
// next to the source files; a Go binary has no such fixed "next to the
// script" location once installed, so this uses the same XDG convention as
// splitsy (tally's Go sibling) instead - an internal storage-location
// choice, not a user-visible behavior difference.
func dataDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "tally")
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

// ExportCSV mirrors export_csv: a dated ledger export (date, category,
// amount, note, balance) for every transaction, sorted chronologically,
// written next to the data files. Returns the path written.
func ExportCSV(transactions []Transaction) (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "export-"+time.Now().Format("2006-01-02")+".csv")

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	balances := runningBalances(transactions)
	sorted := append([]Transaction{}, transactions...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Date < sorted[j].Date })

	w := csv.NewWriter(f)
	if err := w.Write([]string{"date", "category", "amount", "note", "balance"}); err != nil {
		return "", err
	}
	for _, t := range sorted {
		row := []string{t.Date, t.Category, formatFixed2(t.Amount), t.Note, formatFixed2(balances[t.ID])}
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
