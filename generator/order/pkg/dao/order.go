package dao

import "time"

// OrderState ...
type OrderState string

// Order states
const (
	New      = OrderState("open")
	Invoiced = OrderState("invoiced")
	Paid     = OrderState("paid")
	Expired  = OrderState("expired")
)

// OrderItem DAO
type OrderItem struct {
	ID    string
	Name  string
	Qty   int
	Price int
}

// Order DAO
type Order struct {
	ID           string
	Date         time.Time
	State        OrderState
	CustomerID   string
	CustomerName string
	MerchantID   string
	MerchantName string
	Items        []*OrderItem
	Total        int
}
