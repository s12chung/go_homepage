package view

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/russross/blackfriday"

	"github.com/s12chung/go_homepage/settings"
	"github.com/s12chung/go_homepage/utils"
	"github.com/s12chung/go_homepage/view/webpack"
)

const templatePath = "./templates"

var imgRegex = regexp.MustCompile(`<img (src="([^"]*)")`)

type Renderer struct {
	Settings *settings.TemplateSettings
	webpack  *webpack.Webpack
	log      logrus.FieldLogger
}

func NewRenderer(generatedPath string, settings *settings.TemplateSettings, log logrus.FieldLogger) *Renderer {
	webpack := webpack.NewWebpack(generatedPath, settings, log)
	return &Renderer{
		settings,
		webpack,
		log,
	}
}

func (renderer *Renderer) browserAssetsPath() string {
	return regexp.MustCompile("\\A.*/").ReplaceAllString(renderer.Settings.AssetsPath, "/")
}

func (renderer *Renderer) webpackUrl(key string) string {
	return renderer.browserAssetsPath() + "/" + renderer.webpack.ManifestValue(key)
}

func (renderer *Renderer) processHTML(html string) string {
	return imgRegex.ReplaceAllStringFunc(html, func(imgTag string) string {
		matches := imgRegex.FindStringSubmatch(imgTag)
		responsiveImage := renderer.webpack.GetResponsiveImage(matches[2])

		attributes := []string{fmt.Sprintf(`src="%v"`, responsiveImage.Src)}
		if responsiveImage.SrcSet != "" {
			attributes = append(attributes, fmt.Sprintf(`srcset="%v"`, responsiveImage.SrcSet))
		}
		return strings.Replace(imgTag, matches[1], strings.Join(attributes, " "), 1)
	})
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

func (renderer *Renderer) partialPaths() ([]string, error) {
	filePaths, err := utils.FilePaths(".tmpl", templatePath)
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
	tgFuncs := template.FuncMap{
		"webpack":     renderer.webpackUrl,
		"markdown":    renderer.parseMarkdownPath,
		"processHTML": renderer.processHTML,
		"title": func(data interface{}) string {
			title := utils.GetStringField(data, "Title")
			if title == "" {
				title = defaultTitle
			}

			if title == "" {
				return renderer.Settings.WebsiteTitle
			}
			return fmt.Sprintf("%v - %v", strings.Title(title), renderer.Settings.WebsiteTitle)
		},
	}

	defaults := defaultTemplateFuncs()
	for k, v := range tgFuncs {
		defaults[k] = v
	}
	return defaults
}

func (renderer *Renderer) Render(name, defaultTitle string, data interface{}) ([]byte, error) {
	partialPaths, err := renderer.partialPaths()
	if err != nil {
		return nil, err
	}

	templatePaths := append(partialPaths, []string{
		path.Join(templatePath, "layout.tmpl"),
		path.Join(templatePath, name+".tmpl"),
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
