package actor

import (
	"fmt"
)

// Processor of a message
// type Processor func(actor *Actor, in <-chan interface{}, out chan<- interface{})
type Processor func(worker int, actor *Actor, in interface{}) interface{}

// Actor ...
type Actor struct {
	name    string
	inbox   chan interface{}
	outbox  chan interface{}
	process Processor
}

// New instance of an Actor with cap as inbox maximum capacity
func New(process Processor, cap int) *Actor {
	actor := &Actor{
		inbox:   make(chan interface{}, cap),
		outbox:  make(chan interface{}, cap),
		process: process,
	}

	return actor
}

// AddProcessor which must wait for a message from inbox
// Each process works asynchronously and will report the result into `result` channel
// func (actor *Actor) AddProcessor(process Processor, result chan<- interface{}) {
// 	go process(actor, actor.inbox, result)
// }

// Start processing with n number of worker
func (actor *Actor) Start(n int, out chan<- interface{}) {
	for i := 1; i <= n; i++ {
		go func(n int) {
			for message := range actor.inbox {
				result := actor.process(n, actor, message)

				fmt.Println("worker", n, "-", actor.name, "receive", message, "processed as", result, "send to?", out)
				if out != nil {
					out <- result
				}
			}
		}(i)
	}
}

// Queue a message to inbox
func (actor *Actor) Queue(messages ...interface{}) {
	for _, message := range messages {
		actor.inbox <- message
	}
}

// Inbox exposes the actor's inbox
func (actor *Actor) Inbox() chan<- interface{} {
	return actor.inbox
}

// Stop all inbox
func (actor *Actor) Stop() {
	close(actor.inbox)
}
