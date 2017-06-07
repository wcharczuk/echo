package main

import (
	"log"

	logger "github.com/blendlabs/go-logger"
	web "github.com/blendlabs/go-web"
)

func main() {
	agent := logger.NewFromEnvironment()
	agent.EnableEvent(logger.EventInfo)
	agent.EnableEvent(logger.EventDebug)
	agent.EnableEvent(logger.EventError)
	agent.EnableEvent(logger.EventFatalError)
	agent.EnableEvent(logger.EventWebRequest)
	agent.DisableEvent(logger.EventWebResponse)

	app := web.New()
	app.SetLogger(agent)
	app.GET("/", func(r *web.Ctx) web.Result {
		body := r.Request.URL.Path
		if len(body) == 0 {
			return r.RawWithContentType(web.ContentTypeText, []byte("nada."))
		}
		return r.RawWithContentType(r.Request.Header.Get(web.HeaderContentType), []byte(body))
	})
	app.POST("/", func(r *web.Ctx) web.Result {
		body, err := r.PostBody()
		if err != nil {
			return r.JSON().InternalError(err)
		}
		if len(body) == 0 {
			return r.RawWithContentType(web.ContentTypeText, []byte("nada."))
		}
		return r.RawWithContentType(r.Request.Header.Get(web.HeaderContentType), body)
	})

	log.Fatal(app.Start())
}
