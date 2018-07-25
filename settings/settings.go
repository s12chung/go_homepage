package settings

import (
	"encoding/json"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
)

type Settings struct {
	GeneratedPath string            `json:"generated_path,omitempty"`
	PostsPath     string            `json:"generated_path,omitempty"`
	DraftsPath    string            `json:"generated_path,omitempty"`
	Concurrency   int               `json:"concurrency,omitempty"`
	ServerPort    int               `json:"server_port,omitempty"`
	Template      TemplateSettings  `json:"template,omitempty"`
	Goodreads     GoodreadsSettings `json:"goodreads,omitempty"`
}

type TemplateSettings struct {
	WebsiteTitle     string `json:"website_title,omitempty"`
	AssetsFolder     string `json:"assets_folder,omitempty"`
	ManifestFilename string `json:"manifest_filename,omitempty"`
	MarkdownsPath    string `json:"markdowns_path,omitempty"`
}

type GoodreadsSettings struct {
	CachePath  string `json:"cache_path,omitempty"`
	ApiKey     string `json:"api_key,omitempty"`
	UserId     int    `json:"user_id,omitempty"`
	PerPage    int    `json:"per_page,omitempty"`
	MaxPerPage int    `json:"max_per_page,omitempty"`
	RateLimit  int    `json:"rate_limit,omitempty"`
}

const settingsPath = "settings.json"

func ReadFromFile() *Settings {
	settings := Settings{
		"./generated",
		"./content/posts",
		"./content/drafts",
		10,
		3000,
		TemplateSettings{
			"Your Website Title",
			"assets",
			"manifest.json",
			"./assets/markdowns",
		},
		GoodreadsSettings{
			"cache",
			"",
			0,
			50,
			200,
			1000,
		},
	}
	_, err := os.Stat(settingsPath)
	if os.IsNotExist(err) {
		log.Infof("%v not found, using defaults...", settingsPath)
		return &settings
	}

	file, err := ioutil.ReadFile(settingsPath)
	if err != nil {
		log.Warnf("error reading %v, using defaults...", settingsPath)
		return &settings
	}

	err = json.Unmarshal(file, &settings)
	if err != nil {
		log.Warnf("error reading %v, using defaults...", settingsPath)
		return &settings
	}

	return &settings
}
