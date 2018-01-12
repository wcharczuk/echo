package logger

import (
	"bytes"
	"fmt"
	"time"
)

// Timedf returns a timed message event.
func Timedf(flag Flag, elapsed time.Duration, format string, args ...interface{}) TimedEvent {
	return TimedEvent{
		flag:    flag,
		ts:      time.Now().UTC(),
		message: fmt.Sprintf(format, args...),
		elapsed: elapsed,
	}
}

// NewTimedEventListener returns a new timed event listener.
func NewTimedEventListener(listener func(te TimedEvent)) Listener {
	return func(e Event) {
		if typed, isTyped := e.(TimedEvent); isTyped {
			listener(typed)
		}
	}
}

// TimedEvent is a message event with an elapsed time.
type TimedEvent struct {
	flag    Flag
	ts      time.Time
	message string
	elapsed time.Duration
}

// Flag returns the logger flag.
func (te TimedEvent) Flag() Flag {
	return te.flag
}

// Timestamp returns the event timestamp.
func (te TimedEvent) Timestamp() time.Time {
	return te.ts
}

// Message returns the string message.
func (te TimedEvent) Message() string {
	return te.message
}

// Elapsed returns the elapsed time.
func (te TimedEvent) Elapsed() time.Duration {
	return te.elapsed
}

// String implements fmt.Stringer
func (te TimedEvent) String() string {
	return fmt.Sprintf("%s (%v)", te.message, te.elapsed)
}

// WriteText implements TextWritable.
func (te TimedEvent) WriteText(tf TextFormatter, buf *bytes.Buffer) {
	buf.WriteString(te.String())
}

// WriteJSON implements JSONWritable.
func (te TimedEvent) WriteJSON() JSONObj {
	return JSONObj{
		JSONFieldMessage: te.message,
		JSONFieldElapsed: Milliseconds(te.elapsed),
	}
}
