package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// seedIfEmpty mirrors main.py's seed_if_empty: four sample transactions,
// only written when transactions.json didn't already have any rows.
func seedIfEmpty(transactions []Transaction) []Transaction {
	if len(transactions) > 0 {
		return transactions
	}
	sample := []Transaction{
		NewTransaction("2026-07-01", "Food", 240.0, "Groceries"),
		NewTransaction("2026-07-02", "Transport", 80.0, "Auto rides"),
		NewTransaction("2026-07-03", "Food", 150.0, "Dinner out"),
		NewTransaction("2026-07-05", "Entertainment", 500.0, "Concert ticket"),
	}
	_ = SaveTransactions(sample)
	return sample
}

func main() {
	transactions, err := LoadTransactions()
	if err != nil {
		fmt.Fprintln(os.Stderr, "finz: failed to load transactions:", err)
		os.Exit(1)
	}
	transactions = seedIfEmpty(transactions)

	settings, err := LoadSettings()
	if err != nil {
		fmt.Fprintln(os.Stderr, "finz: failed to load settings:", err)
		os.Exit(1)
	}

	p := tea.NewProgram(NewModel(transactions, settings), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "finz:", err)
		os.Exit(1)
	}
}
