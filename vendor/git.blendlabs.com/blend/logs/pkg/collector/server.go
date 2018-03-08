package collector

import (
	"fmt"
	"io"
	"strings"
	"time"

	"git.blendlabs.com/blend/logs/pkg/protoutil"
	collectorv1 "git.blendlabs.com/blend/protos/collector/v1"
	logv1 "git.blendlabs.com/blend/protos/log/v1"
	exception "github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-util/collections"
	"github.com/blendlabs/go-util/env"
	"github.com/blendlabs/go-util/uuid"
)

const (
	// EnvVarPodname is an env var.
	EnvVarPodname = "POD_NAME"
	// EnvVarNodeName is an env var.
	EnvVarNodeName = "NODE_NAME"
	// EnvVarNamespace is an env var.
	EnvVarNamespace = "NAMESPACE"

	// LabelCollectorPodname is a message meta label.
	LabelCollectorPodname = "collector-pod"
	// LabelCollectorNodeName is a message meta label.
	LabelCollectorNodeName = "collector-node"
	// LabelNamespace is a message meta label.
	LabelNamespace = "namespace"
)

// NewServer returns a new server.
func NewServer(buffer *collections.AutoflushBuffer) *Server {
	return &Server{buffer: buffer}
}

// Server is the grpc shim.
type Server struct {
	log    *logger.Logger
	buffer *collections.AutoflushBuffer
	labels Labels
}

// WithLabelsFromEnv returns labels from the environment.
func (s *Server) WithLabelsFromEnv() *Server {
	if s.labels == nil {
		s.labels = Labels{}
	}
	if env.Env().Has(EnvVarPodname) {
		s.labels[LabelCollectorPodname] = env.Env().String(EnvVarPodname)
	}
	if env.Env().Has(EnvVarNodeName) {
		s.labels[LabelCollectorNodeName] = env.Env().String(EnvVarNodeName)
	}
	if env.Env().Has(EnvVarNamespace) {
		s.labels[LabelNamespace] = env.Env().String(EnvVarNamespace)
	}
	return s
}

// WithLabels adds labels to the server.
func (s *Server) WithLabels(labels Labels) *Server {
	if s.labels == nil {
		s.labels = Labels{}
	}
	for key, value := range labels {
		s.labels[key] = value
	}
	return s
}

// WithLogger sets the server logger.
func (s *Server) WithLogger(log *logger.Logger) *Server {
	s.log = log
	return s
}

// Logger returns the logger.
func (s *Server) Logger() *logger.Logger {
	return s.log
}

// Buffer returns the autoflush buffer.
func (s *Server) Buffer() *collections.AutoflushBuffer {
	return s.buffer
}

func (s *Server) err(err error) error {
	if s.log != nil && err != nil {
		s.log.Error(err)
	}
	return err
}

func (s *Server) debugf(format string, args ...interface{}) {
	if s.log != nil {
		s.log.Debugf(format, args...)
	}
}

func (s *Server) infof(format string, args ...interface{}) {
	if s.log != nil {
		s.log.Infof(format, args...)
	}
}

// Push processes messages.
func (s *Server) Push(stream collectorv1.Collector_PushServer) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = s.err(exception.New("Push recovered from panic").WithMessagef("panic: %v", r))
		}
	}()

	var processed int
	start := time.Now()
	stats := map[logv1.MessageType]int{}
	increment := func(mt logv1.MessageType) {
		if value, hasValue := stats[mt]; hasValue {
			stats[mt] = value + 1
		} else {
			stats[mt] = 1
		}
	}

	var msg *logv1.Message
	for {
		msg, err = stream.Recv()
		if err == io.EOF {
			elapsed := time.Since(start)
			s.debugf("processed %d messages in %v", processed, elapsed)
			if len(stats) > 0 {
				var typeStats []string
				for messageType, count := range stats {
					typeStats = append(typeStats, fmt.Sprintf("%v: %d", messageType, count))
				}
				s.debugf("processed message types: %s", strings.Join(typeStats, ", "))
			}
			return stream.SendAndClose(&collectorv1.ReceiveSummary{
				MessageCount: int32(processed),
				Elapsed:      protoutil.MarshalDuration(elapsed),
			})
		}
		if err != nil {
			return s.err(exception.Wrap(err))
		}

		increment(msg.Type)

		msg.Meta.Uid = uuid.V4().String()
		if protoutil.UnmarshalTimestamp(msg.Meta.Timestamp).IsZero() {
			msg.Meta.Timestamp = protoutil.MarshalTimestamp(time.Now().UTC())
		}
		if msg.Meta.Annotations == nil {
			msg.Meta.Annotations = map[string]string{}
		}
		msg.Meta.Annotations["collectedAt"] = time.Now().UTC().Format(time.RFC3339)

		s.AddLabels(msg.Meta)
		s.buffer.Add(msg)
		processed++
	}
}

// AddLabels adds default labels to the meta.
func (s *Server) AddLabels(meta *logv1.Meta) {
	if meta.Labels == nil {
		meta.Labels = map[string]string{}
	}
	for key, value := range s.labels {
		meta.Labels[key] = value
	}
}
