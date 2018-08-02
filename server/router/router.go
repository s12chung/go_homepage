package router

import (
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/s12chung/go_homepage/settings"
	"github.com/s12chung/go_homepage/view"
)

const WildcardUrlPattern = "*"
const RootUrlPattern = "/"

var IsRootUrlPart = func(urlParts []string) bool { return len(urlParts) == 0 }

type Context interface {
	Renderer() *view.Renderer

	Render(data interface{}) error
	Respond(bytes []byte) error

	Settings() *settings.Settings
	Log() logrus.FieldLogger
	SetLog(log logrus.FieldLogger)
	UrlParts() []string
	Url() string
	TemplateName() string
	SetTemplateName(templateName string)
}

type Router interface {
	Around(handler func(ctx Context, handler func(ctx Context) error) error)
	GetWildcardHTML(handler func(ctx Context) error)
	GetRootHTML(handler func(ctx Context) error)
	GetHTML(pattern string, handler func(ctx Context) error)

	StaticRoutes() []string
	Requester() Requester
}

type Requester interface {
	Get(url string) ([]byte, error)
}

func panicDuplicateRoute(route string) {
	panic(fmt.Sprintf("%v is a duplicate route", route))
}

func urlParts(url string) ([]string, error) {
	var parts []string
	for _, part := range strings.Split(url, "/") {
		if part != "" {
			parts = append(parts, part)
		}
	}
	if len(parts) > 1 {
		return nil, fmt.Errorf("currently can't handle more than 1 UrlPart - %v", url)
	}
	return parts, nil
}

func callArounds(arounds []func(ctx Context, handler func(ctx Context) error) error, handler func(ctx Context) error, ctx Context) error {
	aroundToNext := make([]func(ctx Context) error, len(arounds))
	for index := range arounds {
		reverseIndex := len(arounds) - 1 - index
		around := arounds[reverseIndex]
		if index == 0 {
			aroundToNext[reverseIndex] = func(ctx Context) error {
				return around(ctx, handler)
			}
		} else {
			aroundToNext[reverseIndex] = func(ctx Context) error {
				return around(ctx, aroundToNext[reverseIndex+1])
			}
		}
	}
	return aroundToNext[0](ctx)
}

func templateName(ctx Context, templateName string) string {
	if templateName == "" {
		// assume len of <= 1: https://github.com/s12chung/go_homepage/blob/aa77eaf3ffff669b6abaab35078fb65ee3ffb17c/server/router/router.go#L52
		if IsRootUrlPart(ctx.UrlParts()) {
			ctx.Log().Panicf("No TemplateName given for root route")
			return ""
		}
		return ctx.UrlParts()[0]
	} else {
		return templateName
	}
}

func renderTemplate(ctx Context, data interface{}) ([]byte, error) {
	templateName := ctx.TemplateName()
	defaultTitle := templateName
	if IsRootUrlPart(ctx.UrlParts()) {
		defaultTitle = ""
	}

	ctx.Log().Infof("Rendering template %v", templateName)
	return ctx.Renderer().Render(templateName, defaultTitle, data)
}
