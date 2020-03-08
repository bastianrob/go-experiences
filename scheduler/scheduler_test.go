package scheduler

import (
	"testing"
	"time"
)

func Test_Scheduler(t *testing.T) {
	one := time.Now().Add(1 * time.Second) // next 1 sec
	two := one.Add(1 * time.Second)        // next 2 sec
	tri := two.Add(1 * time.Second)        // next 3 sec
	att := []Attachment{
		{Name: "Here!"}, // we need this to test immutability,
		// we'll later change it to THERE!
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
