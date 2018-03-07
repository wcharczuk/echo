package protoutil

import (
	"fmt"
	"strings"

	logv1 "git.blendlabs.com/blend/protos/log/v1"
	exception "github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
)

const (
	rawFlag    logger.Flag = "raw"
	valuesFlag logger.Flag = "values"
)

// SerializeText serializes messages as text.
func SerializeText(msg *logv1.Message, writer *logger.TextWriter) error {
	switch msg.Type {
	case logv1.MessageType_RAW:
		return SerializeRawText(msg, writer)
	case logv1.MessageType_VALUES:
		return SerializeValuesText(msg, writer)
	case logv1.MessageType_ERROR:
		return SerializeErrorText(msg, writer)
	case logv1.MessageType_HTTP:
		return SerializeHTTPRequestText(msg, writer)
	case logv1.MessageType_INFO:
		return SerializeInfoText(msg, writer)
	default:
		return exception.NewFromErr(ErrUnsupportedMessageType).WithMessagef("message type: %v", msg.Type)
	}
}

// SerializeRawText serializes a raw message as text.
func SerializeRawText(msg *logv1.Message, writer *logger.TextWriter) error {
	if msg.Raw == nil {
		return exception.NewFromErr(ErrUnset).WithMessagef("field: Raw")
	}

	return writer.Write(logger.Messagef(rawFlag, string(msg.Raw.Body)).WithTimestamp(UnmarshalTimestamp(msg.Meta.Timestamp)))
}

// SerializeValuesText serializes a values as text.
func SerializeValuesText(msg *logv1.Message, writer *logger.TextWriter) error {
	if msg.Values == nil {
		return exception.NewFromErr(ErrUnset).WithMessagef("field: Values")
	}

	buf := writer.GetBuffer()
	defer writer.PutBuffer(buf)

	for key, value := range msg.Values.Values {
		buf.WriteRune(logger.RuneNewline)
		buf.WriteString(fmt.Sprintf("%s: %s", key, value))
	}

	return writer.Write(logger.Messagef(valuesFlag, buf.String()).WithTimestamp(UnmarshalTimestamp(msg.Meta.Timestamp)))
}

// SerializeErrorText serializes an error as text.
func SerializeErrorText(msg *logv1.Message, writer *logger.TextWriter) error {
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

	return writer.Write(logger.Messagef(logger.Error, buf.String()).WithTimestamp(UnmarshalTimestamp(msg.Meta.Timestamp)))
}

// SerializeHTTPRequestText serializes an http request as text.
func SerializeHTTPRequestText(msg *logv1.Message, writer *logger.TextWriter) error {
	if msg.HttpRequest == nil {
		return exception.NewFromErr(ErrUnset).WithMessagef("field: HttpRequest")
	}

	event := UnmarshalHTTPRequestAsEvent(msg.HttpRequest)
	event.WithTimestamp(UnmarshalTimestamp(msg.Meta.Timestamp))
	buf := writer.GetBuffer()
	defer writer.PutBuffer(buf)
	event.WriteText(writer, buf)
	return nil
}

// SerializeInfoText serializes an info message as text.
func SerializeInfoText(msg *logv1.Message, writer *logger.TextWriter) error {
	if msg.Info == nil {
		return exception.NewFromErr(ErrUnset).WithMessagef("field: Info")
	}
	return writer.Write(logger.Messagef(logger.Flag(msg.Info.Flag), msg.Info.Message).WithTimestamp(UnmarshalTimestamp(msg.Meta.Timestamp)))
}
