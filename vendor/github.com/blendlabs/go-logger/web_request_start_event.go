package logger

import (
	"bytes"
	"net/http"
	"time"
)

// NewWebRequestStart creates a new web request start event.
func NewWebRequestStart(req *http.Request) WebRequestStartEvent {
	return WebRequestStartEvent{
		ts:  time.Now().UTC(),
		req: req,
	}
}

// NewWebRequestStartEventListener returns a new web request start event listener.
func NewWebRequestStartEventListener(listener func(wrse WebRequestStartEvent)) Listener {
	return func(e Event) {
		if typed, isTyped := e.(WebRequestStartEvent); isTyped {
			listener(typed)
		}
	}
}

// WebRequestStartEvent is an event type for http responses.
type WebRequestStartEvent struct {
	ts  time.Time
	req *http.Request
}

// Flag returns the event flag.
func (wre WebRequestStartEvent) Flag() Flag {
	return WebRequestStart
}

// Timestamp returns the event timestamp.
func (wre WebRequestStartEvent) Timestamp() time.Time {
	return wre.ts
}

// Request returns the request metadata.
func (wre WebRequestStartEvent) Request() *http.Request {
	return wre.req
}

// WriteText implements TextWritable.
func (wre WebRequestStartEvent) WriteText(formatter TextFormatter, buf *bytes.Buffer) {
	TextWriteRequestStart(formatter, buf, wre.req)
}

// WriteJSON implements JSONWritable.
func (wre WebRequestStartEvent) WriteJSON() JSONObj {
	return JSONWriteRequestStart(wre.Request())
}
