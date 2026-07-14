package main

import "strconv"

// currencySymbols mirrors currency.py's CURRENCIES dict exactly - display
// only, switching currency relabels amounts, it does not convert values.
var currencySymbols = map[string]string{
	"USD": "$",
	"EUR": "€",
	"GBP": "£",
	"INR": "₹",
	"JPY": "¥",
	"CNY": "¥",
	"KRW": "₩",
	"AUD": "A$",
	"CAD": "C$",
	"CHF": "CHF",
	"BTC": "₿",
}

// currencyOrder is the fixed iteration order currency.py's dict literal
// establishes (Python dicts preserve insertion order) - used wherever the
// currency list needs to be walked in the same order the palette shows it.
var currencyOrder = []string{"USD", "EUR", "GBP", "INR", "JPY", "CNY", "KRW", "AUD", "CAD", "CHF", "BTC"}

// symbolFor mirrors currency.py's symbol_for: unknown codes fall back to "$".
func symbolFor(code string) string {
	if s, ok := currencySymbols[code]; ok {
		return s
	}
	return "$"
}

// formatFixed2 mirrors Python's f"{x:.2f}" - fixed two decimals, no
// thousands grouping. Used by CSV export, same as export_csv's row values.
func formatFixed2(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}

// formatComma2 mirrors Python's f"{x:,.2f}" - fixed two decimals with a
// thousands separator on the integer part, sign (if any) kept in front.
func formatComma2(v float64) string {
	neg := v < 0
	if neg {
		v = -v
	}
	s := strconv.FormatFloat(v, 'f', 2, 64)
	dot := len(s) - 3 // ".XX" is always the last 3 bytes
	out := groupThousands(s[:dot]) + s[dot:]
	if neg {
		out = "-" + out
	}
	return out
}

func groupThousands(digits string) string {
	n := len(digits)
	if n <= 3 {
		return digits
	}
	lead := n % 3
	if lead == 0 {
		lead = 3
	}
	out := digits[:lead]
	for i := lead; i < n; i += 3 {
		out += "," + digits[i:i+3]
	}
	return out
}

// fmtAmount mirrors FinanceApp.fmt exactly: symbol immediately followed by
// the comma-grouped signed number - e.g. fmt(-970) -> "₹-970.00" (symbol
// before the sign), which is deliberately different from the ledger
// table's amount column, which prepends its own +/- before an
// always-positive fmt() call (see refreshTable in app.go).
func fmtAmount(amount float64, currency string) string {
	return symbolFor(currency) + formatComma2(amount)
}
