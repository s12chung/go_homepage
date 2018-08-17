package router

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type WebHandler func(w http.ResponseWriter, r *http.Request) error

//
// Context
//
type WebContext struct {
	log logrus.FieldLogger

	url      string
	urlParts []string

	responseWriter http.ResponseWriter
	request        *http.Request
}

func NewWebContext(log logrus.FieldLogger) *WebContext {
	return &WebContext{log: log}
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

func (ctx *WebContext) Respond(bytes []byte) error {
	_, err := ctx.responseWriter.Write(bytes)
	return err
}

//
// Router
//
type WebRouter struct {
	serveMux *http.ServeMux
	log      logrus.FieldLogger

	arounds []AroundHandler
	routes  map[string]bool

	rootHandler     http.HandlerFunc
	wildcardHandler http.HandlerFunc

	port int
}

func NewWebRouter(port int, log logrus.FieldLogger) *WebRouter {
	defaultHandler := func(w http.ResponseWriter, r *http.Request) {
		s := fmt.Sprintf("%v is not being handled", r.URL)
		log.Errorf(s)
		http.Error(w, s, http.StatusBadRequest)
		return
	}

	router := &WebRouter{
		http.NewServeMux(),
		log,
		nil,
		make(map[string]bool),
		defaultHandler,
		defaultHandler,
		port,
	}
	router.handleWildcard()
	return router
}

func (router *WebRouter) Around(handler AroundHandler) {
	router.arounds = append(router.arounds, handler)
}

func (router *WebRouter) GetWildcardHTML(handler ContextHandler) {
	router.checkAndSetRoutes(WildcardUrlPattern)
	router.wildcardHandler = router.getRequestHandler(router.htmlHandler(handler))
}

func (router *WebRouter) GetRootHTML(handler ContextHandler) {
	router.checkAndSetRoutes(RootUrlPattern)
	router.rootHandler = router.getRequestHandler(router.htmlHandler(handler))
}

func (router *WebRouter) GetHTML(pattern string, handler ContextHandler) {
	router.checkAndSetRoutes(pattern)
	router.get(pattern, router.htmlHandler(handler))
}

func (router *WebRouter) Get(pattern, mimeType string, handler ContextHandler) {
	router.checkAndSetRoutes(pattern)
	router.get(pattern, router.handler(mimeType, handler))
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
	return newWebRequester(router.port)
}

func (router *WebRouter) htmlHandler(handler ContextHandler) WebHandler {
	return router.handler(mime.TypeByExtension(".html"), handler)
}

func (router *WebRouter) handler(mimeType string, handler ContextHandler) WebHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Content-Type", mimeType)

		ctx := NewWebContext(router.log)
		ctx.url = r.URL.String()
		parts, err := urlParts(ctx.url)
		if err != nil {
			return err
		}
		ctx.urlParts = parts

		ctx.responseWriter = w
		ctx.request = r

		return callArounds(router.arounds, handler, ctx)
	}
}

func (router *WebRouter) getRequestHandler(handler WebHandler) http.HandlerFunc {
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
	router.get(pattern, func(w http.ResponseWriter, r *http.Request) error {
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

func (router *WebRouter) get(pattern string, handler WebHandler) {
	if pattern == RootUrlPattern {
		router.log.Errorf("Can not use pattern that touches root, use GetRootHTML or GetWildcardHTML instead")
		return
	}

	router.serveMux.HandleFunc(pattern, router.getRequestHandler(handler))
}

func (router *WebRouter) Run() error {
	router.log.Infof("Running server at http://localhost:%v/", router.port)
	server := &http.Server{Addr: ":" + strconv.Itoa(router.port), Handler: router.serveMux}
	return server.ListenAndServe()
}

//
// Requester
//
type WebRequester struct {
	hostname string
	port     int
}

func newWebRequester(port int) *WebRequester {
	return &WebRequester{
		"localhost",
		port,
	}
}

func (requester *WebRequester) Get(url string) (*Response, error) {
	response, err := http.Get(fmt.Sprintf("http://%v:%v%v", requester.hostname, requester.port, url))
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf(strings.TrimSpace(string(body)))
	}
	return NewResponse(body, response.Header.Get("Content-Type")), nil
}
