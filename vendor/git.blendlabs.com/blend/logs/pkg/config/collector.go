package config

import (
	"fmt"
	"time"

	logger "github.com/blendlabs/go-logger"
	util "github.com/blendlabs/go-util"
)

const (
	// DefaultBufferCapacity is the default buffer length before a flush is triggered.
	DefaultBufferCapacity = 500

	// DefaultBufferFlushInterval is the default buffer autoflush interval.
	DefaultBufferFlushInterval = 500 * time.Millisecond

	// DefaultPort is the default listen port.
	DefaultPort = 5125
)

// Collector is the collector config.
type Collector struct {
	Meta `json:",inline" yaml:",inline"`

	// StreamName is the name of the log stream to write to.
	StreamName string `json:"streamName" yaml:"streamName"`

	BufferCapacity      int           `json:"bufferCapacity" yaml:"bufferCapacity"`
	BufferFlushInterval time.Duration `json:"bufferFlushInterval" yaml:"bufferFlushInterval"`

	BindAddr    string `json:"bindAddr" yaml:"bindAddr"`
	Port        int32  `json:"port" yaml:"port"`
	TLSCAFile   string `json:"tlsCAFile" yaml:"tlsCAFile"`
	TLSCertFile string `json:"tlsCertFile" yaml:"tlsCertFile"`
	TLSKeyFile  string `json:"tlsKeyFile" yaml:"tlsKeyFile"`

	Aws    Aws           `json:"aws" yaml:"aws"`
	Logger logger.Config `json:"logger" yaml:"logger"`
}

// GetStreamName gets the stream name for the kinesis stream.
func (c Collector) GetStreamName(inherited ...string) string {
	return util.Coalesce.String(
		c.StreamName,
		util.String.Tokenize(DefaultStreamNameTemplate, Labels{"env": c.Meta.GetEnvironment()}),
		inherited...,
	)
}

// GetPort gets the port or a default.
func (c Collector) GetPort(inherited ...int32) int32 {
	return util.Coalesce.Int32(c.Port, DefaultPort, inherited...)
}

// GetBindAddr gets a bind addr if set, otherwise returns the port formatted as a bindaddr.
func (c Collector) GetBindAddr(inherited ...string) string {
	if len(c.BindAddr) > 0 {
		return c.BindAddr
	}
	if len(inherited) > 0 {
		return inherited[0]
	}

	return fmt.Sprintf(":%d", c.GetPort())
}

// GetBufferCapacity gets the buffer capacity.
func (c Collector) GetBufferCapacity(inherited ...int) int {
	return util.Coalesce.Int(c.BufferCapacity, DefaultBufferCapacity, inherited...)
}

// GetBufferFlushInterval gets the buffer flush interval.
func (c Collector) GetBufferFlushInterval(inherited ...time.Duration) time.Duration {
	return util.Coalesce.Duration(c.BufferFlushInterval, DefaultBufferFlushInterval, inherited...)
}
