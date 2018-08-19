package models

import "github.com/sirupsen/logrus"

type Factory struct {
	Settings *Settings
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
