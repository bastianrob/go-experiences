package dto

// Invoice dto
type Invoice struct {
	ID       string
	Order    string
	Customer string
	Promo    string
	Subtotal int
	Discount int
	Total    int
}
