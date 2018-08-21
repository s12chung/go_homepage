package html

import (
	"bytes"
	"fmt"
	"html/template"
	"path"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/s12chung/go_homepage/go/lib/utils"
)

type Renderer struct {
	settings *Settings
	plugins  []Plugin
	log      logrus.FieldLogger
}

func NewRenderer(settings *Settings, plugins []Plugin, log logrus.FieldLogger) *Renderer {
	return &Renderer{
		settings,
		plugins,
		log,
	}
}

type Plugin interface {
	TemplateFuncs() template.FuncMap
}

func (renderer *Renderer) partialPaths() ([]string, error) {
	filePaths, err := utils.FilePaths(".tmpl", renderer.settings.TemplatePath)
	if err != nil {
		return nil, err
	}

	var partialPaths []string
	for _, filePath := range filePaths {
		if strings.HasPrefix(filepath.Base(filePath), "_") {
			partialPaths = append(partialPaths, filePath)
		}
	}
	return partialPaths, nil
}

func (renderer *Renderer) templateFuncs(defaultTitle string) template.FuncMap {
	defaults := defaultTemplateFuncs()
	// add tests to ./testdata/renderer_funcs.tmpl (excluding title)
	mergeFuncMap(defaults, template.FuncMap{
		"title": func(data interface{}) string {
			title := utils.GetStringField(data, "Title")
			if title == "" {
				title = defaultTitle
			}

			if title == "" {
				return renderer.settings.WebsiteTitle
			}
			return fmt.Sprintf("%v - %v", strings.Title(title), renderer.settings.WebsiteTitle)
		},
	})
	for _, plugin := range renderer.plugins {
		mergeFuncMap(defaults, plugin.TemplateFuncs())
	}
	return defaults
}

func mergeFuncMap(dest, src template.FuncMap) {
	for k, v := range src {
		dest[k] = v
	}
}

func (renderer *Renderer) Render(name, defaultTitle string, data interface{}) ([]byte, error) {
	partialPaths, err := renderer.partialPaths()
	if err != nil {
		return nil, err
	}

	templatePaths := append(partialPaths, []string{
		path.Join(renderer.settings.TemplatePath, "layout.tmpl"),
		path.Join(renderer.settings.TemplatePath, name+".tmpl"),
	}...)

	tmpl, err := template.New("self").
		Funcs(renderer.templateFuncs(defaultTitle)).
		ParseFiles(templatePaths...)
	if err != nil {
		return nil, err
	}

	buffer := &bytes.Buffer{}
	err = tmpl.ExecuteTemplate(buffer, "layout.tmpl", data)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
