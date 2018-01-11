package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	logger "github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-util/env"
	web "github.com/blendlabs/go-web"
)

func main() {
	agent := logger.All().WithWriter(logger.NewWriterFromEnv())

	appStart := time.Now()

	contents, err := ioutil.ReadFile(env.Env().String("CONFIG_PATH", "/var/secrets/config.yml"))
	if err != nil {
		agent.Warning(err)
	}

	app := web.New().WithLogger(agent)
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
		return r.JSON().Result(env.Env())
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
	app.GET("/config", func(r *web.Ctx) web.Result {
		r.Response.Header().Set("Content-Type", "application/yaml") // but is it really?
		return r.Raw(contents)
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

	agent.SyncFatalExit(app.Start())
}
