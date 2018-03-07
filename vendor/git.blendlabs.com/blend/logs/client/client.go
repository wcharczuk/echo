package client

import (
	"context"
	"net"
	"strings"
	"time"

	collectorv1 "git.blendlabs.com/blend/protos/collector/v1"
	logv1 "git.blendlabs.com/blend/protos/log/v1"
	exception "github.com/blendlabs/go-exception"
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

// Close closes the underlying connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Push sends a batch of messages.
func (c *Client) Push(ctx context.Context, messages []logv1.Message) (rs *collectorv1.ReceiveSummary, err error) {
	stream, openStreamErr := c.grpcSender.Push(ctx)
	if openStreamErr != nil {
		err = exception.Wrap(openStreamErr)
		return
	}

	var streamErr error
	for _, msg := range messages {
		c.injectDefaultLabels(&msg)
		streamErr = stream.Send(&msg)
		if streamErr != nil {
			err = exception.Wrap(streamErr)
			return
		}
	}

	summary, closeErr := stream.CloseAndRecv()
	if closeErr != nil {
		err = exception.Wrap(closeErr)
		return
	}
	rs = summary
	return
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
