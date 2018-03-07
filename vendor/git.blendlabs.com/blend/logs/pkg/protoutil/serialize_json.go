package protoutil

import (
	"encoding/json"
	"time"

	logv1 "git.blendlabs.com/blend/protos/log/v1"
	exception "github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
)

const (
	// ErrUnset is a common error.
	ErrUnset Error = "required field unset"
	// ErrUnsupportedMessageType is a common error.
	ErrUnsupportedMessageType Error = "unsupported message type"
)

// SerializeJSON serializes messages as json.
func SerializeJSON(msg *logv1.Message, enc *json.Encoder) error {
	switch msg.Type {
	case logv1.MessageType_RAW:
		return SerializeRawJSON(msg, enc)
	case logv1.MessageType_VALUES:
		return SerializeValuesJSON(msg, enc)
	case logv1.MessageType_ERROR:
		return SerializeErrorJSON(msg, enc)
	case logv1.MessageType_HTTP:
		return SerializeHTTPRequestJSON(msg, enc)
	case logv1.MessageType_INFO:
		return SerializeInfoJSON(msg, enc)
	default:
		return exception.NewFromErr(ErrUnsupportedMessageType).WithMessagef("message type: %v", msg.Type)
	}
}

// SerializeRawJSON serializes a raw message as json.
func SerializeRawJSON(msg *logv1.Message, enc *json.Encoder) error {
	if msg.Raw == nil {
		return exception.NewFromErr(ErrUnset).WithMessagef("field: Raw")
	}
	finalValues := marshalMetaFields(msg.Meta)
	finalValues["body"] = string(msg.Raw.Body)
	return exception.Wrap(enc.Encode(finalValues))
}

// SerializeValuesJSON serializes a values message as json.
func SerializeValuesJSON(msg *logv1.Message, enc *json.Encoder) error {
	if msg.Values == nil {
		return exception.NewFromErr(ErrUnset).WithMessagef("field: Values")
	}
	finalValues := marshalMetaFields(msg.Meta)
	finalValues["values"] = msg.Values.Values
	return exception.Wrap(enc.Encode(finalValues))
}

// SerializeErrorJSON serializes an error message as json.
func SerializeErrorJSON(msg *logv1.Message, enc *json.Encoder) error {
	if msg.Error == nil {
		return exception.NewFromErr(ErrUnset).WithMessagef("field: Error")
	}
	finalValues := marshalMetaFields(msg.Meta)
	finalValues["class"] = msg.Error.Class
	finalValues["message"] = msg.Error.Message
	finalValues["stack"] = msg.Error.Stack
	finalValues["inner"] = marshalErrorJSON(msg.Error.Inner)
	return exception.Wrap(enc.Encode(finalValues))
}

// SerializeHTTPRequestJSON serializes an http request message as json.
func SerializeHTTPRequestJSON(msg *logv1.Message, enc *json.Encoder) error {
	if msg.HttpRequest == nil {
		return exception.NewFromErr(ErrUnset).WithMessagef("field: Raw")
	}

	event := UnmarshalHTTPRequestAsEvent(msg.HttpRequest)
	finalValues := event.WriteJSON()
	finalValues = mergeValues(finalValues, marshalMetaFields(msg.Meta))
	return exception.Wrap(enc.Encode(finalValues))
}

// SerializeInfoJSON serializes an info message as json.
func SerializeInfoJSON(msg *logv1.Message, enc *json.Encoder) error {
	if msg.Info == nil {
		return exception.NewFromErr(ErrUnset).WithMessagef("field: Info")
	}
	finalValues := marshalMetaFields(msg.Meta)
	finalValues[logger.JSONFieldMessage] = msg.Info.Message
	finalValues[logger.JSONFieldFlag] = msg.Info.Flag
	return exception.Wrap(enc.Encode(finalValues))
}

func mergeValues(a, b map[string]interface{}) map[string]interface{} {
	for key, value := range b {
		a[key] = value
	}
	return a
}

func marshalMetaFields(meta *logv1.Meta) map[string]interface{} {
	values := map[string]interface{}{}
	values[JSONFieldUID] = meta.Uid
	values[JSONFieldTimestamp] = UnmarshalTimestamp(meta.Timestamp).Format(time.RFC3339)
	values[JSONFieldLabels] = meta.Labels
	return values
}

// marshalErrorJSON marshals an error as json values.
func marshalErrorJSON(err *logv1.Error) map[string]interface{} {
	if err == nil {
		return nil
	}
	return map[string]interface{}{
		"class":   err.Class,
		"message": err.Message,
		"stack":   err.Stack,
		"inner":   marshalErrorJSON(err.Inner),
	}
}
