package content

import (
	"fmt"
	"mime"

	"github.com/sirupsen/logrus"

	"github.com/s12chung/go_homepage/go/app"
	"github.com/s12chung/go_homepage/go/content/models"
	"github.com/s12chung/go_homepage/go/content/routes"
	"github.com/s12chung/go_homepage/go/lib/atom"
	"github.com/s12chung/go_homepage/go/lib/html"
	"github.com/s12chung/go_homepage/go/lib/markdown"
	"github.com/s12chung/go_homepage/go/lib/router"
	"github.com/s12chung/go_homepage/go/lib/webpack"
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

	w := webpack.NewWebpack(generatedPath, log)
	md := markdown.NewMarkdown(settings.Markdown, log)
	htmlRenderer := html.NewRenderer(settings.Template, []html.Plugin{w, md}, log)
	atomRenderer := atom.NewHtmlRenderer(settings.Atom)
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

func (content *Content) SetRoutes(r router.Router, tracker *app.Tracker) {
	for _, route := range content.routes {
		route.SetRoutes(r, tracker)
	}
}

func (content *Content) WildcardUrls() ([]string, error) {
	wildcardUrls := []string{}
	wildcardUrlsMap := make(map[string]bool)
	for _, route := range content.routes {
		urls, err := route.WildcardUrls()
		if err != nil {
			return nil, err
		}
		for _, url := range urls {
			if wildcardUrlsMap[url] {
				return nil, fmt.Errorf("duplicate wildcar url found: %v", url)
			}
			wildcardUrlsMap[url] = true
		}

		wildcardUrls = append(wildcardUrls, urls...)
	}
	return wildcardUrls, nil
}

func (content *Content) AssetsUrl() string {
	return webpack.AssetsUrl()
}

func (content *Content) GeneratedAssetsPath() string {
	return content.helper.Webpack.GeneratedAssetsPath()
}
