package content

import (
	"mime"

	"github.com/sirupsen/logrus"

	"github.com/s12chung/go_homepage/go/app"
	"github.com/s12chung/go_homepage/go/content/models"
	"github.com/s12chung/go_homepage/go/content/routes"
	"github.com/s12chung/go_homepage/go/lib/atom"
	"github.com/s12chung/go_homepage/go/lib/html"
	"github.com/s12chung/go_homepage/go/lib/router"
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
	helper *routes.Helper
}

type Route interface {
	SetRoutes(r router.Router, tracker *app.Tracker)
	WildcardUrls() ([]string, error)
}

func NewContent(generatedPath string, settings *Settings, log logrus.FieldLogger) *Content {
	models.Config(settings.Models, log.WithFields(logrus.Fields{
		"type": "models",
	}))
	for ext, mimeType := range ExtraMimeTypes {
		mime.AddExtensionType(ext, mimeType)
	}

	htmlRenderer := html.NewRenderer(generatedPath, settings.Template, log)
	atomRenderer := atom.NewHtmlRenderer(settings.Atom)
	helper := routes.NewHelper(settings.Goodreads, htmlRenderer, atomRenderer)

	return &Content{
		settings,
		log,
		allRoutes(helper),
		helper,
	}
}

func allRoutes(helper *routes.Helper) []Route {
	return []Route{
		routes.NewAllRoutes(helper),
	}
}

func (content *Content) SetRoutes(r router.Router, tracker *app.Tracker) {
	for _, route := range content.routes {
		route.SetRoutes(r, tracker)
	}
}

func (content *Content) WildcardUrls() ([]string, error) {
	var wildcardUrls []string
	for _, route := range content.routes {
		urls, err := route.WildcardUrls()
		if err != nil {
			return nil, err
		}
		wildcardUrls = append(wildcardUrls, urls...)
	}
	return wildcardUrls, nil
}

func (content *Content) AssetsUrl() string {
	return content.helper.HtmlRenderer.AssetsUrl()
}

func (content *Content) GeneratedAssetsPath() string {
	return content.helper.HtmlRenderer.GeneratedAssetsPath()
}
