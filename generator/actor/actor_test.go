package actor

import (
	"fmt"
	"runtime"
	"testing"
)

func baleSays(actor *Actor, in <-chan interface{}, out chan<- interface{}) {
	for msg := range in {
		out <- msg
	}
}

func baneSays(actor *Actor, in <-chan interface{}, out chan<- interface{}) {
	for dialog := range in {
		switch {
		case dialog == "I AM VENGEANCE":
			out <- "I AM INEVITABLE"
		case dialog == "I AM THE NIGHT":
			out <- "I AM BANE"
		case dialog == "I'M BATMAN":
			out <- "I WILL BREAK YOU"
		default:
			out <- "WHATEVER YOU SAY"
		}
	}
}

func print(actor *Actor, in <-chan interface{}, out chan<- interface{}) {
	for dialog := range in {
		fmt.Println("Actor:", actor, "Says", dialog)
	}
}

func Test_ActionSolo(t *testing.T) {
	// bale := New(1)
	// bane := New(1)
	// printer := New(1)

	// // what baleSays will be put to bane's inbox
	// bale.AddProcessor(baleSays, bane.Inbox())
	// // what baneSays will be put to printer's inbox
	// bane.AddProcessor(baneSays, printer.Inbox())
	// printer.AddProcessor(print, nil)

	// bale.Queue("I AM VENGEANCE", "I AM THE NIGHT", "I'M BATMAN", "HEY HO!")

	// time.Sleep(1 * time.Second)
	// bale.Stop()
	// bane.Stop()
	// printer.Stop()

	runtime.GOMAXPROCS(100)
	bale := New(func(w int, actor *Actor, in interface{}) interface{} {
		return in
	}, 1)
	bane := New(func(w int, actor *Actor, in interface{}) interface{} {
		switch {
		case in == "I AM VENGEANCE":
			return "I AM INEVITABLE"
		case in == "I AM THE NIGHT":
			return "I AM BANE"
		case in == "I'M BATMAN":
			return "I WILL BREAK YOU"
		default:
			return "WHATEVER YOU SAY"
		}
	}, 1)
	printer := New(func(w int, actor *Actor, in interface{}) interface{} {
		return nil
	}, 1)

	fmt.Println(bale.Inbox(), bane.Inbox(), printer.Inbox())

	bale.name = "Bale"
	bane.name = "Bane"
	printer.name = "Printer"
	// what baleSays will be put to bane's inbox
	bale.Start(1, bane.Inbox())
	bale.Start(1, bane.Inbox())
	bale.Start(1, bane.Inbox())
	// what baneSays will be put to printer's inbox
	bane.Start(3, printer.Inbox())
	printer.Start(3, nil)

	bale.Queue("I AM VENGEANCE", "I AM THE NIGHT", "I'M BATMAN", "HEY HO!")

	bale.Stop()
	bane.Stop()
	printer.Stop()
}
