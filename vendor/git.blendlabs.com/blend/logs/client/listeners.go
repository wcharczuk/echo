package client

import (
	"context"

	logv1 "git.blendlabs.com/blend/protos/log/v1"
	logger "github.com/blendlabs/go-logger"
)

// CreateLoggerListenerHTTPRequest returns a logger listener.
func CreateLoggerListenerHTTPRequest(c *Client) logger.Listener {
	return logger.NewWebRequestEventListener(func(wr *logger.WebRequestEvent) {
		c.Push(context.TODO(), []logv1.Message{
			NewMessageHTTPRequest(wr),
		})
	})
}

// CreateLoggerListenerInfo returns a logger listener.
func CreateLoggerListenerInfo(c *Client) logger.Listener {
	return logger.NewMessageEventListener(func(me *logger.MessageEvent) {
		c.Push(context.TODO(), []logv1.Message{
			NewMessageInfo(me),
		})
	})
}

// CreateLoggerListenerError returns a logger listener.
func CreateLoggerListenerError(c *Client) logger.Listener {
	return logger.NewErrorEventListener(func(ee *logger.ErrorEvent) {
		c.Push(context.TODO(), []logv1.Message{
			NewMessageError(ee),
		})
	})
}
