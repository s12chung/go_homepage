package server

import (
	"net/http"
	"strconv"

	"fmt"
	"github.com/Sirupsen/logrus"
	"io"
	"mime"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/s12chung/go_homepage/settings"
	"github.com/s12chung/go_homepage/view"
)

func RunFileServer(targetDir string, port int, log logrus.FieldLogger) error {
	log.Infof("Serving files from '%v' at http://localhost:%v/", targetDir, port)
	handler := http.FileServer(http.Dir(targetDir))
	return http.ListenAndServe(":"+strconv.Itoa(port), handler)
}

type Context struct {
	*view.Renderer
	*settings.Settings
	Log            logrus.FieldLogger
	Query          map[string][]string
	responseWriter http.ResponseWriter
}

func (ctx *Context) Write(bytes []byte) error {
	_, err := ctx.responseWriter.Write(bytes)
	return err
}

type Router struct {
	defaultContext  *Context
	serveMux        *http.ServeMux
	log             logrus.FieldLogger
	rootHandler     func(w http.ResponseWriter, r *http.Request)
	wildcardHandler func(w http.ResponseWriter, r *http.Request)
}

func NewRouter(renderer *view.Renderer, settings *settings.Settings, log logrus.FieldLogger) *Router {
	defaultContext := Context{
		renderer,
		settings,
		log,
		nil,
		nil,
	}

	defaultHandler := func(w http.ResponseWriter, r *http.Request) {
		s := fmt.Sprintf("%v %v is not being handled", r.Method, r.URL)
		log.Errorf(s)
		http.Error(w, s, http.StatusBadRequest)
		return
	}

	router := &Router{
		&defaultContext,
		http.DefaultServeMux,
		log,
		defaultHandler,
		defaultHandler,
	}
	router.handleWildcard()
	return router
}

func (router *Router) htmlHandler(handler func(ctx *Context) error) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Content-Type", mime.TypeByExtension(".html"))

		ctx := *router.defaultContext
		ctx.Query = r.URL.Query()
		ctx.responseWriter = w
		return handler(&ctx)
	}
}

func (router *Router) getRequestHandler(handler func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			err := handler(w, r)
			if err != nil {
				router.log.Warnf("Server Error for %v %v: %v", r.Method, r.URL, err)
				http.Error(w, err.Error(), http.StatusBadRequest)
			} else {
				router.log.Infof("Success for %v %v", r.Method, r.URL)
			}
		}
	}
}

func (router *Router) handleWildcard() {
	router.serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/" {
			router.rootHandler(w, r)
		} else {
			router.wildcardHandler(w, r)
		}
	})
}

func (router *Router) FileServe(pattern, dirPath string) {
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

func (router *Router) GetWildcardHTML(handler func(ctx *Context) error) {
	router.wildcardHandler = router.getRequestHandler(router.htmlHandler(handler))
}

func (router *Router) GetRootHTML(handler func(ctx *Context) error) {
	router.rootHandler = router.getRequestHandler(router.htmlHandler(handler))
}

func (router *Router) GetHTML(pattern string, handler func(ctx *Context) error) {
	router.Get(pattern, router.htmlHandler(handler))
}

func (router *Router) Get(pattern string, handler func(w http.ResponseWriter, r *http.Request) error) {
	if pattern == "/" {
		router.log.Errorf("Can not use pattern that touches root, use GetRootHTML or GetWildcardHTML instead")
		return
	}

	router.serveMux.HandleFunc(pattern, router.getRequestHandler(handler))
}

func (router *Router) Run(port int) error {
	router.log.Infof("Running server at http://localhost:%v/", port)
	return http.ListenAndServe(":"+strconv.Itoa(port), router.serveMux)
}
