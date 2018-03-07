package protoutil

import (
	"net/http"
	"net/url"
	"time"

	logv1 "git.blendlabs.com/blend/protos/log/v1"
	logger "github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-util/uuid"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// Labels is a loose type alias to map[string]string
type Labels = map[string]string

// NewMessageMeta creates a new message meta.
func NewMessageMeta(labels Labels) *logv1.Meta {
	return &logv1.Meta{
		Uid:       uuid.V4().String(),
		Timestamp: MarshalTimestamp(time.Now().UTC()),
		Labels:    labels,
	}
}

// UnmarshalHTTPRequest unmarshals a log http request as a golang http request.
func UnmarshalHTTPRequest(req *logv1.HttpRequest) *http.Request {
	return &http.Request{
		Method:     req.Method,
		Host:       req.Host,
		RemoteAddr: req.RemoteAddr,
		Proto:      req.Scheme,
		URL: &url.URL{
			Scheme:   req.Scheme,
			Host:     req.Host,
			Path:     req.Path,
			RawQuery: req.QueryString,
		},
		Header: http.Header{
			"Content-Type":     []string{req.ContentType},
			"Content-Encoding": []string{req.ContentEncoding},
		},
		ContentLength: req.ContentLength,
	}
}

// UnmarshalHTTPRequestAsEvent unmarshals an http request message as a logger event.
func UnmarshalHTTPRequestAsEvent(req *logv1.HttpRequest) *logger.WebRequestEvent {
	return logger.NewWebRequest(UnmarshalHTTPRequest(req)).
		WithStatusCode(int(req.StatusCode)).
		WithElapsed(UnmarshalDuration(req.Elapsed)).
		WithContentType(req.ResponseContentType).
		WithContentEncoding(req.ResponseContentEncoding).
		WithContentLength(req.ResponseContentLength)
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
