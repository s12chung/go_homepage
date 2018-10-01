package routes

import (
	"path"
	"strings"

	"github.com/s12chung/gostatic/go/lib/html"
	"github.com/s12chung/gostatic/go/lib/router"
	"github.com/s12chung/gostatic/go/lib/webpack"

	"github.com/s12chung/gostatic-packages/atom"
	"github.com/s12chung/gostatic-packages/goodreads"
)

type Helper interface {
	ManifestURL(key string) string
	RespondAtom(ctx router.Context, feedName, logoUrl string, htmlEntries []*atom.HTMLEntry) error
	RespondHTML(ctx router.Context, templateName string, data interface{}) error
	GoodreadsSettings() *goodreads.Settings
}

type BaseHelper struct {
	goodreadsSettings *goodreads.Settings
	Webpack           *webpack.Webpack
	HtmlRenderer      *html.Renderer
	AtomRenderer      *atom.HTMLRenderer
}

func NewBaseHelper(goodReadSettings *goodreads.Settings, w *webpack.Webpack, htmlRenderer *html.Renderer, atomRenderer *atom.HTMLRenderer) *BaseHelper {
	return &BaseHelper{goodReadSettings, w, htmlRenderer, atomRenderer}
}

func (helper *BaseHelper) ManifestURL(key string) string {
	return helper.Webpack.ManifestURL(key)
}

func (helper *BaseHelper) GoodreadsSettings() *goodreads.Settings {
	return helper.goodreadsSettings
}

func (helper *BaseHelper) RespondAtom(ctx router.Context, feedName, logoUrl string, htmlEntries []*atom.HTMLEntry) error {
	bytes, err := helper.AtomRenderer.Render(feedName, ctx.URL(), logoUrl, htmlEntries)
	if err != nil {
		return err
	}
	ctx.Respond(bytes)
	return nil
}

func (helper *BaseHelper) RespondHTML(ctx router.Context, tmplName string, layoutD interface{}) error {
	tmplName = templateName(tmplName)

	ctx.Log().Infof("Rendering template %v", tmplName)

	bytes, err := helper.HtmlRenderer.Render(tmplName, layoutD)
	if err != nil {
		return err
	}
	ctx.Respond(bytes)
	return nil
}

func templateName(tmplName string) string {
	return strings.TrimRight(path.Base(tmplName), path.Ext(tmplName))
}
