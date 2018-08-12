package html

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/russross/blackfriday"
	"github.com/sirupsen/logrus"

	"github.com/s12chung/go_homepage/go/lib/html/webpack"
	"github.com/s12chung/go_homepage/go/lib/utils"
)

var imgRegex = regexp.MustCompile(`<img (src="([^"]*)")`)

type Renderer struct {
	settings *Settings
	w        *webpack.Webpack
	log      logrus.FieldLogger
}

func NewRenderer(generatedPath string, settings *Settings, log logrus.FieldLogger) *Renderer {
	w := webpack.NewWebpack(generatedPath, log)
	return &Renderer{
		settings,
		w,
		log,
	}
}

func (renderer *Renderer) Webpack() *webpack.Webpack {
	return renderer.w
}

func (renderer *Renderer) AssetsUrl() string {
	return webpack.AssetsUrl()
}

func (renderer *Renderer) GeneratedAssetsPath() string {
	return renderer.Webpack().GeneratedAssetsPath()
}

func (renderer *Renderer) processHTML(html string) string {
	return imgRegex.ReplaceAllStringFunc(html, func(imgTag string) string {
		matches := imgRegex.FindStringSubmatch(imgTag)
		responsiveImage := renderer.Webpack().GetResponsiveImage(matches[2])

		attributes := []string{fmt.Sprintf(`src="%v"`, responsiveImage.Src)}
		if responsiveImage.SrcSet != "" {
			attributes = append(attributes, fmt.Sprintf(`srcset="%v"`, responsiveImage.SrcSet))
		}
		return strings.Replace(imgTag, matches[1], strings.Join(attributes, " "), 1)
	})
}

func (renderer *Renderer) parseMarkdownPath(filename string) template.HTML {
	filePath := path.Join(renderer.settings.MarkdownsPath, filename)
	input, err := ioutil.ReadFile(filePath)
	if err != nil {
		renderer.log.Error(err)
		return ""
	}
	return template.HTML(blackfriday.Run(input))
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
	tgFuncs := template.FuncMap{
		"webpackPath": renderer.Webpack().ManifestPath,
		"markdown":    renderer.parseMarkdownPath,
		"processHTML": renderer.processHTML,
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
