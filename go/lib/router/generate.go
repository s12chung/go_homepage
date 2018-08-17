package router

import (
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
	"mime"
)

func RunFileServer(targetDir string, port int, log logrus.FieldLogger) error {
	log.Infof("Serving files from '%v' at http://localhost:%v/", targetDir, port)
	handler := http.FileServer(http.Dir(targetDir))
	return http.ListenAndServe(":"+strconv.Itoa(port), handler)
}

//
// Context
//
type GenerateContext struct {
	log logrus.FieldLogger

	url      string
	urlParts []string

	response []byte
}

func NewGenerateContext(log logrus.FieldLogger) *GenerateContext {
	return &GenerateContext{log: log}
}

func (ctx *GenerateContext) Log() logrus.FieldLogger {
	return ctx.log
}

func (ctx *GenerateContext) SetLog(log logrus.FieldLogger) {
	ctx.log = log
}

func (ctx *GenerateContext) UrlParts() []string {
	return ctx.urlParts
}

func (ctx *GenerateContext) Url() string {
	return ctx.url
}

func (ctx *GenerateContext) Respond(bytes []byte) error {
	ctx.response = bytes
	return nil
}

//
// Router
//
type GenerateRoute struct {
	MimeType string
	handler  ContextHandler
}

func NewGenerateRoute(mimeType string, handler ContextHandler) *GenerateRoute {
	return &GenerateRoute{mimeType, handler}
}

type GenerateRouter struct {
	log    logrus.FieldLogger
	routes map[string]*GenerateRoute

	arounds []AroundHandler
}

func NewGenerateRouter(log logrus.FieldLogger) *GenerateRouter {
	return &GenerateRouter{
		log,
		make(map[string]*GenerateRoute),
		nil,
	}
}

func (router *GenerateRouter) Around(handler AroundHandler) {
	router.arounds = append(router.arounds, handler)
}

func (router *GenerateRouter) GetWildcardHTML(handler ContextHandler) {
	router.checkAndSetHTMLRoutes(WildcardUrlPattern, handler)
}

func (router *GenerateRouter) GetRootHTML(handler ContextHandler) {
	router.checkAndSetHTMLRoutes(RootUrlPattern, handler)
}

func (router *GenerateRouter) GetHTML(pattern string, handler ContextHandler) {
	router.checkAndSetHTMLRoutes(pattern, handler)
}

func (router *GenerateRouter) Get(pattern, mimeType string, handler ContextHandler) {
	router.checkAndSetRoutes(pattern, mimeType, handler)
}

func (router *GenerateRouter) checkAndSetHTMLRoutes(pattern string, handler ContextHandler) {
	router.checkAndSetRoutes(pattern, mime.TypeByExtension(".html"), handler)
}

func (router *GenerateRouter) checkAndSetRoutes(pattern, mimeType string, handler ContextHandler) {
	_, has := router.routes[pattern]
	if has {
		panicDuplicateRoute(pattern)
	}
	router.routes[pattern] = NewGenerateRoute(mimeType, handler)
}

func (router *GenerateRouter) get(url string) (*Response, error) {
	route := router.routes[url]
	if route == nil {
		route = router.routes[WildcardUrlPattern]
	}

	ctx := NewGenerateContext(router.log)
	ctx.url = url
	parts, err := urlParts(ctx.url)
	if err != nil {
		return nil, err
	}
	ctx.urlParts = parts

	err = callArounds(router.arounds, route.handler, ctx)
	if err != nil {
		return nil, err
	}
	return NewResponse(ctx.response, route.MimeType), nil
}

func (router *GenerateRouter) StaticRoutes() []string {
	var staticRoutes []string
	for k := range router.routes {
		if k != WildcardUrlPattern {
			staticRoutes = append(staticRoutes, k)
		}
	}
	return staticRoutes
}

func (router *GenerateRouter) Requester() Requester {
	return &GenerateRequester{
		router,
	}
}

//
// Requester
//
type GenerateRequester struct {
	router *GenerateRouter
}

func (requester *GenerateRequester) Get(url string) (*Response, error) {
	return requester.router.get(url)
}
