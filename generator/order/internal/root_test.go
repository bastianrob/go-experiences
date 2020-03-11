package order

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/bastianrob/go-experiences/generator/actor"
	"github.com/bastianrob/go-experiences/generator/mock"
	"github.com/bastianrob/go-experiences/generator/order/pkg/command"
	"github.com/bastianrob/go-experiences/generator/order/pkg/dao"
	"github.com/bastianrob/go-experiences/generator/order/pkg/dto"
)

// We'll do test
func Test_OrderAsAggregateRoot(t *testing.T) {
	customerAPIMock := &mock.APIClient{
		GetFunc: func(id string) (interface{}, error) {
			time.Sleep(20 * time.Millisecond) // simulate 20ms latency
			return &dto.Customer{
				ID:   id,
				Name: "I am your customer",
			}, nil
		},
	}
	merchantAPIMock := &mock.APIClient{
		GetFunc: func(id string) (interface{}, error) {
			time.Sleep(20 * time.Millisecond) // simulate 20ms latency
			return &dto.Merchant{
				ID:   id,
				Name: "I am your merchant",
			}, nil
		},
	}
	promotionAPIMock := &mock.APIClient{
		GetFunc: func(id string) (interface{}, error) {
			time.Sleep(20 * time.Millisecond) // simulate 20ms latency
			return &dto.Promotion{
				ID:       id,
				Name:     "10% discount",
				Discount: 10,
			}, nil
		},
	}
	invoiceAPIMock := &mock.APIClient{
		CreateFunc: func(obj interface{}) error {
			time.Sleep(20 * time.Millisecond) // simulate 20ms latency
			inv := obj.(*dto.Invoice)
			inv.ID = "INV-001"
			return nil
		},
	}
	orderAPIMock := &mock.APIClient{
		CreateFunc: func(obj interface{}) error {
			time.Sleep(20 * time.Millisecond) // simulate 20ms latency
			inv := obj.(*dao.Order)
			inv.ID = "INV-001"
			return nil
		},
	}
	paymentAPIMock := &mock.APIClient{
		CreateFunc: func(obj interface{}) error {
			time.Sleep(20 * time.Millisecond) // simulate 20ms latency
			pay := obj.(*dto.Payment)
			pay.ID = "PMT-001"
			return nil
		},
	}
	productAPIMock := &mock.APIClient{
		GetFunc: func(id string) (interface{}, error) {
			time.Sleep(20 * time.Millisecond) // simulate 20ms latency
			switch id {
			case "ITEM-001":
				return &dto.Product{
					ID:   id,
					Name: "I am item 001",
				}, nil
			case "ITEM-002":
				return &dto.Product{
					ID:   id,
					Name: "I am item 002",
				}, nil
			}
			return nil, errors.New("404")
		},
	}

	// root is order actor which acts as an aggregate root
	// have by default 10 workers
	root := NewAggregateRoot(&Config{
		Worker: 20,
		Services: Services{
			Customer: customerAPIMock,
			Merchant: merchantAPIMock,
			Invoice:  invoiceAPIMock,
			Order:    orderAPIMock,
			Payment:  paymentAPIMock,
			Product:  productAPIMock,
			Promo:    promotionAPIMock,
		},
	})

	// waiter is an actor which hears from root and keep tracks of how many order have we done processing
	wg := &sync.WaitGroup{}
	wg.Add(100)
	waiter := actor.New(
		func(w int, a *actor.Actor, message interface{}) (interface{}, error) {
			wg.Done()
			return nil, nil
		},
		func(w int, a *actor.Actor, err error) {
			wg.Done()
		},
		&actor.Options{Worker: 10},
	)

	// clock in to check how long we're porcessing 100 command
	start := time.Now()

	// place order 100 times
	var orders []interface{}
	for i := 0; i < 100; i++ {
		orders = append(orders, &command.PlaceOrder{
			Customer: "CUST-001",
			Merchant: "MRCN-001",
			Payment:  "CARD-001",
			Promo:    "DISC-10",
			Items: []command.LineItem{{
				ID:  "ITEM-001",
				Qty: 1,
			}, {
				ID:  "ITEM-002",
				Qty: 1,
			}},
		})
	}
	root.Queue(orders...)

	actor.Direct(root.Actor, waiter)
	fmt.Println("We are waiting")
	wg.Wait()

	// we have at least 7 fake services and each takes simulated 20ms to complete
	// so total time it takes to complete 100 command * 140ms = 14sec
	// 14 sec if we only have 1 worker.
	// Ideally we can cut it down to minimal of 0.7sec because we have 20 workers
	dur := time.Since(start)
	if dur.Seconds() >= 1. {
		t.Error("Total processing time should not exceed 1sec")
	}

	fmt.Println("Duration:", dur)
}
