package view

import (
	"os"
	"bufio"
	"html/template"
	"path"
)

type Template struct {
	Name string
	Path string
}

func NewTemplate(name, path string) *Template {
	return &Template{
		name,
		path,
	}
}

func (tr *Template)Render() (error) {
	file, err := os.Create(tr.Path + ".html")
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	defer writer.Flush()

	tmpl, err := template.ParseFiles(
		path.Join("templates", "layout.tmpl"),
		path.Join("templates", tr.Name + ".tmpl"),
	)
	if err != nil {
		return err
	}

	tmpl.Execute(writer, map[string]interface{}{})

	return nil
}