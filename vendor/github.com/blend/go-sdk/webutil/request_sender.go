package webutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var (
	// ErrURLUnset is a (hopefully) uncommon error.
	ErrURLUnset = fmt.Errorf("request url unset")

	// DefaultRequestTimeout is the default webhook timeout.
	DefaultRequestTimeout = 10 * time.Second

	// DefaultRequestMethod is the default webhook method.
	DefaultRequestMethod = "POST"
)

// NewRequestSender creates a new request sender.
func NewRequestSender(hookURL *url.URL) *RequestSender {
	transport := &http.Transport{
		DisableCompression: false,
		DisableKeepAlives:  false,
	}

	return &RequestSender{
		url:       hookURL,
		method:    DefaultRequestMethod,
		transport: transport,
		headers:   http.Header{},
		client: &http.Client{
			Transport: transport,
			Timeout:   DefaultRequestTimeout,
		},
	}
}

// RequestSender is a slack webhook sender.
type RequestSender struct {
	url       *url.URL
	transport *http.Transport
	close     bool
	method    string
	client    *http.Client
	headers   http.Header
	tracer    RequestTracer
}

// WithTracer sets the request tracer.
func (rs *RequestSender) WithTracer(tracer RequestTracer) *RequestSender {
	rs.tracer = tracer
	return rs
}

// Tracer returns the request tracer.
func (rs *RequestSender) Tracer() RequestTracer {
	return rs.tracer
}

// WithClose sets if we should close the connection.
func (rs *RequestSender) WithClose(close bool) *RequestSender {
	rs.close = close
	rs.transport.DisableKeepAlives = close
	return rs
}

// Close returns if we should close the connection.
func (rs *RequestSender) Close() bool {
	return rs.close
}

// WithHeader sets headers.
func (rs *RequestSender) WithHeader(key, value string) *RequestSender {
	rs.headers.Set(key, value)
	return rs
}

// Headers returns the headers.
func (rs *RequestSender) Headers() http.Header {
	return rs.headers
}

// WithMethod sets the webhook method (defaults to POST).
func (rs *RequestSender) WithMethod(method string) *RequestSender {
	rs.method = method
	return rs
}

// Method is the webhook method.
// It defaults to "POST".
func (rs *RequestSender) Method() string {
	return rs.method
}

// SendJSON sends a message to the webhook with a given msg body as json.
func (rs *RequestSender) SendJSON(msg interface{}) (res *http.Response, err error) {
	var req *http.Request
	req, err = rs.reqJSON(msg)
	if err != nil {
		return
	}
	if rs.tracer != nil {
		tf := rs.tracer.Start(req)
		if tf != nil {
			defer func() { tf.Finish(req, res, err) }()
		}
	}
	res, err = rs.client.Do(req)
	return
}

func (rs *RequestSender) reqJSON(msg interface{}) (*http.Request, error) {
	contents, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return rs.req(contents), nil
}

func (rs *RequestSender) req(contents []byte) *http.Request {
	return &http.Request{
		Method:        rs.method,
		Body:          ioutil.NopCloser(bytes.NewBuffer(contents)),
		ContentLength: int64(len(contents)),
		Close:         rs.close,
		URL:           rs.url,
		Header:        rs.headers,
	}
}
