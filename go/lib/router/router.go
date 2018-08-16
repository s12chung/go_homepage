package router

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

const WildcardUrlPattern = "*"
const RootUrlPattern = "/"

var IsRootUrlPart = func(urlParts []string) bool { return len(urlParts) == 0 }

type ContextHandler func(ctx Context) error
type AroundHandler func(ctx Context, handler ContextHandler) error

type Context interface {
	Respond(bytes []byte) error

	Log() logrus.FieldLogger
	SetLog(log logrus.FieldLogger)

	UrlParts() []string
	Url() string
}

type Router interface {
	Around(handler AroundHandler)
	GetWildcardHTML(handler ContextHandler)
	GetRootHTML(handler ContextHandler)
	GetHTML(pattern string, handler ContextHandler)
	Get(pattern, mimeType string, handler ContextHandler)

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

func callArounds(arounds []AroundHandler, handler ContextHandler, ctx Context) error {
	aroundToNext := make([]ContextHandler, len(arounds))
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
