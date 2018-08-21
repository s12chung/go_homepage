package content

import (
	"github.com/s12chung/go_homepage/go/content/models"
	"github.com/s12chung/go_homepage/go/lib/atom"
	"github.com/s12chung/go_homepage/go/lib/goodreads"
	"github.com/s12chung/go_homepage/go/lib/html"
	"github.com/s12chung/go_homepage/go/lib/markdown"
	"github.com/s12chung/go_homepage/go/lib/webpack"
)

type Settings struct {
	Models    *models.Settings    `json:"models,omitempty"`
	Template  *html.Settings      `json:"template,omitempty"`
	Atom      *atom.Settings      `json:"atom,omitempty"`
	Goodreads *goodreads.Settings `json:"goodreads,omitempty"`
	Markdown  *markdown.Settings  `json:"markdown,omitempty"`
	Webpack   *webpack.Settings   `json:"webpack,omitempty"`
}

func DefaultSettings() *Settings {
	return &Settings{
		models.DefaultSettings(),
		html.DefaultSettings(),
		atom.DefaultSettings(),
		goodreads.DefaultSettings(),
		markdown.DefaultSettings(),
		webpack.DefaultSettings(),
	}
}
