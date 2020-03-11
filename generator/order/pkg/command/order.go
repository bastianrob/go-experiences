package command

// LineItem individual ordered item & qty
type LineItem struct {
	ID  string
	Qty int
}

// PlaceOrder command
type PlaceOrder struct {
	Customer string
	Merchant string
	Payment  string
	Promo    string
	Items    []LineItem
}
