package client

import (
	"os"
	"time"

	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/logger"

	logv1 "git.blendlabs.com/blend/protos/log/v1"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// HasCollectorUnixSocket returns if the unix socket is present for a given config.
func HasCollectorUnixSocket(cfg *Config) bool {
	socketPath := cfg.GetUnixSocketPath()
	if len(socketPath) == 0 {
		return false
	}

	if _, err := os.Stat(socketPath); err == nil {
		return true
	}
	return false
}

// NewMessageValues returns a new values logging message
func NewMessageValues(values map[string]string) *logv1.Values {
	return &logv1.Values{
		Values: values,
	}
}

// NewMessageInfo returns a new info message.
func NewMessageInfo(me *logger.MessageEvent) *logv1.Info {
	return &logv1.Info{
		Label:   string(me.Flag()),
		Message: me.Message(),
	}
}

// NewMessageError returns a new error.
func NewMessageError(err error) *logv1.Error {
	if err == nil {
		return nil
	}

	if typed, isTyped := err.(exception.Exception); isTyped {
		return &logv1.Error{
			Class:   typed.Class(),
			Message: typed.Message(),
			Stack:   typed.Stack().Strings(),
			Inner:   NewMessageError(typed.Inner()),
		}
	}
	return &logv1.Error{
		Class: err.Error(),
	}
}

// NewMessageHTTPRequest returns an http request log message.
func NewMessageHTTPRequest(wr *logger.WebRequestEvent) *logv1.HttpRequest {
	return &logv1.HttpRequest{
		ContentLength:   wr.Request().ContentLength,
		ContentType:     wr.Request().Header.Get("Content-Type"),
		ContentEncoding: wr.Request().Header.Get("Content-Encoding"),
		Elapsed:         MarshalDuration(wr.Elapsed()),
		Host:            wr.Request().Host,
		Method:          wr.Request().Method,
		Path:            wr.Request().URL.Path,
		QueryString:     wr.Request().URL.RawQuery,
		Referrer:        wr.Request().Referer(),
		RemoteAddr:      wr.Request().RemoteAddr,
		RemoteIP:        logger.GetIP(wr.Request()),
		Scheme:          wr.Request().URL.Scheme,
		StatusCode:      int32(wr.StatusCode()),
		UserAgent:       wr.Request().UserAgent(),
		Url:             wr.Request().URL.String(),
		Route:           wr.Route(),
		ResponseContentLength:   wr.ContentLength(),
		ResponseContentType:     wr.ContentType(),
		ResponseContentEncoding: wr.ContentEncoding(),
	}
}

// NewMessageAudit returns a new audit message.
func NewMessageAudit(ae *logger.AuditEvent) *logv1.Audit {
	return &logv1.Audit{
		Principal: ae.Principal(),
		Verb:      ae.Verb(),
		Noun:      ae.Noun(),
		Subject:   ae.Subject(),
		Property:  ae.Property(),
		Extra:     ae.Extra(),
	}
}

// NewQueryEvent returns a log pipeline message for a db event.
func NewQueryEvent(e *logger.QueryEvent) *logv1.Query {
	return &logv1.Query{
		Engine:   e.Engine(),
		Database: e.Database(),
		Label:    e.QueryLabel(),
		Body:     e.Body(),
		Elapsed:  MarshalDuration(e.Elapsed()),
	}
}

// MarshalTimestamp marshals a timestamp as proto.
func MarshalTimestamp(t time.Time) *timestamp.Timestamp {
	tv, _ := ptypes.TimestampProto(t)
	return tv
}

// UnmarshalTimestamp marshals a timestamp as proto.
func UnmarshalTimestamp(t *timestamp.Timestamp) time.Time {
	tv, _ := ptypes.Timestamp(t)
	return tv
}

// MarshalDuration marshals a duration as proto.
func MarshalDuration(d time.Duration) *duration.Duration {
	dv := ptypes.DurationProto(d)
	return dv
}

// UnmarshalDuration unmarshals a duration.
func UnmarshalDuration(d *duration.Duration) time.Duration {
	dv, _ := ptypes.Duration(d)
	return dv
}
