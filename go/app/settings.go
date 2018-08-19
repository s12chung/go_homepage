package app

import (
	"os"

	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

type Settings struct {
	GeneratedPath  string      `json:"generated_path,omitempty"`
	Concurrency    int         `json:"concurrency,omitempty"`
	ServerPort     int         `json:"server_port,omitempty"`
	FileServerPort int         `json:"file_server_port,omitempty"`
	Content        interface{} `json:"content,omitempty"`
}

const settingsPath = "settings.json"

func DefaultSettings() *Settings {
	return &Settings{
		"./generated",
		10,
		8080,
		3000,
		nil,
	}
}

func ReadFromFile(settings *Settings, log logrus.FieldLogger) {
	generatedPath := os.Getenv("GENERATED_PATH")
	if generatedPath != "" {
		settings.GeneratedPath = generatedPath
	}

	_, err := os.Stat(settingsPath)
	if os.IsNotExist(err) {
		log.Infof("%v not found, using defaults...", settingsPath)
		return
	}

	file, err := ioutil.ReadFile(settingsPath)
	if err != nil {
		log.Warnf("error reading %v, using defaults...", settingsPath)
		return
	}

	// set through settingsPath
	err = json.Unmarshal(file, settings)
	if err != nil {
		log.Warnf("error reading %v, using defaults...", settingsPath)
		return
	}
}
