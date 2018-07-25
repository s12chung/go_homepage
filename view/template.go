package view

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"path"
	"regexp"
	"strings"

	"github.com/russross/blackfriday"

	log "github.com/Sirupsen/logrus"

	"github.com/s12chung/go_homepage/settings"
	"github.com/s12chung/go_homepage/view/webpack"
)

//
// TemplateGenerator
//
type TemplateGenerator struct {
	ManifestMap map[string]string
	Settings    settings.TemplateSettings
}

func NewTemplateGenerator(settings settings.TemplateSettings) (*TemplateGenerator, error) {
	manifestMap, err := webpack.ReadManifest(path.Join(settings.AssetsPath, settings.ManifestFilename))
	if err != nil {
		return nil, err
	}

	return &TemplateGenerator{
		manifestMap,
		settings,
	}, nil
}

func (tg *TemplateGenerator) browserAssetsPath() string {
	return regexp.MustCompile("\\A.*/").ReplaceAllString(tg.Settings.AssetsPath, "/")
}

func (tg *TemplateGenerator) webpackUrl(manifestKey string) string {
	manifestValue := tg.ManifestMap[manifestKey]

	if manifestValue == "" {
		log.Error("webpack manifestValue not found for key: ", manifestKey)
	}

	return tg.browserAssetsPath() + "/" + manifestValue
}

func (tg *TemplateGenerator) parseMarkdownPath(filename string) template.HTML {
	filePath := path.Join(tg.Settings.MarkdownsPath, filename)
	input, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Error(err)
		return ""
	}
	return template.HTML(blackfriday.Run(input))
}

func (tg *TemplateGenerator) TemplateFuncs() template.FuncMap {
	defaults := defaultTemplateFuncs()
	tgFuncs := template.FuncMap{
		"Webpack":  tg.webpackUrl,
		"Markdown": tg.parseMarkdownPath,
	}
	for k, v := range tgFuncs {
		defaults[k] = v
	}
	return defaults
}

func (tg *TemplateGenerator) NewTemplate(name string) *Template {
	return NewTemplate(name, tg)
}

func (tg *TemplateGenerator) RenderNewTemplate(name string, data interface{}) ([]byte, error) {
	return tg.NewTemplate(name).Render(data)
}

//
// Template
//
type Template struct {
	Name              string
	templateGenerator *TemplateGenerator
}

func NewTemplate(name string, templateGenerator *TemplateGenerator) *Template {
	return &Template{
		name,
		templateGenerator,
	}
}

func (t *Template) funcs() template.FuncMap {
	tgFuncs := t.templateGenerator.TemplateFuncs()
	tFuncs := template.FuncMap{
		"Title": func() string {
			s := fmt.Sprintf("%v - %v", strings.Title(t.Name), t.templateGenerator.Settings.WebsiteTitle)
			if t.Name == "index" {
				s = t.templateGenerator.Settings.WebsiteTitle
			}
			return s
		},
	}
	for k, v := range tFuncs {
		tgFuncs[k] = v
	}
	return tgFuncs
}

func (t *Template) Render(data interface{}) ([]byte, error) {
	tmpl, err := template.New("self").
		Funcs(t.funcs()).
		ParseFiles(
			path.Join("templates", "layout.tmpl"),
			path.Join("templates", t.Name+".tmpl"),
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
