package config

import (
	util "github.com/blendlabs/go-util"
	"github.com/blendlabs/go-util/env"
)

const (
	// DefaultAWSRegion is a default.
	DefaultAWSRegion = "us-east-1"
)

// NewAwsFromEnv returns a new aws config from the environment.
func NewAwsFromEnv() *Aws {
	var aws Aws
	env.Env().ReadInto(&aws)
	return &aws
}

// Aws is a config object.
type Aws struct {
	Region          string `json:"region,omitempty" yaml:"region,omitempty" env:"AWS_REGION_NAME"`
	AccessKeyID     string `json:"accessKeyID,omitempty" yaml:"accessKeyID,omitempty" env:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string `json:"secretAccessKey,omitempty" yaml:"secretAccessKey,omitempty" env:"AWS_SECRET_ACCESS_KEY"`
	Token           string `json:"token,omitempty" yaml:"token,omitempty" env:"AWS_SECURITY_TOKEN"`
}

// IsZero returns if the aws config is set.
func (a Aws) IsZero() bool {
	return len(a.GetAccessKeyID()) == 0 || len(a.GetSecretAccessKey()) == 0
}

// GetRegion gets a property or a default.
func (a Aws) GetRegion(inherited ...string) string {
	return util.Coalesce.String(a.Region, DefaultAWSRegion, inherited...)
}

// GetAccessKeyID gets a property or a default.
func (a Aws) GetAccessKeyID(inherited ...string) string {
	return util.Coalesce.String(a.AccessKeyID, "", inherited...)
}

// GetSecretAccessKey gets a property or a default.
func (a Aws) GetSecretAccessKey(inherited ...string) string {
	return util.Coalesce.String(a.SecretAccessKey, "", inherited...)
}

// GetToken gets a property or a default.
func (a Aws) GetToken(inherited ...string) string {
	return util.Coalesce.String(a.Token, "", inherited...)
}
