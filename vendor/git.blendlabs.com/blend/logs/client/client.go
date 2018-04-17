package client

import (
	"context"
	"net"
	"strings"
	"time"

	// go-sdk
	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/uuid"

	collectorv1 "git.blendlabs.com/blend/protos/collector/v1"
	logv1 "git.blendlabs.com/blend/protos/log/v1"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// New creates a new collector client.
func New(cfg *Config) (*Client, error) {
	var opts []grpc.DialOption
	var err error

	addr := cfg.GetAddr()
	if strings.HasPrefix(addr, "unix://") {
		opts = append(opts,
			grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
				return net.DialTimeout("unix", addr, timeout)
			}))
		addr = strings.TrimPrefix(addr, "unix://")
	}

	if cfg.GetUseTLS() {
		var creds credentials.TransportCredentials
		if len(cfg.GetCAFile()) > 0 {
			creds, err = credentials.NewClientTLSFromFile(cfg.GetCAFile(), cfg.GetServerName())
			if err != nil {
				return nil, exception.Wrap(err)
			}
		} else {
			creds = credentials.NewClientTLSFromCert(nil, cfg.GetServerName())
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, exception.Wrap(err)
	}

	return &Client{
		conn:               conn,
		defaultLabels:      map[string]string{},
		defaultAnnotations: map[string]string{},
		grpcSender:         collectorv1.NewCollectorClient(conn),
	}, nil
}

// Client is a wrapping client for the collector endpoint.
type Client struct {
	conn               *grpc.ClientConn
	defaultLabels      map[string]string
	defaultAnnotations map[string]string
	grpcSender         collectorv1.CollectorClient
}

// Close closes the client.
func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// WithDefaultLabel sets a default label to inject into collected messages.
func (c *Client) WithDefaultLabel(key, value string) *Client {
	c.defaultLabels[key] = value
	return c
}

// WithDefaultAnnotation sets a default label to inject into collected messages.
func (c *Client) WithDefaultAnnotation(key, value string) *Client {
	c.defaultAnnotations[key] = value
	return c
}

// Send sends a message.
func (c *Client) Send(ctx context.Context, message proto.Message, optionalMeta ...*Meta) (err error) {
	msg, processErr := c.processMessage(message, 0, optionalMeta...)
	if processErr != nil {
		err = exception.Wrap(processErr)
		return
	}

	_, err = c.grpcSender.Send(ctx, msg)
	return
}

// SendMany sends a list of messages.
func (c *Client) SendMany(ctx context.Context, messages []proto.Message, optionalMetas ...*Meta) (err error) {
	stream, openStreamErr := c.grpcSender.SendMany(ctx)
	if openStreamErr != nil {
		err = exception.Wrap(openStreamErr)
		return
	}

	var streamErr error
	for index, msgContents := range messages {
		msg, processErr := c.processMessage(msgContents, index, optionalMetas...)
		if processErr != nil {
			err = exception.Wrap(streamErr)
			return
		}
		streamErr = stream.Send(msg)
		if streamErr != nil {
			err = exception.Wrap(streamErr)
			return
		}
	}

	_, closeErr := stream.CloseAndRecv()
	if closeErr != nil {
		err = exception.Wrap(closeErr)
		return
	}
	return
}

func (c *Client) processMessage(msgContents proto.Message, index int, optionalMetas ...*Meta) (msg *logv1.Message, err error) {
	msg = c.newMessage(msgContents)
	var marshalErr error
	msg.Body, marshalErr = proto.Marshal(msgContents)
	if marshalErr != nil {
		err = exception.Wrap(marshalErr)
		return
	}
	c.injectDefaultLabels(msg)
	c.injectDefaultAnnoations(msg)
	c.injectMessageMeta(msg, c.resolveMessageMeta(index, optionalMetas))
	return
}

func (c *Client) newMessage(msg proto.Message) *logv1.Message {
	return &logv1.Message{
		Meta: &logv1.Meta{
			Timestamp: MarshalTimestamp(time.Now().UTC()),
			Uid:       uuid.V4().String(),
			Type:      proto.MessageName(msg),
		},
	}
}

func (c *Client) injectDefaultLabels(msg *logv1.Message) {
	if msg.Meta.Labels == nil {
		msg.Meta.Labels = map[string]string{}
	}
	for key, value := range c.defaultLabels {
		msg.Meta.Labels[key] = value
	}
}

func (c *Client) injectDefaultAnnoations(msg *logv1.Message) {
	if msg.Meta.Annotations == nil {
		msg.Meta.Annotations = map[string]string{}
	}
	for key, value := range c.defaultAnnotations {
		msg.Meta.Annotations[key] = value
	}
}

func (c *Client) resolveMessageMeta(index int, metas []*Meta) *Meta {
	if len(metas) == 0 || index < 0 {
		return nil
	}

	var meta *Meta
	if len(metas) == 1 {
		meta = metas[0]
	} else if index < len(metas) {
		meta = metas[index]
	}
	return meta
}

func (c *Client) injectMessageMeta(msg *logv1.Message, meta *Meta) {
	if meta == nil {
		return
	}

	if msg.Meta.Labels == nil {
		msg.Meta.Labels = map[string]string{}
	}
	for key, value := range meta.Labels {
		msg.Meta.Labels[key] = value
	}
	if msg.Meta.Annotations == nil {
		msg.Meta.Annotations = map[string]string{}
	}
	for key, value := range meta.Annotations {
		msg.Meta.Annotations[key] = value
	}
}
