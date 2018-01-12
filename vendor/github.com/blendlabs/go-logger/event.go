package logger

import (
	"time"
)

// Event is an interface representing methods necessary to trigger listeners.
type Event interface {
	Flag() Flag
	Timestamp() time.Time
}

// EventErrorWritable determines if we should write the event to the error stream.
type EventErrorWritable interface {
	IsError() bool
}

// EventWritable lets us disable implicit writing for some events.
type EventWritable interface {
	ShouldWrite() bool
}
