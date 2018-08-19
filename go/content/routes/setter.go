package routes

import (
	"github.com/sirupsen/logrus"

	"github.com/s12chung/go_homepage/go/app"
	"github.com/s12chung/go_homepage/go/content/models"
	"github.com/s12chung/go_homepage/go/lib/atom"
	"github.com/s12chung/go_homepage/go/lib/html"
	"github.com/s12chung/go_homepage/go/lib/router"
)

type Setter struct {
	Settings *Settings
	Log      logrus.FieldLogger

	HtmlRenderer *html.Renderer
	AtomRenderer *atom.HtmlRenderer
}

func NewSetter(generatedPath string, settings *Settings, log logrus.FieldLogger) *Setter {
	htmlRenderer := html.NewRenderer(generatedPath, settings.Template, log)
	atomRenderer := atom.NewHtmlRenderer(settings.Atom)
	models.Config(settings.Models, log.WithFields(logrus.Fields{
		"type": "models",
	}))
	return &Setter{settings, log, htmlRenderer, atomRenderer}
}

func (setter *Setter) SetRoutes(r router.Router, tracker *app.Tracker) {
	setter.setAllRoutes(r, tracker)
}

func (setter *Setter) AssetsUrl() string {
	return setter.HtmlRenderer.AssetsUrl()
}

func (setter *Setter) GeneratedAssetsPath() string {
	return setter.HtmlRenderer.GeneratedAssetsPath()
}

func (setter *Setter) WildcardRoutes() ([]string, error) {
	return setter.WildcardPostRoutes()
}

func (setter Setter) RespondAtom(ctx router.Context, feedName, logoUrl string, htmlEntries []*atom.HtmlEntry) error {
	bytes, err := setter.AtomRenderer.Render(feedName, ctx.Url(), logoUrl, htmlEntries)
	if err != nil {
		return err
	}
	return ctx.Respond(bytes)
}

func (setter Setter) RespondUrlHTML(ctx router.Context, data interface{}) error {
	return setter.RespondHTML(ctx, "", data)
}

func (setter Setter) RespondHTML(ctx router.Context, templateName string, data interface{}) error {
	bytes, err := setter.renderHTML(ctx, templateName, data)
	if err != nil {
		return err
	}
	return ctx.Respond(bytes)
}

func (setter Setter) renderHTML(ctx router.Context, tmplName string, data interface{}) ([]byte, error) {
	tmplName = templateName(ctx, tmplName)
	defaultTitle := defaultTitle(ctx, tmplName)
	ctx.Log().Infof("Rendering template %v with title %v", tmplName, defaultTitle)
	return setter.HtmlRenderer.Render(tmplName, defaultTitle, data)
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
