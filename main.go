package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/blend/go-sdk/env"
	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/web"
)

func main() {
	log := logger.All()

	appStart := time.Now()

	app := web.NewFromConfig(web.NewConfigFromEnv()).WithLogger(log)
	app.GET("/", func(r *web.Ctx) web.Result {
		return r.Text().Result("echo")
	})
	app.GET("/headers", func(r *web.Ctx) web.Result {
		contents, err := json.Marshal(r.Request().Header)
		if err != nil {
			return r.View().InternalError(err)
		}
		return r.Text().Result(string(contents))
	})
	app.GET("/env", func(r *web.Ctx) web.Result {
		return r.JSON().Result(env.Env().Vars())
	})
	app.GET("/error", func(r *web.Ctx) web.Result {
		return r.JSON().InternalError(exception.New("This is only a test").WithMessagef("this is a message").WithInner(exception.New("inner exception")))
	})
	app.GET("/proxy/*filepath", func(r *web.Ctx) web.Result {
		return r.JSON().Result("OK!")
	})
	app.GET("/status", func(r *web.Ctx) web.Result {
		if time.Since(appStart) > 12*time.Second {
			return r.Text().Result("OK!")
		}
		return r.Text().InternalError(fmt.Errorf("not ready"))
	})

	app.GET("/long/:seconds", func(r *web.Ctx) web.Result {
		seconds, err := r.RouteParamInt("seconds")
		if err != nil {
			return r.Text().BadRequest(err)
		}

		r.Response().WriteHeader(http.StatusOK)
		timeout := time.After(time.Duration(seconds) * time.Second)
		ticker := time.NewTicker(500 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				{
					fmt.Fprintf(r.Response(), "tick\n")
					r.Response().InnerResponse().(http.Flusher).Flush()
				}
			case <-timeout:
				{
					fmt.Fprintf(r.Response(), "timeout\n")
					r.Response().InnerResponse().(http.Flusher).Flush()
					return nil
				}
			}
		}
	})

	app.GET("/echo/*filepath", func(r *web.Ctx) web.Result {
		body := r.Request().URL.Path
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

	app.WithMethodNotAllowedHandler(func(r *web.Ctx) web.Result {
		log.Infof("headers: %#v", r.Request().Header)
		body, _ := r.PostBodyAsString()
		log.Infof("body: %s", body)
		return r.JSON().OK()
	})

	app.WithNotFoundHandler(func(r *web.Ctx) web.Result {
		log.Infof("headers: %#v", r.Request().Header)
		body, _ := r.PostBodyAsString()
		log.Infof("body: %s", body)
		return r.JSON().OK()
	})

	web.GracefulShutdown(app)
}
