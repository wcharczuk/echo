package collector

import (
	"strconv"

	"git.blendlabs.com/blend/logs/pkg/config"
	logv1 "git.blendlabs.com/blend/protos/log/v1"
	exception "github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-util/collections"
	"github.com/golang/protobuf/proto"
	kinesis "github.com/sendgridlabs/go-kinesis"
)

const (
	// MaxKinesisRetries is the maximum number of times to retry a message.
	MaxKinesisRetries = 5
)

// NewKinesisFlusher returns a new kinsesis flusher.
func NewKinesisFlusher(cfg *config.Collector) *KinesisFlusher {
	auth := kinesis.NewAuth(cfg.Aws.GetAccessKeyID(), cfg.Aws.GetSecretAccessKey(), cfg.Aws.GetToken())
	handler := kinesis.New(auth, cfg.Aws.GetRegion())

	return &KinesisFlusher{
		streamName: cfg.GetStreamName(),
		handler:    handler,
	}
}

// KinesisFlusher writes messages to a given kinesis stream.
// It does not currently support retries, so some messages may fail to send.
type KinesisFlusher struct {
	retryBuffer *collections.AutoflushBuffer
	log         *logger.Logger
	streamName  string
	handler     *kinesis.Kinesis
}

// WithRetryBuffer sets the retry buffer and returns a reference to the flusher.
func (kf *KinesisFlusher) WithRetryBuffer(buffer *collections.AutoflushBuffer) *KinesisFlusher {
	kf.retryBuffer = buffer
	return kf
}

// RetryBuffer returns the retry buffer.
func (kf *KinesisFlusher) RetryBuffer() *collections.AutoflushBuffer {
	return kf.retryBuffer
}

// WithLogger sets the kinesis flusher logger.
func (kf *KinesisFlusher) WithLogger(log *logger.Logger) *KinesisFlusher {
	kf.log = log
	return kf
}

// Logger returns the logger.
func (kf *KinesisFlusher) Logger() *logger.Logger {
	return kf.log
}

// SendMany sends many proto messages.
func (kf *KinesisFlusher) SendMany(messages []Any) {
	args := kinesis.NewArgs()
	args.Add("StreamName", kf.streamName)

	for _, msg := range messages {
		if typed, isTyped := msg.(*logv1.Message); isTyped {
			contents, err := proto.Marshal(typed)
			if err != nil {
				kf.err(err)
				continue
			}
			if len(contents) == 0 {
				kf.err(exception.New("empty contents for protobuf message"))
				continue
			}
			args.AddRecord(contents, typed.Type.String())
		} else {
			kf.err(exception.New("message is not a protobuf message").WithMessagef("%T", msg))
		}
	}

	res, err := kf.handler.PutRecords(args)
	if err != nil {
		kf.err(exception.Wrap(err))
	}

	if res != nil {
		kf.infof("flushed %d messages to `%s`, failed: %d", len(messages), kf.streamName, res.FailedRecordCount)
		if kf.retryBuffer != nil {
			for index, resMsg := range res.Records {
				if len(resMsg.ErrorCode) > 0 {
					msg := messages[index]
					if typed, isTyped := msg.(*logv1.Message); isTyped {
						kf.debugf("requeing %s.%s due to %s", typed.Type.String(), typed.Meta.Uid, resMsg.ErrorMessage)
						kf.incrementRetryCount(typed)
						if kf.isBelowMaxRetries(typed) {
							kf.retryBuffer.Add(msg)
						}
					}
				}
			}
		}
	}
}

func (kf *KinesisFlusher) isBelowMaxRetries(msg *logv1.Message) bool {
	if msg == nil {
		return false
	}
	if value, hasValue := msg.Meta.Annotations["collector-retries"]; hasValue {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return false
		}
		return parsed < MaxKinesisRetries
	}
	return true
}

func (kf *KinesisFlusher) incrementRetryCount(msg *logv1.Message) {
	if msg == nil {
		return
	}
	if value, hasValue := msg.Meta.Annotations["collector-retries"]; hasValue {
		parsed, _ := strconv.Atoi(value)
		msg.Meta.Annotations["collector-retries"] = strconv.Itoa(parsed + 1)
	} else {
		msg.Meta.Annotations["collector-retries"] = "1"
	}
}

func (kf *KinesisFlusher) infof(format string, args ...Any) {
	if kf.log != nil {
		kf.log.Infof(format, args...)
	}
}

func (kf *KinesisFlusher) debugf(format string, args ...Any) {
	if kf.log != nil {
		kf.log.Debugf(format, args...)
	}
}

func (kf *KinesisFlusher) err(err error) {
	if kf.log != nil {
		kf.log.Error(err)
	}
}
