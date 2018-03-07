package client

import (
	"context"

	logger "github.com/blendlabs/go-logger"
)

// CreateLoggerListenerHTTPRequest returns a logger listener.
func CreateLoggerListenerHTTPRequest(c *Client) logger.Listener {
	return logger.NewWebRequestEventListener(func(wr *logger.WebRequestEvent) {
		c.Send(context.TODO(), NewMessageHTTPRequest(wr))
	})
}

// CreateLoggerListenerInfo returns a logger listener.
func CreateLoggerListenerInfo(c *Client) logger.Listener {
	return logger.NewMessageEventListener(func(me *logger.MessageEvent) {
		c.Send(context.TODO(), NewMessageInfo(me))
	})
}

// CreateLoggerListenerError returns a logger listener.
func CreateLoggerListenerError(c *Client) logger.Listener {
	return logger.NewErrorEventListener(func(ee *logger.ErrorEvent) {
		c.Send(context.TODO(), NewMessageError(ee))
	})
}
