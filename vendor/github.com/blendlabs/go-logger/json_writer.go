package logger

import (
	"encoding/json"
	"io"
	"os"

	"github.com/blendlabs/go-util/env"
)

const (
	// JSONFieldLabel is a common json field.
	JSONFieldLabel = "label"
	// JSONFieldFlag is a common json field.
	JSONFieldFlag = "flag"
	// JSONFieldTimestamp is a common json field.
	JSONFieldTimestamp = "ts"
	// JSONFieldMessage is a common json field.
	JSONFieldMessage = "message"
	// JSONFieldElapsed is a common json field.
	JSONFieldElapsed = "elapsed"
	// JSONFieldErr is a common json field.
	JSONFieldErr = "err"
)

// JSONObj is a type alias for map[string]interface{}
type JSONObj = map[string]interface{}

// JSONWritable is a type with a custom formater for json writing.
type JSONWritable interface {
	WriteJSON() JSONObj
}

// NewJSONWriterFromEnv returns a new json writer from the environment.
func NewJSONWriterFromEnv() *JSONWriter {
	return &JSONWriter{
		output:      NewInterlockedWriter(os.Stdout),
		errorOutput: NewInterlockedWriter(os.Stderr),
		pretty:      env.Env().Bool(EnvVarJSONPretty),
	}
}

// JSONWriter is a json output format.
type JSONWriter struct {
	output      io.Writer
	errorOutput io.Writer
	label       string
	pretty      bool
}

// WithOutput sets the primary output.
func (jw *JSONWriter) WithOutput(output io.Writer) *JSONWriter {
	jw.output = NewInterlockedWriter(output)
	return jw
}

// WithErrorOutput sets the error output.
func (jw *JSONWriter) WithErrorOutput(errorOutput io.Writer) *JSONWriter {
	jw.errorOutput = NewInterlockedWriter(errorOutput)
	return jw
}

// Label returns a descriptive label for the writer.
func (jw *JSONWriter) Label() string {
	return jw.label
}

// WithLabel sets the writer label.
func (jw *JSONWriter) WithLabel(label string) Writer {
	jw.label = label
	return jw
}

// Output returns an io.Writer for the ouptut stream.
func (jw *JSONWriter) Output() io.Writer {
	return jw.output
}

// ErrorOutput returns an io.Writer for the error stream.
func (jw *JSONWriter) ErrorOutput() io.Writer {
	if jw.errorOutput != nil {
		return jw.errorOutput
	}
	return jw.output
}

// Write writes to stdout.
func (jw *JSONWriter) Write(e Event) error {
	return jw.write(jw.output, e)
}

// WriteError writes to stderr (or stdout if .errorOutput is unset).
func (jw *JSONWriter) WriteError(e Event) error {
	return jw.write(jw.ErrorOutput(), e)
}

func (jw *JSONWriter) write(output io.Writer, e Event) error {
	encoder := json.NewEncoder(output)
	if jw.pretty {
		encoder.SetIndent("", "\t")
	}

	if typed, isTyped := e.(JSONWritable); isTyped {
		fields := typed.WriteJSON()
		if len(jw.label) > 0 {
			fields[JSONFieldLabel] = jw.label
		}
		fields[JSONFieldFlag] = e.Flag()
		fields[JSONFieldTimestamp] = e.Timestamp()
		return encoder.Encode(fields)
	}

	return encoder.Encode(e)
}
