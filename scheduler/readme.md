# Making Simple Scheduler using Go Routine & Channel

> Learn how to schedule an event to be running at specific time

Recently `Product` team in my job requires me to have a scheduled event that might or might not run at a specific set of interval.

And the story goes like this:

> As a [role] I want [resource] document to be tracked after it was `Assigned` to [role], for every [N] minutes until it was `Handled`

So let's do a case study on Go Routine, and Go Channel!

## Designing the Scheduler

* It have to be re-usable and shared across all projects
* It have to survive server restart
* It have to be able to re-schedule event

So now, we have these barebone requirements. We can start to write some code.

## Event Object

First we have to write `Event` object. It is the data that a `Scheduler` receive

```go
package scheduler

import (
	"time"
)

// Attachment data associated with an event
// Can be anything stored in bytes
type Attachment struct {
	Name        string
	ContentType string
	Body        []byte
}

// Event which will run on scheduler
type Event struct {
	datetime    string // RFC3339 please
	attachments []Attachment
}

// NewEvent create a new instance of immutable Event
func NewEvent(d string, att []Attachment) *Event {
	// we copy the attachment slice to another memory to avoid mutability
	cpy := make([]Attachment, len(att))
	copy(cpy, att)

	return &Event{
		datetime:    d,
		attachments: cpy,
	}
}

// Date get event datetime, parsed into RFC3339 format
func (e *Event) Date() (time.Time, error) {
	return time.Parse(time.RFC3339, e.datetime)
}

// Attachments returns a copy of attachments slice
// This is done to ensure immutability of event
func (e *Event) Attachments() []Attachment {
	cpy := make([]Attachment, len(e.attachments))
    copy(cpy, e.attachments)

	return cpy
}
```

Event is designed to be immutable so **none of these event can be changed once scheduled**.
Each event contains 2 fields
* `datetime` = RFC3339 string representation of a date & time
* `attachments` = List of any object associated to an Event

Both of those fields are private and can only be set from the `NewEvent` factory.
But both can get fetched from 2 getter methods called `Date` dan `Attachments`.

## The Scheduler

Next in the scheduler, we know that we have to create a method called `Schedule` which takes an `Event` as its parameter:

```go
package scheduler

// Scheduler ...
type Scheduler struct {}

// Schedule an event
func (s *Scheduler) Schedule(e *Event) error {
    // TODO:
}
```

### TODO-1 Validate Event

```go
// Scheduler error collection
var (
	ErrEventInPast = errors.New("Event datetime is in the past")
	ErrTimeInvalid = errors.New("Datetime format is not in RFC3339")
)

// Schedule an event
func (s *Scheduler) Schedule(e *Event) error {
    date, err := e.Date()
    if err != nil {
        return ErrTimeInvalid
    }

    now := time.Now()
    if date.Unix() <= now.Unix() {
        return ErrEventInPast
    }

    ...
}
```

### TODO-2 Use Go routine to wait for event

```go
// Schedule an event
func (s *Scheduler) Schedule(e *Event) error {
    ...
    // fire a go routine
    go func(e *Event) {
        now := time.Now()
        target, _ := e.Date()
        waitDuration := target.Sub(now) // compare

        select {
            case <-time.After(waitDuration):
        }
    }(e)
}
```

### TODO-3 Add Event Handler

This line below:

```go
select {
    case <-time.After(waitDuration):
}
```

Will block the goroutine for `waitDuration` long, and then executes whatever code under the `case`

But we don't have the event handler handler for now, so it's time to design it.

* The event handler must be supplied by the caller
* Each `Scheduler` can only have 1 handler
* Everytime an `Event` is triggered, we'll call the registered handler

The concept really is close with `Java` delegates so that's what we'll call it.

```go
// EventHandler delegates
type EventHandler func(*Scheduler, *Event)

// Scheduler ...
type Scheduler struct {
	delegate EventHandler
}

// New instance of scheduler
func New(d EventHandler) *Scheduler {
	return &Scheduler{
		delegate: d,
	}
}
```

Again, we make the `Scheduler` immutable so the `delegate` can only be instantiated via `New` factory function, and can't be changed while it's running.

The `EventHandler` delegate accepts 2 parameter, `Scheduler` and `Event`. This means any `delegate` will always receive the correct reference to its `Scheduler` and `Event`.

And then, we *complete* our `Schedule` method:

```go
// Schedule an event
func (s *Scheduler) Schedule(e *Event) error {
    date, err := e.Date()
    if err != nil {
        return ErrTimeInvalid
    }

    now := time.Now()
    if date.Unix() <= now.Unix() {
        return ErrEventInPast
    }

    // fire a go routine
    go func(e *Event) {
        now := time.Now()
        target, _ := e.Date()
        waitDuration := target.Sub(now) // compare

        select {
            case <-time.After(waitDuration):
                s.delegate(s, e)
        }
    }(e)
    return  nil
}
```

### Reviewing the Requirements

It looks like we're done with the `Scheduler` but let's go back and look at the requirements:

* It have to survive server restart

So, we'll at least have to consider `Stop` method which stops all scheduled events, collects it, and report it back to caller:

```go
// Stop all running scheduler and report all pending events
func (s *Scheduler) Stop() (events []*Event) {
    // TODO:
}
```

### TODO-1 Stop all scheduled events!

So in the `select case` statement in `Schedule` method. Not only we have to wait for `Event` datetime, we also have to listen to another channel which signals as out `Stop` event. So lets call it just that:

```go
// Scheduler ...
type Scheduler struct {
	delegate EventHandler
	stop     chan struct{}
}

// New instance of scheduler
func New(d EventHandler) *Scheduler {
	return &Scheduler{
		delegate: d,
        // initialize stop channel
		stop:     make(chan struct{}),
	}
}
```

In the `Schedule` method we'll listen to the `stop` channel:

```go
// Schedule an event
func (s *Scheduler) Schedule(e *Event) error {
    ...

    // fire a go routine
    go func(e *Event) {
        now := time.Now()
        target, _ := e.Date()
        waitDuration := target.Sub(now) // compare

        select {
            case <-time.After(waitDuration):
                s.delegate(s, e)
            case <-s.stop:
                // TODO:
        }
    }(e)
    return  nil
}
```

And then in the `Stop` method:

```go
// Stop all running scheduler and report all pending events
func (s *Scheduler) Stop() (events []*Event) {
    // close stop channel & it will be broadcasted to all consumer
    close(s.stop)

    // TODO:
}
```

### TODO-2 Collecting pending events

So far we have orchestrated to stop all running go routine but we haven't collected all the pending events yet.

To do this, we'll once again utilize channel. Let's call it `pendings`

```go
// Scheduler ...
type Scheduler struct {
	delegate EventHandler
	stop     chan struct{}
	pendings chan *Event
}

// New instance of scheduler
func New(d EventHandler) *Scheduler {
	return &Scheduler{
        delegate: d,
        // initialize stop channel
        stop:     make(chan struct{}),
        // initialize buffered event channel
		pendings: make(chan *Event, 3),
	}
}
```

`pendings` is a buffered channel with length = 3 which means it can hold up to 3 value until producer have to wait for any consumer to fetch a value, freeing up a buffer.

```go
// Schedule an event
func (s *Scheduler) Schedule(e *Event) error {
    ...
    // fire a go routine
    go func(e *Event) {
        ...
        select {
            ...
            case <-s.stop:
                s.pendings <- e
        }
    }(e)
    return  nil
}
```

And then in the `Stop` function we'll have to collect all pending `Events` being put in `pendings` channel


```go
// Stop all running scheduler and report all pending events
func (s *Scheduler) Stop() (events []*Event) {
	close(s.stop)

	for e := range s.pendings {
		events = append(events, e)
	}

	return events
}
```

Looks good? but wait!

```go
for e := range s.pendings
```

Is iterating over a channel. 
BUT NOBODY is CLOSING THE `pendings` CHANNEL SO IT WON'T EVER QUIT!

### TODO-3 Closing the pendings channel

So how do we know when to close `pendings` channel? We'll need to:
* Manually count how many go routine was fired.
* Decreement it everytime the go routine exit. Either it is done handling event, or forcefully stopped.
* Watch the counter to go down to zero, to then close the `pendings` channel

> !@#E!T#T!V@$Y

Luckily, we have `sync.WaitGroup` to our rescue! It does everything we listed above so let's code right away:

```go
// Scheduler ...
type Scheduler struct {
	delegate EventHandler
	stop     chan struct{}
	pendings chan *Event
	wg       *sync.WaitGroup
}

// New instance of scheduler
func New(d EventHandler) *Scheduler {
	return &Scheduler{
		delegate: d,
        // initialize stop channel
        stop:     make(chan struct{}),
        // initialize buffered event channel
		pendings: make(chan *Event, 3),
		wg:       &sync.WaitGroup{},
	}
}
```

Next, we want to call `wg.Add(1)` every time we fire a go routine.
And We want to call `wg.Done()` every time a go routine exits.

```go
// Schedule an event
func (s *Scheduler) Schedule(e *Event) error {
    ...
    s.wg.Add(1)
    // fire a go routine
    go func(e *Event) {
        ...
        defer s.wg.Done()
        select {
            ...
            case <-s.stop:
                s.pendings <- e
        }
    }(e)
    return  nil
}
```

Lastly! we want to watch and wait for the counter to go down to zero.

```go
// Stop all running scheduler and report all pending events
func (s *Scheduler) Stop() (events []*Event) {
    ...
	go func() {
		s.wg.Wait()
		close(s.pendings)
	}()
    ...
}
```

In the `Stop` method, we are waiting for `wg` counter to drop to zero by calling `wg.Wait()`.
And we do it in another go routine so it doesn't block the Stop execution which collects pending events.

### Reviewing the Requirements again

* It have to be re-usable and shared across all projects
  
  This is a standalone `go package` called `scheduler` and can be shared to whoever needed a scheduler

* It have to survive server restart
  
  Actually because `scheduler` is a standalone `go package` we'll **invert the responsibilities** of persisting pending events to the user / caller. 
  
  All `scheduler` can provide is just a `Stop` method which reports any pending events to the caller.

* It have to be able to re-schedule event
  
  By designing the `EventHandler` delegate to accept `Scheduler` and `Event` as its parameter. We leave the re-schedule implementation to the delegate.

  The delegate can just call `Scheduler.Schedule` method to re-schedule any event.

## Full Code of the Scheduler

```go
package scheduler

import (
	"errors"
	"sync"
	"time"
)

// Scheduler error collection
var (
	ErrEventInPast = errors.New("Event datetime is in the past")
	ErrTimeInvalid = errors.New("Datetime format is not in RFC3339")
)

// EventHandler delegates
type EventHandler func(*Scheduler, *Event)

// Scheduler ...
type Scheduler struct {
	delegate EventHandler
	stop     chan struct{}
	pendings chan *Event
	wg       *sync.WaitGroup
}

// New instance of scheduler
func New(d EventHandler) *Scheduler {
	return &Scheduler{
		delegate: d,
		// initialize stop channel
		stop: make(chan struct{}),
		// initialize buffered event channel
		pendings: make(chan *Event, 3),
		wg:       &sync.WaitGroup{},
	}
}

// Schedule an event
func (s *Scheduler) Schedule(e *Event) error {
	date, err := e.Date()
	if err != nil {
		return ErrTimeInvalid
	}

	now := time.Now()
	if date.Unix() <= now.Unix() {
		return ErrEventInPast
	}

	s.wg.Add(1)
	// fire a go routine
	go func(e *Event) {
		now := time.Now()
		target, _ := e.Date()
		waitDuration := target.Sub(now)

		defer s.wg.Done()
		select {
		case <-time.After(waitDuration):
			s.delegate(s, e)
		case <-s.stop:
			s.pendings <- e
		}
	}(e)
	return nil
}

// Stop all running scheduler and report all pending events
func (s *Scheduler) Stop() (events []*Event) {
	close(s.stop)
	go func() {
		s.wg.Wait()
		close(s.pendings)
	}()

	for e := range s.pendings {
		events = append(events, e)
	}

	return events
}

```

## Testing our Scheduler

```go
func Test_Scheduler(t *testing.T) {
	one := time.Now().Add(1 * time.Second) // next 1 sec
	two := one.Add(1 * time.Second)        // next 2 sec
	tri := two.Add(1 * time.Second)        // next 3 sec
	att := []Attachment{
		{Name: "Here!"}, // we need this to test immutability
	}

	sch := New(func(s *Scheduler, e *Event) {
		if e.attachments[0].Name == "THERE!" {
			t.Error("Name should be Here! not THERE!")
		}
	})
	ev1 := NewEvent(one.Format(time.RFC3339), att)
	sch.Schedule(ev1)

	ev2 := NewEvent(two.Format(time.RFC3339), att)
	sch.Schedule(ev2)

	ev3 := NewEvent(tri.Format(time.RFC3339), att)
	sch.Schedule(ev3)

    // Name changed after all events have been scheduled
	att[0].Name = "THERE!"

    // sleep for 2 secs to leave only 3rd event as pending
    time.Sleep(2 * time.Second)
    // stop scheduler and collect the pending events
	pendings := sch.Stop()
	if len(pendings) != 1 {
		t.Error("Pendings should only left the last event")
	} else {
		last := pendings[0]
		if last != ev3 {
			t.Error("Pendings[0] should equal to ev3")
		}
	}
}
```

And test result shows:

```text
Running tool: go test -timeout 30s -run ^(Test_Scheduler)$

PASS
ok  	/Go/pkg/scheduler	2.512s
Success: Tests passed.
```

---

Hope you guys enjoy this case study :v: