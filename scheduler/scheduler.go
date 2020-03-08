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
