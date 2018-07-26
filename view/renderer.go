package view

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"path"
	"regexp"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/russross/blackfriday"

	"github.com/s12chung/go_homepage/settings"
	"github.com/s12chung/go_homepage/view/webpack"
)

type Renderer struct {
	ManifestMap map[string]string
	Settings    *settings.TemplateSettings
	log         logrus.FieldLogger
}

func NewRenderer(settings *settings.TemplateSettings, log logrus.FieldLogger) (*Renderer, error) {
	manifestMap, err := webpack.ReadManifest(path.Join(settings.AssetsPath, settings.ManifestFilename))
	if err != nil {
		return nil, err
	}

	return &Renderer{
		manifestMap,
		settings,
		log,
	}, nil
}

func (renderer *Renderer) browserAssetsPath() string {
	return regexp.MustCompile("\\A.*/").ReplaceAllString(renderer.Settings.AssetsPath, "/")
}

func (renderer *Renderer) webpackUrl(manifestKey string) string {
	manifestValue := renderer.ManifestMap[manifestKey]

	if manifestValue == "" {
		renderer.log.Error("webpack manifestValue not found for key: ", manifestKey)
	}

	return renderer.browserAssetsPath() + "/" + manifestValue
}

func (renderer *Renderer) parseMarkdownPath(filename string) template.HTML {
	filePath := path.Join(renderer.Settings.MarkdownsPath, filename)
	input, err := ioutil.ReadFile(filePath)
	if err != nil {
		renderer.log.Error(err)
		return ""
	}
	return template.HTML(blackfriday.Run(input))
}

func (renderer *Renderer) TemplateFuncs(name string) template.FuncMap {
	defaults := defaultTemplateFuncs()
	tgFuncs := template.FuncMap{
		"Webpack":  renderer.webpackUrl,
		"Markdown": renderer.parseMarkdownPath,
		"Title": func() string {
			s := fmt.Sprintf("%v - %v", strings.Title(name), renderer.Settings.WebsiteTitle)
			if name == "index" {
				s = renderer.Settings.WebsiteTitle
			}
			return s
		},
	}
	for k, v := range tgFuncs {
		defaults[k] = v
	}
	return defaults
}

func (renderer *Renderer) Render(name string, data interface{}) ([]byte, error) {
	tmpl, err := template.New("self").
		Funcs(renderer.TemplateFuncs(name)).
		ParseFiles(
			path.Join("templates", "layout.tmpl"),
			path.Join("templates", name+".tmpl"),
		)
	if err != nil {
		return nil, err
	}

	writer := bytes.Buffer{}
	err = tmpl.ExecuteTemplate(&writer, "layout.tmpl", data)
	if err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}
