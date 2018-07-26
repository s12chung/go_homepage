package router

import (
	"github.com/Sirupsen/logrus"
	"github.com/s12chung/go_homepage/settings"
)

type Context interface {
	Render(name string, data interface{}) error

	Settings() *settings.Settings
	Log() logrus.FieldLogger
	UrlParts() []string
	Url() string
}

type Router interface {
	Around(handler func(ctx *WebContext, handler func(ctx *WebContext) error) error)
	GetWildcardHTML(handler func(ctx *WebContext) error)
	GetRootHTML(handler func(ctx *WebContext) error)
	GetHTML(pattern string, handler func(ctx *WebContext) error)
}
