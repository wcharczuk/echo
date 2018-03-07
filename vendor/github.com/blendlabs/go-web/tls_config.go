package web

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
)

// TLSConfig is a config for app tls settings.
type TLSConfig struct {
	Cert     []byte `json:"cert" yaml:"cert" env:"TLS_CERT"`
	CertPath string `json:"certPath" yaml:"certPath" env:"TLS_CERT_PATH"`
	Key      []byte `json:"key" yaml:"key" env:"TLS_KEY"`
	KeyPath  string `json:"keyPath" yaml:"keyPath" env:"TLS_KEY_PATH"`

	CAPaths []string `json:"caPaths" yaml:"caPaths" env:"TLS_CA_PATHS,csv"`
}

// GetCert returns a tls cert.
func (tc TLSConfig) GetCert(defaults ...[]byte) []byte {
	return util.Coalesce.Bytes(tc.Cert, nil, defaults...)
}

// GetCertPath returns a tls cert path.
func (tc TLSConfig) GetCertPath(defaults ...string) string {
	return util.Coalesce.String(tc.CertPath, "", defaults...)
}

// GetKey returns a tls key.
func (tc TLSConfig) GetKey(defaults ...[]byte) []byte {
	return util.Coalesce.Bytes(tc.Key, nil, defaults...)
}

// GetKeyPath returns a tls key path.
func (tc TLSConfig) GetKeyPath(defaults ...string) string {
	return util.Coalesce.String(tc.KeyPath, "", defaults...)
}

// GetCAPaths returns a list of ca paths to add.
func (tc TLSConfig) GetCAPaths(defaults ...[]string) []string {
	return util.Coalesce.Strings(tc.CAPaths, nil, defaults...)
}

// GetConfig returns a stdlib tls config for the config.
func (tc TLSConfig) GetConfig() (*tls.Config, error) {
	if (len(tc.GetCertPath()) == 0 && len(tc.GetKeyPath()) == 0) ||
		(len(tc.GetCert()) == 0 && len(tc.GetKey()) == 0) {
		return nil, nil
	}

	var cert tls.Certificate
	var err error

	if len(tc.GetCertPath()) > 0 {
		cert, err = tls.LoadX509KeyPair(
			tc.GetCertPath(),
			tc.GetKeyPath(),
		)
	} else {
		cert, err = tls.X509KeyPair(tc.GetCert(), tc.GetKey())
	}

	if err != nil {
		return nil, exception.Wrap(err)
	}

	if len(tc.GetCAPaths()) == 0 {
		return &tls.Config{
			Certificates: []tls.Certificate{cert},
		}, nil
	}

	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, exception.Wrap(err)
	}
	for _, caPath := range tc.GetCAPaths() {
		caCert, err := ioutil.ReadFile(caPath)
		if err != nil {
			return nil, exception.Wrap(err)
		}
		certPool.AppendCertsFromPEM(caCert)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	}, nil
}