package actor

import (
	"sync"
)

// Processor is the delegate which process a message
// @worker is its assigned worker number (starts from 1) in case we make more than 1 worker
// @actor is the reference to which actor that receives the message
// @message is the current individual message from actor's inbox
type Processor func(worker int, actor *Actor, message interface{}) (interface{}, error)

// Exception handler in case processor produce an error
// @worker is its assigned worker number (starts from 1) in case we make more than 1 worker
// @actor is the reference to which actor that receives the message
// @err is the error that happened after trying to process a message
type Exception func(worker int, actor *Actor, err error)

// Options when initializeing an Actor
type Options struct {
	Name        string       // actor's name
	Worker      int          // number of worker / processor go routine, defaults = 1
	Output      *Actor       // output actor, on which source actor will send a message after process is done
	FailChannel chan<- error // failure channel, on which Actor will send in case there is an error
}

func (opt *Options) configure() {
	if opt.Worker <= 0 {
		opt.Worker = 1
	}
}

// Actor ...
type Actor struct {
	// metadata
	name string

	// actor mechanism
	inbox  chan interface{}
	outbox *Actor

	failure   chan error
	process   Processor
	exception Exception

	// exit mechanism
	exit       chan struct{}
	workgroup  *sync.WaitGroup // worker wait group
	inboxgroup *sync.WaitGroup // inbox wait group
}

// New instance of an Actor with w as number of worker
func New(p Processor, e Exception, opt *Options) *Actor {
	opt.configure()

	actor := &Actor{
		name:      opt.Name,
		inbox:     make(chan interface{}, opt.Worker),
		outbox:    opt.Output,
		process:   p,
		exception: e,

		exit:       make(chan struct{}),
		workgroup:  &sync.WaitGroup{},
		inboxgroup: &sync.WaitGroup{},
	}

	actor.start(0, opt.Worker)
	return actor
}

// start the actor with n number of worker
func (actor *Actor) start(idx, n int) {
	if idx == n {
		return
	}

	// worker number starts from 1
	go actor.work(idx + 1)
	actor.start(idx+1, n)
}

func (actor *Actor) work(w int) {
	actor.workgroup.Add(1)       // worker group is added
	defer actor.workgroup.Done() // defer worker group done

	for {
		select {
		case message := <-actor.inbox: // waits for message to come from inbox
			result, err := actor.process(w, actor, message)

			if err != nil && actor.exception != nil {
				actor.exception(w, actor, err)
				actor.inboxgroup.Done() // flag 1 message as done
				continue
			}

			if actor.outbox != nil {
				actor.outbox.Queue(result)
				actor.inboxgroup.Done() // flag 1 message as done
				continue
			}

			// flag 1 message as done
			actor.inboxgroup.Done()
		case <-actor.exit: // listen on exit signal
			return
		}
	}
}

// Queue a message to inbox
func (actor *Actor) Queue(messages ...interface{}) {
	// add length of message to inbox wait group
	actor.inboxgroup.Add(len(messages))
	go func() {
		for _, message := range messages {
			actor.inbox <- message
		}
	}()
}

// Stop actor from processing any message
func (actor *Actor) Stop() (pendings []interface{}) {
	// stop all worker from processing any inbox
	close(actor.exit)
	actor.workgroup.Wait()

	// gather pending messages inside inbox and flag it as done
	go func() {
		for message := range actor.inbox {
			pendings = append(pendings, message)
			actor.inboxgroup.Done()
		}
	}()

	// wait for pending messages gathering to be completed and close the inbox channel
	actor.inboxgroup.Wait()
	close(actor.inbox)

	// return gathered pending messages
	return pendings
}
