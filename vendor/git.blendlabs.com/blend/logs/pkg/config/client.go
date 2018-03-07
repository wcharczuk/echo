package config

import util "github.com/blendlabs/go-util"

const (
	// DefaultClientAddr is the default client addr.
	DefaultClientAddr = "unix://var/run/log-collector.sock"
)

// Client is the client config.
type Client struct {
	Addr string `json:"addr" yaml:"addr"`

	TLSCAFile   string `json:"tlsCAFile" yaml:"tlsCAFile"`
	TLSCertFile string `json:"tlsCertFile" yaml:"tlsCertFile"`
	TLSKeyFile  string `json:"tlsKeyFile" yaml:"tlsKeyFile"`
}

// GetAddr gets an addr or a default.
func (c Client) GetAddr(inherited ...string) string {
	return util.Coalesce.String(c.Addr, DefaultClientAddr, inherited...)
}

// GetTLSCAFile gets a property or a default.
func (c Client) GetTLSCAFile(inherited ...string) string {
	return util.Coalesce.String(c.TLSCAFile, "", inherited...)
}

// GetTLSCertFile gets a property or a default.
func (c Client) GetTLSCertFile(inherited ...string) string {
	return util.Coalesce.String(c.TLSCertFile, "", inherited...)
}

// GetTLSKeyFile gets a property or a default.
func (c Client) GetTLSKeyFile(inherited ...string) string {
	return util.Coalesce.String(c.TLSKeyFile, "", inherited...)
}
