package client

import (
	"context"

	logger "github.com/blendlabs/go-logger"
)

// NewLoggerListenerHTTPRequest returns a logger listener.
func NewLoggerListenerHTTPRequest(c *Client) logger.Listener {
	return logger.NewWebRequestEventListener(func(wr *logger.WebRequestEvent) {
		c.Send(context.TODO(), NewMessageHTTPRequest(wr))
	})
}

// NewLoggerListenerInfo returns a logger listener.
func NewLoggerListenerInfo(c *Client) logger.Listener {
	return logger.NewMessageEventListener(func(me *logger.MessageEvent) {
		c.Send(context.TODO(), NewMessageInfo(me))
	})
}

// NewLoggerListenerError returns a logger listener.
func NewLoggerListenerError(c *Client) logger.Listener {
	return logger.NewErrorEventListener(func(ee *logger.ErrorEvent) {
		c.Send(context.TODO(), NewMessageError(ee.Err()))
	})
}

// NewLoggerListenerAudit returns a logger listener.
func NewLoggerListenerAudit(c *Client) logger.Listener {
	return logger.NewAuditEventListener(func(me *logger.AuditEvent) {
		c.Send(context.TODO(), NewMessageAudit(me))
	})
}

// NewLoggerListenerQuery returns a logger listener.
func NewLoggerListenerQuery(c *Client) logger.Listener {
	return logger.NewQueryEventListener(func(e *logger.QueryEvent) {
		c.Send(context.TODO(), NewQueryEvent(e))
	})
}

// AddListeners adds logger listeners that pipe to the log collector.
func AddListeners(agent *logger.Logger, logsCfg *Config) (*Client, error) {
	if HasCollectorUnixSocket(logsCfg) {
		logs, err := New(logsCfg)
		if err != nil {
			return nil, err
		}

		agent.Infof("Using log collector: %s", logsCfg.GetCollectorAddr())
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
	agent.Infof("Collector socket missing: %s", logsCfg.GetCollectorAddr())
	return nil, nil
}
