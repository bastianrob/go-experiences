package dto

// Payment dto
type Payment struct {
	ID        string
	InvoiceID string
	MethodID  string // card, cash, balance, whatever
	Amount    int
}
