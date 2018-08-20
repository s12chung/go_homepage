package models

import (
	"path"

	"github.com/sirupsen/logrus"
)

type Factory struct {
	settings *Settings
	log      logrus.FieldLogger
}

var factory *Factory

func Config(settings *Settings, log logrus.FieldLogger) {
	factory = NewFactory(settings, log)
}

func NewFactory(settings *Settings, log logrus.FieldLogger) *Factory {
	return &Factory{
		settings,
		log,
	}
}

func TestConfig(relative string, log logrus.FieldLogger) {
	ResetPostMap()
	settings := DefaultSettings()
	settings.PostsPath = path.Join(relative, "posts")
	settings.DraftsPath = path.Join(relative, "drafts")
	Config(settings, log)
}

func TestSetPostDirEmpty(log logrus.FieldLogger) {
	settings := DefaultSettings()
	settings.PostsPath = "."
	settings.DraftsPath = "."
	Config(settings, log)
}
