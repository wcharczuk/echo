package client

import (
	util "github.com/blendlabs/go-util"
	"github.com/blendlabs/go-util/env"
)

const (
	// DefaultAddr is the default client addr.
	DefaultAddr = "unix://var/run/log-collector.sock"
)

// NewConfigFromEnv returns a new config from the environment.
func NewConfigFromEnv() *Config {
	var cfg Config
	env.Env().ReadInto(&cfg)
	return &cfg
}

// Config is the client config.
type Config struct {
	Addr string `json:"addr" yaml:"addr" env:"CLIENT_ADDR"`

	ServerName string `json:"serverName" yaml:"serverName" env:"CLIENT_SERVER_NAME"`
	UseTLS     *bool  `json:"useTLS" yaml:"useTLS" env:"CLIENT_USE_TLS"`
	CAFile     string `json:"caFile" yaml:"caFile" env:"CLIENT_TLS_CA_FILE"`
}

// GetAddr gets an addr or a default.
func (c Config) GetAddr(inherited ...string) string {
	return util.Coalesce.String(c.Addr, DefaultAddr, inherited...)
}

// GetServerName gets an addr or a default.
func (c Config) GetServerName(inherited ...string) string {
	return util.Coalesce.String(c.ServerName, DefaultAddr, inherited...)
}

// GetUseTLS sets the client to use tls.
func (c Config) GetUseTLS(inherited ...bool) bool {
	return util.Coalesce.Bool(c.UseTLS, false, inherited...)
}

// GetCAFile gets a property or a default.
func (c Config) GetCAFile(inherited ...string) string {
	return util.Coalesce.String(c.CAFile, "", inherited...)
}
