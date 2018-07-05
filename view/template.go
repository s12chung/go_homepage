package view

import (
	"bufio"
	"html/template"
	"os"
	"path"

	"github.com/s12chung/go_homepage/view/webpack"
)

//
// TemplateGenerator
//
type TemplateGenerator struct {
	Webpack *webpack.Helper
}

func NewTemplateGenerator(assetsPath string) (*TemplateGenerator, error) {
	helper, err := webpack.NewHelper(assetsPath)
	if err != nil {
		return nil, err
	}

	return &TemplateGenerator{
		helper,
	}, nil
}

func (tg *TemplateGenerator) NewTemplate(name, path string) *Template {
	return NewTemplate(name, path, tg.Webpack)
}

//
// Template
//
type Template struct {
	Name    string
	Path    string
	Webpack *webpack.Helper
}

func NewTemplate(name, path string, w *webpack.Helper) *Template {
	return &Template{
		name,
		path,
		w,
	}
}

func (t *Template) funcs() template.FuncMap {
	return template.FuncMap{
		"webpack": t.Webpack.Url,
	}
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
