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
const assetsFolder = "assets"
const markdownsPath = "./assets/markdowns"

//
// TemplateGenerator
//
type TemplateGenerator struct {
	generatedPath string
	manifestMap   map[string]string
}

func NewTemplateGenerator(generatedPath string) (*TemplateGenerator, error) {
	manifestMap, err := webpack.ReadManifest(path.Join(generatedPath, assetsFolder, manifestFilename))
	if err != nil {
		return nil, err
	}

	return &TemplateGenerator{
		generatedPath,
		manifestMap,
	}, nil
}

func (tg *TemplateGenerator) generatedAssetsPath() string {
	return path.Join(tg.generatedPath, assetsFolder)
}

func (tg *TemplateGenerator) browserAssetsPath() string {
	return regexp.MustCompile("\\A.*/").ReplaceAllString(tg.generatedAssetsPath(), "/")
}

func (tg *TemplateGenerator) webpackUrl(manifestKey string) string {
	manifestValue := tg.manifestMap[manifestKey]

	if manifestValue == "" {
		log.Error("webpack manifestValue not found for key: ", manifestKey)
	}

	return tg.browserAssetsPath() + "/" + manifestValue
}

func makeSlice(args ...interface{}) []interface{} {
	return args
}

func (tg *TemplateGenerator) parseMarkdownPath(filename string) template.HTML {
	filePath := path.Join(markdownsPath, filename)
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

func (tg *TemplateGenerator) NewTemplate(name string) *Template {
	return NewTemplate(name, tg)
}

func (tg *TemplateGenerator) RenderNewTemplate(name string) error {
	err := tg.NewTemplate(name).Render()
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
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
	return t.templateGenerator.TemplateFuncs()
}

func (t *Template) path() string {
	return path.Join(t.templateGenerator.generatedPath, t.Name+".html")
}

func (t *Template) Render() error {
	file, err := os.Create(t.path())
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
