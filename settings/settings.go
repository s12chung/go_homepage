package settings

import (
	"encoding/json"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
)

type Settings struct {
	GeneratedPath string            `json:"generated_path,omitempty"`
	Concurrency   int               `json:"concurrency,omitempty"`
	ServerPort    int               `json:"server_port,omitempty"`
	Template      TemplateSettings  `json:"template,omitempty"`
	Goodreads     GoodreadsSettings `json:"goodreads,omitempty"`
}

type TemplateSettings struct {
	AssetsFolder     string `json:"assets_folder,omitempty"`
	ManifestFilename string `json:"manifest_filename,omitempty"`
	MarkdownsPath    string `json:"markdowns_path,omitempty"`
}

type GoodreadsSettings struct {
	CachePath string `json:"cache_path,omitempty"`
	ApiKey    string `json:"api_key,omitempty"`
	UserId    int    `json:"user_id,omitempty"`
}

const settingsPath = "settings.json"

func ReadFromFile() *Settings {
	_, err := os.Stat(settingsPath)

	settings := Settings{
		"./generated",
		10,
		3000,
		TemplateSettings{
			"assets",
			"manifest.json",
			"./assets/markdowns",
		},
		GoodreadsSettings{
			"cache",
			"",
			0,
		},
	}
	if os.IsNotExist(err) {
		log.Warn("settings.json in root directory not found, using defaults...")
	} else {
		file, err := ioutil.ReadFile(settingsPath)
		if err != nil {
			log.Warn("error reading settings.json, using defaults...")
		}

		json.Unmarshal(file, &settings)
	}
	return &settings
}
