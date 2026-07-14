package main

import "testing"

func sampleTransactions() []Transaction {
	return []Transaction{
		NewTransaction("2026-07-01", "Food", 240.0, "Groceries"),
		NewTransaction("2026-07-02", "Transport", 80.0, "Auto rides"),
		NewTransaction("2026-07-03", "Food", 150.0, "Dinner out"),
		NewTransaction("2026-08-01", "Salary", -1000.0, "Paycheck"),
	}
}

func TestTotalSpentIgnoresIncome(t *testing.T) {
	got := totalSpent(sampleTransactions())
	if got != 470.0 {
		t.Fatalf("totalSpent = %v, want 470", got)
	}
}

func TestTotalIncomeIgnoresExpenses(t *testing.T) {
	got := totalIncome(sampleTransactions())
	if got != 1000.0 {
		t.Fatalf("totalIncome = %v, want 1000", got)
	}
}

func TestNetBalance(t *testing.T) {
	got := netBalance(sampleTransactions())
	if got != 1000.0-470.0 {
		t.Fatalf("netBalance = %v, want %v", got, 1000.0-470.0)
	}
}

func TestExpenseTotalsByCategoryForMonthExcludesOtherMonths(t *testing.T) {
	totals := expenseTotalsByCategoryForMonth(sampleTransactions(), "2026-07")
	if len(totals) != 2 {
		t.Fatalf("expected 2 categories in July, got %d: %v", len(totals), totals)
	}
	if totals["Food"] != 390.0 {
		t.Fatalf("Food total = %v, want 390", totals["Food"])
	}
	if _, ok := totals["Salary"]; ok {
		t.Fatal("expected income category to be excluded (amount < 0 never counted)")
	}
}

func TestMonthsPresent(t *testing.T) {
	got := monthsPresent(sampleTransactions())
	want := []string{"2026-07", "2026-08"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("monthsPresent = %v, want %v", got, want)
	}
}

func TestRunningBalancesChronological(t *testing.T) {
	txs := sampleTransactions()
	balances := runningBalances(txs)
	// 240 spent -> -240, +80 -> -320, +150 -> -470, income 1000 -> +530
	want := map[int]float64{0: -240, 1: -320, 2: -470, 3: 530}
	for i, w := range want {
		if got := balances[txs[i].ID]; got != w {
			t.Fatalf("balance[%d] = %v, want %v", i, got, w)
		}
	}
}

func TestRunningBalancesTiesKeepOriginalOrder(t *testing.T) {
	// Same date, different insertion order - Python's sorted(..., reverse
	// won't apply here since running_balances sorts ascending) preserves
	// original relative order for ties; sort.SliceStable must too.
	a := NewTransaction("2026-07-01", "A", 10, "")
	b := NewTransaction("2026-07-01", "B", 20, "")
	balances := runningBalances([]Transaction{a, b})
	if balances[a.ID] != -10 {
		t.Fatalf("balance[a] = %v, want -10 (processed first)", balances[a.ID])
	}
	if balances[b.ID] != -30 {
		t.Fatalf("balance[b] = %v, want -30 (processed second)", balances[b.ID])
	}
}
