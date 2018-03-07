package config

import (
	"time"

	logger "github.com/blendlabs/go-logger"
	util "github.com/blendlabs/go-util"
)

const (
	// DefaultConsumerHeartbeatInterval is a default heartbeat interval.
	DefaultConsumerHeartbeatInterval = 500 * time.Millisecond
)

// Consumer is the config for log consumers.
type Consumer struct {
	Meta              `json:",inline" yaml:",inline"`
	StreamURL         string        `json:"streamURL" yaml:"streamURL"`
	StreamName        string        `json:"streamName" yaml:"streamName"`
	HeartbeatInterval time.Duration `json:"heartbeatInterval" yaml:"heartbeatInterval"`
	Aws               Aws           `json:"aws" yaml:"aws"`
	Logger            logger.Config `json:"logger" yaml:"logger"`
}

// GetStreamURL gets the url for the kinesis stream.
func (c Consumer) GetStreamURL(inherited ...string) string {
	return util.Coalesce.String(
		c.StreamURL,
		util.String.Tokenize(DefaultStreamURLTemplate, Labels{"region": c.Aws.GetRegion()}),
		inherited...,
	)
}

// GetStreamName gets the stream name for the kinesis stream.
func (c Consumer) GetStreamName(inherited ...string) string {
	return util.Coalesce.String(
		c.StreamName,
		util.String.Tokenize(DefaultStreamNameTemplate, Labels{"env": c.Meta.GetEnvironment()}),
		inherited...,
	)
}

// GetHeartbeatInterval gets the worker heartbeat.
func (c Consumer) GetHeartbeatInterval(inherited ...time.Duration) time.Duration {
	return util.Coalesce.Duration(c.HeartbeatInterval, DefaultConsumerHeartbeatInterval, inherited...)
}
