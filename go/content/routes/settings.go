package routes

import (
	"github.com/s12chung/go_homepage/go/content/models"
	"github.com/s12chung/go_homepage/go/lib/atom"
	"github.com/s12chung/go_homepage/go/lib/goodreads"
	"github.com/s12chung/go_homepage/go/lib/html"
)

type Settings struct {
	Models    *models.Settings    `json:"model,omitempty"`
	Template  *html.Settings      `json:"template,omitempty"`
	Atom      *atom.Settings      `json:"atom,omitempty"`
	Goodreads *goodreads.Settings `json:"goodreads,omitempty"`
}

func DefaultSettings() *Settings {
	return &Settings{
		models.DefaultSettings(),
		html.DefaultSettings(),
		atom.DefaultSettings(),
		goodreads.DefaultSettings(),
	}
}
