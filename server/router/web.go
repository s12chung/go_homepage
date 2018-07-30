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
	"io/ioutil"
)

//
// Context
//
type WebContext struct {
	r        *view.Renderer
	settings *settings.Settings
	log      logrus.FieldLogger

	url      string
	urlParts []string
	tmplName string

	responseWriter http.ResponseWriter
	request        *http.Request
}

func (ctx *WebContext) renderer() *view.Renderer {
	return ctx.r
}

func (ctx *WebContext) Settings() *settings.Settings {
	return ctx.settings
}

func (ctx *WebContext) Log() logrus.FieldLogger {
	return ctx.log
}

func (ctx *WebContext) SetLog(log logrus.FieldLogger) {
	ctx.log = log
}

func (ctx *WebContext) UrlParts() []string {
	return ctx.urlParts
}

func (ctx *WebContext) Url() string {
	return ctx.url
}

func (ctx *WebContext) TemplateName() string {
	return templateName(ctx, ctx.tmplName)
}

func (ctx *WebContext) SetTemplateName(templateName string) {
	ctx.tmplName = templateName
}

func (ctx *WebContext) Render(data interface{}) error {
	bytes, err := renderTemplate(ctx, data)
	if err != nil {
		return err
	}
	_, err = ctx.responseWriter.Write(bytes)
	return err
}

//
// Router
//
type WebRouter struct {
	defaultContext *WebContext
	serveMux       *http.ServeMux
	log            logrus.FieldLogger

	arounds []func(ctx Context, handler func(ctx Context) error) error
	routes  map[string]bool

	rootHandler     func(w http.ResponseWriter, r *http.Request)
	wildcardHandler func(w http.ResponseWriter, r *http.Request)
}

func NewWebRouter(renderer *view.Renderer, settings *settings.Settings, log logrus.FieldLogger) *WebRouter {
	defaultContext := &WebContext{
		r:        renderer,
		settings: settings,
		log:      log,
	}

	defaultHandler := func(w http.ResponseWriter, r *http.Request) {
		s := fmt.Sprintf("%v is not being handled", r.URL)
		log.Errorf(s)
		http.Error(w, s, http.StatusBadRequest)
		return
	}

	router := &WebRouter{
		defaultContext,
		http.DefaultServeMux,
		log,
		nil,
		make(map[string]bool),
		defaultHandler,
		defaultHandler,
	}
	router.handleWildcard()
	return router
}

func (router *WebRouter) Around(handler func(ctx Context, handler func(ctx Context) error) error) {
	router.arounds = append(router.arounds, handler)
}

func (router *WebRouter) GetWildcardHTML(handler func(ctx Context) error) {
	router.checkAndSetRoutes(WildcardUrlPattern)
	router.wildcardHandler = router.getRequestHandler(router.htmlHandler(handler))
}

func (router *WebRouter) GetRootHTML(handler func(ctx Context) error) {
	router.checkAndSetRoutes(RootUrlPattern)
	router.rootHandler = router.getRequestHandler(router.htmlHandler(handler))
}

func (router *WebRouter) GetHTML(pattern string, handler func(ctx Context) error) {
	router.checkAndSetRoutes(pattern)
	router.Get(pattern, router.htmlHandler(handler))
}

func (router *WebRouter) checkAndSetRoutes(pattern string) error {
	_, has := router.routes[pattern]
	if has {
		panicDuplicateRoute(pattern)
	}
	router.routes[pattern] = true
	return nil
}

func (router *WebRouter) StaticRoutes() []string {
	var staticRoutes []string
	for k := range router.routes {
		if k != WildcardUrlPattern {
			staticRoutes = append(staticRoutes, k)
		}
	}
	return staticRoutes
}

func (router *WebRouter) Requester() Requester {
	return newWebRequester(router.defaultContext.settings)
}

func (router *WebRouter) htmlHandler(handler func(ctx Context) error) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Content-Type", mime.TypeByExtension(".html"))

		ctx := *router.defaultContext
		ctx.url = r.URL.String()
		parts, err := urlParts(ctx.url)
		if err != nil {
			return err
		}
		ctx.urlParts = parts

		ctx.responseWriter = w
		ctx.request = r

		return callArounds(router.arounds, handler, &ctx)
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
	router.serveMux.HandleFunc(RootUrlPattern, func(w http.ResponseWriter, r *http.Request) {
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

func (router *WebRouter) Get(pattern string, handler func(w http.ResponseWriter, r *http.Request) error) {
	if pattern == RootUrlPattern {
		router.log.Errorf("Can not use pattern that touches root, use GetRootHTML or GetWildcardHTML instead")
		return
	}

	router.serveMux.HandleFunc(pattern, router.getRequestHandler(handler))
}

func (router *WebRouter) Run(port int) error {
	router.log.Infof("Running server at http://localhost:%v/", port)
	return http.ListenAndServe(":"+strconv.Itoa(port), router.serveMux)
}

//
// Requester
//
type WebRequester struct {
	host string
	port int
}

func newWebRequester(s *settings.Settings) *WebRequester {
	return &WebRequester{
		"localhost",
		s.ServerPort,
	}
}

func (requester *WebRequester) Get(url string) ([]byte, error) {
	response, err := http.Get(fmt.Sprintf("http://%v:%v%v", requester.host, requester.port, url))
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	return ioutil.ReadAll(response.Body)

}
