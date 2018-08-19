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

func DefaultSettings() *Settings {
	return &Settings{
		"./generated",
		10,
		8080,
		3000,
		nil,
	}
}

func ReadFromFile(path string, settings *Settings, log logrus.FieldLogger) {
	generatedPath := os.Getenv("GENERATED_PATH")
	if generatedPath != "" {
		settings.GeneratedPath = generatedPath
	}

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		log.Infof("%v not found, using defaults...", path)
		return
	}

	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Warnf("error reading %v, using defaults...", path)
		return
	}

	err = json.Unmarshal(file, settings)
	if err != nil {
		log.Warnf("error reading %v, using defaults...", path)
		return
	}
}
