package models

import "github.com/Sirupsen/logrus"

type Factory struct {
	postsPath  string
	draftsPath string
	log        logrus.FieldLogger
}

var factory *Factory

func Config(postsPath, draftsPath string, log logrus.FieldLogger) {
	factory = NewFactory(postsPath, draftsPath, log)
}

func NewFactory(postsPath, draftsPath string, log logrus.FieldLogger) *Factory {
	return &Factory{
		postsPath,
		draftsPath,
		log,
	}
}
