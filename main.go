package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	logs "git.blendlabs.com/blend/logs/client"
	exception "github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-util/env"
	web "github.com/blendlabs/go-web"
)

func main() {
	agent := logger.All()

	appStart := time.Now()

	app := web.NewFromConfig(web.NewConfigFromEnv()).WithLogger(agent)
	app.GET("/", func(r *web.Ctx) web.Result {
		return r.Text().Result("echo")
	})
	app.GET("/headers", func(r *web.Ctx) web.Result {
		contents, err := json.Marshal(r.Request.Header)
		if err != nil {
			return r.View().InternalError(err)
		}
		return r.Text().Result(string(contents))
	})
	app.GET("/env", func(r *web.Ctx) web.Result {
		return r.JSON().Result(env.Env().Vars())
	})
	app.GET("/error", func(r *web.Ctx) web.Result {
		return r.JSON().InternalError(exception.Newf("This is only a test").WithMessagef("this is a message").WithInner(exception.Newf("inner exception")))
	})
	app.GET("/proxy/*filepath", func(r *web.Ctx) web.Result {
		return r.JSON().Result("OK!")
	})
	app.GET("/status", func(r *web.Ctx) web.Result {
		if time.Since(appStart) > 12*time.Second {
			return r.Text().Result("OK!")
		}
		return r.Text().BadRequest(fmt.Errorf("not ready"))
	})
	app.GET("/long/:seconds", func(r *web.Ctx) web.Result {
		seconds, err := r.RouteParamInt("seconds")
		if err != nil {
			return r.Text().BadRequest(err)
		}

		timeout := time.After(time.Duration(seconds) * time.Second)
		ticker := time.NewTicker(500 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				{
					fmt.Fprintf(r.Response, "tick\n")
				}
			case <-timeout:
				{
					return r.Raw([]byte("timeout\n"))
				}
			}
		}
	})
	app.GET("/echo/*filepath", func(r *web.Ctx) web.Result {
		body := r.Request.URL.Path
		if len(body) == 0 {
			return r.RawWithContentType(web.ContentTypeText, []byte("no response."))
		}
		return r.RawWithContentType(web.ContentTypeText, []byte(body))
	})
	app.POST("/echo/*filepath", func(r *web.Ctx) web.Result {
		body, err := r.PostBody()
		if err != nil {
			return r.JSON().InternalError(err)
		}
		if len(body) == 0 {
			return r.RawWithContentType(web.ContentTypeText, []byte("nada."))
		}
		return r.RawWithContentType(web.ContentTypeText, body)
	})

	logsCfg := logs.NewConfigFromEnv()
	if logs.HasUnixSocket(logsCfg) {
		logsClient, err := logs.New(logs.NewConfigFromEnv())
		if err != nil {
			agent.SyncFatalExit(err)
		}
		defer logsClient.Close()
		logsClient.WithLogger(agent)

		agent.Infof("Using log collector: %s", logsCfg.GetAddr())
		logsClient.WithDefaultLabel("service", "echo-private")
		logsClient.WithDefaultLabel("service-pod", env.Env().String("HOSTNAME"))
		agent.Listen(web.FlagAppStart, "log-collector", web.NewAppStartEventListener(func(was web.AppStartEvent) {
			logsClient.Send(context.TODO(), logs.NewMessageInfo(logger.Messagef(was.Flag(), was.String())))
		}))
		agent.Listen(web.FlagAppStartComplete, "log-collector", web.NewAppStartCompleteEventListener(func(was web.AppStartCompleteEvent) {
			logsClient.Send(context.TODO(), logs.NewMessageInfo(logger.Messagef(was.Flag(), was.String())))
		}))
		agent.Listen(web.FlagAppExit, "log-collector", web.NewAppExitEventListener(func(was web.AppExitEvent) {
			logsClient.Send(context.TODO(), logs.NewMessageInfo(logger.Messagef(was.Flag(), was.String())))
		}))
		agent.Listen(logger.WebRequest, "log-collector", logs.CreateLoggerListenerHTTPRequest(logsClient))
		agent.Listen(logger.Silly, "log-collector", logs.CreateLoggerListenerInfo(logsClient))
		agent.Listen(logger.Info, "log-collector", logs.CreateLoggerListenerInfo(logsClient))
		agent.Listen(logger.Debug, "log-collector", logs.CreateLoggerListenerInfo(logsClient))
		agent.Listen(logger.Warning, "log-collector", logs.CreateLoggerListenerError(logsClient))
		agent.Listen(logger.Error, "log-collector", logs.CreateLoggerListenerError(logsClient))
		agent.Listen(logger.Fatal, "log-collector", logs.CreateLoggerListenerError(logsClient))
	} else {
		agent.Infof("Collector socket missing: %s", logsCfg.GetAddr())
	}

	agent.SyncFatalExit(app.Start())
}
