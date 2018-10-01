package content

import (
	"mime"

	"github.com/sirupsen/logrus"

	"github.com/s12chung/go_homepage/go/content/models"
	"github.com/s12chung/go_homepage/go/content/routes"
	"github.com/s12chung/gostatic/go/app"
	"github.com/s12chung/gostatic/go/lib/html"
	"github.com/s12chung/gostatic/go/lib/router"
	"github.com/s12chung/gostatic/go/lib/webpack"

	"github.com/s12chung/gostatic-packages/atom"
	"github.com/s12chung/gostatic-packages/markdown"
)

var ExtraMimeTypes = map[string]string{
	".atom": "application/xml",
	".ico":  "image/x-icon",
	".txt":  "text/plain; charset=utf-8",
}

type Content struct {
	Settings *Settings
	Log      logrus.FieldLogger

	routes []Route
	helper *routes.BaseHelper
}

type Route interface {
	SetRoutes(r router.Router, tracker *app.Tracker) error
}

func NewContent(generatedPath string, settings *Settings, log logrus.FieldLogger) *Content {
	models.Config(settings.Models, log.WithFields(logrus.Fields{
		"type": "models",
	}))
	for ext, mimeType := range ExtraMimeTypes {
		mime.AddExtensionType(ext, mimeType)
	}

	w := webpack.NewWebpack(generatedPath, settings.Webpack, log)
	md := markdown.NewMarkdown(settings.Markdown, log)
	htmlRenderer := html.NewRenderer(settings.Html, []html.Plugin{w, md}, log)
	atomRenderer := atom.NewHTMLRenderer(settings.Atom)
	helper := routes.NewBaseHelper(settings.Goodreads, w, htmlRenderer, atomRenderer)

	return &Content{
		settings,
		log,
		allRoutes(helper),
		helper,
	}
}

func allRoutes(helper routes.Helper) []Route {
	return []Route{
		routes.NewAllRoutes(helper),
	}
}

func (content *Content) SetRoutes(r router.Router, tracker *app.Tracker) error {
	for _, route := range content.routes {
		err := route.SetRoutes(r, tracker)
		if err != nil {
			return err
		}
	}
	return nil
}

func (content *Content) AssetsURL() string {
	return content.helper.Webpack.AssetsURL()
}

func (content *Content) GeneratedAssetsPath() string {
	return content.helper.Webpack.GeneratedAssetsPath()
}
