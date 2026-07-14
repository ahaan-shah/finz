package main

// Transaction mirrors the Python Transaction dataclass exactly: amount > 0
// is money spent (an expense), amount < 0 is money received (income or a
// refund) - there's no separate type field, the sign carries the meaning
// throughout analytics.go and app.go.
type Transaction struct {
	Date     string  `json:"date"`
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Note     string  `json:"note"`
	ID       string  `json:"id"`
}

// NewTransaction builds a Transaction with a freshly generated id, mirroring
// the Python dataclass's `field(default_factory=lambda: uuid4().hex)`.
func NewTransaction(date, category string, amount float64, note string) Transaction {
	return Transaction{Date: date, Category: category, Amount: amount, Note: note, ID: newID()}
}
