package view

import (
	"bufio"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/russross/blackfriday"

	log "github.com/Sirupsen/logrus"
	"github.com/s12chung/go_homepage/view/webpack"
)

const manifestFilename = "manifest.json"
const markdownPath = "./assets/markdowns"

//
// TemplateGenerator
//
type TemplateGenerator struct {
	browserAssetsPath string
	manifestMap       map[string]string
}

func NewTemplateGenerator(generatedAssetsPath string) (*TemplateGenerator, error) {
	manifestMap, err := webpack.ReadManifest(path.Join(generatedAssetsPath, manifestFilename))

	if err != nil {
		return nil, err
	}

	return &TemplateGenerator{
		regexp.MustCompile("\\A.*/").ReplaceAllString(generatedAssetsPath, "/"),
		manifestMap,
	}, nil
}

func (tg *TemplateGenerator) webpackUrl(manifestKey string) string {
	manifestValue := tg.manifestMap[manifestKey]

	if manifestValue == "" {
		log.Error("webpack manifestValue not found for key: ", manifestKey)
	}

	return tg.browserAssetsPath + "/" + manifestValue
}

func makeSlice(args ...interface{}) []interface{} {
	return args
}

func (tg *TemplateGenerator) parseMarkdownPath(filename string) template.HTML {
	filePath := path.Join(markdownPath, filename)
	input, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Error(err)
		return ""
	}
	return template.HTML(string(blackfriday.Run(input)))
}

func (tg *TemplateGenerator) TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"Slice":    makeSlice,
		"ToLower":  strings.ToLower,
		"Webpack":  tg.webpackUrl,
		"Markdown": tg.parseMarkdownPath,
	}
}

func (tg *TemplateGenerator) NewTemplate(name, path string) *Template {
	return NewTemplate(name, path, tg)
}

//
// Template
//
type Template struct {
	Name              string
	Path              string
	templateGenerator *TemplateGenerator
}

func NewTemplate(name, path string, templateGenerator *TemplateGenerator) *Template {
	return &Template{
		name,
		path,
		templateGenerator,
	}
}

func (t *Template) funcs() template.FuncMap {
	return t.templateGenerator.TemplateFuncs()
}

func (t *Template) Render() error {
	file, err := os.Create(t.Path + ".html")
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	defer writer.Flush()

	tmpl, err := template.New("self").
		Funcs(t.funcs()).
		ParseFiles(
			path.Join("templates", "layout.tmpl"),
			path.Join("templates", t.Name+".tmpl"),
		)
	if err != nil {
		return err
	}

	tmpl.ExecuteTemplate(writer, "layout.tmpl", nil)

	return nil
}
