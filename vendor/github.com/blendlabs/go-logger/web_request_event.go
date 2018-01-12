package logger

import (
	"bytes"
	"net/http"
	"time"
)

// NewWebRequest creates a new web request event.
func NewWebRequest(req *http.Request, statusCode, contentLength int, elapsed time.Duration) WebRequestEvent {
	return WebRequestEvent{
		flag:          WebRequest,
		ts:            time.Now().UTC(),
		req:           req,
		statusCode:    statusCode,
		contentLength: contentLength,
		elapsed:       elapsed,
	}
}

// NewWebRequestEventListener returns a new web request event listener.
func NewWebRequestEventListener(listener func(wrse WebRequestEvent)) Listener {
	return func(e Event) {
		if typed, isTyped := e.(WebRequestEvent); isTyped {
			listener(typed)
		}
	}
}

// WebRequestEvent is an event type for http responses.
type WebRequestEvent struct {
	flag          Flag
	ts            time.Time
	req           *http.Request
	statusCode    int
	contentLength int
	elapsed       time.Duration
}

// Flag returns the event flag.
func (wre WebRequestEvent) Flag() Flag {
	return wre.flag
}

// Timestamp returns the event timestamp.
func (wre WebRequestEvent) Timestamp() time.Time {
	return wre.ts
}

// Request returns the request metadata.
func (wre WebRequestEvent) Request() *http.Request {
	return wre.req
}

// StatusCode is the HTTP status code of the response.
func (wre WebRequestEvent) StatusCode() int {
	return wre.statusCode
}

// ContentLength is the size of the response.
func (wre WebRequestEvent) ContentLength() int {
	return wre.contentLength
}

// Elapsed is the duration of the request.
func (wre WebRequestEvent) Elapsed() time.Duration {
	return wre.elapsed
}

// WriteText implements TextWritable.
func (wre WebRequestEvent) WriteText(formatter TextFormatter, buf *bytes.Buffer) {
	TextWriteRequest(formatter, buf, wre.req, wre.statusCode, wre.contentLength, wre.elapsed)
}

// WriteJSON implements JSONWritable.
func (wre WebRequestEvent) WriteJSON() JSONObj {
	return JSONWriteRequest(wre.req, wre.statusCode, wre.contentLength, wre.elapsed)
}
