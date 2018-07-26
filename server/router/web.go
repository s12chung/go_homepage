package router

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/s12chung/go_homepage/settings"
	"github.com/s12chung/go_homepage/view"
)

type WebContext struct {
	renderer *view.Renderer
	settings *settings.Settings
	log      logrus.FieldLogger
	urlParts []string

	url            string
	responseWriter http.ResponseWriter
	request        *http.Request
}

func (ctx *WebContext) Settings() *settings.Settings {
	return ctx.settings
}

func (ctx *WebContext) Log() logrus.FieldLogger {
	return ctx.log
}

func (ctx *WebContext) UrlParts() []string {
	return ctx.urlParts
}

func (ctx *WebContext) Url() string {
	return ctx.url
}

func (ctx *WebContext) Render(name string, data interface{}) error {
	ctx.Log().Infof("Rendering template: %v", ctx.Url())

	bytes, err := ctx.renderer.Render("index", data)
	if err != nil {
		return err
	}
	_, err = ctx.responseWriter.Write(bytes)
	return err
}

type WebRouter struct {
	defaultContext *WebContext
	serveMux       *http.ServeMux
	log            logrus.FieldLogger

	arounds []func(ctx *WebContext, handler func(ctx *WebContext) error) error

	rootHandler     func(w http.ResponseWriter, r *http.Request)
	wildcardHandler func(w http.ResponseWriter, r *http.Request)
}

func NewWebRouter(renderer *view.Renderer, settings *settings.Settings, log logrus.FieldLogger) *WebRouter {
	defaultContext := WebContext{
		renderer,
		settings,
		log,
		nil,
		"",
		nil,
		nil,
	}

	defaultHandler := func(w http.ResponseWriter, r *http.Request) {
		s := fmt.Sprintf("%v is not being handled", r.URL)
		log.Errorf(s)
		http.Error(w, s, http.StatusBadRequest)
		return
	}

	router := &WebRouter{
		&defaultContext,
		http.DefaultServeMux,
		log,
		nil,
		defaultHandler,
		defaultHandler,
	}
	router.handleWildcard()
	return router
}

func (router *WebRouter) Around(handler func(ctx *WebContext, handler func(ctx *WebContext) error) error) {
	router.arounds = append(router.arounds, handler)
}

func (router *WebRouter) htmlHandler(handler func(ctx *WebContext) error) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Content-Type", mime.TypeByExtension(".html"))

		ctx := *router.defaultContext
		ctx.responseWriter = w
		ctx.request = r
		ctx.url = r.URL.String()

		arounds := router.arounds
		aroundToNext := make([]func(ctx *WebContext) error, len(arounds))
		for index := range arounds {
			reverseIndex := len(arounds) - 1 - index
			around := arounds[reverseIndex]
			if index == 0 {
				aroundToNext[reverseIndex] = func(ctx *WebContext) error {
					return around(ctx, handler)
				}
			} else {
				aroundToNext[reverseIndex] = func(ctx *WebContext) error {
					return around(ctx, aroundToNext[reverseIndex+1])
				}
			}
		}
		return aroundToNext[0](&ctx)
	}
}

func (router *WebRouter) getRequestHandler(handler func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			err := handler(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		}
	}
}

func (router *WebRouter) handleWildcard() {
	router.serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/" {
			router.rootHandler(w, r)
		} else {
			router.wildcardHandler(w, r)
		}
	})
}

func (router *WebRouter) FileServe(pattern, dirPath string) {
	router.Get(pattern, func(w http.ResponseWriter, r *http.Request) error {
		regex := regexp.MustCompile(strings.Replace(`^/`+pattern+`/`, "//", "/", -1))
		assetFilePath := path.Join(dirPath, regex.ReplaceAllString(r.URL.String(), ""))

		file, err := os.Open(assetFilePath)
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(assetFilePath)))
		_, err = io.Copy(w, file)
		return err
	})
}

func (router *WebRouter) GetWildcardHTML(handler func(ctx *WebContext) error) {
	router.wildcardHandler = router.getRequestHandler(router.htmlHandler(func(ctx *WebContext) error {
		url := ctx.request.URL.String()

		var parts []string
		for _, part := range strings.Split(url, "/") {
			if part != "" {
				parts = append(parts, part)
			}
		}
		if len(parts) > 1 {
			return fmt.Errorf("currently can't handle more than 1 UrlPart - %v", url)
		}

		ctx.urlParts = parts
		return handler(ctx)
	}))
}

func (router *WebRouter) GetRootHTML(handler func(ctx *WebContext) error) {
	router.rootHandler = router.getRequestHandler(router.htmlHandler(handler))
}

func (router *WebRouter) GetHTML(pattern string, handler func(ctx *WebContext) error) {
	router.Get(pattern, router.htmlHandler(handler))
}

func (router *WebRouter) Get(pattern string, handler func(w http.ResponseWriter, r *http.Request) error) {
	if pattern == "/" {
		router.log.Errorf("Can not use pattern that touches root, use GetRootHTML or GetWildcardHTML instead")
		return
	}

	router.serveMux.HandleFunc(pattern, router.getRequestHandler(handler))
}

func (router *WebRouter) Run(port int) error {
	router.log.Infof("Running server at http://localhost:%v/", port)
	return http.ListenAndServe(":"+strconv.Itoa(port), router.serveMux)
}
