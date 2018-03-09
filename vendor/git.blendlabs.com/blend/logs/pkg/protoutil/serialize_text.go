package protoutil

import (
	"fmt"
	"strings"

	logv1 "git.blendlabs.com/blend/protos/log/v1"
	exception "github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
	util "github.com/blendlabs/go-util"
)

const (
	rawFlag    logger.Flag = "raw"
	valuesFlag logger.Flag = "values"
)

// SerializeTextOptions is a set of options to serialize text with.
type SerializeTextOptions struct {
	EventLabelTemplate string
}

// SerializeText serializes messages as text.
func SerializeText(opts *SerializeTextOptions, msg *logv1.Message, writer *logger.TextWriter) error {
	switch msg.Type {
	case logv1.MessageType_RAW:
		return SerializeRawText(opts, msg, writer)
	case logv1.MessageType_VALUES:
		return SerializeValuesText(opts, msg, writer)
	case logv1.MessageType_ERROR:
		return SerializeErrorText(opts, msg, writer)
	case logv1.MessageType_HTTP:
		return SerializeHTTPRequestText(opts, msg, writer)
	case logv1.MessageType_INFO:
		return SerializeInfoText(opts, msg, writer)
	default:
		return exception.NewFromErr(ErrUnsupportedMessageType).WithMessagef("message type: %v", msg.Type)
	}
}

// SerializeRawText serializes a raw message as text.
func SerializeRawText(opts *SerializeTextOptions, msg *logv1.Message, writer *logger.TextWriter) error {
	if msg.Raw == nil {
		return exception.NewFromErr(ErrUnset).WithMessagef("field: Raw")
	}
	event := logger.Messagef(rawFlag, string(msg.Raw.Body)).WithTimestamp(UnmarshalTimestamp(msg.Meta.Timestamp))
	if len(opts.EventLabelTemplate) > 0 {
		event = event.WithLabel(util.String.Tokenize(opts.EventLabelTemplate, msg.Meta.Labels))
	}
	return writer.Write(event)
}

// SerializeValuesText serializes a values as text.
func SerializeValuesText(opts *SerializeTextOptions, msg *logv1.Message, writer *logger.TextWriter) error {
	if msg.Values == nil {
		return exception.NewFromErr(ErrUnset).WithMessagef("field: Values")
	}

	buf := writer.GetBuffer()
	defer writer.PutBuffer(buf)

	for key, value := range msg.Values.Values {
		buf.WriteRune(logger.RuneNewline)
		buf.WriteString(fmt.Sprintf("%s: %s", key, value))
	}
	event := logger.Messagef(valuesFlag, buf.String()).WithTimestamp(UnmarshalTimestamp(msg.Meta.Timestamp))
	if len(opts.EventLabelTemplate) > 0 {
		event = event.WithLabel(util.String.Tokenize(opts.EventLabelTemplate, msg.Meta.Labels))
	}
	return writer.Write(event)
}

// SerializeErrorText serializes an error as text.
func SerializeErrorText(opts *SerializeTextOptions, msg *logv1.Message, writer *logger.TextWriter) error {
	if msg.Error == nil {
		return exception.NewFromErr(ErrUnset).WithMessagef("field: Error")
	}

	buf := writer.GetBuffer()
	defer writer.PutBuffer(buf)

	buf.WriteString(msg.Error.Class)
	if len(msg.Error.Message) > 0 {
		buf.WriteRune(logger.RuneNewline)
		buf.WriteString(msg.Error.Message)
	}
	if len(msg.Error.Stack) > 0 {
		buf.WriteRune(logger.RuneNewline)
		buf.WriteString(strings.Join(msg.Error.Stack, "\n"))
	}
	event := logger.Messagef(logger.Error, buf.String()).WithTimestamp(UnmarshalTimestamp(msg.Meta.Timestamp))
	if len(opts.EventLabelTemplate) > 0 {
		event = event.WithLabel(util.String.Tokenize(opts.EventLabelTemplate, msg.Meta.Labels))
	}
	return writer.Write(event)
}

// SerializeHTTPRequestText serializes an http request as text.
func SerializeHTTPRequestText(opts *SerializeTextOptions, msg *logv1.Message, writer *logger.TextWriter) error {
	if msg.HttpRequest == nil {
		return exception.NewFromErr(ErrUnset).WithMessagef("field: HttpRequest")
	}

	event := UnmarshalHTTPRequestAsEvent(msg.HttpRequest)
	if len(opts.EventLabelTemplate) > 0 {
		event = event.WithLabel(util.String.Tokenize(opts.EventLabelTemplate, msg.Meta.Labels))
	}
	event.WithTimestamp(UnmarshalTimestamp(msg.Meta.Timestamp))
	return writer.Write(event)
}

// SerializeInfoText serializes an info message as text.
func SerializeInfoText(opts *SerializeTextOptions, msg *logv1.Message, writer *logger.TextWriter) error {
	if msg.Info == nil {
		return exception.NewFromErr(ErrUnset).WithMessagef("field: Info")
	}
	event := logger.Messagef(logger.Flag(msg.Info.Flag), msg.Info.Message).WithTimestamp(UnmarshalTimestamp(msg.Meta.Timestamp))
	if len(opts.EventLabelTemplate) > 0 {
		event = event.WithLabel(util.String.Tokenize(opts.EventLabelTemplate, msg.Meta.Labels))
	}
	return writer.Write(event)
}
