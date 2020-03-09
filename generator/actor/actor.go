package actor

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
	Worker      int              // number of worker / processor go routine, defaults = 1
	Output      chan interface{} // output channel, on which Actor will sned in after process is done
	FailChannel chan<- error     // failure channel, on which Actor will send in case there is an error
}

func (opt *Options) configure() {
	if opt.Worker <= 0 {
		opt.Worker = 1
	}
}

// Actor ...
type Actor struct {
	name      string
	inbox     chan interface{}
	outbox    chan interface{}
	failure   chan error
	process   Processor
	exception Exception
}

// New instance of an Actor with w as number of worker
func New(p Processor, e Exception, opt *Options) *Actor {
	opt.configure()

	actor := &Actor{
		inbox:     make(chan interface{}, opt.Worker),
		outbox:    opt.Output,
		process:   p,
		exception: e,
	}

	actor.start(0, opt.Worker)
	return actor
}

// start the actor with n number of worker
func (actor *Actor) start(idx, n int) {
	if idx == n {
		return
	}

	go func(w int) {
		for message := range actor.inbox {
			result, err := actor.process(w, actor, message)

			if err != nil && actor.exception != nil {
				actor.exception(w, actor, err)
				continue
			}

			if actor.outbox != nil {
				actor.outbox <- result
				continue
			}
		}
	}(idx + 1) // worker number starts from 1

	actor.start(idx+1, n)
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

// Outbox exposes the actor's outbox
func (actor *Actor) Outbox() <-chan interface{} {
	return actor.outbox
}

// Stop all inbox
func (actor *Actor) Stop() {
	close(actor.inbox)
}
