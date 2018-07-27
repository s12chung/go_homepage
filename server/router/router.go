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

type Context interface {
	Render(name string, data interface{}) error
	renderer() *view.Renderer

	Settings() *settings.Settings
	Log() logrus.FieldLogger
	UrlParts() []string
	Url() string
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

func renderTemplate(ctx Context, name string, data interface{}) ([]byte, error) {
	ctx.Log().Infof("Rendering template: %v", ctx.Url())
	return ctx.renderer().Render(name, data)
}
