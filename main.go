package main

import (
	"io/ioutil"
	"log"
	"time"

	exception "github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-util/env"
	web "github.com/blendlabs/go-web"
)

func main() {
	agent := logger.NewFromEnvironment()

	appStart := time.Now()

	contents, err := ioutil.ReadFile(env.Env().String("CONFIG_PATH", "/var/secrets/config.yml"))
	if err != nil {
		log.Fatal(exception.New(err))
	}

	app := web.New()
	app.SetLogger(agent)
	app.GET("/", func(r *web.Ctx) web.Result {
		var x int
		for i := 0; i < 1<<20; i++ {
			x++
		}
		return r.Text().Result("echo")
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
	app.GET("/aws/config", func(r *web.Ctx) web.Result {
		contents, err := ioutil.ReadFile(env.Env().String("AWS_PATH_CONFIG", "/root/.aws/config"))
		if err != nil {
			return r.JSON().InternalError(err)
		}
		r.Response.Header().Set("Content-Type", "application/yaml")
		return r.Raw(contents)
	})
	app.GET("/aws/lease", func(r *web.Ctx) web.Result {
		contents, err := ioutil.ReadFile(env.Env().String("AWS_PATH_LEASE", "/root/.aws/lease"))
		if err != nil {
			return r.JSON().InternalError(err)
		}
		r.Response.Header().Set("Content-Type", "application/yaml")
		return r.Raw(contents)
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

	app.GET("/stat/secrets", func(r *web.Ctx) web.Result {
		files, err := ioutil.ReadDir("/var/secrets-agent")
		if err != nil {
			return r.JSON().InternalError(err)
		}
		allFiles := ""

		for _, file := range files {
			allFiles += file.Name()
		}

		return r.RawWithContentType(web.ContentTypeText, []byte(allFiles))
	})

	log.Fatal(app.Start())
}
