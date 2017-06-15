package main

import (
	"log"
	"time"

	logger "github.com/blendlabs/go-logger"
	web "github.com/blendlabs/go-web"
)

func main() {
	agent := logger.NewFromEnvironment()

	appStart := time.Now()

	app := web.New()
	app.SetLogger(agent)
	app.GET("/", func(r *web.Ctx) web.Result {
		return r.Text().Result("echo")
	})
	app.GET("/status", func(r *web.Ctx) web.Result {
		if time.Since(appStart) > 10*time.Second {
			return r.Text().Result("OK!")
		}
		return r.Text().BadRequest("not ready")
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
