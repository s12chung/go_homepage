package markdown

import (
	"html/template"
	"io/ioutil"
	"path"

	"github.com/russross/blackfriday"
	"github.com/sirupsen/logrus"
)

type Markdown struct {
	settings *Settings
	log      logrus.FieldLogger
}

func NewMarkdown(settings *Settings, log logrus.FieldLogger) *Markdown {
	return &Markdown{settings, log}
}

func (markdown *Markdown) parseMarkdownPath(filename string) string {
	filePath := path.Join(markdown.settings.MarkdownsPath, filename)
	input, err := ioutil.ReadFile(filePath)
	if err != nil {
		markdown.log.Error(err)
		return ""
	}
	return string(blackfriday.Run(input))
}

func (markdown *Markdown) TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"markdown": markdown.parseMarkdownPath,
	}
}
