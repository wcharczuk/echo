package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/blend/go-sdk/env"
	"github.com/blend/go-sdk/ex"
	"github.com/blend/go-sdk/graceful"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/web"
)

func main() {
	log := logger.All(logger.OptPath("echo"))

	appStart := time.Now()

	app := web.New(web.OptConfigFromEnv(), web.OptLog(log))
	app.GET("/", func(r *web.Ctx) web.Result {
		return web.Text.Result("echo")
	})
	app.GET("/headers", func(r *web.Ctx) web.Result {
		contents, err := json.Marshal(r.Request.Header)
		if err != nil {
			return r.Views.InternalError(err)
		}
		return web.Text.Result(string(contents))
	})
	app.GET("/env", func(r *web.Ctx) web.Result {
		return web.JSON.Result(env.Env().Vars())
	})
	app.GET("/error", func(r *web.Ctx) web.Result {
		return web.JSON.InternalError(ex.New("This is only a test", ex.OptMessagef("this is a message"), ex.OptInner(ex.New("inner exception"))))
	})
	app.GET("/status", func(r *web.Ctx) web.Result {
		if time.Since(appStart) > 12*time.Second {
			return web.Text.Result("OK!")
		}
		return web.Text.InternalError(fmt.Errorf("not ready"))
	})

	app.GET("/long/:seconds", func(r *web.Ctx) web.Result {
		seconds, err := web.IntValue(r.RouteParam("seconds"))
		if err != nil {
			return web.Text.BadRequest(err)
		}

		start := time.Now()
		r.Response.WriteHeader(http.StatusOK)
		timeout := time.After(time.Duration(seconds) * time.Second)
		ticker := time.NewTicker(500 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				{
					fmt.Fprintf(r.Response, "%v tick\n", time.Since(start))
					r.Response.Flush()
				}
			case <-timeout:
				{
					fmt.Fprintf(r.Response, "timeout\n")
					r.Response.Flush()
					return nil
				}
			}
		}
	})

	if err := graceful.Shutdown(app); err != nil {
		logger.FatalExit(err)
	}
}
