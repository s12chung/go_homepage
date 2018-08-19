package app

import (
	"os"
)

type Settings struct {
	GeneratedPath  string      `json:"generated_path,omitempty"`
	Concurrency    int         `json:"concurrency,omitempty"`
	ServerPort     int         `json:"server_port,omitempty"`
	FileServerPort int         `json:"file_server_port,omitempty"`
	Content        interface{} `json:"content,omitempty"`
}

func DefaultSettings() *Settings {
	generatedPath := os.Getenv("GENERATED_PATH")
	if generatedPath != "" {
		generatedPath = "./generated"
	}
	return &Settings{
		generatedPath,
		10,
		8080,
		3000,
		nil,
	}
}
