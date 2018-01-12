package logger

import (
	"io"
	"strings"

	"github.com/blendlabs/go-util/env"
)

// Writer is a type that can consume events.
type Writer interface {
	Label() string
	WithLabel(string) Writer
	Write(Event) error
	WriteError(Event) error
	Output() io.Writer
	ErrorOutput() io.Writer
}

// NewWriterFromEnv returns a new writer based on the environment variable `LOG_FORMAT`.
func NewWriterFromEnv() Writer {
	if format := env.Env().String(EnvVarFormat); len(format) > 0 {
		switch strings.ToLower(format) {
		case "json":
			return NewJSONWriterFromEnv()
		case "text":
			return NewTextWriterFromEnv()
		}
	}
	return NewTextWriterFromEnv()
}
