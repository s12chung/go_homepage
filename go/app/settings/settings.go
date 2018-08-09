package settings

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/Sirupsen/logrus"

	"github.com/s12chung/go_homepage/go/lib/goodreads"
	"github.com/s12chung/go_homepage/go/lib/view"
	"github.com/s12chung/go_homepage/go/lib/view/atom"
)

type Settings struct {
	GeneratedPath  string             `json:"generated_path,omitempty"`
	PostsPath      string             `json:"generated_path,omitempty"`
	DraftsPath     string             `json:"generated_path,omitempty"`
	GithubUrl      string             `json:"github_url,omitempty"`
	Concurrency    int                `json:"concurrency,omitempty"`
	ServerPort     int                `json:"server_port,omitempty"`
	FileServerPort int                `json:"server_port,omitempty"`
	Template       view.Settings      `json:"template,omitempty"`
	Goodreads      goodreads.Settings `json:"goodreads,omitempty"`
	Atom           atom.Settings      `json:"atom,omitempty"`
}

const settingsPath = "settings.json"

func DefaultSettings() *Settings {
	return &Settings{
		"./generated",
		"./content/posts",
		"./content/drafts",
		"",
		10,
		8080,
		3000,
		*view.DefaultSettings(),
		*goodreads.DefaultSettings(),
		*atom.DefaultSettings(),
	}
}

func ReadFromFile(log logrus.FieldLogger) *Settings {
	// also programmatically set in defaults
	settings := DefaultSettings()

	// set here
	generatedPath := os.Getenv("GENERATED_PATH")
	if generatedPath != "" {
		settings.GeneratedPath = generatedPath
	}

	_, err := os.Stat(settingsPath)
	if os.IsNotExist(err) {
		log.Infof("%v not found, using defaults...", settingsPath)
		return settings
	}

	file, err := ioutil.ReadFile(settingsPath)
	if err != nil {
		log.Warnf("error reading %v, using defaults...", settingsPath)
		return settings
	}

	// set through settingsPath
	err = json.Unmarshal(file, settings)
	if err != nil {
		log.Warnf("error reading %v, using defaults...", settingsPath)
		return settings
	}

	// set here
	settings.Atom.AuthorUri = settings.Atom.Url()
	return settings
}
