package config

const (
	// DefaultStreamNameTemplate is the default template for the stream name.
	DefaultStreamNameTemplate = "${env}-message-bus"

	// DefaultStreamURLTemplate is the default template for the stream url.
	DefaultStreamURLTemplate = "https://kinesis.${region}.amazonaws.com"
)
