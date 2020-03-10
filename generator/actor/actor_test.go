package actor

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func Test_Actor(t *testing.T) {
	words := [...]string{"One", "Two", "Three"}
	actor := New(func(w int, actor *Actor, message interface{}) (interface{}, error) {
		result := words[w-1]

		fmt.Println("worker", w,
			"receive", message,
			"processed as", result,
			"send to?", actor.outbox)

		return result, nil
	}, func(w int, actor *Actor, err error) {
		fmt.Println(err)
	}, &Options{Worker: 3})

	wg := sync.WaitGroup{}
	for i := 0; i <= 10; i++ {
		wg.Add(1)
		go func(i int) {
			idx := i % 3
			word := words[idx]
			actor.Queue(word)
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func Test_ActorStop(t *testing.T) {
	mux := sync.Mutex{}
	var processed []interface{}
	actor := New(func(w int, actor *Actor, message interface{}) (interface{}, error) {
		mux.Lock()
		processed = append(processed, message)
		mux.Unlock()

		return nil, nil
	}, func(w int, actor *Actor, err error) {
		fmt.Println(err)
	}, &Options{Worker: 5})

	expected := 0
	for i := 1; i <= 100; i++ {
		go actor.Queue(i)
		expected = expected + i
	}

	pendings := actor.Stop()
	combined := append(processed, pendings...)

	sum := 0
	for _, e := range combined {
		sum = sum + e.(int)
	}
	if sum != expected {
		t.Error("Sum of 1-100 must be", expected, "but received", sum)
	}

	fmt.Println("PEND", pendings)
	fmt.Println("PROC", processed)
}

func Test_ActorDirected(t *testing.T) {
	errPrinter := func(w int, actor *Actor, err error) {
		fmt.Println("worker", w, "-", "err:", err)
	}

	opt := &Options{
		Worker: 3,
	}

	bale := New(func(w int, actor *Actor, in interface{}) (interface{}, error) {
		return in, nil
	}, errPrinter, opt)
	bane := New(func(w int, actor *Actor, in interface{}) (interface{}, error) {
		switch {
		case in == "I AM VENGEANCE":
			return "I AM INEVITABLE", nil
		case in == "I AM THE NIGHT":
			return "I AM BANE", nil
		case in == "I'M BATMAN":
			return "I WILL BREAK YOU", nil
		default:
			return nil, errors.New("WHATEVER YOU SAY")
		}
	}, errPrinter, opt)
	printer := New(func(w int, actor *Actor, in interface{}) (interface{}, error) {
		return nil, nil
	}, errPrinter, opt)

	fmt.Println(bale.Inbox(), bane.Inbox(), printer.Inbox())
	fmt.Println(bale.Outbox(), bane.Outbox(), printer.Outbox())

	bale.name = "Bale"
	bane.name = "Bane"
	printer.name = "Printer"

	Direct(bale, bane, printer)

	fmt.Println(bale.Inbox(), bane.Inbox(), printer.Inbox())
	fmt.Println(bale.Outbox(), bane.Outbox(), printer.Outbox())

	bale.Queue("I AM VENGEANCE", "I AM THE NIGHT", "I'M BATMAN", "HEY HO!")
	// time.Sleep(1 * time.Second)

	// bale.Stop()
	// bane.Stop()
	// printer.Stop()
}