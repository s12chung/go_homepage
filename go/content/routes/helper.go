package routes

import (
	"github.com/s12chung/go_homepage/go/lib/atom"
	"github.com/s12chung/go_homepage/go/lib/goodreads"
	"github.com/s12chung/go_homepage/go/lib/html"
	"github.com/s12chung/go_homepage/go/lib/router"
)

type Helper interface {
	ManifestUrl(key string) string
	RespondAtom(ctx router.Context, feedName, logoUrl string, htmlEntries []*atom.HtmlEntry) error
	RespondUrlHTML(ctx router.Context, data interface{}) error
	RespondHTML(ctx router.Context, templateName string, data interface{}) error
	GoodreadsSettings() *goodreads.Settings
}

type BaseHelper struct {
	goodreadsSettings *goodreads.Settings
	HtmlRenderer      *html.Renderer
	AtomRenderer      *atom.HtmlRenderer
}

func NewBaseHelper(goodReadSettings *goodreads.Settings, htmlRenderer *html.Renderer, atomRenderer *atom.HtmlRenderer) *BaseHelper {
	return &BaseHelper{goodReadSettings, htmlRenderer, atomRenderer}
}

func (helper *BaseHelper) ManifestUrl(key string) string {
	return helper.HtmlRenderer.Webpack().ManifestUrl(key)
}

func (helper *BaseHelper) GoodreadsSettings() *goodreads.Settings {
	return helper.goodreadsSettings
}

func (helper *BaseHelper) RespondAtom(ctx router.Context, feedName, logoUrl string, htmlEntries []*atom.HtmlEntry) error {
	bytes, err := helper.AtomRenderer.Render(feedName, ctx.Url(), logoUrl, htmlEntries)
	if err != nil {
		return err
	}
	return ctx.Respond(bytes)
}

func (helper *BaseHelper) RespondUrlHTML(ctx router.Context, data interface{}) error {
	return helper.RespondHTML(ctx, "", data)
}

func (helper *BaseHelper) RespondHTML(ctx router.Context, templateName string, data interface{}) error {
	bytes, err := helper.renderHTML(ctx, templateName, data)
	if err != nil {
		return err
	}
	return ctx.Respond(bytes)
}

func (helper *BaseHelper) renderHTML(ctx router.Context, tmplName string, data interface{}) ([]byte, error) {
	tmplName = templateName(ctx, tmplName)
	defaultTitle := defaultTitle(ctx, tmplName)
	ctx.Log().Infof("Rendering template %v with title %v", tmplName, defaultTitle)
	return helper.HtmlRenderer.Render(tmplName, defaultTitle, data)
}

func templateName(ctx router.Context, templateName string) string {
	if templateName == "" {
		// assume len of <= 1: https://github.com/s12chung/go_homepage/blob/aa77eaf3ffff669b6abaab35078fb65ee3ffb17c/server/router/router.go#L52
		if router.IsRootUrlPart(ctx.UrlParts()) {
			ctx.Log().Panicf("No TemplateName given for root respond")
			return ""
		}
		return ctx.UrlParts()[0]
	} else {
		return templateName
	}
}

func defaultTitle(ctx router.Context, templateName string) string {
	defaultTitle := templateName
	if router.IsRootUrlPart(ctx.UrlParts()) {
		defaultTitle = ""
	}
	return defaultTitle
}
