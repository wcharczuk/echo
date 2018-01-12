package logger

import (
	"bytes"
	"fmt"
	"time"
)

// Errorf returns a new error event based on format and arguments.
func Errorf(flag Flag, format string, args ...interface{}) ErrorEvent {
	return ErrorEvent{
		flag: flag,
		ts:   time.Now().UTC(),
		err:  fmt.Errorf(format, args...),
	}
}

// ErrorfWithFlagTextColor returns a new error event based on format and arguments with a given flag text color.
func ErrorfWithFlagTextColor(flag Flag, flagColor AnsiColor, format string, args ...interface{}) ErrorEvent {
	return ErrorEvent{
		flag:      flag,
		flagColor: flagColor,
		ts:        time.Now().UTC(),
		err:       fmt.Errorf(format, args...),
	}
}

// NewError returns a new error event.
func NewError(flag Flag, err error) ErrorEvent {
	return ErrorEvent{
		flag: flag,
		ts:   time.Now().UTC(),
		err:  err,
	}
}

// NewErrorWithState returns a new error event with state.
func NewErrorWithState(flag Flag, err error, state interface{}) ErrorEvent {
	return ErrorEvent{
		flag:  flag,
		ts:    time.Now().UTC(),
		err:   err,
		state: state,
	}
}

// NewErrorEventListener returns a new error event listener.
func NewErrorEventListener(listener func(me ErrorEvent)) Listener {
	return func(e Event) {
		if typed, isTyped := e.(ErrorEvent); isTyped {
			listener(typed)
		}
	}
}

// ErrorEvent is an event that wraps an error.
type ErrorEvent struct {
	flag      Flag
	flagColor AnsiColor
	ts        time.Time
	err       error
	state     interface{}
}

// IsError indicates if we should write to the error writer or not.
func (ee ErrorEvent) IsError() bool {
	return true
}

// Timestamp returns the event timestamp.
func (ee ErrorEvent) Timestamp() time.Time {
	return ee.ts
}

// Flag returns the event flag.
func (ee ErrorEvent) Flag() Flag {
	return ee.flag
}

// Err returns the underlying error.
func (ee ErrorEvent) Err() error {
	return ee.err
}

// State returns underlying state, typically an http.Request.
func (ee ErrorEvent) State() interface{} {
	return ee.state
}

// FlagTextColor returns a custom color for the flag.
func (ee ErrorEvent) FlagTextColor() AnsiColor {
	return ee.flagColor
}

// WriteText implements TextWritable.
func (ee ErrorEvent) WriteText(formatter TextFormatter, buf *bytes.Buffer) {
	buf.WriteString(fmt.Sprintf("%+v", ee.err))
}

// WriteJSON implements JSONWritable.
func (ee ErrorEvent) WriteJSON() JSONObj {
	return JSONObj{
		JSONFieldErr: ee.err,
	}
}
