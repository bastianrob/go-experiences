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
