package main

import "sort"

// Convention (mirrors analytics.py exactly): amount > 0 is money spent (an
// expense); amount < 0 is money received (income or a refund).

func totalSpent(transactions []Transaction) float64 {
	total := 0.0
	for _, t := range transactions {
		if t.Amount > 0 {
			total += t.Amount
		}
	}
	return total
}

func totalIncome(transactions []Transaction) float64 {
	total := 0.0
	for _, t := range transactions {
		if t.Amount < 0 {
			total += -t.Amount
		}
	}
	return total
}

func netBalance(transactions []Transaction) float64 {
	return totalIncome(transactions) - totalSpent(transactions)
}

func totalsByCategory(transactions []Transaction) map[string]float64 {
	totals := make(map[string]float64)
	for _, t := range transactions {
		totals[t.Category] += t.Amount
	}
	return totals
}

// monthTotal pairs a "YYYY-MM" period with a signed amount total - the
// return shape totalsByMonth needs since Go maps don't preserve order the
// way Python's dict(sorted(...)) does.
type monthTotal struct {
	Month string
	Total float64
}

func totalsByMonth(transactions []Transaction) []monthTotal {
	totals := make(map[string]float64)
	for _, t := range transactions {
		totals[dateMonth(t.Date)] += t.Amount
	}
	months := make([]string, 0, len(totals))
	for m := range totals {
		months = append(months, m)
	}
	sort.Strings(months)
	out := make([]monthTotal, len(months))
	for i, m := range months {
		out[i] = monthTotal{Month: m, Total: totals[m]}
	}
	return out
}

// expenseTotalsByCategoryForMonth mirrors expense_totals_by_category_for_month:
// expense-only (amount > 0), scoped to one "YYYY-MM" period. tui.py
// deliberately reads this (never the raw signed totalsByCategory) anywhere
// a user sees a number, since a net figure goes negative in an
// income-heavy month and reads as a confusing "spent -₹7,470".
func expenseTotalsByCategoryForMonth(transactions []Transaction, month string) map[string]float64 {
	totals := make(map[string]float64)
	for _, t := range transactions {
		if t.Amount > 0 && dateMonth(t.Date) == month {
			totals[t.Category] += t.Amount
		}
	}
	return totals
}

// monthsPresent mirrors months_present: every distinct "YYYY-MM" period any
// transaction falls in, sorted ascending.
func monthsPresent(transactions []Transaction) []string {
	set := make(map[string]bool)
	for _, t := range transactions {
		set[dateMonth(t.Date)] = true
	}
	months := make([]string, 0, len(set))
	for m := range set {
		months = append(months, m)
	}
	sort.Strings(months)
	return months
}

// dateYear/dateMonth slice a "YYYY-MM-DD" date string the same way Python's
// t.date[:4] / t.date[:7] do.
func dateYear(d string) string {
	if len(d) < 4 {
		return d
	}
	return d[:4]
}

func dateMonth(d string) string {
	if len(d) < 7 {
		return d
	}
	return d[:7]
}

// runningBalances mirrors running_balances: chronological cumulative
// balance keyed by transaction id. Expenses (amount > 0) decrease the
// balance; income (amount < 0) increases it. Ties on date break in
// original-list order, matching Python's stable sort on (date, index).
func runningBalances(transactions []Transaction) map[string]float64 {
	type indexed struct {
		index int
		t     Transaction
	}
	ordered := make([]indexed, len(transactions))
	for i, t := range transactions {
		ordered[i] = indexed{index: i, t: t}
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].t.Date < ordered[j].t.Date
	})

	balances := make(map[string]float64, len(transactions))
	balance := 0.0
	for _, it := range ordered {
		balance -= it.t.Amount
		balances[it.t.ID] = balance
	}
	return balances
}
