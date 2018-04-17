package client

import (
	"context"

	"github.com/blend/go-sdk/logger"
)

// NewLoggerListenerHTTPRequest returns a logger listener.
func NewLoggerListenerHTTPRequest(c *Client) logger.Listener {
	return logger.NewWebRequestEventListener(func(e *logger.WebRequestEvent) {
		c.Send(context.TODO(), NewMessageHTTPRequest(e), &Meta{Labels: e.Labels(), Annotations: e.Annotations()})
	})
}

// NewLoggerListenerInfo returns a logger listener.
func NewLoggerListenerInfo(c *Client) logger.Listener {
	return logger.NewMessageEventListener(func(e *logger.MessageEvent) {
		c.Send(context.TODO(), NewMessageInfo(e), &Meta{Labels: e.Labels(), Annotations: e.Annotations()})
	})
}

// NewLoggerListenerError returns a logger listener.
func NewLoggerListenerError(c *Client) logger.Listener {
	return logger.NewErrorEventListener(func(e *logger.ErrorEvent) {
		c.Send(context.TODO(), NewMessageError(e.Err()), &Meta{Labels: e.Labels(), Annotations: e.Annotations()})
	})
}

// NewLoggerListenerAudit returns a logger listener.
func NewLoggerListenerAudit(c *Client) logger.Listener {
	return logger.NewAuditEventListener(func(e *logger.AuditEvent) {
		c.Send(context.TODO(), NewMessageAudit(e), &Meta{Labels: e.Labels(), Annotations: e.Annotations()})
	})
}

// NewLoggerListenerQuery returns a logger listener.
func NewLoggerListenerQuery(c *Client) logger.Listener {
	return logger.NewQueryEventListener(func(e *logger.QueryEvent) {
		c.Send(context.TODO(), NewQueryEvent(e), &Meta{Labels: e.Labels(), Annotations: e.Annotations()})
	})
}

// AddListeners adds logger listeners that pipe to the log collector.
func AddListeners(agent *logger.Logger, logsCfg *Config) (*Client, error) {
	if !HasCollectorUnixSocket(logsCfg) {
		agent.Infof("Collector socket missing: %s", logsCfg.GetAddr())
		return nil, nil
	}
	logs, err := New(logsCfg)
	if err != nil {
		return nil, err
	}

	agent.Infof("Using log collector: %s", logsCfg.GetAddr())
	for key, value := range logsCfg.GetDefaultLabels() {
		logs.WithDefaultLabel(key, value)
	}

	agent.Listen(logger.WebRequest, LoggerListenerName, NewLoggerListenerHTTPRequest(logs))
	agent.Listen(logger.Audit, LoggerListenerName, NewLoggerListenerAudit(logs))
	agent.Listen(logger.Silly, LoggerListenerName, NewLoggerListenerInfo(logs))
	agent.Listen(logger.Info, LoggerListenerName, NewLoggerListenerInfo(logs))
	agent.Listen(logger.Debug, LoggerListenerName, NewLoggerListenerInfo(logs))
	agent.Listen(logger.Warning, LoggerListenerName, NewLoggerListenerError(logs))
	agent.Listen(logger.Error, LoggerListenerName, NewLoggerListenerError(logs))
	agent.Listen(logger.Fatal, LoggerListenerName, NewLoggerListenerError(logs))
	agent.Listen(logger.Query, LoggerListenerName, NewLoggerListenerQuery(logs))
	return logs, nil
}
