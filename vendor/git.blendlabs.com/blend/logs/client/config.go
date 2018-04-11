package client

import (
	"strings"

	"github.com/blend/go-sdk/env"
	"github.com/blend/go-sdk/util"
)

const (
	// DefaultAddr is the default client addr.
	DefaultAddr = "unix:///var/run/log-collector/collector.sock"
)

// NewConfigFromEnv returns a new config from the environment.
func NewConfigFromEnv() *Config {
	var cfg Config
	if err := env.Env().ReadInto(&cfg); err != nil {
		panic(err)
	}
	return &cfg
}

// Config is the client config.
type Config struct {
	Addr       string `json:"addr" yaml:"addr" env:"LOGS_ADDR"`
	ServerName string `json:"serverName" yaml:"serverName" env:"LOGS_SERVER_NAME"`
	UseTLS     *bool  `json:"useTLS" yaml:"useTLS" env:"LOGS_USE_TLS"`
	CAFile     string `json:"caFile" yaml:"caFile" env:"LOGS_TLS_CA_FILE"`

	ServiceName string `json:"serviceName" yaml:"serviceName" env:"SERVICE_NAME"`
	Hostname    string `json:"hostname" yaml:"hostname" env:"HOSTNAME"`

	DefaultLabels map[string]string `json:"defaultLabels" yaml:"defaultLabels"`
}

// GetUnixSocketPath gets the unix socket path.
func (c Config) GetUnixSocketPath() string {
	if strings.HasPrefix(c.GetAddr(), "unix://") {
		return strings.TrimPrefix(c.GetAddr(), "unix://")
	}
	return ""
}

// GetAddr gets an addr or a default.
func (c Config) GetAddr(inherited ...string) string {
	return util.Coalesce.String(c.Addr, DefaultAddr, inherited...)
}

// GetServerName gets an addr or a default.
func (c Config) GetServerName(inherited ...string) string {
	return util.Coalesce.String(c.ServerName, "", inherited...)
}

// GetUseTLS sets the client to use tls.
func (c Config) GetUseTLS(inherited ...bool) bool {
	return util.Coalesce.Bool(c.UseTLS, false, inherited...)
}

// GetCAFile gets a property or a default.
func (c Config) GetCAFile(inherited ...string) string {
	return util.Coalesce.String(c.CAFile, "", inherited...)
}

// GetServiceName gets a property or default.
func (c Config) GetServiceName(inherited ...string) string {
	return util.Coalesce.String(c.ServiceName, "", inherited...)
}

// GetHostname gets a property or default.
func (c Config) GetHostname(inherited ...string) string {
	return util.Coalesce.String(c.Hostname, "", inherited...)
}

// GetDefaultLabels returns the default labels set.
func (c Config) GetDefaultLabels() map[string]string {
	output := map[string]string{}
	for key, value := range c.DefaultLabels {
		output[key] = value
	}
	if len(c.GetServiceName()) > 0 {
		output[LabelService] = c.GetServiceName()
	}
	if len(c.GetHostname()) > 0 {
		output[LabelServicePod] = c.GetHostname()
	}
	return output
}
