package client

import (
	"strings"

	util "github.com/blendlabs/go-util"
	"github.com/blendlabs/go-util/env"
)

const (
	// DefaultAddr is the default client addr.
	DefaultAddr = "unix:////var/run/log-collector/collector.sock"
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
	CollectorAddr       string `json:"collectorAddr" yaml:"collectorAddr" env:"LOG_CLIENT_ADDR"`
	CollectorServerName string `json:"collectorServerName" yaml:"collectorServerName" env:"LOG_CLIENT_SERVER_NAME"`
	UseTLS              *bool  `json:"useTLS" yaml:"useTLS" env:"LOG_CLIENT_USE_TLS"`
	CAFile              string `json:"caFile" yaml:"caFile" env:"LOG_CLIENT_TLS_CA_FILE"`

	ServiceName string `json:"serviceName" yaml:"serviceName" env:"SERVICE_NAME"`
	Hostname    string `json:"hostname" yaml:"hostname" env:"HOSTNAME"`

	DefaultLabels map[string]string `json:"defaultLabels" yaml:"defaultLabels"`
}

// GetCollectorUnixSocketPath gets the unix socket path.
func (c Config) GetCollectorUnixSocketPath() string {
	if strings.HasPrefix(c.GetCollectorAddr(), "unix://") {
		return strings.TrimPrefix(c.GetCollectorAddr(), "unix://")
	}
	return ""
}

// GetCollectorAddr gets an addr or a default.
func (c Config) GetCollectorAddr(inherited ...string) string {
	return util.Coalesce.String(c.CollectorAddr, DefaultAddr, inherited...)
}

// GetCollectorServerName gets an addr or a default.
func (c Config) GetCollectorServerName(inherited ...string) string {
	return util.Coalesce.String(c.CollectorServerName, DefaultAddr, inherited...)
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
