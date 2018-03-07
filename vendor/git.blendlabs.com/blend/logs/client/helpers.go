package client

import (
	"time"

	"git.blendlabs.com/blend/logs/pkg/protoutil"

	logv1 "git.blendlabs.com/blend/protos/log/v1"
	"github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
)

// NewMessageRaw returns a new raw logging message
func NewMessageRaw(ts time.Time, body []byte) logv1.Message {
	return logv1.Message{
		Type: logv1.MessageType_RAW,
		Meta: &logv1.Meta{
			Timestamp: protoutil.MarshalTimestamp(ts),
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
			Timestamp: protoutil.MarshalTimestamp(ts),
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
			Timestamp: protoutil.MarshalTimestamp(me.Timestamp()),
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
				Timestamp: protoutil.MarshalTimestamp(ee.Timestamp()),
			},
			Type:  logv1.MessageType_ERROR,
			Error: newMessageExceptionInner(typed),
		}
	}
	return logv1.Message{
		Type: logv1.MessageType_ERROR,
		Meta: &logv1.Meta{
			Timestamp: protoutil.MarshalTimestamp(ee.Timestamp()),
		},
		Error: &logv1.Error{
			Class: ee.Err().Error(),
		},
	}
}

// NewMessageHTTPRequest returns an http request log message.
func NewMessageHTTPRequest(wr *logger.WebRequestEvent) logv1.Message {
	return logv1.Message{
		Type: logv1.MessageType_HTTP,
		Meta: &logv1.Meta{
			Timestamp: protoutil.MarshalTimestamp(wr.Timestamp()),
		},
		HttpRequest: &logv1.HttpRequest{
			ContentLength:   wr.Request().ContentLength,
			ContentType:     wr.Request().Header.Get("Content-Type"),
			ContentEncoding: wr.Request().Header.Get("Content-Encoding"),
			Elapsed:         protoutil.MarshalDuration(wr.Elapsed()),
			Host:            wr.Request().Host,
			Method:          wr.Request().Method,
			Path:            wr.Request().URL.Path,
			QueryString:     wr.Request().URL.RawQuery,
			Referrer:        wr.Request().Referer(),
			RemoteAddr:      wr.Request().RemoteAddr,
			Scheme:          wr.Request().URL.Scheme,
			StatusCode:      int32(wr.StatusCode()),
			UserAgent:       wr.Request().UserAgent(),
			Url:             wr.Request().URL.String(),
			ResponseContentLength:   wr.ContentLength(),
			ResponseContentType:     wr.ContentType(),
			ResponseContentEncoding: wr.ContentEncoding(),
		},
	}
}

func newMessageExceptionInner(ex *exception.Ex) *logv1.Error {
	if ex == nil {
		return nil
	}
	return &logv1.Error{
		Class:   ex.Class(),
		Message: ex.Message(),
		Stack:   ex.StackTrace().AsStringSlice(),
		Inner:   newMessageExceptionInner(ex),
	}
}
