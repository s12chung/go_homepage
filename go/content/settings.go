package content

import (
	"github.com/s12chung/go_homepage/go/content/models"

	"github.com/s12chung/gostatic/go/lib/html"
	"github.com/s12chung/gostatic/go/lib/webpack"

	"github.com/s12chung/gostatic-packages/atom"
	"github.com/s12chung/gostatic-packages/goodreads"
	"github.com/s12chung/gostatic-packages/markdown"
)

type Settings struct {
	Models    *models.Settings    `json:"models,omitempty"`
	HTML      *html.Settings      `json:"html,omitempty"`
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
