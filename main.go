package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	logger "github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-util/env"
	web "github.com/blendlabs/go-web"
)

func main() {
	agent := logger.NewFromEnvironment()

	appStart := time.Now()

	contents, err := ioutil.ReadFile(env.Env().String("CONFIG_PATH", "/var/secrets/config.yml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}

	app := web.New()
	app.SetLogger(agent)
	app.GET("/", func(r *web.Ctx) web.Result {
		return r.Text().Result("echo")
	})
	app.GET("/headers", func(r *web.Ctx) web.Result {
		contents, err := json.MarshalString(r.Header())
		if err != nil {
			return r.View().InternalError(err)
		}
		return r.Text().Result(string(contents))
	})
	app.GET("/status", func(r *web.Ctx) web.Result {
		if time.Since(appStart) > 12*time.Second {
			return r.Text().Result("OK!")
		}
		return r.Text().BadRequest("not ready")
	})
	app.GET("/config", func(r *web.Ctx) web.Result {
		r.Response.Header().Set("Content-Type", "application/yaml") // but is it really?
		return r.Raw(contents)
	})
	app.GET("/long", func(r *web.Ctx) web.Result {
		ticker := time.NewTicker(500 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				{
					fmt.Fprintf(r.Response, "tick\n")
					r.Response.Flush()
				}
			}
		}

		return nil
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

	log.Fatal(app.Start())
}
