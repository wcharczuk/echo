package logger

import (
	"bytes"
	"fmt"
	"time"
)

// Messagef returns a new Message Event.
func Messagef(flag Flag, format string, args ...interface{}) MessageEvent {
	return MessageEvent{
		flag:    flag,
		ts:      time.Now().UTC(),
		message: fmt.Sprintf(format, args...),
	}
}

// MessagefWithFlagTextColor returns a new Message Event with a given flag text color.
func MessagefWithFlagTextColor(flag Flag, flagColor AnsiColor, format string, args ...interface{}) MessageEvent {
	return MessageEvent{
		flag:      flag,
		flagColor: flagColor,
		ts:        time.Now().UTC(),
		message:   fmt.Sprintf(format, args...),
	}
}

// NewMessageEventListener returns a new message event listener.
func NewMessageEventListener(listener func(me MessageEvent)) Listener {
	return func(e Event) {
		if typed, isTyped := e.(MessageEvent); isTyped {
			listener(typed)
		}
	}
}

// MessageEvent is a common type of message.
type MessageEvent struct {
	flag      Flag
	flagColor AnsiColor
	ts        time.Time
	message   string
}

// Flag returns the message flag.
func (me MessageEvent) Flag() Flag {
	return me.flag
}

// Timestamp returns the message timestamp.
func (me MessageEvent) Timestamp() time.Time {
	return me.ts
}

// Message returns the message.
func (me MessageEvent) Message() string {
	return me.message
}

// FlagTextColor returns a custom color for the flag.
func (me MessageEvent) FlagTextColor() AnsiColor {
	return me.flagColor
}

// WriteText implements TextWritable.
func (me MessageEvent) WriteText(formatter TextFormatter, buf *bytes.Buffer) {
	buf.WriteString(me.message)
}

// WriteJSON implements JSONWriteable.
func (me MessageEvent) WriteJSON() JSONObj {
	return JSONObj{
		JSONFieldMessage: me.message,
	}
}

// String returns the message event body.
func (me MessageEvent) String() string {
	return me.message
}
