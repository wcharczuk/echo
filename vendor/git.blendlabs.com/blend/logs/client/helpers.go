package client

import (
	"os"
	"time"

	logv1 "git.blendlabs.com/blend/protos/log/v1"
	"github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// HasCollectorUnixSocket returns if the unix socket is present for a given config.
func HasCollectorUnixSocket(cfg *Config) bool {
	socketPath := cfg.GetCollectorUnixSocketPath()
	if len(socketPath) == 0 {
		return false
	}

	if _, err := os.Stat(socketPath); err == nil {
		return true
	}
	return false
}

// NewMessageRaw returns a new raw logging message
func NewMessageRaw(ts time.Time, body []byte) logv1.Message {
	return logv1.Message{
		Type: logv1.MessageType_RAW,
		Meta: &logv1.Meta{
			Timestamp: MarshalTimestamp(ts),
		},
		Raw: &logv1.Raw{
			Body: body,
		},
	}
}

// NewMessageValues returns a new values logging message
func NewMessageValues(ts time.Time, values map[string]string) logv1.Message {
	return logv1.Message{
		Type: logv1.MessageType_VALUES,
		Meta: &logv1.Meta{
			Timestamp: MarshalTimestamp(ts),
		},
		Values: &logv1.Values{
			Values: values,
		},
	}
}

// NewMessageInfo returns a new info message.
func NewMessageInfo(me *logger.MessageEvent) logv1.Message {
	return logv1.Message{
		Type: logv1.MessageType_INFO,
		Meta: &logv1.Meta{
			Timestamp: MarshalTimestamp(me.Timestamp()),
		},
		Info: &logv1.Info{
			Flag:    string(me.Flag()),
			Message: me.Message(),
		},
	}
}

// NewMessageError returns a new error logging message.
func NewMessageError(ee *logger.ErrorEvent) logv1.Message {
	if typed, isTyped := ee.Err().(*exception.Ex); isTyped {
		return logv1.Message{
			Meta: &logv1.Meta{
				Timestamp: MarshalTimestamp(ee.Timestamp()),
			},
			Type:  logv1.MessageType_ERROR,
			Error: newMessageExceptionInner(typed),
		}
	}
	return logv1.Message{
		Type: logv1.MessageType_ERROR,
		Meta: &logv1.Meta{
			Timestamp: MarshalTimestamp(ee.Timestamp()),
		},
		Error: &logv1.Error{
			Class: ee.Err().Error(),
		},
	}
}

func newMessageExceptionInner(err error) *logv1.Error {
	if err == nil {
		return nil
	}
	if typed, isTyped := err.(*exception.Ex); isTyped {
		return &logv1.Error{
			Class:   typed.Class(),
			Message: typed.Message(),
			Stack:   typed.StackTrace().AsStringSlice(),
			Inner:   newMessageExceptionInner(typed.Inner()),
		}
	}
	return &logv1.Error{
		Class: err.Error(),
	}
}

// NewMessageHTTPRequest returns an http request log message.
func NewMessageHTTPRequest(wr *logger.WebRequestEvent) logv1.Message {
	return logv1.Message{
		Type: logv1.MessageType_HTTP,
		Meta: &logv1.Meta{
			Timestamp: MarshalTimestamp(wr.Timestamp()),
		},
		HttpRequest: &logv1.HttpRequest{
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
		},
	}
}

// NewMessageAudit returns a new audit message.
func NewMessageAudit(ae *logger.AuditEvent) logv1.Message {
	return logv1.Message{
		Type: logv1.MessageType_AUDIT,
		Meta: &logv1.Meta{
			Timestamp: MarshalTimestamp(ae.Timestamp()),
		},
		Audit: &logv1.Audit{
			Principal: ae.Principal(),
			Verb:      ae.Verb(),
			Noun:      ae.Noun(),
			Subject:   ae.Subject(),
			Property:  ae.Property(),
			Extra:     ae.Extra(),
		},
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

// MetaProvider is a type that has message meta.
type MetaProvider interface {
	GetMeta() *logv1.Meta
}
