package router

import (
	"github.com/Sirupsen/logrus"

	"github.com/s12chung/go_homepage/go/app/settings"
	"github.com/s12chung/go_homepage/go/lib/view"
)

//
// Context
//
type GenerateContext struct {
	r        *view.Renderer
	settings *settings.Settings
	log      logrus.FieldLogger

	url      string
	urlParts []string
	tmplName string

	response []byte
}

func (ctx *GenerateContext) Renderer() *view.Renderer {
	return ctx.r
}

func (ctx *GenerateContext) Settings() *settings.Settings {
	return ctx.settings
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

func (ctx *GenerateContext) TemplateName() string {
	return templateName(ctx, ctx.tmplName)
}

func (ctx *GenerateContext) SetTemplateName(templateName string) {
	ctx.tmplName = templateName
}

func (ctx *GenerateContext) Render(data interface{}) error {
	bytes, err := renderTemplate(ctx, data)
	if err != nil {
		return err
	}
	return ctx.Respond(bytes)
}

func (ctx *GenerateContext) Respond(bytes []byte) error {
	ctx.response = bytes
	return nil
}

//
// Router
//
type GenerateRouter struct {
	defaultContext *GenerateContext
	log            logrus.FieldLogger
	routes         map[string]func(ctx Context) error

	arounds []func(ctx Context, handler func(ctx Context) error) error
}

func NewGenerateRouter(renderer *view.Renderer, settings *settings.Settings, log logrus.FieldLogger) *GenerateRouter {
	defaultContext := &GenerateContext{
		r:        renderer,
		settings: settings,
		log:      log,
	}
	return &GenerateRouter{
		defaultContext,
		log,
		make(map[string]func(ctx Context) error),
		nil,
	}
}

func (router *GenerateRouter) Around(handler func(ctx Context, handler func(ctx Context) error) error) {
	router.arounds = append(router.arounds, handler)
}

func (router *GenerateRouter) GetWildcardHTML(handler func(ctx Context) error) {
	router.checkAndSetRoutes(WildcardUrlPattern, handler)
}

func (router *GenerateRouter) GetRootHTML(handler func(ctx Context) error) {
	router.checkAndSetRoutes(RootUrlPattern, handler)
}

func (router *GenerateRouter) GetHTML(pattern string, handler func(ctx Context) error) {
	router.checkAndSetRoutes(pattern, handler)
}

func (router *GenerateRouter) checkAndSetRoutes(pattern string, handler func(ctx Context) error) {
	_, has := router.routes[pattern]
	if has {
		panicDuplicateRoute(pattern)
	}
	router.routes[pattern] = handler
}

func (router *GenerateRouter) get(url string) ([]byte, error) {
	handler := router.routes[url]
	if handler == nil {
		handler = router.routes[WildcardUrlPattern]
	}

	ctx := *router.defaultContext
	ctx.url = url
	parts, err := urlParts(ctx.url)
	if err != nil {
		return nil, err
	}
	ctx.urlParts = parts

	err = callArounds(router.arounds, handler, &ctx)
	if err != nil {
		return nil, err
	}
	return ctx.response, nil
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

func (requester *GenerateRequester) Get(url string) ([]byte, error) {
	return requester.router.get(url)
}