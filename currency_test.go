package main

import "testing"

func TestSymbolForKnownAndUnknown(t *testing.T) {
	if s := symbolFor("INR"); s != "₹" {
		t.Fatalf("symbolFor(INR) = %q, want ₹", s)
	}
	if s := symbolFor("XYZ"); s != "$" {
		t.Fatalf("symbolFor(unknown) = %q, want $ fallback", s)
	}
}

func TestFormatComma2MatchesPythonStyle(t *testing.T) {
	cases := map[float64]string{
		970:      "970.00",
		-970:     "-970.00",
		14000:    "14,000.00",
		1234567:  "1,234,567.00",
		0:        "0.00",
		-1000000: "-1,000,000.00",
	}
	for in, want := range cases {
		if got := formatComma2(in); got != want {
			t.Fatalf("formatComma2(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestFmtAmountSymbolBeforeSign(t *testing.T) {
	// Mirrors FinanceApp.fmt: f"{symbol}{amount:,.2f}" - the symbol comes
	// before the sign, unlike the ledger table's amount column, which
	// prepends its own +/- before an always-positive fmt() call.
	if got := fmtAmount(-970, "INR"); got != "₹-970.00" {
		t.Fatalf("fmtAmount(-970, INR) = %q, want ₹-970.00", got)
	}
	if got := fmtAmount(14000, "INR"); got != "₹14,000.00" {
		t.Fatalf("fmtAmount(14000, INR) = %q, want ₹14,000.00", got)
	}
}

func TestFormatFixed2NoGrouping(t *testing.T) {
	if got := formatFixed2(14000); got != "14000.00" {
		t.Fatalf("formatFixed2(14000) = %q, want 14000.00 (no thousands separator)", got)
	}
}
