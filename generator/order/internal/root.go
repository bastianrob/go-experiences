package order

import (
	"errors"
	"fmt"
	"time"

	"github.com/bastianrob/go-experiences/generator/order/pkg/dao"
	"github.com/bastianrob/go-experiences/generator/order/pkg/dto"

	"github.com/bastianrob/go-experiences/generator/actor"
	"github.com/bastianrob/go-experiences/generator/mock"
	"github.com/bastianrob/go-experiences/generator/order/pkg/command"
)

// Services collection
type Services struct {
	Customer mock.CRUD
	Invoice  mock.CRUD
	Merchant mock.CRUD
	Order    mock.CRUD
	Payment  mock.CRUD
	Product  mock.CRUD
	Promo    mock.CRUD
}

// Config for order service
type Config struct {
	Worker   int
	Services Services
}

// Root aggregate root of order
type Root struct {
	*actor.Actor
	services Services
}

// NewAggregateRoot for order
func NewAggregateRoot(cfg *Config) *Root {
	root := &Root{
		services: cfg.Services,
	}

	n := cfg.Worker
	if n <= 0 {
		n = 10
	}

	worker := &actor.Options{Worker: n}
	root.Actor = actor.New(root.processor, root.exception, worker)

	return root
}

func (root *Root) processor(w int, a *actor.Actor, msg interface{}) (interface{}, error) {
	if msg == nil {
		return nil, errors.New("Order message is empty")
	}

	var customer *dto.Customer
	var merchant *dto.Merchant
	var promo *dto.Promotion

	// 1. Converts message to command
	cmd := msg.(*command.PlaceOrder)

	// 2. Fetch required information
	// Uses goroutine because we all have verbose if err
	errc := make(chan error)
	go func(errc chan<- error) {
		cust, err := root.services.Customer.Get(cmd.Customer)
		if err != nil {
			errc <- err
			return
		}
		customer = cust.(*dto.Customer)

		mcr, err := root.services.Merchant.Get(cmd.Merchant)
		if err != nil {
			errc <- err
			return
		}
		merchant = mcr.(*dto.Merchant)

		prm, err := root.services.Promo.Get(cmd.Promo)
		if err != nil {
			errc <- err
			return
		}
		promo = prm.(*dto.Promotion)

		errc <- nil
	}(errc)

	// 3. Wait for fetch to complete and listen to any error occurred
	if err := <-errc; err != nil {
		return nil, err
	}

	// 4. Get product details and calculate the total
	order := &dao.Order{
		Date:         time.Now(),
		State:        dao.New,
		CustomerID:   customer.ID,
		CustomerName: customer.Name,
		MerchantID:   merchant.ID,
		MerchantName: merchant.Name,
		Items:        make([]*dao.OrderItem, len(cmd.Items)),
	}
	for i, entry := range cmd.Items {
		it, err := root.services.Product.Get(entry.ID)
		if err != nil {
			return nil, errors.New("Failed to get item with ID: " + entry.ID)
		}

		item := it.(*dto.Product)
		order.Items[i] = &dao.OrderItem{
			ID:    item.ID,
			Name:  item.Name,
			Qty:   entry.Qty,
			Price: item.Price,
		}
		order.Total += (entry.Qty * item.Price)
	}

	// 5. Persist the order data to database
	err := root.services.Order.Create(order)
	if err != nil {
		return nil, errors.New("Failed to create a new order: " + err.Error())
	}

	// 6. Create the invoice through API
	discount := order.Total * promo.Discount / 100
	invoice := &dto.Invoice{
		Order:    order.ID,
		Customer: order.CustomerID,
		Promo:    promo.ID,
		Subtotal: order.Total,
		Discount: discount,
		Total:    (order.Total - discount),
	}
	err = root.services.Invoice.Create(invoice)
	if err != nil {
		// TODO: Recovery strategy, delete the order? or flag it if you wish
		return nil, errors.New("Failed to create a payment: " + err.Error())
	}

	// 7. Make a payment through API call
	payment := &dto.Payment{
		InvoiceID: invoice.ID,
		MethodID:  cmd.Payment,
		Amount:    invoice.Total,
	}
	err = root.services.Payment.Create(payment)
	if err != nil {
		// TODO: Recovery strategy to both order and invoice
		return nil, errors.New("Failed to create a payment: " + err.Error())
	}

	return order, nil
}

func (root *Root) exception(w int, a *actor.Actor, err error) {
	fmt.Println("Exception occurred at worker:", w, "with err:", err)
}
