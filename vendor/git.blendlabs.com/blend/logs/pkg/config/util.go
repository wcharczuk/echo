package config

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	exception "github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util/env"
	yaml "gopkg.in/yaml.v2"
)

const (
	// EnvVarConfigPath is the env var for configs.
	EnvVarConfigPath = "CONFIG_PATH"

	// DefaultConfigMountPath is the default mount path for configs.
	DefaultConfigMountPath = "/var/run/secrets/config"

	// DefaultConfigPath is the default path for configs.
	DefaultConfigPath = "/var/run/secrets/config/config"

	// StateKeyConfig is the web app state key for a config.
	StateKeyConfig = "cluster_config"

	// ExtensionJSON is a file extension.
	ExtensionJSON = ".json"
	// ExtensionYAML is a file extension.
	ExtensionYAML = ".yaml"
	// ExtensionYML is a file extension.
	ExtensionYML = ".yml"
)

// Path returns the config path.
func Path() string {
	if env.Env().Has(EnvVarConfigPath) {
		return env.Env().String(EnvVarConfigPath)
	}
	return DefaultConfigPath
}

// ReadConfigFromPath reads a config from a given path.
func ReadConfigFromPath(path string, ref interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return exception.Wrap(err)
	}
	defer f.Close()
	return DeserializeConfig(filepath.Ext(path), f, ref)
}

// SerializeConfig serializes a config.
func SerializeConfig(cfg interface{}) ([]byte, error) {
	return json.Marshal(cfg)
}

// DeserializeConfig deserializes a config.
func DeserializeConfig(ext string, r io.Reader, ref interface{}) error {
	switch strings.ToLower(ext) {
	case ExtensionJSON:
		return exception.Wrap(json.NewDecoder(r).Decode(ref))
	case ExtensionYAML, ExtensionYML:
		contents, err := ioutil.ReadAll(r)
		if err != nil {
			return exception.Wrap(err)
		}
		return exception.Wrap(yaml.Unmarshal(contents, ref))
	default:
		return exception.Wrap(json.NewDecoder(r).Decode(ref))
	}
}
