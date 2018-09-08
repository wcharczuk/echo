package web

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/blend/go-sdk/util"
)

// PathRedirectHandler returns a handler for AuthManager.RedirectHandler based on a path.
func PathRedirectHandler(path string) func(*Ctx) *url.URL {
	return func(ctx *Ctx) *url.URL {
		u := *ctx.Request().URL
		u.Path = fmt.Sprintf("/login")
		return &u
	}
}

// NestMiddleware reads the middleware variadic args and organizes the calls recursively in the order they appear.
func NestMiddleware(action Action, middleware ...Middleware) Action {
	if len(middleware) == 0 {
		return action
	}

	var nest = func(a, b Middleware) Middleware {
		if b == nil {
			return a
		}
		return func(action Action) Action {
			return a(b(action))
		}
	}

	var metaAction Middleware
	for _, step := range middleware {
		metaAction = nest(step, metaAction)
	}
	return metaAction(action)
}

// NewSessionID returns a new session id.
// It is not a uuid; session ids are generated using a secure random source.
// SessionIDs are generally 64 bytes.
func NewSessionID() string {
	return util.String.MustSecureRandom(32)
}

// Base64URLDecode decodes a base64 string.
func Base64URLDecode(raw string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(raw)
}

// Base64URLEncode base64 encodes data.
func Base64URLEncode(raw []byte) string {
	return base64.URLEncoding.EncodeToString(raw)
}

// PortFromBindAddr returns a port number as an integer from a bind addr.
func PortFromBindAddr(bindAddr string) int32 {
	if len(bindAddr) == 0 {
		return 0
	}
	parts := strings.SplitN(bindAddr, ":", 2)
	if len(parts) == 0 {
		return 0
	}
	if len(parts) < 2 {
		return ParseInt32(parts[0])
	}
	return ParseInt32(parts[1])
}

// ParseInt32 parses an int32.
func ParseInt32(v string) int32 {
	parsed, _ := strconv.Atoi(v)
	return int32(parsed)
}

// NewMockRequest creates a mock request.
func NewMockRequest(method, path string) *http.Request {
	return &http.Request{
		Method:     method,
		Proto:      "http",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Host:       "localhost",
		RemoteAddr: "127.0.0.1:8080",
		Header: http.Header{
			HeaderUserAgent: []string{"go-web test"},
		},
		URL: &url.URL{
			Scheme:  "http",
			Host:    "localhost",
			Path:    path,
			RawPath: path,
		},
	}
}

// NewMockRequestWithCookie creates a mock request with a cookie attached to it.
func NewMockRequestWithCookie(method, path, cookieName, cookieValue string) *http.Request {
	req := NewMockRequest(method, path)
	req.AddCookie(NewCookie(cookieName, cookieValue))
	return req
}

// NewCookie returns a new name + value pair cookie.
func NewCookie(name, value string) *http.Cookie {
	return &http.Cookie{Name: name, Value: value}
}

// BoolValue parses a value as an bool.
// If the input error is set it short circuits.
func BoolValue(value string, inputErr error) (output bool, err error) {
	if inputErr != nil {
		err = inputErr
		return
	}
	switch strings.ToLower(value) {
	case "1", "true", "yes":
		output = true
	case "0", "false", "no":
		output = false
	default:
		err = fmt.Errorf("invalid boolean value")
	}
	return
}

// IntValue parses a value as an int.
// If the input error is set it short circuits.
func IntValue(value string, inputErr error) (output int, err error) {
	if inputErr != nil {
		err = inputErr
		return
	}
	output, err = strconv.Atoi(value)
	return
}

// Int64Value parses a value as an int64.
// If the input error is set it short circuits.
func Int64Value(value string, inputErr error) (output int64, err error) {
	if inputErr != nil {
		err = inputErr
		return
	}
	output, err = strconv.ParseInt(value, 10, 64)
	return
}

// Float64Value parses a value as an float64.
// If the input error is set it short circuits.
func Float64Value(value string, inputErr error) (output float64, err error) {
	if inputErr != nil {
		err = inputErr
		return
	}
	output, err = strconv.ParseFloat(value, 64)
	return
}

// DurationValue parses a value as an time.Duration.
// If the input error is set it short circuits.
func DurationValue(value string, inputErr error) (output time.Duration, err error) {
	if inputErr != nil {
		err = inputErr
		return
	}
	output, err = time.ParseDuration(value)
	return
}

// StringValue just returns the string directly from a value error pair.
func StringValue(value string, _ error) string {
	return value
}

// CSVValue just returns the string directly from a value error pair.
func CSVValue(value string, err error) ([]string, error) {
	if err != nil {
		return nil, err
	}
	return strings.Split(value, ","), nil
}
