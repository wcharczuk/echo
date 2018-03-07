package config

import util "github.com/blendlabs/go-util"

const (
	// DefaultNamespace is the default namespace.
	DefaultNamespace = "blend-system"

	// DefaultEnvironment is the default environment.
	DefaultEnvironment = "sandbox"

	// ServiceEnvSandbox is a service env.
	ServiceEnvSandbox = "sandbox"
	// ServiceEnvProduction is a service env.
	ServiceEnvProduction = "prod"
)

// CurrentVersion is the current application version.
var CurrentVersion = "1.0.0"

// Meta is the cluster config meta.
type Meta struct {
	// Version is the semantic version of the cluster.
	Version string `json:"version" yaml:"version" env:"VERSION"`
	// Namespace is the namespace to root resources in.
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty" env:"NAMESPACE"`
	// Environment is the environment of the cluster (sandbox, prod etc.)
	Environment string `json:"environment,omitempty" yaml:"environment,omitempty" env:"SERVICE_ENV"`
	// GitRef is the currently deployed git ref.
	GitRef string `json:"gitRef,omitempty" yaml:"gitRef,omitempty" env:"GIT_REF"`
}

// GetVersion gets a value or a default.
func (m Meta) GetVersion(defaults ...string) string {
	return util.Coalesce.String(m.Version, CurrentVersion, defaults...)
}

// GetNamespace gets a value or a default.
func (m Meta) GetNamespace(defaults ...string) string {
	return util.Coalesce.String(m.Namespace, DefaultNamespace, defaults...)
}

// GetEnvironment returns the cluster environment.
func (m Meta) GetEnvironment(defaults ...string) string {
	return util.Coalesce.String(m.Environment, DefaultEnvironment, defaults...)
}

// IsProdlike returns if the cluster meta environment is prodlike.
func (m Meta) IsProdlike(defaults ...string) bool {
	return m.GetEnvironment(defaults...) == ServiceEnvProduction
}

// GetGitRef returns the current git ref of the collector.
func (m Meta) GetGitRef(defaults ...string) string {
	return util.Coalesce.String(m.GitRef, "", defaults...)
}
